package version

import (
	"fmt"
	"runtime"
)

var (
	version   string
	buildDate string
	otelCol   string
)

// Version holds this Controller's version as well as the version of some of the components it uses.
type Version struct {
	Controller string `json:"k8s-policy-controller"`
	BuildDate  string `json:"build-date"`
	Go         string `json:"go-version"`
}

// Get returns the Version object with the relevant information.
func Get() Version {
	return Version{
		Controller: version,
		BuildDate:  buildDate,
		Go:         runtime.Version(),
	}
}

func (v Version) String() string {
	return fmt.Sprintf(
		"Version(Controller='%v', BuildDate='%v', Go='%v')",
		v.Controller,
		v.BuildDate,
		v.Go,
	)
}
