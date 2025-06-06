package service

import (
	"context"
	"gorm.io/gorm"
	"main/internal/model"
	"main/internal/targeting_rule/repository"
	"main/pkg/errors"
	oredis "main/pkg/redis"
	"sync"
)

type Service struct {
	repo  repository.Repository
	redis oredis.Cache
}

var (
	syncOnce sync.Once
	service  *Service
)

func NewService(repo *repository.Repository, redis *oredis.Redis) *Service {
	syncOnce.Do(func() {
		service = &Service{
			repo:  *repo,
			redis: redis,
		}
	})

	return service
}

func (s *Service) GetTargetingRule(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (model.TargetingRule, errors.Error) {
	return s.repo.Get(ctx, filter, scopes...)
}

func (s *Service) GetTargetingRules(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (model.TargetingRules, errors.Error) {
	return s.repo.GetAll(ctx, filter, scopes...)
}
