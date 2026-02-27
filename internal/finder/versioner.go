package finder

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/flavio/kuberlr/internal/common"
	"github.com/flavio/kuberlr/internal/downloader"
	"github.com/flavio/kuberlr/internal/kubehelper"

	"github.com/blang/semver/v4"
	"k8s.io/klog"
)

type downloadHelper interface {
	GetKubectlBinary(version semver.Version, destination string) error
	UpstreamStableVersion() (semver.Version, error)
}

type kubeAPIHelper interface {
	Version(timeout int64) (semver.Version, error)
}

type iFinder interface {
	AllKubectlBinaries(reverseSort bool) KubectlBinaries
}

// Versioner is used to manage the local kubectl binaries used by kuberlr.
type Versioner struct {
	kFinder                           iFinder
	downloader                        downloadHelper
	apiServer                         kubeAPIHelper
	preventRecursiveInvocationEnvName string
}

// NewVersioner is an helper function that creates a new Versioner instance.
func NewVersioner(f iFinder) *Versioner {
	return &Versioner{
		kFinder:                           f,
		downloader:                        &downloader.Downloder{},
		apiServer:                         &kubehelper.KubeAPI{},
		preventRecursiveInvocationEnvName: PreventRecursiveInvocationEnvName,
	}
}

const PreventRecursiveInvocationEnvName = "KUBERLR_RESOLVING_VERSION"

// KubectlVersionToUse returns the kubectl version to be used to interact with
// the remote server. The method takes into account different failure scenarios
// and acts accordingly.
func (v *Versioner) KubectlVersionToUse(timeout int64) (semver.Version, error) {
	// We use Kubernetes client-go to interact with the remote server to obtain its version.
	// Depending on the cluster configuration, the client-go library might shell out and invoke
	// the kubectl binary to authenticate to the server.
	// Since the kubectl binary is an alias to kuberlr, we can enter a recursion loop
	// where kuberlr calls kubectl that in turn calls kuberlr again.
	//
	// To avoid this, we set an environment variable to signal that we are currently resolving the kubernetes version.

	_, recursiveInvocationDetected := os.LookupEnv(v.preventRecursiveInvocationEnvName)
	if recursiveInvocationDetected {
		klog.V(common.VerbosityTwo).Info("client-go invoked kubectl to authenticate. Preventing kuberlr endless recursion loop.")
		return v.mostRecentKubectlVersionAvailableOrLatestFromUpstream()
	}

	if err := os.Setenv(v.preventRecursiveInvocationEnvName, "1"); err != nil {
		return semver.Version{}, fmt.Errorf("failed to set environment variable %s: %w", v.preventRecursiveInvocationEnvName, err)
	}
	defer os.Unsetenv(v.preventRecursiveInvocationEnvName)

	version, err := v.apiServer.Version(timeout)
	if err != nil {
		if isUnreachable(err) {
			// the remote server is unreachable, let's get
			// the latest version of kubectl that is available on the system
			klog.V(common.VerbosityTwo).Info("Remote kubernetes server unreachable")
		} else {
			klog.V(common.VerbosityOne).Info(err)
		}
		return v.mostRecentKubectlVersionAvailableOrLatestFromUpstream()
	}
	return version, err
}

// mostRecentKubectlVersionAvailableOrLatestFromUpstream returns the most recent version of kubectl
// available on the system. If no kubectl binary is found, it will download the
// latest stable version from the upstream mirror.
func (v *Versioner) mostRecentKubectlVersionAvailableOrLatestFromUpstream() (semver.Version, error) {
	bins := v.kFinder.AllKubectlBinaries(true)
	if kubectl, err := mostRecentKubectlAvailable(bins); err == nil {
		return kubectl.Version, nil
	}

	klog.V(common.VerbosityTwo).Info("No local kubectl binary found, fetching latest stable release version")
	return v.downloader.UpstreamStableVersion()
}

// EnsureCompatibleKubectlAvailable ensures the kubectl binary with the specified
// version is available on the system. It will return the full path to the
// binary.
func (v *Versioner) EnsureCompatibleKubectlAvailable(version semver.Version, allowDownload bool, useLatestIfNoCompatible bool) (string, error) {
	bins := v.kFinder.AllKubectlBinaries(true)
	kubectl, err := findCompatibleKubectl(version, bins)
	if err == nil {
		return kubectl.Path, nil
	}

	if !allowDownload {
		if useLatestIfNoCompatible {
			all := v.kFinder.AllKubectlBinaries(true /* reverseSort */)
			if len(all) > 0 {
				return all[0].Path, nil
			}
		}
		return "", errors.New("the right kubectl is missing, binary downloads from kubernetes' upstream mirror are disabled")
	}

	klog.Infof("Right kubectl missing, downloading version %s", version.String())

	// download the right kubectl to the local cache
	filename := filepath.Join(
		common.LocalDownloadDir(),
		common.BuildKubectlNameForLocalBin(version))

	if err = v.downloader.GetKubectlBinary(version, filename); err != nil {
		if useLatestIfNoCompatible {
			all := v.kFinder.AllKubectlBinaries(true) // newest-first
			if len(all) > 0 {
				klog.Infof("download failed (%v); falling back to newest local kubectl %s at %s",
					err, all[0].Version, all[0].Path)
				return all[0].Path, nil
			}
		}
		return "", fmt.Errorf("failed to download compatible kubectl: %w", err)
	}

	return filename, nil
}

func isUnreachable(err error) bool {
	var e *url.Error
	return os.IsTimeout(err) || errors.As(err, &e)
}

// findCompatibleKubectl returns a kubectl binary compatible with the
// version given via the `requestedVersion` parameter.
// Important: the `bins` parameter must be sorted in descending order.
func findCompatibleKubectl(requestedVersion semver.Version, bins KubectlBinaries) (KubectlBinary, error) {
	if len(bins) == 0 {
		return KubectlBinary{}, &common.NoVersionFoundError{}
	}

	lowerBound := lowerBoundVersion(requestedVersion)
	upperBound := upperBoundVersion(requestedVersion)
	rangeRule := fmt.Sprintf(">=%s <%s", lowerBound.String(), upperBound.String())

	validRange, err := semver.ParseRange(rangeRule)
	if err != nil {
		return KubectlBinary{}, err
	}

	for _, b := range bins {
		if validRange(b.Version) {
			return b, nil
		}
	}

	return KubectlBinary{}, &common.NoVersionFoundError{}
}

// mostRecentKubectlAvailable returns the most recent version of
// kubectl available on the system. It could be something downloaded
// by kuberlr or something already available on the system.
func mostRecentKubectlAvailable(bins KubectlBinaries) (KubectlBinary, error) {
	if len(bins) == 0 {
		return KubectlBinary{}, &common.NoVersionFoundError{}
	}

	return bins[0], nil
}

func lowerBoundVersion(v semver.Version) semver.Version {
	res := v

	res.Patch = 0
	if v.Minor > 0 {
		res.Minor = v.Minor - 1
	}

	return res
}

func upperBoundVersion(v semver.Version) semver.Version {
	//nolint: mnd // we are setting the patch version to 0
	return semver.Version{
		Major: v.Major,
		Minor: v.Minor + 2,
		Patch: 0,
	}
}
