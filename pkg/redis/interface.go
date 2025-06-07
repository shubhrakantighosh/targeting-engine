package redis

import (
	"context"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) (done bool, err error)
	Get(ctx context.Context, key string, out interface{}) (found bool, err error)
	MSet(ctx context.Context, keyValue map[string]any) error
	MGet(ctx context.Context, keys []string) ([]interface{}, error)
	Unlink(ctx context.Context, keys []string) (int64, error)
}
