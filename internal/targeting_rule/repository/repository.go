package repository

import (
	"main/internal/model"
	"main/internal/repository"
	"main/pkg/db/postgres"
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
