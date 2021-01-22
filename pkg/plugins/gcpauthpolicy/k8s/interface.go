package k8s

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"

	gcpauth_api "github.com/nxmatic/admission-webhook-controller/apis/gcpauthpolicyprofile/v1alpha1"
	"github.com/pkg/errors"
	core_api "k8s.io/api/core/v1"
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8s_spi "github.com/nxmatic/admission-webhook-controller/pkg/plugins/spi/k8s"
	"k8s.io/apimachinery/pkg/runtime"
)

type (
	Interface struct {
		*k8s_spi.Interface
	}
)

func NewInterface(client dynamic.Interface) *Interface {
	return &Interface{
		k8s_spi.NewInterface(client),
	}
}

func (s *Interface) ResolveProfile(namespace *meta_api.ObjectMeta, resource *meta_api.ObjectMeta) (*gcpauth_api.GCPAuthPolicyProfile, error) {
	annotations := make(map[string]string)
	annotations = s.MergeAnnotations(annotations, namespace)
	annotations = s.MergeAnnotations(annotations, resource)

	if name, ok := annotations[gcpauth_api.ProfileKey.String()]; ok {
		return s.GetProfile(name)
	}
	return nil, errors.New("Annotation not found")

}

var (
	GCPAuthpolicyprofilesResource = gcpauth_api.GCPAuthpolicyprofilesResource
)

func (s *Interface) GetProfile(name string) (*gcpauth_api.GCPAuthPolicyProfile, error) {
	resp, err := s.Interface.Resource(GCPAuthpolicyprofilesResource).Get(context.TODO(), name, meta_api.GetOptions{})
	if err != nil {
		return nil, err
	}
	profile := &gcpauth_api.GCPAuthPolicyProfile{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), profile)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

var (
	SecretsResource = core_api.SchemeGroupVersion.WithResource("secrets")
)

func (i *Interface) DeleteSecrets(profile *gcpauth_api.GCPAuthPolicyProfile) error {
	return i.Interface.Resource(SecretsResource).DeleteCollection(context.TODO(),
		meta_api.DeleteOptions{},
		meta_api.ListOptions{})
}

func (f *Interface) ListSecrets(profile *gcpauth_api.GCPAuthPolicyProfile) (SecretIterator, error) {
	options := meta_api.ListOptions{
		LabelSelector: labels.Set{
			gcpauth_api.ProfileKey.String(): profile.ObjectMeta.Name,
		}.String(),
	}

	resp, err := f.Interface.Resource(SecretsResource).List(context.TODO(), options)
	if err != nil {
		return SecretIterator{}, err
	}
	return SecretIterator{resp}, nil
}

func (f *Interface) EnsureSecretExist(namespace string, profile *gcpauth_api.GCPAuthPolicyProfile) error {
	options := meta_api.ListOptions{
		LabelSelector: labels.Set{
			gcpauth_api.ProfileKey.String(): profile.ObjectMeta.Name,
		}.String(),
	}
	resp, err := f.Interface.Resource(SecretsResource).Namespace(namespace).List(context.TODO(), options)
	if err != nil {
		return err
	}
	if len(resp.Items) > 0 {
		return nil
	}
	secret := core_api.Secret{
		ObjectMeta: meta_api.ObjectMeta{
			Name:      profile.ObjectMeta.Name,
			Namespace: namespace,
			Labels: map[string]string{
				gcpauth_api.ProfileKey.String(): profile.ObjectMeta.Name,
				gcpauth_api.WatchKey.String():   "true",
			},
		},
		Type: core_api.SecretTypeDockerConfigJson,
		Data: map[string][]uint8{
			core_api.DockerConfigJsonKey: []uint8{
				123, 125, 10,
			},
		},
	}
	_, err = f.NewReplicator().CreateReplicatedSecret(&secret, &profile.Spec)
	return err
}

type (
	SecretIterator struct {
		*unstructured.UnstructuredList
	}
	SecretConsumer func(secret *core_api.Secret) error
)

func (i *SecretIterator) Apply(consumer SecretConsumer) error {
	for _, data := range i.UnstructuredList.Items {
		secret := &core_api.Secret{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(data.UnstructuredContent(), secret)
		if err != nil {
			return err
		}
		err = consumer(secret)
		if err != nil {
			return err
		}
	}
	return nil
}
