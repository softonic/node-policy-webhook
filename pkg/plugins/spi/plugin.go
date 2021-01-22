package spi

import (
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type (
	Supplier func() Plugin

	Plugin interface {
		Name() string
		Add(manager manager.Manager, client dynamic.Interface) error
	}
)
