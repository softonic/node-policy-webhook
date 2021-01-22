package reconciler

import (
	gcpauthpolicy_api "github.com/nxmatic/admission-webhook-controller/apis/gcpauthpolicyprofile/v1alpha1"
	"github.com/nxmatic/admission-webhook-controller/pkg/plugins/gcpauthpolicy/k8s"

	"github.com/pkg/errors"
	core_api "k8s.io/api/core/v1"
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
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
		Type: &gcpauthpolicy_api.GCPAuthPolicyProfile{
			TypeMeta: meta_api.TypeMeta{
				APIVersion: gcpauthpolicy_api.SchemeGroupVersion.String(),
				Kind:       gcpauthpolicy_api.GCPAuthpolicyprofileKind.String(),
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
	err = c.Watch(secretResource, &enqueueRequestForOwner{})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
