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
	matchedCampaignIDs, cusErr := s.targetingRuleService.FilterMatchingCampaigns(
		ctx,
		params.DimensionTypeMapValue(),
	)
	if cusErr.Exists() || len(matchedCampaignIDs) == 0 {
		return nil, cusErr
	}

	return s.campaignService.FetchCampaignsByIDs(ctx, matchedCampaignIDs)
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
