package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"


	"github.com/golang/glog"
	"github.com/softonic/node-policy-webhook/pkg/admission"
	h "github.com/softonic/node-policy-webhook/pkg/http"
	"github.com/softonic/node-policy-webhook/pkg/version"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

	goflag "flag"
	

	"github.com/spf13/pflag"
)

type params struct {
	version     bool
	certificate string
	privateKey  string
}

const DEFAULT_BIND_ADDRESS = ":8443"

var handler *h.HttpHandler

func init() {
	handler = getHttpHandler()
}

func main() {
	var params params

	commandName := path.Base(os.Args[0])
	
	pflag.CommandLine.AddGoFlag(goflag.CommandLine.Lookup("v"))


	pflag.CommandLine.Parse([]string{})

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



	rootCmd.Flags().BoolVarP(&params.version, "version", "n", false, "print version and exit")
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
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler.HealthCheckHandler(w, r)
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
	address := os.Getenv("BIND_ADDRESS")
	if address == "" {
		address = DEFAULT_BIND_ADDRESS
	}
	glog.V(2).Infof("Starting server, bound at %v", address)
	glog.Infof("Listening to address %v", address)
	srv := &http.Server{
		Addr:         address,
		Handler:      mux,
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	glog.Fatalf("Could not start server: %v", srv.ListenAndServeTLS(params.certificate, params.privateKey))

}

func getHttpHandler() *h.HttpHandler {
	return h.NewHttpHanlder(getNodePolicyAdmissionReviewer())
}

func getNodePolicyAdmissionReviewer() *admission.AdmissionReviewer {
	client, err := getRestClient()
	if err != nil {
		panic(err.Error())
	}
	return admission.NewNodePolicyAdmissionReviewer(
		getFetcher(client),
		admission.NewPatcher(),
	)
}

func getFetcher(client dynamic.Interface) admission.FetcherInterface {
	return admission.NewNodePolicyProfileFetcher(client)
}
func getRestClient() (dynamic.Interface, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.New("Error configuring client")
	}
	return dynamic.NewForConfig(cfg)
}
