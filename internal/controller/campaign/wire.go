//go:build wireinject
// +build wireinject

package campaign

import (
	"context"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"main/pkg/db/postgres"
)

func Wire(ctx context.Context, db *postgres.Db, redis *redis.Client) *Controller {
	panic(wire.Build(ProviderSet))
}
