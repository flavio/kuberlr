package kubectl_versioner

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/flavio/kuberlr/internal/common"
	"github.com/flavio/kuberlr/internal/downloader"

	"github.com/blang/semver"
	"k8s.io/klog"
)

const KUBECTL_STABLE_URL = "https://storage.googleapis.com/kubernetes-release/release/stable.txt"

func BuildKubectNameFromVersion(v semver.Version) string {
	return fmt.Sprintf("kubectl-%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func LocalDownloadDir() string {
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	return filepath.Join(
		common.HomeDir(),
		".kuberlr",
		platform,
	)
}

func IsKubectlAvailable(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

func SetupLocalDirs() error {
	return os.MkdirAll(LocalDownloadDir(), os.ModePerm)
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

	klog.Info("Correct kubectl version missing, downloading...")
	err = downloader.Download(downloadUrl, filename, 0755)
	if err != nil {
		return "", err
	}
	return filename, nil
}

func MostRecentKubectlDownloaded() (semver.Version, error) {
	var versions []semver.Version

	kubectlBins, err := ioutil.ReadDir(LocalDownloadDir())
	if err != nil {
		if os.IsNotExist(err) {
			err = &NoVersionFoundError{}
		}
		return semver.Version{}, err
	}

	for _, f := range kubectlBins {
		var major, minor, patch uint64
		n, err := fmt.Sscanf(f.Name(), "kubectl-%d.%d.%d", &major, &minor, &patch)
		if n == 3 && err == nil {
			sv := semver.Version{
				Major: major,
				Minor: minor,
				Patch: patch,
			}
			versions = append(versions, sv)
		}
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
