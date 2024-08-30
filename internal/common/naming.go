package common

import (
	"fmt"

	"github.com/flavio/kuberlr/internal/osexec"

	"github.com/blang/semver/v4"
)

// KubectlLocalNamingScheme holds the scheme used to name the kubectl binaries
// downloaded by kuberlr.
const KubectlLocalNamingScheme = "kubectl%d.%d.%d"

// KubectlSystemNamingScheme holds the scheme used to name the kubectl binaries
// installed system-wide.
const KubectlSystemNamingScheme = "kubectl%d.%d"

// BuildKubectlNameForLocalBin returns how kuberlr will name the kubectl binary
// with the specified version when downloading that to the user home.
func BuildKubectlNameForLocalBin(v semver.Version) string {
	return fmt.Sprintf(KubectlLocalNamingScheme+osexec.Ext, v.Major, v.Minor, v.Patch)
}

// BuildKubectlNameForSystemBin returns how kuberlr expects system-wide
// kubectl binaries to be named.
func BuildKubectlNameForSystemBin(version semver.Version) string {
	return fmt.Sprintf(KubectlSystemNamingScheme+osexec.Ext, version.Major, version.Minor)
}
