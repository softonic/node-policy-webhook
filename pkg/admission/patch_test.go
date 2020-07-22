package admission

import (
	"github.com/softonic/node-policy-webhook/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

var p = &Patcher{}

func expectPatch(t *testing.T, expectedPatch []PatchOperation, patch *[]PatchOperation) {
	if reflect.DeepEqual(expectedPatch, patch) {
		t.Errorf("Patch should match expected patch %v, got %v", expectedPatch, patch)
	}
}

func TestCreatePatchWhenNodeSelectorNotSpecified(t *testing.T) {
	pod := getPodWithNodeSelector(map[string]string{
		"type": "foobar",
	})
	tolerations := &[]v1.Toleration{
		{
			Key:      "foo",
			Operator: "equals",
			Value:    "bar",
			Effect:   "NoSchedule",
		},
	}
	nodePolicyProfile := getNodePolicyProfileWithTolerations(tolerations)

	patch := p.CreatePatch(pod, nodePolicyProfile)

	expectedPatch := []PatchOperation{
		{
			Op:    "replace",
			Path:  "/spec/toleration",
			Value: tolerations,
		},
	}
	expectPatch(t, expectedPatch, patch)
}

func TestCreatePatchWhenNodeSelectorEmpty(t *testing.T) {
	nodeSelector := map[string]string{
		"type": "foobar",
	}
	pod := getPodWithNodeSelector(nodeSelector)
	nodePolicyProfile := getNodePolicyProfileWithNodeSelector(map[string]string{})

	patch := p.CreatePatch(pod, nodePolicyProfile)

	expectedPatch := []PatchOperation{
		{
			Op:    "replace",
			Path:  "/spec/toleration",
			Value: nodeSelector,
		},
	}
	expectPatch(t, expectedPatch, patch)
}

func TestCreatePatchWhenNodeSelectorSameKeyDifferentValue(t *testing.T) {
	podNodeSelector := map[string]string{
		"type": "foobar",
	}
	pod := getPodWithNodeSelector(podNodeSelector)

	profileNodeSelector := map[string]string{
		"type": "barfoo",
	}
	nodePolicyProfile := getNodePolicyProfileWithNodeSelector(profileNodeSelector)

	patch := p.CreatePatch(pod, nodePolicyProfile)
	expectedPatch := []PatchOperation{
		{
			Op:    "replace",
			Path:  "/spec/toleration",
			Value: profileNodeSelector,
		},
	}
	expectPatch(t, expectedPatch, patch)
}

func TestCreatePatchWhenNodeSelectorDifferentKey(t *testing.T) {
	podNodeSelector := map[string]string{
		"type": "foobar",
	}
	pod := getPodWithNodeSelector(podNodeSelector)

	profileNodeSelector := map[string]string{
		"anotherkey": "barfoo",
	}
	nodePolicyProfile := getNodePolicyProfileWithNodeSelector(profileNodeSelector)

	patch := p.CreatePatch(pod, nodePolicyProfile)
	expectedPatch := []PatchOperation{
		{
			Op:    "replace",
			Path:  "/spec/toleration",
			Value: profileNodeSelector,
		},
	}
	expectPatch(t, expectedPatch, patch)
}

func TestCreatePatchWhenNodeAffinity(t *testing.T) {


	nodeSelectorPod := v1.NodeSelector{
		NodeSelectorTerms: []v1.NodeSelectorTerm{
			{
				MatchExpressions: []v1.NodeSelectorRequirement{
					{
						Key:      "bar",
						Operator: "equals",
						Values: []string{
							"foo",
						},
					},
				},
			},
		}}
	nodeAffinityPod := v1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &nodeSelectorPod,
	}

	Affinity := v1.Affinity{
		NodeAffinity: &nodeAffinityPod,
	}

	pod := getPodWithAffinity(Affinity)


	nodeSelector := v1.NodeSelector{
		NodeSelectorTerms: []v1.NodeSelectorTerm{
			{
				MatchExpressions: []v1.NodeSelectorRequirement{
					{
						Key:      "foo",
						Operator: "equals",
						Values: []string{
							"bar",
						},
					},
				},
			},
		}}
	nodeAffinity := v1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &nodeSelector,
	}
	nodePolicyProfile := getNodePolicyProfileWithNodeAffinity(nodeAffinity)

	patch := p.CreatePatch(pod, nodePolicyProfile)

	expectedPatch := []PatchOperation{
		{
			Op:    "replace",
			Path:  "/spec/affinity",
			Value: nodeAffinity,
		},
	}
	expectPatch(t, expectedPatch, patch)
}

