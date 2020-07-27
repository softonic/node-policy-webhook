package admission

import (
	"encoding/json"
	"errors"
	"github.com/softonic/node-policy-webhook/api/v1alpha1"
	"gotest.tools/assert"
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

type MockPatcher struct {
	mockPatch *[]PatchOperation
	t         *testing.T
}

func (p MockPatcher) CreatePatch(pod *v1.Pod, nodePolicyProfile *v1alpha1.NodePolicyProfile) *[]PatchOperation {
	return p.mockPatch
}

func getMockPatcher(patch *[]PatchOperation, t *testing.T) PatcherInterface {
	return &MockPatcher{
		patch,
		t,
	}
}

type MockFetcher struct {
	mockNodePolicyProfile *v1alpha1.NodePolicyProfile
	err                   error
	t                     *testing.T
	expectedProfile       string
}

func (f MockFetcher) Get(name string) (*v1alpha1.NodePolicyProfile, error) {
	if f.expectedProfile != "" {
		assert.Equal(f.t, f.expectedProfile, name)
	}
	return f.mockNodePolicyProfile, f.err
}

func (f MockFetcher) ExpectProfile(expectedProfile string) {
	f.expectedProfile = expectedProfile
}

func getMockFetcher(nodePolicyProfile *v1alpha1.NodePolicyProfile, err error, t *testing.T) FetcherInterface {
	return MockFetcher{mockNodePolicyProfile: nodePolicyProfile, err: err, t: t}
}

func TestFailResponseIfAdmissionReviewRequestEmpty(t *testing.T) {
	admissionReview := v1beta1.AdmissionReview{}

	reviewer := NewNodePolicyAdmissionReviewer(getMockFetcher(
		&v1alpha1.NodePolicyProfile{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       v1alpha1.NodePolicyProfileSpec{},
			Status:     v1alpha1.NodePolicyProfileStatus{},
		}, nil, t),
		getMockPatcher(&[]PatchOperation{
			{
				Op:    "mock",
				Path:  "/mock/me",
				Value: nil,
			},
		}, t),
	)

	reviewer.PerformAdmissionReview(&admissionReview)

	if admissionReview.Response.Result.Status != "Fail" {
		t.Errorf("Status should be Fail, but got %v", admissionReview.Response.Result.Status)
	}
}

func TestAllowResponseIfAdmissionReviewRequestPodWithNoAnnotation(t *testing.T) {
	pod := v1.Pod{}
	rawPod, _ := json.Marshal(pod)
	admissionReview := v1beta1.AdmissionReview{
		Request: &v1beta1.AdmissionRequest{
			Object: runtime.RawExtension{
				Raw: rawPod,
			},
		},
	}

	reviewer := NewNodePolicyAdmissionReviewer(getMockFetcher(
		&v1alpha1.NodePolicyProfile{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       v1alpha1.NodePolicyProfileSpec{},
			Status:     v1alpha1.NodePolicyProfileStatus{},
		}, nil, t),
		getMockPatcher(&[]PatchOperation{
			{
				Op:    "mock",
				Path:  "/mock/me",
				Value: nil,
			},
		}, t),
	)

	reviewer.PerformAdmissionReview(&admissionReview)
	if !admissionReview.Response.Allowed {
		t.Errorf("Admission review should return true, but got %v", admissionReview.Response.Allowed)
	}
}

func TestModifyResponseIfAdmissionReviewRequestPodWithAnnotation(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				PROFILE_ANNOTATION: "testProfile",
			},
		},
	}
	mockFetcher := getMockFetcher(
		&v1alpha1.NodePolicyProfile{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       v1alpha1.NodePolicyProfileSpec{},
			Status:     v1alpha1.NodePolicyProfileStatus{},
		}, nil, t)

	m, _ := mockFetcher.(MockFetcher)
	m.ExpectProfile("testProfile")

	rawPod, _ := json.Marshal(pod)
	admissionReview := v1beta1.AdmissionReview{
		Request: &v1beta1.AdmissionRequest{
			Object: runtime.RawExtension{
				Raw: rawPod,
			},
		},
	}
	patch := &[]PatchOperation{
		{
			Op:    "mock",
			Path:  "/mock/me",
			Value: nil,
		},
	}
	jsonPatch, _ := json.Marshal(patch)
	reviewer := NewNodePolicyAdmissionReviewer(
		mockFetcher,
		getMockPatcher(patch, t),
	)

	reviewer.PerformAdmissionReview(&admissionReview)
	if !admissionReview.Response.Allowed {
		t.Errorf("Admission review should return true, but got %v", admissionReview.Response.Allowed)
	}
	if admissionReview.Response.Result.Status != "Success" {
		t.Errorf("Status should be Success, got %v", admissionReview.Response.Result.Status)
	}
	assert.DeepEqual(t, admissionReview.Response.Patch, jsonPatch)
}

func TestFailResponseIfNodePolicyProfileNotFound(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				PROFILE_ANNOTATION: "testProfile",
			},
		},
	}
	mockFetcher := getMockFetcher(
		nil,
		errors.New("Node Policy Profile not found"),
		t,
	)

	m, _ := mockFetcher.(MockFetcher)
	m.ExpectProfile("testProfile")

	rawPod, _ := json.Marshal(pod)
	admissionReview := v1beta1.AdmissionReview{
		Request: &v1beta1.AdmissionRequest{
			Object: runtime.RawExtension{
				Raw: rawPod,
			},
		},
	}
	patch := &[]PatchOperation{
		{
			Op:    "mock",
			Path:  "/mock/me",
			Value: nil,
		},
	}
	reviewer := NewNodePolicyAdmissionReviewer(
		mockFetcher,
		getMockPatcher(patch, t),
	)

	reviewer.PerformAdmissionReview(&admissionReview)
	if admissionReview.Response.Allowed {
		t.Errorf("Admission review should return false, but got %v", admissionReview.Response.Allowed)
	}
	if admissionReview.Response.Result.Status != "Fail" {
		t.Errorf("Status should be Fail, got %v", admissionReview.Response.Result.Status)
	}
}
