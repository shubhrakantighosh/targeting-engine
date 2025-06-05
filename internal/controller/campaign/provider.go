package campaign

import (
	"github.com/google/wire"
	"main/internal/campaign/repository"
	"main/internal/campaign/service"
)

var ProviderSet = wire.NewSet(
	NewController,
	service.NewService,
	repository.NewRepository,

	// bind each one of the interfaces
	wire.Bind(new(Interface), new(*Controller)),
	wire.Bind(new(service.Interface), new(*service.Service)),
	wire.Bind(new(repository.Interface), new(*repository.Repository)),
)
