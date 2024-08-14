package finder

import (
	"errors"
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
	SystemKubectlBinaries() (KubectlBinaries, error)
	LocalKubectlBinaries() (KubectlBinaries, error)
	AllKubectlBinaries(reverseSort bool) KubectlBinaries
	FindCompatibleKubectl(requestedVersion semver.Version) (KubectlBinary, error)
	MostRecentKubectlAvailable() (KubectlBinary, error)
}

// Versioner is used to manage the local kubectl binaries used by kuberlr
type Versioner struct {
	kFinder    iFinder
	downloader downloadHelper
	apiServer  kubeAPIHelper
}

// NewVersioner is an helper function that creates a new Versioner instance
func NewVersioner(f iFinder) *Versioner {
	return &Versioner{
		kFinder:    f,
		downloader: &downloader.Downloder{},
		apiServer:  &kubehelper.KubeAPI{},
	}
}

const (
	_ = iota
	VerbosityOne
	VerbosityTwo
)

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

	_, recursiveInvocationDetected := os.LookupEnv(PreventRecursiveInvocationEnvName)
	if recursiveInvocationDetected {
		klog.V(VerbosityTwo).Info("clint-go invoked kubectl to authenticate. Preventing kuberlr endless recursion loop.")
		return v.mostRecentKubectlVersionAvailableOrLatestFromUpstream()
	}

	os.Setenv(PreventRecursiveInvocationEnvName, "1")
	defer os.Unsetenv(PreventRecursiveInvocationEnvName)

	version, err := v.apiServer.Version(timeout)
	if err != nil {
		if isUnreachable(err) {
			// the remote server is unreachable, let's get
			// the latest version of kubectl that is available on the system
			klog.V(VerbosityTwo).Info("Remote kubernetes server unreachable")
		} else {
			klog.V(VerbosityOne).Info(err)
		}
		return v.mostRecentKubectlVersionAvailableOrLatestFromUpstream()
	}
	return version, err
}

// mostRecentKubectlVersionAvailableOrLatestFromUpstream returns the most recent version of kubectl
// available on the system. If no kubectl binary is found, it will download the
// latest stable version from the upstream mirror
func (v *Versioner) mostRecentKubectlVersionAvailableOrLatestFromUpstream() (semver.Version, error) {
	if kubectl, err := v.kFinder.MostRecentKubectlAvailable(); err == nil {
		return kubectl.Version, nil
	}

	klog.V(VerbosityTwo).Info("No local kubectl binary found, fetching latest stable release version")
	return v.downloader.UpstreamStableVersion()
}

// EnsureCompatibleKubectlAvailable ensures the kubectl binary with the specified
// version is available on the system. It will return the full path to the
// binary
func (v *Versioner) EnsureCompatibleKubectlAvailable(version semver.Version, allowDownload bool) (string, error) {
	kubectl, err := v.kFinder.FindCompatibleKubectl(version)
	if err == nil {
		return kubectl.Path, nil
	}

	if !allowDownload {
		return "", errors.New("the right kubectl is missing, binary downloads from kubernetes' upstream mirror are disabled")
	}

	klog.Infof("Right kubectl missing, downloading version %s", version.String())

	// download the right kubectl to the local cache
	filename := filepath.Join(
		common.LocalDownloadDir(),
		common.BuildKubectlNameForLocalBin(version))

	if err = v.downloader.GetKubectlBinary(version, filename); err != nil {
		return "", err
	}

	return filename, nil
}

func isUnreachable(err error) bool {
	var e *url.Error
	return os.IsTimeout(err) || errors.As(err, &e)
}
