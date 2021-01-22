package controller

import (
	"github.com/nxmatic/admission-webhook-controller/pkg/plugins/spi"

	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewManagerWithOptions(konfig *rest.Config, opts ctrl.Options, plugin spi.Plugin) (manager.Manager, error) {
	manager, err := manager.New(konfig, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create manager")
	}

	client, err := dynamic.NewForConfig(konfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create dynamic interface")
	}

	opts.Logger.Info("Registering Components.")

	plugin.Add(manager, client)

	return manager, nil
}
