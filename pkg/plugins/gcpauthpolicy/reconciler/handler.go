package reconciler

import (
	"fmt"
	"reflect"

	gcpauthpolicy_api "github.com/nxmatic/admission-webhook-controller/apis/gcpauthpolicyprofile/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type profileDecorator struct {
	handler handler.EventHandler
}

func (e *profileDecorator) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	e.handler.Create(evt, q)
}

func (e *profileDecorator) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	if !reflect.DeepEqual(evt.ObjectOld.(*gcpauthpolicy_api.GCPAuthPolicyProfile).Spec, evt.ObjectNew.(*gcpauthpolicy_api.GCPAuthPolicyProfile).Spec) {
		log.Log.WithValues("policy", evt.ObjectNew.GetName()).Info(
			fmt.Sprintf("%T/%s has been updated", evt.ObjectNew, evt.ObjectNew.GetName()))
	}
	e.handler.Update(evt, q)
}

func (e *profileDecorator) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	e.handler.Delete(evt, q)
}

func (e *profileDecorator) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	e.handler.Generic(evt, q)
}

// enqueueRequestForOwner enqueues a Request for Secrets created by gcpauth.
type enqueueRequestForOwner struct{}

func (e *enqueueRequestForOwner) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	if req := e.getOwnerReconcileRequests(evt.Object); req != nil {
		q.Add(*req)
	}
}

func (e *enqueueRequestForOwner) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	req1 := e.getOwnerReconcileRequests(evt.ObjectOld)
	req2 := e.getOwnerReconcileRequests(evt.ObjectNew)

	if req1 != nil || req2 != nil {
		name := "unknown"
		if req1 != nil {
			name = req1.Name
		}
		if req2 != nil {
			name = req2.Name
		}

		log.Log.WithValues("profile", name).Info(
			fmt.Sprintf("%T/%s has been updated", evt.ObjectNew, evt.ObjectNew.GetName()))
	}

	if req1 != nil {
		q.Add(*req1)
		return
	}
	if req2 != nil {
		q.Add(*req2)
	}
}

func (e *enqueueRequestForOwner) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	if req := e.getOwnerReconcileRequests(evt.Object); req != nil {
		q.Add(*req)
	}
}

func (e *enqueueRequestForOwner) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	if req := e.getOwnerReconcileRequests(evt.Object); req != nil {
		q.Add(*req)
	}
}

func (e *enqueueRequestForOwner) getOwnerReconcileRequests(object metav1.Object) *reconcile.Request {
	if len(object.GetLabels()[gcpauthpolicy_api.ProfileKey.String()]) > 0 {
		return &reconcile.Request{NamespacedName: types.NamespacedName{
			Namespace: object.GetNamespace(),
			Name:      object.GetLabels()[gcpauthpolicy_api.ProfileKey.String()],
		}}
	}
	return nil
}
