package service

import (
	"context"
	"log"
	campaignService "main/internal/campaign/service"
	"main/internal/controller/delivery/request"
	"main/internal/model"
	targetingRuleService "main/internal/targeting_rule/service"
	"main/pkg/apperror"
	"main/util"
	"sync"
)

type Service struct {
	campaignService      campaignService.Interface
	targetingRuleService targetingRuleService.Interface
}

var (
	syncOnce sync.Once
	service  *Service
)

func NewService(
	campaignService campaignService.Interface,
	targetingRuleService targetingRuleService.Interface,
) *Service {
	syncOnce.Do(func() {
		service = &Service{
			campaignService:      campaignService,
			targetingRuleService: targetingRuleService,
		}
	})

	return service
}

func (s *Service) GetMatchingCampaigns(
	ctx context.Context,
	params request.DeliveryRequestParams,
) (matchingCampaigns model.Campaigns, cusErr apperror.Error) {
	logTag := util.LogPrefix(ctx, "GetMatchingCampaigns")

	matchingCampaigns = make(model.Campaigns, 0)
	campaignIDs, cusErr := s.targetingRuleService.GetCampaignIDsByApp(ctx, params.App)
	if cusErr.Exists() {
		log.Println(logTag, "Failed to get campaign IDs for app:", params.App, "-", cusErr)
		return
	}

	if len(campaignIDs) == 0 {
		log.Println(logTag, "No campaigns targeting app:", params.App)
		return
	}

	targetingRules, cusErr := s.targetingRuleService.GetTargetingRulesByCampaigns(
		ctx,
		campaignIDs,
		params.Country,
		params.OS,
	)
	if cusErr.Exists() {
		log.Println(logTag, "Failed to fetch targeting rules:", cusErr)
		return
	}

	if targetingRules.IsEmpty() {
		log.Println(logTag, "No targeting rules found for app:", params.App)
		return
	}

	matchingCampaigns, cusErr = s.filterMatchingCampaigns(ctx, params, campaignIDs, targetingRules)
	if cusErr.Exists() {
		log.Println(logTag, "Rule matching failed:", cusErr)
		return
	}

	return matchingCampaigns, apperror.Error{}
}

func (s *Service) filterMatchingCampaigns(
	ctx context.Context,
	params request.DeliveryRequestParams,
	campaignIDs []uint64,
	targetingRules model.TargetingRules,
) (model.Campaigns, apperror.Error) {
	matchingCampaignIDs := make([]uint64, 0)
	groupedRules := targetingRules.GroupByCampaignID()

	for _, campaignID := range campaignIDs {
		rules, ok := groupedRules[campaignID]
		if !ok || len(rules) < 2 {
			continue
		}

		hasMatchingCountry := false
		hasMatchingOS := false

		for _, rule := range rules {
			switch rule.DimensionType {
			case model.Country:
				if rule.Value == params.Country && rule.Include {
					hasMatchingCountry = true
				}
			case model.OS:
				if rule.Value == params.OS && rule.Include {
					hasMatchingOS = true
				}
			}
		}

		if hasMatchingCountry && hasMatchingOS {
			matchingCampaignIDs = append(matchingCampaignIDs, campaignID)
		}
	}

	if len(matchingCampaignIDs) == 0 {
		return model.Campaigns{}, apperror.Error{}
	}

	return s.campaignService.FetchCampaignsByIDs(ctx, matchingCampaignIDs)
}

func (s *Service) IsAppTargeted(
	ctx context.Context,
	appID string,
) (isExists bool, cusErr apperror.Error) {
	logTag := util.LogPrefix(ctx, "IsAppTargeted")

	campaignIDs, cusErr := s.targetingRuleService.GetCampaignIDsByApp(ctx, appID)
	if cusErr.Exists() {
		log.Println(logTag, "Failed to verify app:", appID, "-", cusErr)

		return
	}

	return len(campaignIDs) > 0, apperror.Error{}
}
