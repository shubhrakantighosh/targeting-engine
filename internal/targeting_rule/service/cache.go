package service

import (
	"context"
	"fmt"
	"log"
	"main/constants"
	"main/internal/model"
	"main/pkg/apperror"
	"main/pkg/redis"
	"main/util"
	"net/http"
)

func (s *Service) GetCampaignIDsByApp(
	ctx context.Context,
	appID string,
) (campaignIDs []uint64, cusErr apperror.Error) {
	logTag := util.LogPrefix(ctx, "GetCampaignIDsByApp")

	campaignIDs = make([]uint64, 0)
	found, err := s.redis.Get(ctx, appCacheKey(appID), &campaignIDs)
	if err != nil {
		log.Println(logTag, "Redis GET error for app:", appID, "-", err)

		cusErr = apperror.New(err, http.StatusBadRequest)
		return
	}

	if found {
		return
	}

	return s.populateCampaignIDsByApp(ctx, appID)
}

func (s *Service) populateCampaignIDsByApp(
	ctx context.Context,
	appID string,
) (campaignIDs []uint64, cusErr apperror.Error) {
	logTag := util.LogPrefix(ctx, "populateCampaignIDsByApp")

	filter := map[string]any{
		constants.DimensionType: model.App,
		constants.Value:         appID,
		constants.Include:       true,
	}
	campaignIDs, cusErr = s.GetDistinctCampaignIDsByFilter(ctx, filter)
	if cusErr.Exists() {
		log.Println(logTag, "Failed to fetch campaign IDs from DB for filter:", filter, "-", cusErr)

		return
	}

	if len(campaignIDs) == 0 {
		return
	}

	if _, err := s.redis.Set(ctx, appCacheKey(appID), campaignIDs, constants.OneDay); err != nil {
		log.Println(logTag, "Failed to cache campaign IDs for app:", appID, "-", err)

		cusErr = apperror.New(err, http.StatusBadRequest)
		return
	}

	return
}

func (s *Service) GetTargetingRulesByCampaigns(
	ctx context.Context,
	campaignIDs []uint64,
	country, os string,
) (rules model.TargetingRules, cusErr apperror.Error) {
	logTag := util.LogPrefix(ctx, "GetTargetingRulesByCampaigns")

	rules = make(model.TargetingRules, 0)
	cacheMissCampaigns := make([]uint64, 0)
	cacheKeys := make([]*redis.KVOut, 0)
	campaignKeyMap := make(map[string]uint64)
	campaignMatchMap := make(map[uint64]bool)

	for _, id := range campaignIDs {
		campaignMatchMap[id] = false

		for _, dimension := range []model.DimensionType{model.Country, model.OS} {
			value := country
			if dimension == model.OS {
				value = os
			}
			key := targetingRuleCacheKey(id, dimension, value)
			cacheKeys = append(cacheKeys, &redis.KVOut{Key: key, Val: &model.TargetingRule{}})
			campaignKeyMap[key] = id
		}
	}

	if err := s.redis.PipedMGet(ctx, cacheKeys); err != nil {
		log.Println(logTag, "Redis PipedMGet error:", err)

		cusErr = apperror.New(err, http.StatusBadRequest)
		return
	}

	for _, kv := range cacheKeys {
		val, ok := kv.Val.(*model.TargetingRule)
		if !kv.OK() || !ok {
			continue
		}

		campaignID := val.CampaignID
		campaignMatchMap[campaignID] = true
		rules = append(rules, *val)
	}

	for id, found := range campaignMatchMap {
		if !found {
			cacheMissCampaigns = append(cacheMissCampaigns, id)
		}
	}

	if len(cacheMissCampaigns) == 0 {
		return
	}

	dbRules, cusErr := s.populateTargetingRulesByCampaigns(ctx, cacheMissCampaigns, country, os)
	if cusErr.Exists() {
		log.Println(logTag, "DB fetch error for missing targeting rules:", cusErr)

		return
	}

	rules = append(rules, dbRules...)
	return
}

func (s *Service) populateTargetingRulesByCampaigns(
	ctx context.Context,
	campaignIDs []uint64,
	country, os string,
) (rules model.TargetingRules, cusErr apperror.Error) {
	logTag := util.LogPrefix(ctx, "populateTargetingRulesByCampaigns")

	rules = make(model.TargetingRules, 0)

	filter := map[string]any{
		constants.CampaignID:    campaignIDs,
		constants.DimensionType: []string{constants.Country, constants.OS},
		constants.Value:         []string{country, os},
		constants.Include:       true,
	}

	rules, cusErr = s.GetTargetingRules(ctx, filter)
	if cusErr.Exists() {
		log.Println(logTag, "DB fetch error:", cusErr, "Filter:", filter)

		return
	}

	if rules.IsEmpty() {
		return rules, apperror.Error{}
	}

	keyMap := make([]redis.KVIn, 0, len(rules))
	for _, rule := range rules {
		key := targetingRuleCacheKey(rule.CampaignID, rule.DimensionType, rule.Value)
		keyMap = append(keyMap, redis.KVIn{Key: key, Val: &rule})
	}

	if err := s.redis.PipedMSet(ctx, keyMap, constants.OneDay); err != nil {
		log.Println(logTag, "Redis MSet error while caching rules:", err)
	}

	return
}

func appCacheKey(appID string) string {
	return fmt.Sprintf("targeting_rule_app_%s", appID)
}

func targetingRuleCacheKey(campaignID uint64, dimension model.DimensionType, value string) string {
	return fmt.Sprintf("targeting_rule_%d_%s_%s", campaignID, dimension.String(), value)
}
