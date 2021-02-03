package gcpauthpolicy

import (
	gcpauth_api "github.com/nuxeo/k8s-policy-controller/apis/gcpauthpolicyprofile/v1alpha1"
	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/gcpauthpolicy/reconciler"
	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/gcpauthpolicy/reviewer"
	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi"
	reviewer_spi "github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi/reviewer"
	"github.com/pkg/errors"
	core_api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	_name                   string                                            = "gcpauthpolicyprofile"
	_serviceaccountResource schema.GroupVersionResource                       = core_api.SchemeGroupVersion.WithResource("serviceaccounts")
	_serviceaccountHook     reviewer_spi.Hook                                 = &serviceaccountHook{}
	_plugin                 spi.Plugin                                        = &plugin{}
	_hooks                  map[schema.GroupVersionResource]reviewer_spi.Hook = map[schema.GroupVersionResource]reviewer_spi.Hook{
		_serviceaccountResource: _serviceaccountHook,
	}
)

type (
	plugin struct {
	}
	serviceaccountHook struct{}
)

func SupplyPlugin() spi.Plugin {
	return _plugin
}

func (p *plugin) Name() string {
	return _name
}

func (p *plugin) Add(manager manager.Manager, client dynamic.Interface) error {
	scheme := manager.GetScheme()
	if err := gcpauth_api.SchemeBuilder.AddToScheme(scheme); err != nil {
		return errors.Wrap(err, "failed to load gcpauthpolicyprofile scheme")
	}
	if err := core_api.SchemeBuilder.AddToScheme(scheme); err != nil {
		return errors.Wrap(err, "failed to load core scheme")
	}

	reconciler.Add(manager, client)
	reviewer_spi.Add(_name, manager, client, _hooks)
	return nil
}

func (h *serviceaccountHook) Review(s *reviewer_spi.GivenStage) *reviewer_spi.WhenStage {
	return reviewer.Given().
		The().RequestedObject(s).And().
		The().RequestedKind().IsAServiceAccount().End().
		The().RequestedProfile().Exists().And().
		The().SecretIsAvailable().End().End()
}
