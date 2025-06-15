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
	return s.repo.FetchCampaignIDsByFilter(ctx, filter, scopes...)
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
