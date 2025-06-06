package repository

import (
	"context"
	"gorm.io/gorm"
	"log"
	"main/pkg/db/postgres"
	"main/utils"
)

type Interface[T any] interface {
	GetAll(
		ctx context.Context,
		filter map[string]interface{},
		scopes ...func(db *gorm.DB) *gorm.DB,
	) (results []T, err error)

	GetAllWithPagination(
		ctx context.Context,
		filter map[string]interface{},
		scopes ...func(db *gorm.DB) *gorm.DB,
	) (results []T, count int64, err error)

	Get(
		ctx context.Context,
		filter map[string]interface{},
		scopes ...func(db *gorm.DB) *gorm.DB,
	) (result T, err error)
}

type Repository[T any] struct {
	Db *postgres.DbCluster
}

func (r *Repository[T]) GetAll(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (results []T, err error) {
	logTag := utils.LogPrefix(ctx, "Repository.GetAll")

	tx := r.Db.GetSlaveDB(ctx).Model(&results).Where(filter).Scopes(scopes...).Find(&results)
	if tx.Error != nil {
		log.Println(logTag, "Error while fetching records:", tx.Error)

		return
	}

	return
}

func (r *Repository[T]) GetAllWithPagination(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (results []T, count int64, err error) {
	logTag := utils.LogPrefix(ctx, "Repository.GetAllWithPagination")

	tx := r.Db.GetSlaveDB(ctx).Model(&results).Where(filter).Scopes(scopes...).Count(&count)
	if tx.Error != nil || count == 0 {
		log.Println(logTag, "Error while fetching records:", tx.Error)

		return nil, 0, tx.Error
	}

	tx = r.Db.GetSlaveDB(ctx).Model(&results).Where(filter).Scopes(scopes...).Find(&results)
	if tx.Error != nil {
		log.Println(logTag, "Error while fetching records:", tx.Error)

		return
	}

	return
}

func (r *Repository[T]) Get(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (result T, err error) {
	logTag := utils.LogPrefix(ctx, "Repository.Get")

	tx := r.Db.GetSlaveDB(ctx).Model(&result).Where(filter).Scopes(scopes...).First(&result)
	if tx.Error != nil {
		log.Println(logTag, "Error while fetching records:", tx.Error)

		return result, tx.Error
	}

	return
}
