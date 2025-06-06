package repository

import (
	"context"
	"gorm.io/gorm"
	"log"
	"main/pkg/apperror"
	"main/pkg/db/postgres"
	"main/util"
	"net/http"
)

type Repository[T any] struct {
	Db *postgres.DbCluster
}

func (r *Repository[T]) GetAll(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (results []T, err apperror.Error) {
	logTag := util.LogPrefix(ctx, "Repository.GetAll")

	tx := r.Db.GetSlaveDB(ctx).Model(&results).Where(filter).Scopes(scopes...).Find(&results)
	if tx.Error != nil {
		log.Println(logTag, "Error while fetching records:", tx.Error)

		return nil, apperror.New(tx.Error, http.StatusBadRequest)
	}

	return
}

func (r *Repository[T]) GetAllWithPagination(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (results []T, count int64, err apperror.Error) {
	logTag := util.LogPrefix(ctx, "Repository.GetAllWithPagination")

	tx := r.Db.GetSlaveDB(ctx).Model(&results).Where(filter).Scopes(scopes...).Count(&count)
	if tx.Error != nil {
		log.Println(logTag, "Error while fetching records:", tx.Error)

		return nil, 0, apperror.New(tx.Error, http.StatusBadRequest)
	}

	if count == 0 {
		return nil, 0, apperror.Error{}
	}

	tx = r.Db.GetSlaveDB(ctx).Model(&results).Where(filter).Scopes(scopes...).Find(&results)
	if tx.Error != nil {
		log.Println(logTag, "Error while fetching records:", tx.Error)

		return nil, 0, apperror.New(tx.Error, http.StatusBadRequest)
	}

	return
}

func (r *Repository[T]) Get(
	ctx context.Context,
	filter map[string]interface{},
	scopes ...func(db *gorm.DB) *gorm.DB,
) (result T, err apperror.Error) {
	//logTag := util.LogPrefix(ctx, "Repository.Get")

	r.Db.GetSlaveDB(ctx).Model(&result).Where(filter).Scopes(scopes...).First(&result)
	//if tx.Error != nil {
	//	log.Println(logTag, "Error while fetching records:", tx.Error)
	//
	//	return result, apperror.New(tx.Error, http.StatusBadRequest)
	//}

	return
}
