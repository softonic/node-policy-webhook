package admission

import (
	"github.com/softonic/node-policy-webhook/api/v1alpha1"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

type MockFetcher struct {
	mockNodePolicyProfile *v1alpha1.NodePolicyProfile
}

func (f MockFetcher) Get(name string) (*v1alpha1.NodePolicyProfile, error) {
	return f.mockNodePolicyProfile, nil
}

func getMockFetcher(nodePolicyProfile *v1alpha1.NodePolicyProfile) NodePolicyProfileFetcherInterface {
	return MockFetcher{mockNodePolicyProfile: nodePolicyProfile}
}

func TestFailResponseIfAdmissionReviewRequestEmpty(t *testing.T) {
	admissionReview := v1beta1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{},
		Request:  nil,
		Response: nil,
	}

	reviewer := NewNodePolicyAdmissionReviewer(getMockFetcher(
		&v1alpha1.NodePolicyProfile{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       v1alpha1.NodePolicyProfileSpec{},
			Status:     v1alpha1.NodePolicyProfileStatus{},
		},
	))

	reviewer.PerformAdmissionReview(&admissionReview)

	if admissionReview.Response.Result.Status != "Fail" {
		t.Errorf("Status should be Fail, but got %v", admissionReview.Response.Result.Status)
	}
}
