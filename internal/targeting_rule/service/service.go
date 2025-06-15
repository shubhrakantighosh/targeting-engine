package service

import (
	"context"
	"gorm.io/gorm"
	"log"
	"main/internal/model"
	"main/internal/targeting_rule/repository"
	"main/pkg/apperror"
	oredis "main/pkg/redis"
	"main/util"
	"sync"
)

type Service struct {
	repo  repository.Repository
	redis oredis.Cache
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

func (s *Service) FilterMatchingCampaigns(
	ctx context.Context,
	clientRequest ClientRequest,
) (matchedCampaignIDs []uint64, cusErr apperror.Error) {
	logTag := util.LogPrefix(ctx, "FilterMatchingCampaigns")

	campaignIDs, cusErr := s.GetCampaignIDsByApp(ctx, clientRequest[model.App])
	if cusErr.Exists() {
		log.Println(logTag, "Failed to get campaign IDs for app:", clientRequest[model.App], "-", cusErr)
		return
	}

	if len(campaignIDs) == 0 {
		log.Println(logTag, "No campaigns targeting app:", clientRequest[model.App])
		return
	}

	targetingRules, cusErr := s.GetTargetingRulesByCampaigns(
		ctx,
		campaignIDs,
	)
	if cusErr.Exists() {
		log.Println(logTag, "Failed to fetch targeting rules:", cusErr)
		return
	}

	if targetingRules.IsEmpty() {
		log.Println(logTag, "No targeting rules found for app:", clientRequest[model.App])
		return
	}

	hierarchy := BuildHierarchy(clientRequest.ToBuildHierarchy()...)
	filter := NewTargetingFilter(targetingRules, hierarchy)
	matchedCampaignIDs = filter.GetMatchingCampaigns(campaignIDs, clientRequest)
	return
}

type DimensionHierarchy struct {
	Dimensions []model.DimensionType                       // Order defines hierarchy (country -> state -> city -> district)
	ParentMap  map[model.DimensionType]model.DimensionType // child -> parent mapping
}

type ClientRequest map[model.DimensionType]string

func (c ClientRequest) ToBuildHierarchy() (dimensionType []model.DimensionType) {
	dimensionType = make([]model.DimensionType, 0)
	if len(c[model.State]) == 0 {
		return
	}

	dimensionType = append(dimensionType, model.Country, model.State)

	if len(c[model.City]) > 0 {
		dimensionType = append(dimensionType, model.City)
	}

	return
}

type filterResult struct {
	CampaignID uint64
	Matched    bool
}

type TargetingFilter struct {
	hierarchy         DimensionHierarchy
	parentChildMap    map[uint64]model.TargetingRules
	campaignRuleIndex map[uint64]model.TargetingRules
}

func Build(rules model.TargetingRules) (map[uint64]model.TargetingRules, map[uint64]model.TargetingRules) {
	parentChildMap := make(map[uint64]model.TargetingRules)
	campaignRuleIndex := make(map[uint64]model.TargetingRules)

	for _, rule := range rules {
		if rule.ParentID > 0 {
			parentChildMap[rule.ParentID] = append(parentChildMap[rule.ParentID], rule)
		}

		campaignRuleIndex[rule.CampaignID] = append(campaignRuleIndex[rule.CampaignID], rule)
	}

	return parentChildMap, campaignRuleIndex
}

func NewTargetingFilter(rules model.TargetingRules, hierarchy DimensionHierarchy) *TargetingFilter {
	parentChildMap, campaignRuleIndex := Build(rules)
	return &TargetingFilter{
		hierarchy:         hierarchy,
		parentChildMap:    parentChildMap,
		campaignRuleIndex: campaignRuleIndex,
	}
}

func BuildHierarchy(dimensions ...model.DimensionType) DimensionHierarchy {
	hierarchy := DimensionHierarchy{
		ParentMap: make(map[model.DimensionType]model.DimensionType),
	}
	if len(dimensions) > 0 {
		hierarchy.Dimensions = dimensions[:len(dimensions)-1]
	}

	for i := 0; i < len(dimensions)-1; i++ {
		hierarchy.ParentMap[dimensions[i]] = dimensions[i+1]
	}

	return hierarchy
}

func (tf *TargetingFilter) GetMatchingCampaigns(campaignIDs []uint64, clientReq ClientRequest) []uint64 {
	var (
		resultChan = make(chan filterResult, len(campaignIDs))
		wg         = new(sync.WaitGroup)
		results    = make([]uint64, 0)
	)

	wg.Add(len(campaignIDs))

	for _, campaignID := range campaignIDs {
		go func(cID uint64) {
			defer wg.Done()
			matched := tf.matchesCampaign(cID, clientReq)
			resultChan <- filterResult{CampaignID: cID, Matched: matched}
		}(campaignID)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for res := range resultChan {
		if res.Matched {
			results = append(results, res.CampaignID)
		}
	}

	return results
}

func (tf *TargetingFilter) matchesCampaign(campaignID uint64, clientReq ClientRequest) bool {
	matchedRules := make(map[model.DimensionType]uint64)
	rules, ok := tf.campaignRuleIndex[campaignID]
	if !ok {
		return false
	}

	campaignRuleMap := rules.GroupByDimensionAndValue()
	for dimension := range model.NonHierarchicalDimensionType {
		val := clientReq[dimension]
		rulesByVal, found := campaignRuleMap[dimension][val]
		if !found {
			return false
		}

		match := false
		for _, rule := range rulesByVal {
			if rule.Include {
				match = true
				matchedRules[rule.DimensionType] = rule.ID
				break
			}
		}

		if !match {
			return false
		}
	}

	if len(tf.hierarchy.Dimensions) > 0 {
		return tf.checkHierarchicalDimensions(clientReq, matchedRules)
	}
	return true
}

func (tf *TargetingFilter) checkHierarchicalDimensions(
	clientReq ClientRequest,
	matchedRules map[model.DimensionType]uint64,
) bool {
	for _, parentDimType := range tf.hierarchy.Dimensions {
		childDimType := tf.hierarchy.ParentMap[parentDimType]

		clientVal := clientReq[childDimType]
		id, ok := matchedRules[parentDimType]
		if !ok {
			return false
		}

		val, ok := tf.parentChildMap[id]
		if !ok {
			return false
		}

		isFound := false
		for _, rule := range val {
			if rule.Include && rule.Value == clientVal {
				matchedRules[rule.DimensionType] = rule.ID
				isFound = true
				break
			}
		}

		if !isFound {
			return false
		}
	}

	return true
}
