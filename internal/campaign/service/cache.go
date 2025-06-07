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
	"strconv"
)

func (s *Service) FetchCampaignsByIDs(
	ctx context.Context,
	ids []uint64,
) (campaigns model.Campaigns, cusErr apperror.Error) {
	logTag := util.LogPrefix(ctx, "FetchCampaignsByIDs")

	campaigns = make(model.Campaigns, 0)
	notFoundIDs := make([]uint64, 0)

	ids = util.DeduplicateSlice(ids)
	if len(ids) == 0 {
		log.Println(logTag, "No IDs provided.")
		return
	}

	for _, id := range ids {
		key := cacheKey(strconv.FormatUint(id, 10))

		var campaign model.Campaign
		found, err := s.redis.Get(ctx, key, &campaign)
		if err != nil {
			log.Printf("%s Redis MGet error: %v", logTag, err)

			cusErr = apperror.New(err, http.StatusBadRequest)
			return
		}

		if found {
			campaigns = append(campaigns, campaign)
			continue
		}

		notFoundIDs = append(notFoundIDs, id)
	}

	if len(notFoundIDs) == 0 {
		return
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

	for _, campaign := range campaigns {
		key := cacheKey(strconv.FormatUint(campaign.ID, 10))

		if _, err := s.redis.Set(ctx, key, campaign, constants.OneDay); err != nil {
			log.Printf("%s Redis MSet error: %v", logTag, err)

			cusErr = apperror.New(err, http.StatusBadRequest)
			return
		}

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
