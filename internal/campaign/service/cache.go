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

func (s *Service) FetchCampaignsByIDs(
	ctx context.Context,
	campaignIDs []uint64,
) (campaigns model.Campaigns, cusErr apperror.Error) {
	logTag := util.LogPrefix(ctx, "FetchCampaignsByIDs")

	campaignIDs = util.DeduplicateSlice(campaignIDs)
	if len(campaignIDs) == 0 {
		log.Println(logTag, "No campaign IDs provided")
		return
	}

	campaigns = make(model.Campaigns, 0)
	cacheMissIDs := make([]uint64, 0)
	idTracker := make(map[uint64]struct{}, len(campaignIDs))

	kvPairs := make([]*redis.KVOut, 0, len(campaignIDs))
	for _, id := range campaignIDs {
		cacheKey := cacheKeyForCampaign(id)
		kvPairs = append(kvPairs, &redis.KVOut{
			Key: cacheKey,
			Val: &model.Campaign{},
		})
		idTracker[id] = struct{}{}
	}

	if err := s.redis.PipedMGet(ctx, kvPairs); err != nil {
		log.Println(logTag, "Redis PipedMGet failed:", err)

		cusErr = apperror.New(err, http.StatusBadRequest)
		return
	}

	for _, kv := range kvPairs {
		campaign, ok := kv.Val.(*model.Campaign)
		if !kv.OK() || !ok {
			continue
		}
		campaigns = append(campaigns, *campaign)
		delete(idTracker, campaign.ID)
	}

	if len(idTracker) == 0 {
		return
	}

	for id := range idTracker {
		cacheMissIDs = append(cacheMissIDs, id)
	}

	missingCampaigns, cusErr := s.fetchAndCacheCampaigns(ctx, cacheMissIDs)
	if cusErr.Exists() {
		log.Println(logTag, "DB fallback failed:", cusErr)
		return
	}

	campaigns = append(campaigns, missingCampaigns...)
	return
}

func (s *Service) fetchAndCacheCampaigns(
	ctx context.Context,
	campaignIDs []uint64,
) (campaigns model.Campaigns, cusErr apperror.Error) {
	logTag := util.LogPrefix(ctx, "fetchAndCacheCampaigns")

	campaigns = make(model.Campaigns, 0)
	filter := map[string]any{
		constants.ID: campaignIDs,
	}

	campaigns, cusErr = s.GetCampaigns(ctx, filter)
	if cusErr.Exists() {
		log.Println(logTag, "DB fetch failed for campaigns:", cusErr, "Filter:", filter)

		return
	}

	if campaigns.IsEmpty() {
		return
	}

	kvPairs := make([]redis.KVIn, 0, len(campaigns))
	for _, campaign := range campaigns {
		cacheKey := cacheKeyForCampaign(campaign.ID)
		kvPairs = append(kvPairs, redis.KVIn{
			Key: cacheKey,
			Val: &campaign,
		})
	}

	if err := s.redis.PipedMSet(ctx, kvPairs, constants.OneDay); err != nil {
		log.Println(logTag, "Redis PipedMSet failed:", err)

		cusErr = apperror.New(err, http.StatusBadRequest)
		return
	}

	return
}

func (s *Service) InvalidateCampaignCache(
	ctx context.Context,
	campaignIDs []uint64,
) (cusErr apperror.Error) {
	logTag := util.LogPrefix(ctx, "InvalidateCampaignCache")

	campaignIDs = util.DeduplicateSlice(campaignIDs)
	if len(campaignIDs) == 0 {
		return
	}

	keys := make([]string, 0, len(campaignIDs))
	for _, id := range campaignIDs {
		keys = append(keys, cacheKeyForCampaign(id))
	}

	if _, err := s.redis.Unlink(ctx, keys); err != nil {
		log.Println(logTag, "Redis Unlink failed:", err)

		return apperror.New(err, http.StatusBadRequest)
	}

	return
}

func cacheKeyForCampaign(campaignID uint64) string {
	return fmt.Sprintf("campaigns_id_%d", campaignID)
}
