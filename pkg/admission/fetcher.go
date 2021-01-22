package admission

import (
	"context"
	"errors"

	"github.com/softonic/node-policy-webhook/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

type FetcherInterface interface {
	GetNamespace(name string) (*v1.Namespace, error)
	Get(profileName string) (*v1alpha1.NodePolicyProfile, error)
}

type NodePolicyProfileFetcher struct {
	client dynamic.Interface
}

func NewNodePolicyProfileFetcher(client dynamic.Interface) FetcherInterface {
	return &NodePolicyProfileFetcher{
		client: client,
	}
}

func (n *NodePolicyProfileFetcher) GetNamespace(name string) (*v1.Namespace, error) {
	resourceScheme := v1alpha1.SchemeBuilder.GroupVersion.WithResource("namespace")
	resp, err := n.client.Resource(resourceScheme).Get(context.TODO(), name, v12.GetOptions{})
	if err != nil {
		klog.Errorf("Error getting Namespace %s (%v)", name, err)
		return nil, errors.New("Error getting Namespace")
	}
	namespace := &v1.Namespace{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), namespace)
	if err != nil {
		return nil, errors.New("Namespace not found")
	}
	return namespace, nil
}

func (n *NodePolicyProfileFetcher) Get(profileName string) (*v1alpha1.NodePolicyProfile, error) {
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
