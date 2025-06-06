package repository

import (
	"main/internal/model"
	"main/internal/repository"
	"main/pkg/db/postgres"
	"sync"
)

type Repository struct {
	repository.Repository[model.Campaign]
}

var (
	syncOnce sync.Once
	repo     *Repository
)

func NewRepository(db *postgres.DbCluster) *Repository {
	syncOnce.Do(func() {
		repo = &Repository{
			Repository: repository.Repository[model.Campaign]{Db: db},
		}
	})

	return repo
}

type Interface interface {
}
