package service

import (
	"context"
	"gorm.io/gorm"
	"main/internal/model"
	"main/pkg/errors"
)

type Interface interface {
	GetTargetingRule(
		ctx context.Context,
		filter map[string]interface{},
		scopes ...func(db *gorm.DB) *gorm.DB,
	) (model.TargetingRule, errors.Error)

	GetTargetingRules(
		ctx context.Context,
		filter map[string]interface{},
		scopes ...func(db *gorm.DB) *gorm.DB,
	) (model.TargetingRules, errors.Error)
}
