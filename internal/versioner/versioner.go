package versioner

import (
	"path/filepath"

	"github.com/flavio/kuberlr/internal/downloader"
	"github.com/flavio/kuberlr/internal/kubehelper"

	"github.com/blang/semver"
	"k8s.io/klog"
)

type localCache interface {
	LocalDownloadDir() string
	SetupLocalDirs() error
	LocalKubectlBinaries() (KubectlBinaries, error)
	SystemKubectlBinaries() (KubectlBinaries, error)
	AllKubectlBinaries(reverseSort bool) KubectlBinaries
	FindCompatibleKubectl(requestedVersion semver.Version) (KubectlBinary, error)
}

type downloadHelper interface {
	GetKubectlBinary(version semver.Version, destination string) error
	UpstreamStableVersion() (semver.Version, error)
}

type kubeAPIHelper interface {
	Version() (semver.Version, error)
}

// Versioner is used to manage the local kubectl binaries used by kuberlr
type Versioner struct {
	cache      localCache
	downloader downloadHelper
	apiServer  kubeAPIHelper
}

// NewVersioner is an helper function that creates a new Versioner instance
func NewVersioner() *Versioner {
	return &Versioner{
		cache:      NewLocalCacheHandler(),
		downloader: &downloader.Downloder{},
		apiServer:  &kubehelper.KubeAPI{},
	}
}

// KubectlVersionToUse returns the kubectl version to be used to interact with
// the remote server. The method takes into account different failure scenarios
// and acts accordingly.
func (v *Versioner) KubectlVersionToUse() (semver.Version, error) {
	version, err := v.apiServer.Version()
	if err != nil {
		if isTimeout(err) {
			// the remote server is unreachable, let's get
			// the latest version of kubectl that is available on the system
			klog.Info("Remote kubernetes server unreachable")
		} else {
			klog.Info(err)
		}
		kubectl, err := v.MostRecentKubectlAvailable()
		if err == nil {
			return kubectl.Version, nil
		} else if isNoVersionFound(err) {
			klog.Info("No local kubectl binary found, fetching latest stable release version")
			return v.downloader.UpstreamStableVersion()
		}
	}
	return version, err
}

func (v *Versioner) kubectlBinary(version semver.Version) string {
	return filepath.Join(
		v.cache.LocalDownloadDir(),
		buildKubectlNameForLocalBin(version))
}

// EnsureCompatibleKubectlAvailable ensures the kubectl binary with the specified
// version is available on the system. It will return the full path to the
// binary
func (v *Versioner) EnsureCompatibleKubectlAvailable(version semver.Version) (string, error) {
	kubectl, err := v.cache.FindCompatibleKubectl(version)
	if err == nil {
		return kubectl.Path, nil
	}

	klog.Infof("Right kubectl missing, downloading version %s", version.String())

	//download the right kubectl to the local cache
	if err := v.cache.SetupLocalDirs(); err != nil {
		return "", err
	}

	filename := v.kubectlBinary(version)
	if err := v.downloader.GetKubectlBinary(version, filename); err != nil {
		return "", err
	}

	return filename, nil
}

// MostRecentKubectlAvailable returns the most recent version of
// kubectl available on the system. It could be something downloaded
// by kuberlr or something already available on the system
func (v *Versioner) MostRecentKubectlAvailable() (KubectlBinary, error) {
	bins := v.cache.AllKubectlBinaries(true)

	if len(bins) == 0 {
		return KubectlBinary{}, &NoVersionFoundError{}
	}

	return bins[0], nil
}

func isTimeout(err error) bool {
	t, ok := err.(interface {
		Timeout() bool
	})
	return ok && t.Timeout()
}

func isNoVersionFound(err error) bool {
	t, ok := err.(noVersionFound)
	return ok && t.NoVersionFound()
}
