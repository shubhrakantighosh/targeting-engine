package repository

import (
	"context"
	"gorm.io/gorm"
	"main/pkg/apperror"
)

type Interface interface {
	FetchCampaignIDsByFilter(
		ctx context.Context,
		filter map[string]interface{},
		scopes ...func(db *gorm.DB) *gorm.DB,
	) (campaignIDs []uint64, cusErr apperror.Error)
}
