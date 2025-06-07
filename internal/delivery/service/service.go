package service

import (
	"context"
	campaignService "main/internal/campaign/service"
	"main/internal/controller/delivery/request"
	"main/internal/model"
	targetingRuleService "main/internal/targeting_rule/service"
	"main/pkg/apperror"
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

// handle no match for one record ( example : device ios but by mistake input come ioss it will work if the secound will match )

func (s *Service) GetMatchingCampaigns(
	ctx context.Context,
	deliveryRequestParams request.DeliveryRequestParams,
) (activeCampaigns model.Campaigns, cusErr apperror.Error) {
	campaignIDs, cusErr := s.targetingRuleService.GetCampaignIDByApp(ctx, deliveryRequestParams.App)
	if cusErr.Exists() {
		return
	}

	if len(campaignIDs) == 0 {
		return
	}

	campaigns, cusErr := s.campaignService.FetchCampaignsByIDs(ctx, campaignIDs)
	if cusErr.Exists() {
		return
	}

	activeCampaignsIDs := campaigns.GetActiveCampaignIDs()
	if len(activeCampaignsIDs) == 0 {
		return
	}

	targetingRules, ruleErr := s.targetingRuleService.GetTargetingRuleByDimensionType(
		ctx,
		activeCampaignsIDs,
		deliveryRequestParams.Country,
		deliveryRequestParams.OS,
	)
	if ruleErr.Exists() {
		cusErr = ruleErr
		return
	}

	if targetingRules.IsEmpty() {
		return
	}

	campaignsIDMap := make(map[uint64]struct{})
	for _, targetingRule := range targetingRules {
		campaignsIDMap[targetingRule.CampaignID] = struct{}{}
	}

	for _, campaign := range campaigns {
		if _, ok := campaignsIDMap[campaign.ID]; ok {
			activeCampaigns = append(activeCampaigns, campaign)
		}
	}

	return
}

func (s *Service) AppExists(
	ctx context.Context,
	app string,
) (isExists bool, cusErr apperror.Error) {
	campaignIDs, cusErr := s.targetingRuleService.GetCampaignIDByApp(ctx, app)
	if cusErr.Exists() {
		return
	}

	if len(campaignIDs) > 0 {
		isExists = true
		return
	}

	return
}
