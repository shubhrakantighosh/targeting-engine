package service

import (
	"context"
	"fmt"
	"log"
	"main/constants"
	"main/internal/model"
	"main/pkg/apperror"
	"main/util"
	"net/http"
)

func (s *Service) GetCampaignIDByApp(ctx context.Context, app string) (campaignIDs []uint64, cusErr apperror.Error) {
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
	// discint only campagnine
	targetingRules, cusErr := s.GetTargetingRules(ctx, filter)
	if cusErr.Exists() {
		return
	}

	if targetingRules.IsEmpty() {
		return
	}

	for _, targetingRule := range targetingRules {
		campaignIDs = append(campaignIDs, targetingRule.CampaignID)
	}

	campaignIDs = util.DeduplicateSlice(campaignIDs)

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

	for _, campaignID := range campaignIDs {
		foundTargetingRules, ruleErr := s.getTargetingRuleByDimensionType(
			ctx,
			campaignID,
			country,
			os,
		)
		if ruleErr.Exists() {
			cusErr = ruleErr
			return
		}

		targetingRules = append(targetingRules, foundTargetingRules...)
	}

	return
}

func (s *Service) getTargetingRuleByDimensionType(
	ctx context.Context,
	campaignID uint64,
	country,
	os string,
) (targetingRules model.TargetingRules, cusErr apperror.Error) {
	targetingRules = make(model.TargetingRules, 0)
	notFoundDimensionTypes := make([]model.DimensionType, 0)
	notFoundValues := make([]string, 0)

	var targetingRule model.TargetingRule
	key := cacheByDimensionTypeKey(campaignID, model.Country, country)
	found, err := s.redis.Get(ctx, key, &targetingRule)
	if err != nil {
		log.Println(err.Error())

		cusErr = apperror.New(err, http.StatusBadRequest)
		return
	}

	if found {
		targetingRules = append(targetingRules, targetingRule)
	} else {
		notFoundDimensionTypes = append(notFoundDimensionTypes, model.Country)
		notFoundValues = append(notFoundValues, country)
	}

	key = cacheByDimensionTypeKey(campaignID, model.OS, os)
	found, err = s.redis.Get(ctx, key, &targetingRule)
	if err != nil {
		log.Println(err.Error())

		cusErr = apperror.New(err, http.StatusBadRequest)
		return
	}

	if found {
		targetingRules = append(targetingRules, targetingRule)
	} else {
		notFoundDimensionTypes = append(notFoundDimensionTypes, model.OS)
		notFoundValues = append(notFoundValues, os)
	}

	filter := map[string]any{
		constants.CampaignID:    campaignID,
		constants.DimensionType: notFoundDimensionTypes,
		constants.Value:         notFoundValues,
		constants.Include:       true,
	}
	targetingRule, cusErr = s.GetTargetingRule(ctx, filter)
	if cusErr.Exists() {
		return
	}

	for i := 0; i < len(notFoundDimensionTypes); i++ {
		notFoundDimensionType := notFoundDimensionTypes[i]
		value := notFoundValues[i]

		key = cacheByDimensionTypeKey(targetingRule.CampaignID, notFoundDimensionType, value)
		_, err = s.redis.Set(ctx, key, targetingRule, constants.OneDay)
		if err != nil {
			log.Println(err.Error())

			cusErr = apperror.New(err, http.StatusBadRequest)
			return
		}
	}

	return
}

func cacheByDimensionTypeKey(campaignID uint64, dimensionType model.DimensionType, value string) string {
	return fmt.Sprintf("targeting_rule_%d_%s_%s", campaignID, dimensionType.String(), value)
}

func cacheKey(str string) string {
	return fmt.Sprintf("targeting_rule_app_%s", str)
}
