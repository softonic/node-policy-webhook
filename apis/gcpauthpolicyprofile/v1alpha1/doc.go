// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:rbac:groups=gcpauthpolicy.nuxeo.io,resources=profiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gcpauthpolicy.nuxeo.io,resources=profiles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=*
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get
// +kubebuilder:webhook:versions={v1,v1beta1},groups=gcpauthpolicy.nuxeo.io,resources=serviceaccounts,verbs="CREATE",name=gcpauthpolicy,path=/mutate-v1-serviceaccounts,mutating=true,failurePolicy=Ignore
package v1alpha1
