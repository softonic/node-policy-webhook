// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:rbac:groups=nodepolicy.nuxeo.io,resources=nodepolicyprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nodepolicy.nuxeo.io,resources=nodepolicyprofiles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get
// +kubebuilder:webhook:name=nodepolicy,versions={v1,v1beta1},groups=nodepolicy.nuxeo.io,resources=pods,verbs="CREATE",path=/mutate-v1-pods,mutating=true,failurePolicy=Ignore
package v1alpha1
