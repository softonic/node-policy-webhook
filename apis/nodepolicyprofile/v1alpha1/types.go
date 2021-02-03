// +k8s:deepcopy-gen=package,register
// +kubebuilder:object:generate=true
// +groupName=nodepolicy.nuxeo.io
// +versionName=v1alpha1
package v1alpha1

import (
	core_api "k8s.io/api/core/v1"
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "nodepolicy.nuxeo.io", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// ResourceName typed gcpauth resource identifiers
type ResourceName string

func (name ResourceName) String() string {
	return string(name)
}

var (
	NodepolicyprofilesResource = SchemeBuilder.GroupVersion.WithResource("profiles")
	PodsResource               = core_api.SchemeGroupVersion.WithResource("pods")
)

// AnnotationName typed gcpauth annotation identifiers
type AnnotationName string

func (name AnnotationName) String() string {
	return string(name)
}

const (
	AnnotationPolicyProfile AnnotationName = "nodepolicy.nuxeo.io/profile"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProfileSpec defines the desired state of NodePolicyProfile
type ProfileSpec struct {
	Tolerations  []core_api.Toleration `json:"tolerations,omitempty"`
	NodeAffinity core_api.NodeAffinity `json:"nodeAffinity,omitempty"`
	NodeSelector map[string]string     `json:"nodeSelector,omitempty"`
}

// ProfileStatus defines the observed state of NodePolicyProfile
type ProfileStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// Profile is the Schema for the nodepolicyprofiles API
type Profile struct {
	meta_api.TypeMeta   `json:",inline"`
	meta_api.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProfileSpec   `json:"spec,omitempty"`
	Status ProfileStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProfileList contains a list of NodePolicyProfile
type ProfileList struct {
	meta_api.TypeMeta `json:",inline"`
	meta_api.ListMeta `json:"metadata,omitempty"`
	Items             []Profile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Profile{}, &ProfileList{})
}
