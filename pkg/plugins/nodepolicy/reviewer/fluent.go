package reviewer

import (
	"encoding/json"

	core_api "k8s.io/api/core/v1"

	nodepolicy_api "github.com/nxmatic/admission-webhook-controller/apis/nodepolicyprofile/v1alpha1"

	"github.com/nxmatic/admission-webhook-controller/pkg/plugins/nodepolicy/k8s"
	reviewer_spi "github.com/nxmatic/admission-webhook-controller/pkg/plugins/spi/reviewer"
)

type (
	RequestedPodStage struct {
		k8s.Interface
		*reviewer_spi.GivenStage
		*core_api.Pod
		Profile *nodepolicy_api.NodePolicyProfile
	}

	RequestedKindStage struct {
		*RequestedPodStage
	}

	RequestedProfileStage struct {
		*RequestedPodStage
		*core_api.Namespace
	}
)

func Given() *RequestedPodStage {
	return &RequestedPodStage{}
}

func (s *RequestedPodStage) RequestedObject(o *reviewer_spi.GivenStage) *RequestedPodStage {
	s.GivenStage = o
	s.Interface = k8s.Interface{Interface: o.Interface}
	return s
}

func (s *RequestedPodStage) The() *RequestedPodStage {
	return s
}

func (s *RequestedPodStage) And() *RequestedPodStage {
	return s
}

func (s *RequestedPodStage) End() *reviewer_spi.WhenStage {
	return &reviewer_spi.WhenStage{
		GivenStage: s.GivenStage,
		Patcher: &patcher{
			s.Pod,
			s.Profile,
			[]reviewer_spi.PatchOperation{},
		},
	}
}

func (r *RequestedPodStage) RequestedKind() *RequestedKindStage {
	return &RequestedKindStage{r}
}

func (s *RequestedKindStage) Or() *RequestedKindStage {
	return s
}

func (s *RequestedKindStage) IsAPod() *RequestedKindStage {

	err := json.Unmarshal(s.AdmissionRequest.Object.Raw, &s.Pod)
	if err != nil {
		s.Allow(nil)
		return s
	}
	s.Logger = s.Logger.WithValues("name", s.Pod.ObjectMeta.Name)
	s.Logger = s.Logger.WithValues("generated-name", s.Pod.ObjectMeta.GenerateName)

	return s
}

func (s *RequestedKindStage) End() *RequestedPodStage {
	return s.RequestedPodStage
}

func (s *RequestedKindStage) And() *RequestedKindStage {
	return s
}

func (s *RequestedPodStage) RequestedProfile() *RequestedProfileStage {
	return &RequestedProfileStage{s, nil}
}

func (s *RequestedProfileStage) Exists() *RequestedProfileStage {
	if !s.CanContinue() {
		return s
	}
	s.Namespace, s.Error = s.Interface.GetNamespace(s.AdmissionRequest.Namespace)
	if s.Error != nil {
		s.Allow(nil)
		return s
	}
	s.Profile, s.Error = s.Interface.ResolveProfile(&s.Namespace.ObjectMeta, &s.Pod.ObjectMeta)
	if s.Error != nil {
		s.Allow(nil)
		return s
	}
	s.Logger = s.Logger.WithValues("profile", s.Profile.ObjectMeta.Name)
	return s
}

func (s *RequestedProfileStage) The() *RequestedProfileStage {
	return s
}

func (s *RequestedProfileStage) And() *RequestedProfileStage {
	return s
}

func (s *RequestedProfileStage) End() *RequestedPodStage {
	return s.RequestedPodStage
}
