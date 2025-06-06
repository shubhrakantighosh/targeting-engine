//go:build wireinject
// +build wireinject

package service

import (
	"context"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"main/pkg/db/postgres"
)

func Wire(ctx context.Context, db *postgres.DbCluster, redis *redis.Client) *Service {
	panic(wire.Build(ProviderSet))
}
