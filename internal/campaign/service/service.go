package service

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"main/internal/campaign/repository"
	"main/internal/model"
	"sync"
	"time"
)

type Service struct {
	repo  repository.Repository
	redis redis.Client
}

var (
	syncOnce sync.Once
	service  *Service
)

func NewService(repo *repository.Repository, redis *redis.Client) *Service {
	syncOnce.Do(func() {
		service = &Service{
			repo:  *repo,
			redis: *redis,
		}
	})

	return service
}

type Interface interface {
}

func (s *Service) GetAll(ctx context.Context, filter map[string]any) (model.Campaigns, int64, error) {
	cm := s.redis.Get(ctx, "hola")
	ok, err := cm.Bool()
	if err != nil {
		return nil, 0, err
	}

	if ok {
		f1, er := cm.Result()
		if er != nil {
			return nil, 0, err
		}

		res := make(model.Campaigns, 0)
		err = json.Unmarshal([]byte(f1), &res)
		if err != nil {
			return nil, 0, err
		}

		return res, 0, nil
	}

	data, _, err := s.repo.Repository.GetAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	s.redis.Set(ctx, "hola", data, time.Second*5)
	return data, 0, nil
}
