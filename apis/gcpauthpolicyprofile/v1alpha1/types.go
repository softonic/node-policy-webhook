// +k8s:deepcopy-gen=package,register
// +kubebuilder:object:generate=true
// +groupName=gcpauthpolicy.nuxeo.io
// +versionName=v1alpha1
package v1alpha1

import (
	"fmt"

	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "gcpauthpolicy.nuxeo.io", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme

	GCPAuthpolicyprofilesResource = SchemeGroupVersion.WithResource("gcpauthpolicyprofiles")
)

type ResourceKind string

func (name ResourceKind) String() string {
	return string(name)
}

const (
	GCPAuthpolicyprofileKind ResourceKind = "GCPAuthPolicyProfile"
)

// Key typed gcpauth annotation identifiers
type Key string

func (name Key) String() string {
	return string(name)
}

const (
	ProfileKey Key = "gcpauthpolicy.nuxeo.io/profile"
	WatchKey   Key = "gcpauthpolicy.nuxeo.io/watch"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// GCPAuthProfile is the schema for the GCPAuthPolicy profile API
type GCPAuthPolicyProfile struct {
	meta_api.TypeMeta   `json:",inline"`
	meta_api.ObjectMeta `json:"metadata,omitempty"`

	Spec   GCPAuthPolicyProfileSpec   `json:"spec,omitempty"`
	Status GCPAuthPolicyProfileStatus `json:"status,omitempty"`
}

type GCPAuthPolicyProfileSpec struct {
	GCPAuthDatasource `json:"datasource,omitempty"`
}

type GCPAuthDatasource struct {
	GCPAuthSecretDatasource `json:"secret,omitempty"`
}

type GCPAuthSecretDatasource struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
}

func (s *GCPAuthPolicyProfileSpec) Path() string {
	return fmt.Sprintf("%s/%s", s.Namespace(), s.Name())
}

func (s *GCPAuthPolicyProfileSpec) Name() string {
	return s.GCPAuthDatasource.GCPAuthSecretDatasource.Name
}

func (s *GCPAuthPolicyProfileSpec) Namespace() string {
	return s.GCPAuthDatasource.GCPAuthSecretDatasource.Namespace
}

// +kubebuilder:object:root=true

// GCPAuthPolicyProfileList contains a list of GCPAuthPolicyProfile
type GCPAuthPolicyProfileList struct {
	meta_api.TypeMeta `json:",inline"`
	meta_api.ListMeta `json:"metadata,omitempty"`
	Items             []GCPAuthPolicyProfile `json:"items"`
}

// GCPAuthPolicyProfileStatus the status
type GCPAuthPolicyProfileStatus struct {
}

func init() {
	SchemeBuilder.Register(&GCPAuthPolicyProfile{}, &GCPAuthPolicyProfileList{})
}
