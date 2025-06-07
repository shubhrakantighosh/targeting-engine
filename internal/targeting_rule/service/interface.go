package service

import (
	"context"
	"gorm.io/gorm"
	"main/internal/model"
	"main/pkg/apperror"
)

type Interface interface {
	GetTargetingRule(
		ctx context.Context,
		filter map[string]interface{},
		scopes ...func(db *gorm.DB) *gorm.DB,
	) (model.TargetingRule, apperror.Error)

	GetTargetingRules(
		ctx context.Context,
		filter map[string]interface{},
		scopes ...func(db *gorm.DB) *gorm.DB,
	) (model.TargetingRules, apperror.Error)

	GetCampaignIDByApp(ctx context.Context, app string) (campaignIDs []uint64, cusErr apperror.Error)

	GetTargetingRuleByDimensionType(
		ctx context.Context,
		campaignIDs []uint64,
		country,
		os string,
	) (targetingRules model.TargetingRules, cusErr apperror.Error)
}
