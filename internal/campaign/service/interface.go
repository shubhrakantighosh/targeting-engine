package service

import (
	"context"
	"gorm.io/gorm"
	"main/internal/model"
	"main/pkg/apperror"
)

type Interface interface {
	GetCampaign(
		ctx context.Context,
		filter map[string]interface{},
		scopes ...func(db *gorm.DB) *gorm.DB,
	) (model.Campaign, apperror.Error)

	GetCampaigns(
		ctx context.Context,
		filter map[string]interface{},
		scopes ...func(db *gorm.DB) *gorm.DB,
	) (model.Campaigns, apperror.Error)

	FetchCampaignsByIDs(
		ctx context.Context,
		ids []uint64,
	) (campaigns model.Campaigns, cusErr apperror.Error)

	InvalidCampaignsByIDs(
		ctx context.Context,
		ids []uint64,
	) (cusErr apperror.Error)
}
