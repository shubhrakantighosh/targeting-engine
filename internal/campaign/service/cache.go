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
	"strconv"
)

func (s *Service) FetchCampaignsByIDs(
	ctx context.Context,
	ids []uint64,
) (campaigns model.Campaigns, cusErr apperror.Error) {
	logTag := util.LogPrefix(ctx, "FetchCampaignsByIDs")

	campaigns = make(model.Campaigns, 0)
	notFoundIDs := make([]uint64, 0)
	idMap := make(map[uint64]struct{})

	ids = util.DeduplicateSlice(ids)
	if len(ids) == 0 {
		log.Println(logTag, "No IDs provided.")
		return
	}

	keyMapValue := make([]*redis.KVOut, 0)
	for _, id := range ids {
		key := cacheKey(strconv.FormatUint(id, 10))
		keyMapValue = append(keyMapValue, &redis.KVOut{
			Key: key,
			Val: &model.Campaign{},
		})
		idMap[id] = struct{}{}
	}

	err := s.redis.PipedMGet(ctx, keyMapValue)
	if err != nil {
		log.Println(err)

		cusErr = apperror.New(err, http.StatusBadRequest)
		return
	}

	for _, v := range keyMapValue {
		val, ok := v.Val.(*model.Campaign)
		if !v.OK() || !ok {
			continue
		}

		campaign := *val
		if _, ok = idMap[campaign.ID]; ok {
			delete(idMap, campaign.ID)
		}

		campaigns = append(campaigns, campaign)
	}

	if len(idMap) == 0 {
		return
	}

	for id, _ := range idMap {
		notFoundIDs = append(notFoundIDs, id)
	}

	missedCampaigns, cusErr := s.fetchAndCacheCampaigns(ctx, notFoundIDs)
	if cusErr.Exists() {
		log.Printf("%s Failed to fetch & cache campaigns: %v", logTag, cusErr)
		return
	}

	campaigns = append(campaigns, missedCampaigns...)
	return
}

func (s *Service) fetchAndCacheCampaigns(
	ctx context.Context,
	ids []uint64,
) (campaigns model.Campaigns, cusErr apperror.Error) {
	logTag := util.LogPrefix(ctx, "fetchAndCacheCampaigns")

	filter := map[string]any{
		constants.ID: ids,
	}
	campaigns, cusErr = s.GetCampaigns(ctx, filter)
	if cusErr.Exists() {
		log.Printf("%s Failed to get campaigns: %v", logTag, cusErr)
		return
	}

	if campaigns.IsEmpty() {
		log.Printf("%s No campaigns found for given IDs , filter :%v", logTag, filter)
		return
	}

	keyMapValue := make([]redis.KVIn, 0)
	for _, campaign := range campaigns {
		key := cacheKey(strconv.FormatUint(campaign.ID, 10))
		keyMapValue = append(keyMapValue, redis.KVIn{
			Key: key,
			Val: &campaign,
		})
	}

	if err := s.redis.PipedMSet(ctx, keyMapValue, constants.OneDay); err != nil {
		log.Printf("%s Redis MSet error: %v", logTag, err)

		cusErr = apperror.New(err, http.StatusBadRequest)
		return
	}

	return
}

func (s *Service) InvalidCampaignsByIDs(
	ctx context.Context,
	ids []uint64,
) (cusErr apperror.Error) {
	logTag := util.LogPrefix(ctx, "InvalidCampaignsByIDs")
	ids = util.DeduplicateSlice(ids)
	if len(ids) == 0 {
		return
	}

	keys := make([]string, 0)
	for _, id := range ids {
		key := cacheKey(strconv.FormatUint(id, 10))
		keys = append(keys, key)
	}

	_, err := s.redis.Unlink(ctx, keys)
	if err != nil {
		log.Printf("%s Redis MUnlink error: %v", logTag, err)

		cusErr = apperror.New(err, http.StatusBadRequest)
		return
	}

	return
}

func cacheKey(str string) string {
	return fmt.Sprintf("campaigns_id_%s", str)
}
