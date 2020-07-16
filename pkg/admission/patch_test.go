package admission

import (
	"bytes"
	"github.com/softonic/node-policy-webhook/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestCreatePatchWhenNodeSelectorNotSpecified(t *testing.T) {
	pod := getPodWithNodeSelector(map[string]string{
		"type": "foobar",
	})
	nodePolicyProfile := getNodePolicyProfileWithTolerations()

	patch, err := createPatch(pod, nodePolicyProfile)

	expectedPatch := []byte("[{\"op\":\"replace\",\"path\":\"/spec/tolerations\",\"value\":[{\"key\":\"foo\",\"operator\":\"equals\",\"value\":\"bar\",\"effect\":\"NoSchedule\"}]}]")

	if err != nil {
		t.Errorf("Expecting patch %s, got error %v", string(expectedPatch), err)
	}

	if bytes.Compare(patch, expectedPatch) != 0 {
		t.Errorf("Patch should match expected patch %s, got %s", string(expectedPatch), string(patch))
	}
}

func TestCreatePatchWhenNodeSelectorEmpty(t *testing.T) {
	pod := getPodWithNodeSelector(map[string]string{
		"type": "foobar",
	})
	nodePolicyProfile := getNodePolicyProfileWithNodeSelector(map[string]string{})

	patch, err := createPatch(pod, nodePolicyProfile)

	expectedPatch := []byte("[{\"op\":\"replace\",\"path\":\"/spec/nodeSelector\",\"value\":{}}]")

	if err != nil {
		t.Errorf("Expecting patch %s, got error %v", string(expectedPatch), err)
	}

	if bytes.Compare(patch, expectedPatch) != 0 {
		t.Errorf("Patch should match expected patch %s, got %s", string(expectedPatch), string(patch))
	}
}

func TestCreatePatchWhenNodeSelectorSameKeyDifferentValue(t *testing.T) {
	pod := getPodWithNodeSelector(map[string]string{
		"type": "foobar",
	})
	nodePolicyProfile := getNodePolicyProfileWithNodeSelector(map[string]string{
		"type": "barfoo",
	})

	patch, err := createPatch(pod, nodePolicyProfile)

	expectedPatch := []byte("[{\"op\":\"replace\",\"path\":\"/spec/nodeSelector\",\"value\":{\"type\":\"barfoo\"}}]")

	if err != nil {
		t.Errorf("Expecting patch %s, got error %v", string(expectedPatch), err)
	}

	if bytes.Compare(patch, expectedPatch) != 0 {
		t.Errorf("Patch should match expected patch %s, got %s", string(expectedPatch), string(patch))
	}
}

func TestCreatePatchWhenNodeSelectorDifferentKey(t *testing.T) {
	pod := getPodWithNodeSelector(map[string]string{
		"type": "foobar",
	})
	nodePolicyProfile := getNodePolicyProfileWithNodeSelector(map[string]string{
		"anotherkey": "barfoo",
	})

	patch, err := createPatch(pod, nodePolicyProfile)

	expectedPatch := []byte("[{\"op\":\"replace\",\"path\":\"/spec/nodeSelector\",\"value\":{\"anotherkey\":\"barfoo\"}}]")

	if err != nil {
		t.Errorf("Expecting patch %s, got error %v", string(expectedPatch), err)
	}

	if bytes.Compare(patch, expectedPatch) != 0 {
		t.Errorf("Patch should match expected patch %s, got %s", string(expectedPatch), string(patch))
	}
}

func TestCreatePatchNodeAffinity(t *testing.T) {
	pod := getPodWithAffinity(v1.Affinity{})

	nodeSelector := v1.NodeSelector{
		NodeSelectorTerms: []v1.NodeSelectorTerm{
		{
			MatchExpressions: []v1.NodeSelectorRequirement{
				{
					Key:      "foo",
					Operator: "equals",
					Values:   []string{
						"bar",
					},
				},
			},
		},
	}}
	nodePolicyProfile := getNodePolicyProfileWithNodeAffinity(v1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &nodeSelector,
	})

	patch, err := createPatch(pod, nodePolicyProfile)

	expectedPatch := []byte("[{\"op\":\"replace\",\"path\":\"/spec/affinity\",\"value\":{\"nodeAffinity\":{\"requiredDuringSchedulingIgnoredDuringExecution\":{\"nodeSelectorTerms\":[{\"matchExpressions\":[{\"key\":\"foo\",\"operator\":\"equals\",\"values\":[\"bar\"]}]}]}}}}]")

	if err != nil {
		t.Errorf("Expecting patch %s, got error %v", string(expectedPatch), err)
	}

	if bytes.Compare(patch, expectedPatch) != 0 {
		t.Errorf("Patch should match expected patch %s, got %s", string(expectedPatch), string(patch))
	}
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

func getNodePolicyProfileWithTolerations() *v1alpha1.NodePolicyProfile {
	tolerations := []v1.Toleration{
		{
			Key:      "foo",
			Operator: "equals",
			Value:    "bar",
			Effect:   "NoSchedule",
		},
	}
	nodePolicyProfile := &v1alpha1.NodePolicyProfile{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: v1alpha1.NodePolicyProfileSpec{
			Tolerations: tolerations,
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