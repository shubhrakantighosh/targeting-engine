package service

import (
	"context"
	"gorm.io/gorm"
	"main/internal/campaign/repository"
	"main/internal/model"
	"main/pkg/apperror"
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

func (s *Service) GetCampaign(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (model.Campaign, apperror.Error) {
	return s.repo.Get(ctx, filter, scopes...)
}

func (s *Service) GetCampaigns(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (model.Campaigns, apperror.Error) {
	return s.repo.GetAll(ctx, filter, scopes...)
}
