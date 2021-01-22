package reconciler

import (
	"context"
	"time"

	core_api "k8s.io/api/core/v1"
	errors_api "k8s.io/apimachinery/pkg/api/errors"

	"github.com/nxmatic/admission-webhook-controller/apis/gcpauthpolicyprofile/v1alpha1"
	"github.com/nxmatic/admission-webhook-controller/pkg/plugins/gcpauthpolicy/k8s"
	k8s_spi "github.com/nxmatic/admission-webhook-controller/pkg/plugins/spi/k8s"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type (
	reconciler struct {
		*k8s.Interface
	}
	replicator struct {
		k8s_spi.Replicator
		k8s_spi.ReplicatorSource
	}
)

var (
	requeueOnError reconcile.Result = reconcile.Result{RequeueAfter: 5 * time.Minute}
)

func (r *reconciler) Reconcile(ctx context.Context, o reconcile.Request) (reconcile.Result, error) {
	profile, err := r.Interface.GetProfile(o.Name)
	if err != nil {
		if !errors_api.IsNotFound(err) {
			return requeueOnError, err
		}
		return r.deleteHandler(profile)
	}
	return r.updateHandler(profile)
}

func (r *reconciler) deleteHandler(profile *v1alpha1.GCPAuthPolicyProfile) (reconcile.Result, error) {
	err := r.Interface.DeleteSecrets(profile)
	if err != nil {
		return requeueOnError, err
	}
	return reconcile.Result{}, nil
}

func (r *reconciler) updateHandler(profile *v1alpha1.GCPAuthPolicyProfile) (reconcile.Result, error) {
	iterator, err := r.Interface.ListSecrets(profile)
	if err != nil {
		return reconcile.Result{}, err
	}
	replicator := r.NewReplicator()
	iterator.Apply(func(secret *core_api.Secret) error {
		_, err := replicator.UpdateReplicatedSecret(secret, &profile.Spec)
		return err
	})

	return reconcile.Result{}, nil
}
