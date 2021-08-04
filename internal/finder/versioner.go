package finder

import (
	"errors"
	"github.com/flavio/kuberlr/internal/common"
	"github.com/flavio/kuberlr/internal/downloader"
	"github.com/flavio/kuberlr/internal/kubehelper"
	"path/filepath"

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

// KubectlVersionToUse returns the kubectl version to be used to interact with
// the remote server. The method takes into account different failure scenarios
// and acts accordingly.
func (v *Versioner) KubectlVersionToUse(timeout int64) (semver.Version, error) {
	version, err := v.apiServer.Version(timeout)
	if err != nil {
		if isTimeout(err) {
			// the remote server is unreachable, let's get
			// the latest version of kubectl that is available on the system
			klog.Info("Remote kubernetes server unreachable")
		} else {
			klog.Info(err)
		}
		kubectl, err := v.kFinder.MostRecentKubectlAvailable()
		if err == nil {
			return kubectl.Version, nil
		} else if common.IsNoVersionFound(err) {
			klog.Info("No local kubectl binary found, fetching latest stable release version")
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
		return "", errors.New("The right kubectl is missing, binary downloads from kubernetes' upstream mirror are disabled")
	}

	klog.Infof("Right kubectl missing, downloading version %s", version.String())

	//download the right kubectl to the local cache
	filename := filepath.Join(
		common.LocalDownloadDir(),
		common.BuildKubectlNameForLocalBin(version))

	if err := v.downloader.GetKubectlBinary(version, filename); err != nil {
		return "", err
	}

	return filename, nil
}

func isTimeout(err error) bool {
	t, ok := err.(interface {
		Timeout() bool
	})
	return ok && t.Timeout()
}
