package controller

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/nxmatic/admission-webhook-controller/pkg/plugins"
)

// ServerOptions from cert-managed with reviewer kind selection
type (
	Options struct {
		MetricsAddr          string
		EnableLeaderElection bool
		PolicyOption
	}

	PolicyOption struct {
		value string
	}
)

func (o *Options) FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("admisssion-webhook-controller", pflag.ExitOnError)
	o.AddFlags(fs)
	return fs
}

// AddFlags inject options command flags
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	pflag.StringVar(&o.MetricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	pflag.BoolVar(&o.EnableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	o.PolicyOption.AddFlags(fs)
}

func (o *PolicyOption) AddFlags(fs *pflag.FlagSet) {
	fs.VarP(o, "kind", "k", "kind, can be 'node' or 'gcpauth'")
}

func (o *PolicyOption) String() string {
	return o.value
}

func (o *PolicyOption) Set(value string) error {
	if !plugins.SupportPolicy(value) {
		return fmt.Errorf("Unsupported kind %s, available options are %s", value, plugins.Policies())
	}
	o.value = value

	return nil
}

func (o *PolicyOption) Type() string {
	return "kind"
}
