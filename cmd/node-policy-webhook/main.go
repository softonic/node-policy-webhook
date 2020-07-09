package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/softonic/node-policy-webhook/api/v1alpha1"
	"github.com/softonic/node-policy-webhook/pkg/version"
	"github.com/spf13/cobra"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"log"
	"net/http"
	"os"
	"path"
)

type params struct {
	version bool
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func main() {
	var params params
	commandName := path.Base(os.Args[0])

	rootCmd := &cobra.Command{
		Use:   commandName,
		Short: fmt.Sprintf("%v handles node policy profiles in kubernetes", commandName),
		Run: func(cmd *cobra.Command, args []string) {
			if params.version {
				fmt.Println("Version:", version.Version)
			} else {
				run(&params)
			}
		},
	}
	rootCmd.Flags().BoolVarP(&params.version, "version", "v", false, "print version and exit")
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(params *params) {

	_, err := tls.LoadX509KeyPair("/etc/webhook/certs/cert.pem", "/etc/webhook/certs/key.pem")
	if err != nil {
		glog.Errorf("Failed to load key pair: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	})
	if err := http.ListenAndServeTLS(":443", "/etc/webhook/certs/cert.pem", "/etc/webhook/certs/key.pem", nil); err != nil {
		log.Println(err)
		glog.Errorf("Failed to listen and serve webhook server: %v", err)

	}
	glog.Info("Server started")
}

func createPatch(pod *corev1.Pod, profileName string) ([]byte, error) {
	var patch []patchOperation

	cfg, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	client, err := dynamic.NewForConfig(cfg)
	if err != nil {
		panic(err.Error())
	}

	resourceScheme := v1alpha1.SchemeBuilder.GroupVersion.WithResource("nodepolicyprofiles")

	resp, err := client.Resource(resourceScheme).Get(context.TODO(), profileName, metav1.GetOptions{})
	if err != nil {
		klog.Fatalf("Error getting NodePolicyProfile %s %v (Resource Scheme %v)", profileName, err, resourceScheme)
	}
	klog.Infof("Got NodePolicyProfile %s: %v", profileName, resp)

	nodePolicyProfile := &v1alpha1.NodePolicyProfile{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), nodePolicyProfile)
	if err != nil {
		panic(err.Error())
	}

	nodeSelector := make(map[string]string)

	for key, value := range nodePolicyProfile.Spec.NodeSelector {
		nodeSelector[key] = value
	}

	patch = append(patch, patchOperation{
		Op:    "add",
		Path:  "/spec/nodeSelector",
		Value: nodeSelector,
	})

//	tolerations := []corev1.Toleration{}

//	append(tolerations, pod.Spec.Tolerations, nodePolicyProfile.Spec.Tolerations)
	// patch = append(patch, patchOperation{
	// 	Op:    "replace",
	// 	Path:  "/spec/containers/0/image",
	// 	Value: "debian",
	// })

	return json.Marshal(patch)
}

func getProfile(metadata *metav1.ObjectMeta) (string, error) {

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	//nodeselector := annotations["softonic.io/profile"]

	if val, ok := annotations["softonic.io/profile"]; ok {
		return val, nil
	}

	return "", errors.New("Annotation not found")
}

func mutate(ar *v1beta1.AdmissionReview) (*v1beta1.AdmissionResponse, error) {

	req := ar.Request
	var pod corev1.Pod
	err := json.Unmarshal(req.Object.Raw, &pod)
	if err != nil {
		glog.Errorf("Could not unmarshal raw object: %v", err)
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}, err
	}

	profile, err := getProfile(&pod.ObjectMeta)
	if err !=nil {
		glog.Infof("Skipping mutation for %s/%s due to policy check", pod.Namespace, pod.Name)

		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}, nil
	}

	patchBytes, err := createPatch(&pod, profile)
	if err != nil {
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}, err
	}

	pT := v1beta1.PatchTypeJSONPatch

	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Status: "Success",
		},
		Patch: patchBytes,
		// PatchType: func() *v1beta1.PatchType {
		// 	pt := v1beta1.PatchTypeJSONPatch
		// 	return &pt
		// }(),

		PatchType: &pT,
		Allowed:   true,
		UID:       ar.Request.UID,
	}, nil

}

func handler(w http.ResponseWriter, r *http.Request) {

	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}

	if r.URL.Path == "/mutate" {

		err := json.NewDecoder(r.Body).Decode(&ar)
		if err != nil {
			glog.Errorf("Can decode body: %v", err)
			log.Println(err)
			admissionResponse = &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
					Status:  "Fail",
				},
			}

		} else {

			admissionResponse, err = mutate(&ar)
			if err != nil {
				log.Println(err)
				glog.Errorf("Can't write response: %v", err)
			} else {
				glog.Infof("Success mutating")
				log.Println("funciono el mutating")
			}

		}

		admissionReview := v1beta1.AdmissionReview{}
		admissionReview.Response = admissionResponse

		resp, err := json.Marshal(admissionReview)
		if _, err := w.Write(resp); err != nil {
			glog.Errorf("Can't write response: %v", err)
			http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
		}

	}

}
