package service

import (
	"context"
	"fmt"
	"log"
	"main/constants"
	"main/internal/model"
	"main/pkg/apperror"
	"main/pkg/redis"
	"net/http"
)

func (s *Service) GetCampaignIDByApp(
	ctx context.Context,
	app string,
) (campaignIDs []uint64, cusErr apperror.Error) {
	campaignIDs = make([]uint64, 0)
	found, err := s.redis.Get(ctx, cacheKey(app), &campaignIDs)
	if err != nil {
		log.Println(err.Error())

		cusErr = apperror.New(err, http.StatusBadRequest)
		return
	}

	if found {
		return
	}

	filter := map[string]any{
		constants.DimensionType: model.App,
		constants.Value:         app,
		constants.Include:       true,
	}
	campaignIDs, cusErr = s.GetDistinctCampaignIDsByFilter(ctx, filter)
	if cusErr.Exists() {
		return
	}

	if len(campaignIDs) == 0 {
		return
	}

	_, err = s.redis.Set(ctx, cacheKey(app), campaignIDs, constants.OneDay)
	if err != nil {
		log.Println(err.Error())

		cusErr = apperror.New(err, http.StatusBadRequest)
		return
	}

	return
}

func (s *Service) GetTargetingRuleByDimensionType(
	ctx context.Context,
	campaignIDs []uint64,
	country,
	os string,
) (targetingRules model.TargetingRules, cusErr apperror.Error) {
	targetingRules = make(model.TargetingRules, 0)
	notFoundCampaignIDs := make([]uint64, 0)
	keyMapCampaignID := make(map[string]uint64)
	idMap := make(map[uint64]bool)

	keyVal := make([]*redis.KVOut, 0)
	for _, campaignID := range campaignIDs {
		countryKey := cacheByDimensionTypeKey(campaignID, model.Country, country)
		osKey := cacheByDimensionTypeKey(campaignID, model.OS, os)
		keyVal = append(keyVal, &redis.KVOut{
			Key: countryKey,
			Val: &model.TargetingRule{},
		})

		keyVal = append(keyVal, &redis.KVOut{
			Key: osKey,
			Val: &model.TargetingRule{},
		})

		keyMapCampaignID[countryKey] = campaignID
		keyMapCampaignID[osKey] = campaignID
		idMap[campaignID] = false
	}

	err := s.redis.PipedMGet(ctx, keyVal)
	if err != nil {
		log.Println(err.Error())

		cusErr = apperror.New(err, http.StatusBadRequest)
		return
	}

	for _, kv := range keyVal {
		val, ok := kv.Val.(*model.TargetingRule)
		if !kv.OK() || !ok {
			continue
		}

		if isFalse, found := idMap[val.CampaignID]; found && !isFalse {
			idMap[val.CampaignID] = true
		}

		targetingRules = append(targetingRules, *val)
	}

	for campaignID, boolean := range idMap {
		if !boolean {
			notFoundCampaignIDs = append(notFoundCampaignIDs, campaignID)
		}
	}

	if len(notFoundCampaignIDs) == 0 {
		return
	}

	filter := map[string]any{
		constants.CampaignID:    notFoundCampaignIDs,
		constants.DimensionType: []string{constants.Country, constants.OS},
		constants.Value:         []string{country, os},
		constants.Include:       true,
	}
	targetingRules, cusErr = s.GetTargetingRules(ctx, filter)
	if cusErr.Exists() {
		return
	}

	keyMap := make([]redis.KVIn, 0)
	for _, targetingRule := range targetingRules {
		key := cacheByDimensionTypeKey(targetingRule.CampaignID, targetingRule.DimensionType, targetingRule.Value)
		keyMap = append(keyMap, redis.KVIn{
			Key: key,
			Val: &targetingRule,
		})
	}

	err = s.redis.PipedMSet(ctx, keyMap, constants.OneDay)
	if err != nil {
		log.Println(err.Error())

		cusErr = apperror.New(err, http.StatusBadRequest)
		return
	}

	return
}

func cacheByDimensionTypeKey(campaignID uint64, dimensionType model.DimensionType, value string) string {
	return fmt.Sprintf("targeting_rule_%d_%s_%s", campaignID, dimensionType.String(), value)
}

func cacheKey(str string) string {
	return fmt.Sprintf("targeting_rule_app_%s", str)
}
