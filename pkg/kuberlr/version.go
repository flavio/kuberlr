package kuberlr

import (
	"fmt"
	"runtime"
)

var (
	Version    string
	BuildDate  string
	Tag        string
	ClosestTag string
)

type KuberlrVersion struct {
	Version   string
	BuildDate string
	Tag       string
	GoVersion string
}

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

func (s KuberlrVersion) String() string {
	if s.Tag == "" {
		return fmt.Sprintf("kuberlr version: %s %s %s", s.Version, s.BuildDate, s.GoVersion)
	}
	return fmt.Sprintf("kuberlr version: %s (tagged as %q) %s %s", s.Version, s.Tag, s.BuildDate, s.GoVersion)
}
