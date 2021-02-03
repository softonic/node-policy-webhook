package reviewer

import (
	"reflect"

	nodepolicy_api "github.com/nuxeo/k8s-policy-controller/apis/nodepolicyprofile/v1alpha1"
	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi/reviewer"

	core_api "k8s.io/api/core/v1"
)

type patcher struct {
	*core_api.Pod
	*nodepolicy_api.Profile
	Patch []reviewer.PatchOperation
}

func (p *patcher) Create() ([]reviewer.PatchOperation, error) {
	p.Patch = make([]reviewer.PatchOperation, 0, 5)
	p.Patch = p.addLabelProfilePatch()
	p.Patch = p.addNodeProfilePatch()
	p.Patch = p.addNodeSelectorPatch()
	p.Patch = p.addTolerationsPatch()
	p.Patch = p.addNodeAffinityPatch()

	return p.Patch, nil
}

func (p *patcher) addLabelProfilePatch() []reviewer.PatchOperation {
	return append(p.Patch, reviewer.PatchOperation{
		Op:    "add",
		Path:  "/metadata/labels/nodepolicy.nuxeo.io~1profile",
		Value: p.Profile.Name})
}

func (p *patcher) addNodeProfilePatch() []reviewer.PatchOperation {
	_, ok := p.Pod.Annotations[nodepolicy_api.AnnotationPolicyProfile.String()]
	if ok {
		return p.Patch
	}

	return append(p.Patch, reviewer.PatchOperation{
		Op:    "add",
		Path:  "/metadata/annotations/nodepolicy.nuxeo.io~1profile",
		Value: p.Profile.Name})
}

func (p *patcher) addNodeAffinityPatch() []reviewer.PatchOperation {
	if reflect.DeepEqual(p.Profile.Spec.NodeAffinity, core_api.NodeAffinity{}) {
		return p.Patch
	}

	affinity := core_api.Affinity{}

	affinity.NodeAffinity = &p.Profile.Spec.NodeAffinity

	if p.Pod.Spec.Affinity != nil {
		if p.Pod.Spec.Affinity.PodAntiAffinity != nil {
			affinity.PodAntiAffinity = p.Pod.Spec.Affinity.PodAntiAffinity
		}

		if p.Pod.Spec.Affinity.PodAffinity != nil {
			affinity.PodAffinity = p.Pod.Spec.Affinity.PodAffinity
		}
	}

	return append(p.Patch, reviewer.PatchOperation{
		Op:    "replace",
		Path:  "/spec/affinity",
		Value: affinity,
	})
}

func (p *patcher) addTolerationsPatch() []reviewer.PatchOperation {
	if p.Pod.Spec.Tolerations == nil && p.Profile.Spec.Tolerations == nil {
		return p.Patch
	}
	tolerations := []core_api.Toleration{}

	tolerations = append(tolerations, p.Pod.Spec.Tolerations...)

	tolerationEqual := false

	for _, tolerationPod := range p.Pod.Spec.Tolerations {
		for _, tolerationProfile := range p.Profile.Spec.Tolerations {
			if reflect.DeepEqual(tolerationPod, tolerationProfile) {
				tolerationEqual = true
			}
		}
	}

	if tolerationEqual == false {
		tolerations = append(tolerations, p.Profile.Spec.Tolerations...)
	}

	return append(p.Patch, reviewer.PatchOperation{
		Op:    "replace",
		Path:  "/spec/tolerations",
		Value: tolerations,
	})
}

func (p *patcher) addNodeSelectorPatch() []reviewer.PatchOperation {
	if p.Profile.Spec.NodeSelector == nil {
		return p.Patch
	}

	nodeSelector := make(map[string]string)

	for key, value := range p.Profile.Spec.NodeSelector {
		nodeSelector[key] = value
	}

	return append(p.Patch, reviewer.PatchOperation{
		Op:    "replace",
		Path:  "/spec/nodeSelector",
		Value: nodeSelector,
	})
}
