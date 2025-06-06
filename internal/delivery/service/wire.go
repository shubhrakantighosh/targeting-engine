//go:build wireinject
// +build wireinject

package service

import (
	"context"
	"github.com/google/wire"
	"main/pkg/db/postgres"
	oredis "main/pkg/redis"
)

func Wire(ctx context.Context, db *postgres.DbCluster, redis *oredis.Redis) *Service {
	panic(wire.Build(ProviderSet))
}
