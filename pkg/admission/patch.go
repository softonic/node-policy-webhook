package admission

import (
	"encoding/json"
	"github.com/softonic/node-policy-webhook/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"reflect"
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func createPatch(pod *v1.Pod, nodePolicyProfile *v1alpha1.NodePolicyProfile) ([]byte, error) {
	var patch = []patchOperation{}

	addNodeSelectorPatch(nodePolicyProfile, &patch)
	addTolerationsPatch(pod, nodePolicyProfile, &patch)
	addNodeAffinityPatch(pod, nodePolicyProfile, &patch)

	return json.Marshal(patch)
}

func addNodeAffinityPatch(pod *v1.Pod, nodePolicyProfile *v1alpha1.NodePolicyProfile, patch *[]patchOperation) {
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

	*patch = append(*patch, patchOperation{
		Op:    "replace",
		Path:  "/spec/affinity",
		Value: affinity,
	})
}

func addTolerationsPatch(pod *v1.Pod, nodePolicyProfile *v1alpha1.NodePolicyProfile, patch *[]patchOperation) {
	if pod.Spec.Tolerations == nil && nodePolicyProfile.Spec.Tolerations == nil {
		return
	}
	tolerations := []v1.Toleration{}

	tolerations = append(tolerations, pod.Spec.Tolerations...)

	tolerations = append(tolerations, nodePolicyProfile.Spec.Tolerations...)

	*patch = append(*patch, patchOperation{
		Op:    "replace",
		Path:  "/spec/tolerations",
		Value: tolerations,
	})
}

func addNodeSelectorPatch(nodePolicyProfile *v1alpha1.NodePolicyProfile, patch *[]patchOperation) {
	if nodePolicyProfile.Spec.NodeSelector == nil {
		return
	}

	nodeSelector := make(map[string]string)

	for key, value := range nodePolicyProfile.Spec.NodeSelector {
		nodeSelector[key] = value
	}

	*patch = append(*patch, patchOperation{
		Op:    "replace",
		Path:  "/spec/nodeSelector",
		Value: nodeSelector,
	})
}
