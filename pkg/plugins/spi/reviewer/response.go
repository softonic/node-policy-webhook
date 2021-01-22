package reviewer

import (
	"encoding/json"

	admission_api "k8s.io/api/admission/v1"
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pkg/errors"
)

func newAllowedResponse(request *admission_api.AdmissionRequest, cause error) *admission_api.AdmissionResponse {
	warnings := make([]string, 1)
	if cause != nil {
		warnings[0] = cause.Error()
	}
	return &admission_api.AdmissionResponse{
		Allowed:  true,
		Warnings: warnings,
	}
}

func newFailureResponse(request *admission_api.AdmissionRequest, cause error) *admission_api.AdmissionResponse {
	message := cause.Error()
	return &admission_api.AdmissionResponse{
		Result: &meta_api.Status{
			Message: message,
			Status:  meta_api.StatusFailure,
			Reason:  meta_api.StatusReasonInvalid,
		},
		Warnings: []string{message},
	}
}

func newJSONPatchResponse(request *admission_api.AdmissionRequest, patch []PatchOperation) *admission_api.AdmissionResponse {
	patchType := admission_api.PatchTypeJSONPatch
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return newAllowedResponse(request, errors.Wrap(err, "Cannot encode patch"))
	}
	return &admission_api.AdmissionResponse{
		Result: &meta_api.Status{
			Status: "Success",
		},
		Patch:     patchBytes,
		PatchType: &patchType,
		Allowed:   true,
		UID:       request.UID,
	}
}
