package kubectl_versioner

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/flavio/kuberlr/internal/common"
	"github.com/flavio/kuberlr/internal/downloader"

	"github.com/blang/semver"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/klog"
)

func SanitizedVersion(v *version.Info) string {
	// We need patch level too, which is not exposed as a string
	// value unlike v.Major and v.Minor.
	// We have to parse the GitVersion for that
	sv, err := semver.ParseTolerant(v.GitVersion)
	if err != nil {
		klog.Warningf(
			"Failed to parse remote API GitVersion",
			v.GitVersion,
			"with error",
			err,
		)
		return fmt.Sprintf("%s.%s.0", v.Major, v.Minor)
	}

	return fmt.Sprintf("%s.%s.%d", v.Major, v.Minor, sv.Patch)
}

func BuildKubectNameFromVersion(v *version.Info) string {
	return fmt.Sprintf("kubectl-%s", SanitizedVersion(v))
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

func EnsureKubectlIsAvailable(v *version.Info) (string, error) {
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

	downloadUrl, err := downloader.KubectlDownloadURL(SanitizedVersion(v))
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
