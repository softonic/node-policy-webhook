package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/softonic/node-policy-webhook/pkg/version"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"path"

	_ "github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		glog.Errorf("Failed to listen and serve webhook server: %v", err)

	}

}

func createPatch(pod *corev1.Pod) ([]byte, error) {
	var patch []patchOperation

	patch = append(patch, patchOperation{
		Op:   "add",
		Path: "/spec/nodeSelector",
		Value: map[string]string{
			"type": "stateless",
		},
	})

	// patch = append(patch, patchOperation{
	// 	Op:    "replace",
	// 	Path:  "/spec/containers/0/image",
	// 	Value: "debian",
	// })

	return json.Marshal(patch)
}

func mutationRequired(metadata *metav1.ObjectMeta) bool {

	//var required bool
	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	//nodeselector := annotations["softonic.io/profile"]

	return true

}

func mutate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {

	req := ar.Request
	var pod corev1.Pod
	err := json.Unmarshal(req.Object.Raw, &pod)
	if err != nil {
		glog.Errorf("Could not unmarshal raw object: %v", err)
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	if !mutationRequired(&pod.ObjectMeta) {
		glog.Infof("Skipping mutation for %s/%s due to policy check", pod.Namespace, pod.Name)
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	patchBytes, err := createPatch(&pod)
	if err != nil {
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
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
	}

}

func handler(w http.ResponseWriter, r *http.Request) {

	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}

	if r.URL.Path == "/mutate" {

		err := json.NewDecoder(r.Body).Decode(&ar)
		if err != nil {
			glog.Errorf("Can decode body: %v", err)
			admissionResponse = &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
					Status:  "Fail",
				},
			}

		} else {

			admissionResponse = mutate(&ar)

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
