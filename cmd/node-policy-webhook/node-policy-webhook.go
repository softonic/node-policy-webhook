package main

import (
	"crypto/tls"
	"fmt"
	_ "github.com/golang/glog"
	h "github.com/softonic/node-policy-webhook/pkg/http"
	"github.com/softonic/node-policy-webhook/pkg/version"
	"github.com/spf13/cobra"
	"k8s.io/klog"
	"log"
	"net/http"
	"os"
	"path"
)

type params struct {
	version bool
	certificate string
	privateKey string
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
	_, err := tls.LoadX509KeyPair(params.certificate, params.privateKey)
	if err != nil {
		klog.Errorf("Failed to load key pair: %v", err)
	}

	http.HandleFunc("/mutate", func(w http.ResponseWriter, r *http.Request) {
		h.MutationHandler(w, r)
	})
	if err := http.ListenAndServeTLS(":443", params.certificate, params.privateKey, nil); err != nil {
		log.Println(err)
		klog.Errorf("Failed to listen and serve webhook server: %v", err)

	}

}

