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

// KubectlVersionToUse returns the kubectl version to be used to interact with
// the remote server. The method takes into account different failure scenarios
// and acts accordingly.
func (v *Versioner) KubectlVersionToUse(timeout int64) (semver.Version, error) {
	// Using kubectl exec plugin, the Kubernetes version client will systematically
	// execute kubectl to obtain credentials to the cluster.
	// Having `kubectl` in the `PATH` substitute by kuberlr, this causes infinite recursion loop.
	// kuberlr/kubectl calls Kubernetes client go that in its tunrn calls kuberlr/kubectl to authenticate to the server.
	// To avoid this, we set an environment variable to signal that we are currently resolving the kubernetes version.
	const protectVersionRecusrionEnvName = "KUBERLR_RESOLVING_VERSION"
	_, ok := os.LookupEnv(protectVersionRecusrionEnvName)
	if ok {
		klog.V(VerbosityTwo).Info("Currently resolving the kubernetes version. Avoid infinite recursion loop and returning the latest stable version")
		return v.downloader.UpstreamStableVersion()
	} else {
		defer os.Unsetenv(protectVersionRecusrionEnvName)
	}
	os.Setenv(protectVersionRecusrionEnvName, "1")

	version, err := v.apiServer.Version(timeout)
	if err != nil {
		if isUnreachable(err) {
			// the remote server is unreachable, let's get
			// the latest version of kubectl that is available on the system
			klog.V(VerbosityTwo).Info("Remote kubernetes server unreachable")
		} else {
			klog.V(VerbosityOne).Info(err)
		}
		kubectl, internalErr := v.kFinder.MostRecentKubectlAvailable()
		if internalErr == nil {
			return kubectl.Version, nil
		} else if common.IsNoVersionFound(internalErr) {
			klog.V(VerbosityTwo).Info("No local kubectl binary found, fetching latest stable release version")
			return v.downloader.UpstreamStableVersion()
		}
	}
	return version, err
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
