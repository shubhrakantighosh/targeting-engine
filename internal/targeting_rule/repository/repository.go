package repository

import (
	"context"
	"gorm.io/gorm"
	"main/constants"
	"main/internal/model"
	"main/internal/repository"
	"main/pkg/apperror"
	"main/pkg/db/postgres"
	"net/http"
	"sync"
)

type Repository struct {
	repository.Repository[model.TargetingRule]
}

var (
	syncOnce sync.Once
	repo     *Repository
)

func NewRepository(db *postgres.DbCluster) *Repository {
	syncOnce.Do(func() {
		repo = &Repository{
			Repository: repository.Repository[model.TargetingRule]{Db: db},
		}
	})

	return repo
}

func (repo *Repository) GetDistinctCampaignIDsByFilter(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (campaignIDs []uint64, cusErr apperror.Error) {
	err := repo.Db.GetSlaveDB(ctx).Model(&model.TargetingRule{}).Where(filter).Scopes(scopes...).
		Distinct(constants.CampaignID).Find(&campaignIDs).Error
	if err != nil {
		cusErr = apperror.New(err, http.StatusBadRequest)
		return
	}

	return
}
