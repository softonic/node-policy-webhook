package http

import (
	"encoding/json"
	"errors"
	"github.com/softonic/node-policy-webhook/pkg/admission"
	"io"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/klog"
	"net/http"
)

type HttpHandler struct {
	reviewer *admission.AdmissionReviewer
}

func NewHttpHanlder(reviewer *admission.AdmissionReviewer) *HttpHandler {
	return &HttpHandler{
		reviewer: reviewer,
	}
}

func (h *HttpHandler) MutationHandler(w http.ResponseWriter, r *http.Request) {

	if err, status := validateRequest(r); err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	resp, err := h.getResponse(r.Body)
	if err != nil {
		failIfError(w, err)
	}

	if _, err := w.Write(resp); err != nil {
		failIfError(w, err)
	}

}

func (h *HttpHandler) getResponse(rawAdmissionReview io.Reader) ([]byte, error) {
	admissionReview := &v1beta1.AdmissionReview{}
	err := json.NewDecoder(rawAdmissionReview).Decode(admissionReview)
	if err != nil {
		return nil, err
	}

	h.reviewer.PerformAdmissionReview(admissionReview)

	resp, err := json.Marshal(admissionReview)
	return resp, err
}

func failIfError(w http.ResponseWriter, err error) {
	if err != nil {
		klog.Errorf("Can't write response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func validateRequest(r *http.Request) (error, int) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return errors.New("invalid Content-Type, expect `application/json`"), http.StatusUnsupportedMediaType
	}
	return nil, 0
}
