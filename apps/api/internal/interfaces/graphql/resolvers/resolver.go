package resolvers

import bootstrap "github.com/JosePintos/tora/apps/api/internal/platform/bootstrap"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	Container *bootstrap.Container
}

func New(container *bootstrap.Container) *Resolver {
	return &Resolver{
		Container: container,
	}
}
