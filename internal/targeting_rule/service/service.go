package service

import (
	"context"
	"gorm.io/gorm"
	"main/constants"
	"main/internal/model"
	"main/internal/targeting_rule/repository"
	"main/pkg/apperror"
	oredis "main/pkg/redis"
	"net/http"
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
) (model.TargetingRule, apperror.Error) {
	return s.repo.Get(ctx, filter, scopes...)
}

func (s *Service) GetTargetingRules(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (model.TargetingRules, apperror.Error) {
	return s.repo.GetAll(ctx, filter, scopes...)
}

// only return
func (s *Service) GetDistinctCampaignIDsByFilter(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (campaignIDs []uint64, cusErr apperror.Error) {
	err := s.repo.Db.GetSlaveDB(ctx).Model(&model.TargetingRule{}).Where(filter).Scopes(scopes...).
		Distinct(constants.CampaignID).Find(&campaignIDs).Error
	if err != nil {
		cusErr = apperror.New(err, http.StatusBadRequest)
		return
	}

	return
}
