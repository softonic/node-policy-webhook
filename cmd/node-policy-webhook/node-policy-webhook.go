package main

import (
	"crypto/tls"
	"fmt"
	_ "github.com/golang/glog"
	"github.com/softonic/node-policy-webhook/pkg/node_policy"
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
}

func main() {
	var params params
	var certificate string
	var privateKey string

	commandName := path.Base(os.Args[0])

	rootCmd := &cobra.Command{
		Use:   commandName,
		Short: fmt.Sprintf("%v handles node policy profiles in kubernetes", commandName),
		Run: func(cmd *cobra.Command, args []string) {
			if params.version {
				fmt.Println("Version:", version.Version)
			} else {
				run(&params, certificate, privateKey)
			}
		},
	}
	rootCmd.Flags().BoolVarP(&params.version, "version", "v", false, "print version and exit")
	rootCmd.Flags().StringVarP(&certificate, "tls-cert", "c", "default", "certificate (required)")
	rootCmd.Flags().StringVarP(&privateKey, "tls-key", "p", "default", "privateKey (required)")

	rootCmd.MarkFlagRequired("tls-cert")
	rootCmd.MarkFlagRequired("tls-key")


	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(params *params, certificate string, privateKey string) {

	//_, err := tls.LoadX509KeyPair("/etc/webhook/certs/cert.pem", "/etc/webhook/certs/key.pem")
	_, err := tls.LoadX509KeyPair(certificate, privateKey)
	if err != nil {
		klog.Errorf("Failed to load key pair: %v", err)
	}

	http.HandleFunc("/mutate", func(w http.ResponseWriter, r *http.Request) {
		node_policy.HttpHandler(w, r)
	})
	if err := http.ListenAndServeTLS(":443", certificate, privateKey, nil); err != nil {
		log.Println(err)
		klog.Errorf("Failed to listen and serve webhook server: %v", err)

	}

}

