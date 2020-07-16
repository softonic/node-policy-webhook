package admission

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/softonic/node-policy-webhook/api/v1alpha1"
	"github.com/softonic/node-policy-webhook/pkg/log"
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

type NodePolicyAdmissionReviewer struct {
	client dynamic.Interface
}

func NewNodePolicyAdmissionReviewer(client dynamic.Interface) *NodePolicyAdmissionReviewer {
	return &NodePolicyAdmissionReviewer{
		client: client,
	}
}
// PerformAdmissionReview : It generates the Adminission Review Response
func (n *NodePolicyAdmissionReviewer) PerformAdmissionReview(admissionReview *v1beta1.AdmissionReview) {
	pod, err := getPod(admissionReview)
	if err != nil {
		admissionReview.Response = newAdmissionError(pod, err)
		return
	}

	profile, err := getProfile(pod)
	if err != nil {
		admissionReview.Response = admissionAllowedResponse(pod)
		return
	}

	nodePolicyProfile, err := n.getNodePolicyProfile(profile)
	if err != nil {
		admissionReview.Response = newAdmissionError(pod, err)
		return
	}

	patchBytes, err := createPatch(pod, nodePolicyProfile)
	if err != nil {
		admissionReview.Response = newAdmissionError(pod, err)
		return
	}

	klog.V(log.INFO).Infof("Patching pod %s/%s", pod.Namespace, pod.Name)
	patchType := v1beta1.PatchTypeJSONPatch

	admissionReview.Response = &v1beta1.AdmissionResponse{
		Result: &v12.Status{
			Status: "Success",
		},
		Patch:            patchBytes,
		PatchType:        &patchType,
		Allowed:          true,
		UID:              admissionReview.Request.UID,
	}
}

func newAdmissionError(pod *v1.Pod, err error) *v1beta1.AdmissionResponse {
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

func admissionAllowedResponse(pod *v1.Pod) *v1beta1.AdmissionResponse {
	klog.V(log.EXTENDED).Infof("Skipping admission review for pod %s/%s", pod.Namespace, pod.Name)
	return &v1beta1.AdmissionResponse{
		Allowed: true,
	}
}

func (n *NodePolicyAdmissionReviewer) getNodePolicyProfile(profileName string) (*v1alpha1.NodePolicyProfile, error) {
	resourceScheme := v1alpha1.SchemeBuilder.GroupVersion.WithResource("nodepolicyprofiles")

	resp, err := n.client.Resource(resourceScheme).Get(context.TODO(), profileName, v12.GetOptions{})
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

func getProfile(pod *v1.Pod) (string, error) {
	annotations := pod.ObjectMeta.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	if profileName, ok := annotations["softonic.io/profile"]; ok {
		klog.V(log.INFO).Infof("Successfully found annotation softonic.io/profile. With profile: %v", profileName)
		return profileName, nil
	}

	return "", errors.New("Annotation not found")
}

func getPod(admissionReview *v1beta1.AdmissionReview) (*v1.Pod, error) {
	var pod v1.Pod
	err := json.Unmarshal(admissionReview.Request.Object.Raw, &pod)
	if err != nil {
		return nil, err
	}
	return &pod, nil
}
