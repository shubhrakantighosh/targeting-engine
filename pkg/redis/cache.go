package redis

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

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

func (r *Redis) MSet(ctx context.Context, keyValue map[string]any) error {
	values := make([]any, 0)
	for k, v := range keyValue {
		b, err := r.serializer.Marshal(v)
		if err != nil {
			log.Printf("[Cache] MSet Serialise Cache key %v failed. err: %s", v, err)

			return err
		}

		values = append(values, k, string(b))
	}

	err := r.Client.MSet(ctx, values...).Err()
	if err != nil {
		log.Printf("[Cache] MSet Cache keys %v failed. err: %s", values, err)

		return err
	}

	return nil
}

func (r *Redis) MGet(ctx context.Context, keys []string) ([]interface{}, error) {
	return r.Client.MGet(ctx, keys...).Result()
}
