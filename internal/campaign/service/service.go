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
	cacheKey := "campaigns_cache"

	// Try fetching from Redis
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil && len(cached) > 0 {
		var result model.Campaigns
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			return result, int64(len(result)), nil
		}
		// If unmarshal fails, log and fallback to DB
	}

	// Fallback to DB
	data, err := s.repo.Repository.GetAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Save to Redis (marshal first)
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, 0, err
	}

	if err := s.redis.Set(ctx, cacheKey, jsonBytes, 10*time.Second).Err(); err != nil {
		// Log the error but don’t fail the request
		// e.g., log.Println("Failed to cache campaigns:", err)
	}

	return data, int64(len(data)), nil
}
