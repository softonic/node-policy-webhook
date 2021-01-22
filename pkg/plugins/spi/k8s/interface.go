package k8s

import (
	"context"

	core_api "k8s.io/api/core/v1"
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

// Interface for interacting with kube resources
type (
	Interface struct {
		dynamic.Interface
	}
)

var (
	PodsResource            = core_api.SchemeGroupVersion.WithResource("pods")
	NamespacesResource      = core_api.SchemeGroupVersion.WithResource("namespaces")
	ServiceaccountsResource = core_api.SchemeGroupVersion.WithResource("serviceaccounts")
)

func NewInterface(itf dynamic.Interface) *Interface {
	return &Interface{
		itf,
	}
}

func (f *Interface) NewReplicator() *Replicator {
	return &Replicator{f.Interface}
}

func (f *Interface) GetNamespace(name string) (*core_api.Namespace, error) {
	resp, err := f.Interface.Resource(NamespacesResource).Get(context.TODO(), name, meta_api.GetOptions{})
	if err != nil {
		return nil, err
	}
	namespace := &core_api.Namespace{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), namespace)
	if err != nil {
		return nil, err
	}
	return namespace, nil
}

func (f *Interface) GetServiceAccount(name string, namespace string) (*core_api.ServiceAccount, error) {
	resp, err := f.Interface.Resource(ServiceaccountsResource).Namespace(namespace).Get(context.TODO(), name, meta_api.GetOptions{})
	if err != nil {
		return nil, err
	}
	serviceaccount := &core_api.ServiceAccount{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), serviceaccount)
	if err != nil {
		return nil, err
	}
	return serviceaccount, nil
}

func (f *Interface) MergeAnnotations(accumulator map[string]string, meta *meta_api.ObjectMeta) map[string]string {
	annotations := meta.Annotations
	if annotations == nil {
		return accumulator
	}
	for k, v := range annotations {
		accumulator[k] = v
	}
	return accumulator
}
