package repository

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewRepository,

	// bind each one of the interfaces
	wire.Bind(new(Interface), new(*Repository)),
)
