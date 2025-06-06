package service

import (
	"context"
	"gorm.io/gorm"
	"main/internal/model"
	"main/pkg/errors"
)

type Interface interface {
	GetCampaign(
		ctx context.Context,
		filter map[string]interface{},
		scopes ...func(db *gorm.DB) *gorm.DB,
	) (model.Campaign, errors.Error)

	GetCampaigns(
		ctx context.Context,
		filter map[string]interface{},
		scopes ...func(db *gorm.DB) *gorm.DB,
	) (model.Campaigns, errors.Error)
}
