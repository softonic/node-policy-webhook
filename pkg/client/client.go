package client

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	konfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

func NewClient() (*rest.Config, dynamic.Interface, error) {
	konfig, err := konfig.GetConfig()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error loading konfig")
	}
	dynamic, err := dynamic.NewForConfig(konfig)
	if err != nil {
		return nil, nil, errors.New("Error configuring client")
	}
	return konfig, dynamic, nil
}
