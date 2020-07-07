package main

import (
	"fmt"
	"github.com/softonic/node-policy-webhook/pkg/version"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"path"
)

type params struct {
	version bool
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
	rootCmd.Execute()
}

func run(params *params) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, world")
	})
	err := http.ListenAndServe(":8080", nil)

	fmt.Println(err)
}
