package admission

import (
	"encoding/json"
	"errors"
	"github.com/softonic/node-policy-webhook/pkg/log"
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const PROFILE_ANNOTATION = "nodepolicy.softonic.io/profile"

type AdmissionReviewer struct {
	fetcher FetcherInterface
	patcher PatcherInterface
}

func NewNodePolicyAdmissionReviewer(fetcher FetcherInterface, patcher PatcherInterface) *AdmissionReviewer {
	return &AdmissionReviewer{
		fetcher: fetcher,
		patcher: patcher,
	}
}

// PerformAdmissionReview : It generates the Adminission Review Response
func (r *AdmissionReviewer) PerformAdmissionReview(admissionReview *v1beta1.AdmissionReview) {
	pod, err := r.getPod(admissionReview)
	if err != nil {
		admissionReview.Response = r.newAdmissionError(pod, err)
		return
	}

	profile, err := r.getProfile(pod)
	if err != nil {
		admissionReview.Response = r.admissionAllowedResponse(pod)
		return
	}

	nodePolicyProfile, err := r.fetcher.Get(profile)
	if err != nil {
		admissionReview.Response = r.newAdmissionError(pod, err)
		return
	}

	patch := r.patcher.CreatePatch(pod, nodePolicyProfile)
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		admissionReview.Response = r.newAdmissionError(pod, err)
		return
	}

	klog.V(log.INFO).Infof("Patching pod %s/%s", pod.Namespace, pod.Name)
	patchType := v1beta1.PatchTypeJSONPatch

	admissionReview.Response = &v1beta1.AdmissionResponse{
		Result: &v12.Status{
			Status: "Success",
		},
		Patch:     patchBytes,
		PatchType: &patchType,
		Allowed:   true,
		UID:       admissionReview.Request.UID,
	}
}

func (r *AdmissionReviewer) newAdmissionError(pod *v1.Pod, err error) *v1beta1.AdmissionResponse {
	if pod != nil {
		klog.Errorf("Pod %s/%s failed admission review: %v", pod.Namespace, pod.Name, err)
	} else {
		klog.Errorf("Failed admission review: %v", err)
	}
	return &v1beta1.AdmissionResponse{
		Result: &v12.Status{
			Message: err.Error(),
			Status:  "Fail",
		},
	}
}

func (r *AdmissionReviewer) admissionAllowedResponse(pod *v1.Pod) *v1beta1.AdmissionResponse {
	klog.Errorf("Skipping admission review for pod %s/%s", pod.Namespace, pod.Name)
	return &v1beta1.AdmissionResponse{
		Allowed: true,
	}
}

func (r *AdmissionReviewer) getProfile(pod *v1.Pod) (string, error) {
	annotations := pod.ObjectMeta.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	if profileName, ok := annotations[PROFILE_ANNOTATION]; ok {
		klog.V(log.INFO).Infof("Successfully found annotation softonic.io/profile. With profile: %v", profileName)
		return profileName, nil
	}

	return "", errors.New("Annotation not found")
}

func (r *AdmissionReviewer) getPod(admissionReview *v1beta1.AdmissionReview) (*v1.Pod, error) {
	var pod v1.Pod
	if admissionReview.Request == nil {
		return nil, errors.New("Request is nil")
	}
	if admissionReview.Request.Object.Raw == nil {
		return nil, errors.New("Request object raw is nil")
	}
	err := json.Unmarshal(admissionReview.Request.Object.Raw, &pod)
	if err != nil {
		return nil, err
	}
	return &pod, nil
}
