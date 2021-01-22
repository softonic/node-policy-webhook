package nodepolicy

import (
	nodepolicy_api "github.com/nxmatic/admission-webhook-controller/apis/nodepolicyprofile/v1alpha1"
	"github.com/nxmatic/admission-webhook-controller/pkg/plugins/nodepolicy/reviewer"
	"github.com/nxmatic/admission-webhook-controller/pkg/plugins/spi"
	reviewer_spi "github.com/nxmatic/admission-webhook-controller/pkg/plugins/spi/reviewer"
	"github.com/pkg/errors"
	core_api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	_name        string                                            = "nodepolicyprofile"
	_podResource schema.GroupVersionResource                       = nodepolicy_api.PodsResource
	_podHook     reviewer_spi.Hook                                 = &podHook{}
	_plugin      spi.Plugin                                        = &plugin{}
	_hooks       map[schema.GroupVersionResource]reviewer_spi.Hook = map[schema.GroupVersionResource]reviewer_spi.Hook{
		_podResource: _podHook,
	}
)

func SupplyPlugin() spi.Plugin {
	return _plugin
}

type (
	plugin struct {
	}
	podHook struct {
	}
)

func (p *plugin) Name() string {
	return _name
}

func (p *plugin) Add(manager manager.Manager, client dynamic.Interface) error {
	scheme := manager.GetScheme()
	if err := nodepolicy_api.SchemeBuilder.AddToScheme(scheme); err != nil {
		return errors.Wrap(err, "failed to setup scheme")
	}
	if err := core_api.SchemeBuilder.AddToScheme(scheme); err != nil {
		return errors.Wrap(err, "failed to load core scheme")
	}
	reviewer_spi.Add(_name, manager, client, _hooks)
	return nil
}

func (h *podHook) Review(s *reviewer_spi.GivenStage) *reviewer_spi.WhenStage {
	return reviewer.Given().
		The().RequestedObject(s).And().
		The().RequestedKind().IsAPod().End().And().
		The().RequestedProfile().Exists().End().
		End()
}
