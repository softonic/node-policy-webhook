package k8s

import (
	"context"

	core_api "k8s.io/api/core/v1"
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

const ReplicateKey string = "replicator.v1.mittwald.de/replicate-from"

type (
	ReplicatorSource interface {
		Name() string
		Namespace() string
	}

	Replicator struct {
		dynamic.Interface
	}
)

var (
	SecretsResource = core_api.SchemeGroupVersion.WithResource("secrets")
)

func (f *Replicator) CreateReplicatedSecret(secret *core_api.Secret, source ReplicatorSource) (*core_api.Secret, error) {
	secret.TypeMeta = meta_api.TypeMeta{
		Kind:       "Secret",
		APIVersion: SecretsResource.Version,
	}
	if secret.ObjectMeta.Annotations == nil {
		secret.ObjectMeta.Annotations = map[string]string{}
	}
	secret.ObjectMeta.Annotations[ReplicateKey] = path(source)

	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(secret)
	if err != nil {
		return nil, err
	}

	resp, err :=
		f.Resource(SecretsResource).
			Namespace(secret.ObjectMeta.Namespace).
			Create(context.TODO(), &unstructured.Unstructured{Object: data}, meta_api.CreateOptions{})
	if err != nil {
		return nil, err
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func (f *Replicator) UpdateReplicatedSecret(secret *core_api.Secret, source ReplicatorSource) (*core_api.Secret, error) {
	secret.ObjectMeta.Annotations[ReplicateKey] = path(source)

	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(secret)
	if err != nil {
		return nil, err
	}

	resourceScheme := core_api.SchemeGroupVersion.WithResource(core_api.ResourceSecrets.String())
	resp, err :=
		f.Resource(resourceScheme).
			Namespace(secret.ObjectMeta.Namespace).
			Update(context.TODO(), &unstructured.Unstructured{Object: data}, meta_api.UpdateOptions{})

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func path(s ReplicatorSource) string {
	return s.Namespace() + "/" + s.Name()
}
