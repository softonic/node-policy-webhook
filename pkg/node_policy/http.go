package node_policy

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/softonic/node-policy-webhook/api/v1alpha1"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"net/http"
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func createPatch(pod *v1.Pod, profileName string) ([]byte, error) {
	patch := make([]patchOperation, 0)

	nodePolicyProfile, err := getNodePolicyProfile(profileName)
	if err != nil {
		return nil, err
	}

	patch = addNodeSelectorPatch(nodePolicyProfile, patch)

	patch = addTolerationsPatch(pod, nodePolicyProfile, patch)

	patch = addNodeAffinityPatch(pod, nodePolicyProfile, patch)


	return json.Marshal(patch)
}

func addNodeAffinityPatch(pod *v1.Pod, nodePolicyProfile *v1alpha1.NodePolicyProfile, patch []patchOperation) []patchOperation {
	affinity := v1.Affinity{}

	affinity.NodeAffinity = &nodePolicyProfile.Spec.NodeAffinity

	if pod.Spec.Affinity != nil {
		if pod.Spec.Affinity.PodAntiAffinity != nil {
			affinity.PodAntiAffinity = pod.Spec.Affinity.PodAntiAffinity
		}

		if pod.Spec.Affinity.PodAffinity != nil {
			affinity.PodAffinity = pod.Spec.Affinity.PodAffinity
		}
	}

	return append(patch, patchOperation{
		Op:    "add",
		Path:  "/spec/affinity",
		Value: affinity,
	})
}

func addTolerationsPatch(pod *v1.Pod, nodePolicyProfile *v1alpha1.NodePolicyProfile, patch []patchOperation) []patchOperation {
	tolerations := []v1.Toleration{}

	tolerations = append(tolerations, pod.Spec.Tolerations...)

	tolerations = append(tolerations, nodePolicyProfile.Spec.Tolerations...)

	return append(patch, patchOperation{
		Op:    "replace",
		Path:  "/spec/tolerations",
		Value: tolerations,
	})
}

func addNodeSelectorPatch(nodePolicyProfile *v1alpha1.NodePolicyProfile, patch []patchOperation) []patchOperation {
	nodeSelector := make(map[string]string)

	for key, value := range nodePolicyProfile.Spec.NodeSelector {
		nodeSelector[key] = value
	}

	return append(patch, patchOperation{
		Op:    "add",
		Path:  "/spec/nodeSelector",
		Value: nodeSelector,
	})
}

func getNodePolicyProfile(profileName string) (*v1alpha1.NodePolicyProfile, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.New("Error configuring client")
	}
	client, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, errors.New("Error creating client")
	}

	resourceScheme := v1alpha1.SchemeBuilder.GroupVersion.WithResource("nodepolicyprofiles")

	resp, err := client.Resource(resourceScheme).Get(context.TODO(), profileName, v12.GetOptions{})
	if err != nil {
		klog.Errorf("Error getting NodePolicyProfile %s (%v)", profileName, err)
		return nil, errors.New("Error getting NodePolicyProfile")
	}

	nodePolicyProfile := &v1alpha1.NodePolicyProfile{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), nodePolicyProfile)
	if err != nil {
		return nil, errors.New("NodePolicyProfile not found")
	}
	return nodePolicyProfile, nil
}

func getProfile(metadata *v12.ObjectMeta) (string, error) {

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	if profileName, ok := annotations["softonic.io/profile"]; ok {
		klog.Infof("Successfully found annotation softonic.io/profile. With profile: %v", profileName)
		return profileName, nil
	}

	return "", errors.New("Annotation not found")
}

func mutate(ar *v1beta1.AdmissionReview) (*v1beta1.AdmissionResponse, error) {

	resp := &v1beta1.AdmissionResponse{}
	req := ar.Request

	var pod v1.Pod
	err := json.Unmarshal(req.Object.Raw, &pod)
	if err != nil {
		return resp, err
	}

	klog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v (%v) UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, pod.Name, req.UID, req.Operation, req.UserInfo)

	profile, err := getProfile(&pod.ObjectMeta)
	if err != nil {
		return admissionAllowedResponse(pod), nil
	}

	patchBytes, err := createPatch(&pod, profile)

	if err != nil {
		return resp, err
	}

	patchType := v1beta1.PatchTypeJSONPatch

	return &v1beta1.AdmissionResponse{
		Result: &v12.Status{
			Status: "Success",
		},
		Patch:            patchBytes,
		PatchType:        &patchType,
		Allowed:          true,
		UID:              ar.Request.UID,
	}, nil

}

func admissionAllowedResponse(pod v1.Pod) *v1beta1.AdmissionResponse {
	klog.Infof("Skipping mutation for %s/%s due to policy check", pod.Namespace, pod.Name)
	return &v1beta1.AdmissionResponse{
		Allowed: true,
	}
}

func HttpHandler(w http.ResponseWriter, r *http.Request) {

	if err, status := validateRequest(r); err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	admissionReview := &v1beta1.AdmissionReview{}
	err := json.NewDecoder(r.Body).Decode(admissionReview)
	failIfError(w, err)

	admissionReview.Response = getAdmissionResponse(admissionReview)

	err = writeResponse(w, admissionReview)
	failIfError(w, err)
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

func writeResponse(w http.ResponseWriter, admissionReview *v1beta1.AdmissionReview) error {
	resp, err := json.Marshal(admissionReview)
	if err != nil {
		return err
	}
	if _, err := w.Write(resp); err != nil {
		return err
	}
	return nil
}

func getAdmissionResponse(admissionReview *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {

	response := &v1beta1.AdmissionResponse{}

	response, err := mutate(admissionReview)
	if err != nil {
		admissionReview.Response = newAdmissionError(err)
	}
	return response
}

func newAdmissionError(err error) *v1beta1.AdmissionResponse {
	klog.Errorf("Error %v", err)
	return &v1beta1.AdmissionResponse{
		Result: &v12.Status{
			Message: err.Error(),
			Status:  "Fail",
		},
	}
}