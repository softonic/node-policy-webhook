package reviewer

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type (
	Hook interface {
		Review(s *GivenStage) *WhenStage
	}
)

func Add(name string, manager manager.Manager, client dynamic.Interface, hooks map[schema.GroupVersionResource]Hook) {
	webhooks := manager.GetWebhookServer()
	logger := manager.GetLogger().WithName(name)
	for resource, hook := range hooks {
		path, webhook := newMutationWebhook(resource, hook, client, logger)
		webhooks.Register(path, webhook)
	}
}

func newMutationWebhook(resource schema.GroupVersionResource, hook Hook, client dynamic.Interface, logger logr.Logger) (string, *webhook.Admission) {
	return mutationWebhookPath(resource), &webhook.Admission{
		Handler: &mutationHandler{
			Reviewer: NewAdmissionReviewer(hook, client, logger),
			Logger:   logger,
		},
	}
}

func mutationWebhookPath(gvr schema.GroupVersionResource) string {
	path := "/mutate"
	if gvr.Group != "" {
		path += "-" + gvr.Group
	}
	path += "-" + gvr.Version
	path += "-" + gvr.Resource
	return path
}

type mutationHandler struct {
	Reviewer *AdmissionReviewer
	Logger   logr.Logger
}

func (h *mutationHandler) Handle(ctx context.Context, request admission.Request) admission.Response {
	response := h.Reviewer.PerformAdmissionReview(&request.AdmissionRequest)
	return admission.Response{
		AdmissionResponse: *response,
	}
}
