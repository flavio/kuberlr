package kuberlr

import (
	"fmt"
	"runtime"
)

var (
	// Version holds the version of kuberlr, this is set at build time.
	Version string //nolint:gochecknoglobals // this is a variable set at build time
	// BuildDate holds the date during which the kuberlr binary was built, this is set at build time.
	BuildDate string //nolint:gochecknoglobals // this is a variable set at build time
	// Tag holds the git tag defined on kuberlr repo when the binary was built, this is set at build time.
	Tag string //nolint:gochecknoglobals // this is a variable set at build time
	// ClosestTag holds the closest git tag defined on kuberlr repo when the binary was built, this is set at build time.
	ClosestTag string //nolint:gochecknoglobals // this is a variable set at build time
)

// KVersion holds the build-time information of kuberlr.
type KVersion struct {
	Version   string
	BuildDate string
	Tag       string
	GoVersion string
}

// CurrentVersion returns the information about the current version of kuberlr.
func CurrentVersion() KVersion {
	kuberlrVersion := KVersion{
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

// String returns the version information nicely formatted.
func (s KVersion) String() string {
	if s.Tag == "" {
		return fmt.Sprintf("kuberlr version: %s %s %s", s.Version, s.BuildDate, s.GoVersion)
	}
	return fmt.Sprintf("kuberlr version: %s (tagged as %q) %s %s", s.Version, s.Tag, s.BuildDate, s.GoVersion)
}
