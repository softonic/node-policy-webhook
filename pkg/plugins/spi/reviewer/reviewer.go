package reviewer

import (
	"github.com/go-logr/logr"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	admission_api "k8s.io/api/admission/v1"
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/dynamic"
)

type AdmissionReviewer struct {
	meta_api.GroupVersionKind
	*admission.Decoder
	dynamic.Interface
	Logger logr.Logger
	Hook
}

// NewAdmissionReviewer allocate a reviewer for processing requested reviews
func NewAdmissionReviewer(hook Hook, service dynamic.Interface, logger logr.Logger) *AdmissionReviewer {
	return &AdmissionReviewer{
		Hook:      hook,
		Interface: service,
		Logger:    logger.WithName("reviewer"),
	}
}

// PerformAdmissionReview : It generates the Adminission Review Response
func (r *AdmissionReviewer) PerformAdmissionReview(request *admission_api.AdmissionRequest) *admission_api.AdmissionResponse {
	logger := r.Logger.WithName("perform")
	given := Given(logger, r.Interface)
	return given.
		A().Request(request).And().
		The().RequestedObject().
		NamespaceIsNot(meta_api.NamespaceSystem).And().
		IsNotNull().End().
		When(r.Hook).
		I().PatchTheRequest().
		Then().
		I().ReturnThePatch().
		OrElse().
		I().ReturnTheStatus().End().
		Response()
}

func (r *AdmissionReviewer) InjectDecoder(d *admission.Decoder) error {
	r.Decoder = d
	return nil
}
