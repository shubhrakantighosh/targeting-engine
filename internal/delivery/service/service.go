package service

import (
	"context"
	"main/constants"
	campaignService "main/internal/campaign/service"
	"main/internal/controller/delivery/request"
	"main/internal/model"
	targetingRuleService "main/internal/targeting_rule/service"
	"main/pkg/errors"
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
	deliveryRequestParams request.DeliveryRequestParams,
) (campaigns model.Campaigns, cudErr errors.Error) {
	filter := map[string]any{
		constants.DimensionType: constants.App,
		constants.Value:         deliveryRequestParams.App,
		constants.Include:       true,
	}
	targetingRule, cusErr := s.targetingRuleService.GetTargetingRule(ctx, filter)
	if cusErr.Exists() {
		return
	}

	if targetingRule.ID == 0 {
		return
	}

	filter = map[string]any{
		constants.CampaignID:    targetingRule.CampaignID,
		constants.Include:       true,
		constants.DimensionType: []string{constants.Country, constants.OS},
		constants.Value:         []string{deliveryRequestParams.Country, deliveryRequestParams.OS},
	}
	targetingRules, cusErr := s.targetingRuleService.GetTargetingRules(ctx, filter)
	if cusErr.Exists() {
		return
	}

	if targetingRules.IsEmpty() {
		return
	}

	filter = map[string]any{
		constants.ID:     targetingRules.GetCampaignIDs(),
		constants.Status: model.Active,
	}
	campaigns, cusErr = s.campaignService.GetCampaigns(ctx, filter)
	if cusErr.Exists() {
		return
	}

	return
}
