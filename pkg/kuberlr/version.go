package kuberlr

import (
	"fmt"
	"runtime"
)

var (
	// Version holds the version of kuberlr, this is set at build time
	Version string
	// BuildDate holds the date during which the kuberlr binary was built, this is set at build time
	BuildDate string
	// Tag holds the git tag defined on kuberlr repo when the binary was built, this is set at build time
	Tag string
	// ClosestTag holds the closest git tag defined on kuberlr repo when the binary was built, this is set at build time
	ClosestTag string
)

// KuberlrVersion holds the build-time information of kuberlr
type KuberlrVersion struct {
	Version   string
	BuildDate string
	Tag       string
	GoVersion string
}

// CurrentVersion returns the information about the current version of kuberlr
func CurrentVersion() KuberlrVersion {
	kuberlrVersion := KuberlrVersion{
		Version:   Version,
		BuildDate: BuildDate,
		Tag:       Tag,
		GoVersion: runtime.Version(),
	}
	if kuberlrVersion.Tag == "" {
		kuberlrVersion.Version = fmt.Sprintf("untagged (%s)", ClosestTag)
	}
	return kuberlrVersion
}

// String returns the version infromation nicely formatted
func (s KuberlrVersion) String() string {
	if s.Tag == "" {
		return fmt.Sprintf("kuberlr version: %s %s %s", s.Version, s.BuildDate, s.GoVersion)
	}
	return fmt.Sprintf("kuberlr version: %s (tagged as %q) %s %s", s.Version, s.Tag, s.BuildDate, s.GoVersion)
}
