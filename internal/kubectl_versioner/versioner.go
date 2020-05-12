package kubectl_versioner

import (
	"path/filepath"

	"github.com/flavio/kuberlr/internal/downloader"
	"github.com/flavio/kuberlr/internal/kubehelper"

	"github.com/blang/semver"
	"k8s.io/klog"
)

type localCache interface {
	LocalDownloadDir() string
	IsKubectlAvailable(filename string) bool
	SetupLocalDirs() error
	LocalKubectlVersions() (semver.Versions, error)
}

type downloadHelper interface {
	GetKubectlBinary(version semver.Version, destination string) error
	UpstreamStableVersion() (semver.Version, error)
}

type kubeAPIHelper interface {
	Version() (semver.Version, error)
}

type Versioner struct {
	cache      localCache
	downloader downloadHelper
	apiServer  kubeAPIHelper
}

func NewVersioner() *Versioner {
	return &Versioner{
		cache:      &localCacheHandler{},
		downloader: &downloader.Downloder{},
		apiServer:  &kubehelper.KubeAPI{},
	}
}

func (v *Versioner) KubectlVersionToUse() (semver.Version, error) {
	version, err := v.apiServer.Version()
	if err != nil && isTimeout(err) {
		// the remote server is unreachable, let's get
		// the latest version of kubectl that is available on the system
		klog.Info("Remote kubernetes server unreachable")
		version, err = v.MostRecentKubectlDownloaded()
		if err != nil && isNoVersionFound(err) {
			klog.Info("No local kubectl binary found, fetching latest stable release version")
			version, err = v.downloader.UpstreamStableVersion()
		}
	}
	return version, err
}

func (v *Versioner) kubectlBinary(version semver.Version) string {
	return filepath.Join(
		v.cache.LocalDownloadDir(),
		BuildKubectNameFromVersion(version))
}

func (v *Versioner) EnsureKubectlIsAvailable(version semver.Version) (string, error) {
	filename := v.kubectlBinary(version)

	if v.cache.IsKubectlAvailable(filename) {
		return filename, nil
	}

	klog.Infof("Right kubectl missing, downloading version %s", version.String())

	//download the right kubectl to the local cache
	if err := v.cache.SetupLocalDirs(); err != nil {
		return "", err
	}

	if err := v.downloader.GetKubectlBinary(version, filename); err != nil {
		return "", err
	}

	return filename, nil
}

func (v *Versioner) MostRecentKubectlDownloaded() (semver.Version, error) {
	versions, err := v.cache.LocalKubectlVersions()
	if err != nil {
		return semver.Version{}, err
	}

	if len(versions) == 0 {
		return semver.Version{}, &NoVersionFoundError{}
	}

	semver.Sort(versions)
	return versions[len(versions)-1], nil
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
