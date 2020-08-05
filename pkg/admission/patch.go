package admission

import (
	"github.com/softonic/node-policy-webhook/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"reflect"
)

type PatcherInterface interface {
	CreatePatch(pod *v1.Pod, nodePolicyProfile *v1alpha1.NodePolicyProfile) *[]PatchOperation
}

type Patcher struct{}

type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func NewPatcher() PatcherInterface {
	return &Patcher{}
}

func (p *Patcher) CreatePatch(pod *v1.Pod, nodePolicyProfile *v1alpha1.NodePolicyProfile) *[]PatchOperation {
	var patch = &[]PatchOperation{}

	p.addNodeSelectorPatch(nodePolicyProfile, patch)
	p.addTolerationsPatch(pod, nodePolicyProfile, patch)
	p.addNodeAffinityPatch(pod, nodePolicyProfile, patch)
	return patch
}

func (p *Patcher) addNodeAffinityPatch(pod *v1.Pod, nodePolicyProfile *v1alpha1.NodePolicyProfile, patch *[]PatchOperation) {
	if reflect.DeepEqual(nodePolicyProfile.Spec.NodeAffinity, v1.NodeAffinity{}) {
		return
	}

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

	*patch = append(*patch, PatchOperation{
		Op:    "replace",
		Path:  "/spec/affinity",
		Value: affinity,
	})
}

func (p *Patcher) addTolerationsPatch(pod *v1.Pod, nodePolicyProfile *v1alpha1.NodePolicyProfile, patch *[]PatchOperation) {
	if pod.Spec.Tolerations == nil && nodePolicyProfile.Spec.Tolerations == nil {
		return
	}
	tolerations := []v1.Toleration{}

	tolerations = append(tolerations, pod.Spec.Tolerations...)

	tolerationEqual := false

	for _, tolerationPod := range pod.Spec.Tolerations {
		for _, tolerationProfile := range nodePolicyProfile.Spec.Tolerations {
			if reflect.DeepEqual(tolerationPod, tolerationProfile) {
				tolerationEqual = true
			}
		}
	}

	if tolerationEqual == false {
		tolerations = append(tolerations, nodePolicyProfile.Spec.Tolerations...)
	}

	*patch = append(*patch, PatchOperation{
		Op:    "replace",
		Path:  "/spec/tolerations",
		Value: tolerations,
	})
}

func (p *Patcher) addNodeSelectorPatch(nodePolicyProfile *v1alpha1.NodePolicyProfile, patch *[]PatchOperation) {
	if nodePolicyProfile.Spec.NodeSelector == nil {
		return
	}

	nodeSelector := make(map[string]string)

	for key, value := range nodePolicyProfile.Spec.NodeSelector {
		nodeSelector[key] = value
	}

	*patch = append(*patch, PatchOperation{
		Op:    "replace",
		Path:  "/spec/nodeSelector",
		Value: nodeSelector,
	})
}