func TestCreatePatchPodWithoutNodeAffinity(t *testing.T) {
	pod := getPodWithAffinity(v1.Affinity{})

	nodeSelector := v1.NodeSelector{
		NodeSelectorTerms: []v1.NodeSelectorTerm{
			{
				MatchExpressions: []v1.NodeSelectorRequirement{
					{
						Key:      "foo",
						Operator: "equals",
						Values: []string{
							"bar",
						},
					},
				},
			},
		}}
	nodeAffinity := v1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &nodeSelector,
	}
	nodePolicyProfile := getNodePolicyProfileWithNodeAffinity(nodeAffinity)

	patch := p.CreatePatch(pod, nodePolicyProfile)

	expectedPatch := []PatchOperation{
		{
			Op:    "replace",
			Path:  "/spec/affinity",
			Value: nodeAffinity,
		},
	}
	expectPatch(t, expectedPatch, patch)
}

func TestCreatePatchWithPodAntiAffinityAndProfileNoAffinity(t *testing.T) {

	WeightedPodAffinityTerm := []v1.WeightedPodAffinityTerm{
		v1.WeightedPodAffinityTerm{
			Weight: 100,
			PodAffinityTerm: v1.PodAffinityTerm{
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "foo",
					"component": "bar",
					"release": "test",
					},
				},
				TopologyKey: "kubernetes.io/hostname",
			},
		},
	}

	podAntiAffinity := v1.PodAntiAffinity{
		PreferredDuringSchedulingIgnoredDuringExecution: WeightedPodAffinityTerm,
	}


	affinity := v1.Affinity{}
	

	expectedAffinity := v1.Affinity{
		PodAntiAffinity: &podAntiAffinity,
	}

	pod := getPodWithAffinity(affinity)

	nodeAffinity := v1.NodeAffinity{}

	nodePolicyProfile := getNodePolicyProfileWithNodeAffinity(nodeAffinity)

	patch := p.CreatePatch(pod, nodePolicyProfile)

	expectedPatch := []PatchOperation{
		{
			Op:    "replace",
			Path:  "/spec/affinity",
			Value: expectedAffinity,
		},
	}
	expectPatch(t, expectedPatch, patch)
}

func TestCreatePatchWithPodAntiAffinity(t *testing.T) {
	pod := getPodWithAffinity(v1.Affinity{})

	nodeSelector := v1.NodeSelector{
		NodeSelectorTerms: []v1.NodeSelectorTerm{
			{
				MatchExpressions: []v1.NodeSelectorRequirement{
					{
						Key:      "foo",
						Operator: "equals",
						Values: []string{
							"bar",
						},
					},
				},
			},
		}}
	nodeAffinity := v1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &nodeSelector,
	}
	nodePolicyProfile := getNodePolicyProfileWithNodeAffinity(nodeAffinity)

	patch := p.CreatePatch(pod, nodePolicyProfile)

	expectedPatch := []PatchOperation{
		{
			Op:    "replace",
			Path:  "/spec/affinity",
			Value: nodeAffinity,
		},
	}
	expectPatch(t, expectedPatch, patch)
}


func getNodePolicyProfileWithNodeSelector(nodeSelector map[string]string) *v1alpha1.NodePolicyProfile {
	nodePolicyProfile := &v1alpha1.NodePolicyProfile{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: v1alpha1.NodePolicyProfileSpec{
			NodeSelector: nodeSelector,
		},
		Status: v1alpha1.NodePolicyProfileStatus{},
	}
	return nodePolicyProfile
}

func getNodePolicyProfileWithTolerations(tolerations *[]v1.Toleration) *v1alpha1.NodePolicyProfile {

	nodePolicyProfile := &v1alpha1.NodePolicyProfile{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: v1alpha1.NodePolicyProfileSpec{
			Tolerations: *tolerations,
		},
		Status: v1alpha1.NodePolicyProfileStatus{},
	}
	return nodePolicyProfile
}

func getNodePolicyProfileWithNodeAffinity(nodeAffinity v1.NodeAffinity) *v1alpha1.NodePolicyProfile {
	return &v1alpha1.NodePolicyProfile{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: v1alpha1.NodePolicyProfileSpec{
			NodeAffinity: nodeAffinity,
		},
		Status: v1alpha1.NodePolicyProfileStatus{},
	}
}

func getPodWithNodeSelector(nodeSelector map[string]string) *v1.Pod {
	pod := &v1.Pod{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: v1.PodSpec{
			NodeSelector: nodeSelector,
		},
		Status: v1.PodStatus{},
	}
	return pod
}

func getPodWithAffinity(affinity v1.Affinity) *v1.Pod {
	return &v1.Pod{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: v1.PodSpec{
			Affinity: &affinity,
		},
		Status: v1.PodStatus{},
	}
}
