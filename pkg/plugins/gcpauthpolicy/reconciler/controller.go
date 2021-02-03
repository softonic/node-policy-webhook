package reconciler

import (
	gcpauthpolicy_api "github.com/nuxeo/k8s-policy-controller/apis/gcpauthpolicyprofile/v1alpha1"
	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/gcpauthpolicy/k8s"

	"github.com/pkg/errors"
	core_api "k8s.io/api/core/v1"
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func Add(mgr manager.Manager, client dynamic.Interface) error {
	reconciler := &reconciler{
		k8s.NewInterface(client),
	}
	return add(mgr, reconciler)
}

// add adds a newReconcilierConfiguration Controller to mgr with r as the reconcile.Reconciler.
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a newReconcilierConfiguration controller
	c, err := controller.New("gcpauthpolicyprofiles", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return errors.WithStack(err)
	}

	// Watch for changes to primary resource GCPAuthPolicyProfile
	decorator := profileDecorator{handler: &handler.EnqueueRequestForObject{}}
	profileResource := &source.Kind{
		Type: &gcpauthpolicy_api.Profile{
			TypeMeta: meta_api.TypeMeta{
				APIVersion: gcpauthpolicy_api.SchemeGroupVersion.String(),
				Kind:       gcpauthpolicy_api.ProfileKind.String(),
			},
		},
	}
	err = c.Watch(profileResource, &decorator)
	if err != nil {
		return errors.WithStack(err)
	}

	// Watch for changes to secondary resource Secrets and requeue the owner
	secretResource := &source.Kind{
		Type: &core_api.Secret{
			TypeMeta: meta_api.TypeMeta{
				APIVersion: core_api.SchemeGroupVersion.String(),
				Kind:       "Secrets",
			},
		},
	}
	predicate, err := predicate.LabelSelectorPredicate(
		meta_api.LabelSelector{
			MatchLabels: map[string]string{
				gcpauthpolicy_api.WatchKey.String(): "true",
			}})
	if err != nil {
		return errors.WithStack(err)
	}
	err = c.Watch(secretResource, &enqueueRequestForOwner{}, predicate)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
