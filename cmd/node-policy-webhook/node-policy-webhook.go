package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	_ "github.com/golang/glog"
	"github.com/softonic/node-policy-webhook/pkg/admission"
	h "github.com/softonic/node-policy-webhook/pkg/http"
	"github.com/softonic/node-policy-webhook/pkg/version"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"log"
	"net/http"
	"os"
	"path"
)

type params struct {
	version     bool
	certificate string
	privateKey  string
}

var handler *h.HttpHandler

func init() {
	handler = getHttpHandler()
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
	rootCmd.Flags().StringVarP(&params.certificate, "tls-cert", "c", "default", "certificate (required)")
	rootCmd.Flags().StringVarP(&params.privateKey, "tls-key", "p", "default", "privateKey (required)")

	rootCmd.MarkFlagRequired("tls-cert")
	rootCmd.MarkFlagRequired("tls-key")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(params *params) {
	mux := http.NewServeMux()

	_, err := tls.LoadX509KeyPair(params.certificate, params.privateKey)
	if err != nil {
		klog.Errorf("Failed to load key pair: %v", err)
	}

	mux.HandleFunc("/mutate", func(w http.ResponseWriter, r *http.Request) {
		handler.MutationHandler(w, r)
	})

	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}
	srv := &http.Server{
		Addr:         ":443",
		Handler:      mux,
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	log.Fatal(srv.ListenAndServeTLS(params.certificate, params.privateKey))

}

func getHttpHandler() *h.HttpHandler {
	return h.NewHttpHanlder(getNodePolicyAdmissionReviewer())
}

func getNodePolicyAdmissionReviewer() *admission.AdmissionReviewer {
	client, err := getRestClient()
	if err != nil {
		panic(err.Error())
	}
	return admission.NewNodePolicyAdmissionReviewer(getNodePolicyProfileFetcher(client))
}

func getNodePolicyProfileFetcher(client dynamic.Interface) admission.NodePolicyProfileFetcherInterface {
	return admission.NewNodePolicyProfileFetcher(client)
}
func getRestClient() (dynamic.Interface, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.New("Error configuring client")
	}
	return dynamic.NewForConfig(cfg)
}
