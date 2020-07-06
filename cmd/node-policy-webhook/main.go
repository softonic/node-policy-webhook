package main

import (
	"fmt"
	"os"
	"path"
	"github.com/spf13/cobra"
	"github.com/softonic/node-policy-webhook/pkg/version"
	"net/http"
)
type params struct {
	version            bool
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
}

func run(params *params) {
	http.HandleFunc("/", func(http.ResponseWriter, *http.Request){
		fmt.Println("Hello world!")
	})
	http.ListenAndServe(":8080", nil)
}
