//go:build wireinject
// +build wireinject

package repository

import (
	"context"
	"github.com/google/wire"
	"main/pkg/db/postgres"
)

func Wire(ctx context.Context, db *postgres.DbCluster) *Repository {
	panic(wire.Build(ProviderSet))
}
