package redis

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) (done bool, err error)
	Get(ctx context.Context, key string, out interface{}) (found bool, err error)
	Unlink(ctx context.Context, keys []string) (int64, error)
}

func (r *Redis) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) (done bool, err error) {
	data, err := r.serializer.Marshal(value)
	if err != nil {
		log.Printf("[Cache] Failed to marshal value for key %s: %v\n", key, err)
		return false, err
	}

	if err = r.Client.Set(ctx, key, data, ttl).Err(); err != nil {
		log.Printf("[Cache] Failed to set key %s in Redis: %v\n", key, err)
		return false, err
	}

	return true, nil
}

func (r *Redis) Get(ctx context.Context, key string, out interface{}) (found bool, err error) {
	cmd := r.Client.Get(ctx, key)
	if err = cmd.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil // key doesn't exist
		}
		log.Printf("[Cache] Failed to get key %s from Redis: %v\n", key, err)
		return false, err
	}

	str, err := cmd.Bytes()
	if err != nil {
		log.Printf("[Cache] Failed to read Redis bytes for key %s: %v\n", key, err)
		return false, err
	}

	if err = r.serializer.Unmarshal(str, out); err != nil {
		log.Printf("[Cache] Failed to unmarshal value for key %s: %v\n", key, err)
		return false, err
	}

	return true, nil
}

func (r *Redis) Unlink(ctx context.Context, keys []string) (int64, error) {
	result := r.Client.Unlink(ctx, keys...)
	if err := result.Err(); err != nil {
		log.Printf("[Cache] Failed to unlink keys %v: %v\n", keys, err)
		return 0, err
	}
	return result.Val(), nil
}
