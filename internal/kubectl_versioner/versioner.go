package kubectl_versioner

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/flavio/kuberlr/internal/downloader"
	"github.com/flavio/kuberlr/internal/kubehelper"

	"github.com/blang/semver"
	"k8s.io/klog"
)

const KUBECTL_STABLE_URL = "https://storage.googleapis.com/kubernetes-release/release/stable.txt"

func KubectlVersionToUse() (semver.Version, error) {
	version, err := kubehelper.ApiVersion()
	if err != nil && isTimeout(err) {
		// the remote server is unreachable, let's get
		// the latest version of kubectl that is available on the system
		klog.Info("Remote kubernetes server unreachable")
		version, err = MostRecentKubectlDownloaded()
		if err != nil && isNoVersionFound(err) {
			klog.Info("No local kubectl binary found, fetching latest stable release version")
			version, err = UpstreamStableVersion()
		}
	}
	return version, err
}

func isTimeout(err error) bool {
	urlError, ok := err.(*url.Error)
	return ok && urlError.Timeout()
}

func isNoVersionFound(err error) bool {
	nvError, ok := err.(*NoVersionFoundError)
	return ok && nvError.NoVersionFound()
}

func EnsureKubectlIsAvailable(v semver.Version) (string, error) {
	filename := filepath.Join(
		LocalDownloadDir(),
		BuildKubectNameFromVersion(v))

	if IsKubectlAvailable(filename) {
		return filename, nil
	}

	//download the right kubectl to the local cache
	if err := SetupLocalDirs(); err != nil {
		return "", err
	}

	downloadUrl, err := downloader.KubectlDownloadURL(v)
	if err != nil {
		return "", err
	}

	klog.Infof("Right kubectl missing, downloading version %s", v.String())
	err = downloader.Download(downloadUrl, filename, 0755)
	if err != nil {
		return "", err
	}
	return filename, nil
}

func MostRecentKubectlDownloaded() (semver.Version, error) {
	versions, err := LocalKubectlVersions()
	if err != nil {
		return semver.Version{}, err
	}

	if len(versions) == 0 {
		return semver.Version{}, &NoVersionFoundError{}
	}

	semver.Sort(versions)
	return versions[len(versions)-1], nil
}

func UpstreamStableVersion() (semver.Version, error) {
	res, err := http.Get(KUBECTL_STABLE_URL)
	if err != nil {
		return semver.Version{}, err
	}
	if res.StatusCode != http.StatusOK {
		return semver.Version{},
			fmt.Errorf(
				"GET %s returned http status %s",
				KUBECTL_STABLE_URL,
				res.Status,
			)
	}

	v, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return semver.Version{}, err
	}

	return semver.ParseTolerant(string(v))
}
