package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	k8sruntime "k8s.io/apimachinery/pkg/runtime"

	"github.com/nxmatic/admission-webhook-controller/pkg/plugins"

	"github.com/nxmatic/admission-webhook-controller/pkg/controller"
	"github.com/nxmatic/admission-webhook-controller/pkg/version"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	scheme = k8sruntime.NewScheme()
)

func main() {
	opts := zap.Options{}
	opts.BindFlags(flag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	// Add flags related to this operator
	v := version.Get()
	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)

	logger.Info("Starting the admission webhook controller",
		"admission-webhook-controller", v.Controller,
		"build-date", v.BuildDate,
		"go-version", v.Go,
		"go-arch", runtime.GOARCH,
		"go-os", runtime.GOOS,
	)

	ctrlOptions := controller.Options{}

	pflag.CommandLine.AddFlagSet(ctrlOptions.FlagSet())

	pflag.Parse()

	plugin := plugins.Get(ctrlOptions.PolicyOption.String())

	mgrOptions := ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: ctrlOptions.MetricsAddr,
		Port:               9443,
		LeaderElection:     ctrlOptions.EnableLeaderElection,
		LeaderElectionID:   fmt.Sprintf("admission-webhook-controller/%s", ctrlOptions.PolicyOption),
		Logger:             logger,
	}

	konfig := ctrl.GetConfigOrDie()

	mgr, err := controller.NewManagerWithOptions(konfig, mgrOptions, plugin)
	if err != nil {
		logger.Error(err, "problem configuring manager")
		os.Exit(1)
	}

	context := ctrl.SetupSignalHandler()

	logger.Info("starting manager")
	if err := mgr.Start(context); err != nil {
		logger.Error(err, "problem running manager")
		os.Exit(1)
	}

}
