package service

import (
	"context"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
	"main/internal/model"
	"main/internal/targeting_rule/repository"
	"main/pkg/apperror"
	oredis "main/pkg/redis"
	"sync"
)

type Service struct {
	repo  repository.Repository
	redis oredis.Cache
	singleflight.Group
}

var (
	syncOnce sync.Once
	service  *Service
)

func NewService(repo *repository.Repository, redis *oredis.Redis) *Service {
	syncOnce.Do(func() {
		service = &Service{
			repo:  *repo,
			redis: redis,
		}
	})

	return service
}

func (s *Service) GetTargetingRule(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (model.TargetingRule, apperror.Error) {
	return s.repo.Get(ctx, filter, scopes...)
}

func (s *Service) GetTargetingRules(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (model.TargetingRules, apperror.Error) {
	return s.repo.GetAll(ctx, filter, scopes...)
}

func (s *Service) GetDistinctCampaignIDsByFilter(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (campaignIDs []uint64, cusErr apperror.Error) {
	return s.repo.GetDistinctCampaignIDsByFilter(ctx, filter, scopes...)
}

type DimensionHierarchy struct {
	Dimensions []model.DimensionType                       // Order defines hierarchy (country -> state -> city -> district)
	ParentMap  map[model.DimensionType]model.DimensionType // child -> parent mapping
}

type ClientRequest map[model.DimensionType]string

type filterResult struct {
	CampaignID uint64
	Matched    bool
}

type TargetingFilter struct {
	rules     model.TargetingRules
	hierarchy DimensionHierarchy
}

func NewTargetingFilter(rules model.TargetingRules, hierarchy DimensionHierarchy) *TargetingFilter {
	return &TargetingFilter{
		rules:     rules,
		hierarchy: hierarchy,
	}
}

func BuildHierarchy(dimensions ...model.DimensionType) DimensionHierarchy {
	hierarchy := DimensionHierarchy{
		Dimensions: dimensions,
		ParentMap:  make(map[model.DimensionType]model.DimensionType),
	}

	for i := 1; i < len(dimensions); i++ {
		hierarchy.ParentMap[dimensions[i]] = dimensions[i-1]
	}

	return hierarchy
}

func (tf *TargetingFilter) filterCampaigns(clientReq ClientRequest) []filterResult {
	var (
		campaignRules = tf.rules.GroupByCampaignID()
		resultChan    = make(chan filterResult, len(campaignRules))
		wg            sync.WaitGroup
		results       = make([]filterResult, 0)
	)

	for campaignID, rules := range campaignRules {
		wg.Add(1)
		go func(cID uint64, cRules model.TargetingRules) {
			defer wg.Done()
			matched := tf.matchesCampaign(clientReq, cRules)
			resultChan <- filterResult{
				CampaignID: cID,
				Matched:    matched,
			}
		}(campaignID, rules)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

func (tf *TargetingFilter) GetMatchingCampaigns(clientReq ClientRequest) []uint64 {
	results := tf.filterCampaigns(clientReq)
	var matchingCampaigns []uint64

	for _, result := range results {
		if result.Matched {
			matchingCampaigns = append(matchingCampaigns, result.CampaignID)
		}
	}

	return matchingCampaigns
}

func (tf *TargetingFilter) matchesCampaign(
	clientReq ClientRequest,
	campaignRules model.TargetingRules,
) bool {
	dimensionRules := make(map[model.DimensionType]model.TargetingRules)
	parentChildMap := make(map[uint64]model.TargetingRules)

	for _, rule := range campaignRules {
		dimensionRules[rule.DimensionType] = append(dimensionRules[rule.DimensionType], rule)
		if rule.ParentID > 0 {
			parentChildMap[rule.ParentID] = append(parentChildMap[rule.ParentID], rule)
		}
	}

	for dimension := range model.NonHierarchicalDimensionType {
		if !dimensionRules[dimension].HasIncludedRuleFor(dimension, clientReq[dimension]) {
			return false
		}
	}

	if len(tf.hierarchy.Dimensions) > 0 {
		return tf.checkHierarchicalDimensions(clientReq, dimensionRules, parentChildMap)
	}

	return true
}

func (tf *TargetingFilter) checkHierarchicalDimensions(
	clientReq ClientRequest,
	dimensionRules map[model.DimensionType]model.TargetingRules,
	parentChildMap map[uint64]model.TargetingRules,
) bool {
	matchedRules := make(map[model.DimensionType]uint64)

	for _, dimType := range tf.hierarchy.Dimensions {
		clientValue := clientReq[dimType] // india
		rules := dimensionRules[dimType]  // all countries data
		if rules.IsEmpty() {
			return false
		}

		//  parent dimension
		var parentID uint64 = 0
		if parentDimType, hasParent := tf.hierarchy.ParentMap[dimType]; hasParent {
			parentID = matchedRules[parentDimType]
		}

		applicableRules := getApplicableRules(rules, parentID, parentChildMap)

		if !applicableRules.HasIncludedRuleFor(dimType, clientValue) {
			return false
		}

		for _, rule := range applicableRules {
			if rule.Value == clientValue && rule.Include {
				matchedRules[dimType] = rule.ID
				break
			}
		}
	}

	return true
}

func getApplicableRules(
	rules model.TargetingRules,
	parentID uint64,
	parentChildMap map[uint64]model.TargetingRules,
) model.TargetingRules {
	if parentID > 0 {
		if childRules, exists := parentChildMap[parentID]; exists && len(childRules) > 0 {
			return childRules
		}
	}

	var generalRules model.TargetingRules
	for _, rule := range rules {
		if rule.ParentID == 0 {
			generalRules = append(generalRules, rule)
		}
	}

	return generalRules
}
