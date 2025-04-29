package downloader

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/flavio/kuberlr/internal/common"
	"github.com/flavio/kuberlr/internal/config"
	"github.com/flavio/kuberlr/internal/osexec"

	"github.com/blang/semver/v4"
	"github.com/schollz/progressbar/v3"
	"k8s.io/klog"
)

func getKubeMirrorURL() (string, error) {
	cfg := config.NewCfg()
	return cfg.GetKubeMirrorURL()
}

// Downloder is a helper class that is used to interact with the
// kubernetes infrastructure holding released binaries and release information.
type Downloder struct{}

func (d *Downloder) getContentsOfURL(url string) (string, error) {
	//nolint: gosec,noctx // the url is built internally
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "",
			fmt.Errorf(
				"GET %s returned http status %s",
				url,
				res.Status,
			)
	}

	v, err := io.ReadAll(res.Body)
	defer func() {
		if e := res.Body.Close(); e != nil {
			klog.V(common.VerbosityTwo).Infof("error closing response body: %v", e)
		}
	}()
	if err != nil {
		return "", err
	}
	return string(v), nil
}

// UpstreamStableVersion returns the latest version of kubernetes that upstream
// considers stable.
func (d *Downloder) UpstreamStableVersion() (semver.Version, error) {
	baseURL, err := getKubeMirrorURL()
	if err != nil {
		return semver.Version{}, err
	}
	url, err := url.Parse(baseURL + "/release/stable.txt")
	if err != nil {
		return semver.Version{}, err
	}

	v, err := d.getContentsOfURL(url.String())
	if err != nil {
		return semver.Version{}, err
	}
	return semver.ParseTolerant(v)
}

// GetKubectlBinary downloads the kubectl binary identified by the given version
// to the specified destination.
func (d *Downloder) GetKubectlBinary(version semver.Version, destination string) error {
	var firstErr error
	const maxNumTries = 3
	const timeToSleepOnRetryPerIter = 10 // seconds

	hashing, hashingErr := NewHashing(version)
	if hashingErr != nil {
		return hashingErr
	}

	for iter := 1; iter <= maxNumTries; iter++ {
		downloadURL, err := d.kubectlDownloadURL(version)
		if err != nil {
			return err
		}

		if _, err = os.Stat(filepath.Dir(destination)); err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(filepath.Dir(destination), os.ModeDir)
			}
			if err != nil {
				return err
			}
		}

		//nolint: mnd // setting the mode to read/write/execute for owner only
		err = d.download(fmt.Sprintf("kubectl%s%s", version, osexec.Ext), downloadURL, hashing, destination, 0o755)
		if err == nil {
			return nil
		}
		if iter == 1 {
			firstErr = err
		}
		if common.IsShaMismatch(err) {
			// Try downloading an older subversion
			fmt.Fprintf(os.Stderr, "Error on download attempt #%d: %s\n", iter, err)
			time.Sleep(time.Duration(iter*timeToSleepOnRetryPerIter) * time.Second)
		} else {
			break
		}
	}
	return firstErr
}

func (d *Downloder) kubectlDownloadURL(version semver.Version) (string, error) {
	// Example: https://storage.googleapis.com/kubernetes-release/release/v1.18.0/bin/linux/amd64/kubectlI
	baseURL, err := getKubeMirrorURL()
	if err != nil {
		return "", err
	}
	url, err := url.Parse(fmt.Sprintf(
		"%s/release/v%d.%d.%d/bin/%s/%s/kubectl%s",
		baseURL,
		version.Major,
		version.Minor,
		version.Patch,
		runtime.GOOS,
		runtime.GOARCH,
		osexec.Ext,
	))
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

func (d *Downloder) download(desc string,
	urlToGet string,
	hashing *Hashing,
	destination string,
	mode os.FileMode,
) error {
	shaURLToGet := urlToGet + hashing.Suffix
	shaExpected, err := d.getContentsOfURL(shaURLToGet)
	if err != nil {
		return fmt.Errorf("error while trying to get contents of %s: %w", shaURLToGet, err)
	}
	shaExpected = strings.TrimRight(shaExpected, "\n")

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, urlToGet, nil)
	if err != nil {
		return fmt.Errorf(
			"error while issuing GET request against %s: %w",
			urlToGet, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf(
			"error while issuing GET request against %s: %w",
			urlToGet, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"GET %s returned http status %s",
			urlToGet,
			resp.Status,
		)
	}
	temporaryDestinationFile, err := os.CreateTemp(os.TempDir(), "kuberlr-kubectl-")
	if err != nil {
		return fmt.Errorf("error trying to create temporary file in %s: %w", os.TempDir(), err)
	}

	tmpname := temporaryDestinationFile.Name()
	defer os.Remove(tmpname)

	// write progress to stderr, writing to stdout would
	// break bash/zsh/shell completion
	fmt.Fprintf(os.Stderr, "Downloading %s\n", urlToGet)
	bar := progressbar.NewOptions(
		int(resp.ContentLength),
		progressbar.OptionSetDescription(desc),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),                  //nolint: mnd // 40 is a good width
		progressbar.OptionThrottle(10*time.Millisecond), //nolint: mnd // 10ms is a good throttle
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprintln(os.Stderr, " done.")
		}),
	)

	_, err = io.Copy(io.MultiWriter(temporaryDestinationFile, bar, hashing.Hasher), resp.Body)
	if err != nil {
		if e := temporaryDestinationFile.Close(); e != nil {
			klog.V(common.VerbosityTwo).Infof("error closing temporary file: %v", e)
		}
		return fmt.Errorf(
			"error while downloading text of %s into file %s: %w",
			urlToGet, tmpname, err)
	}

	// Closing the file handler prior to performing a rename so this process (the
	// open file handler) does not conflict with the rename.
	if err = temporaryDestinationFile.Close(); err != nil {
		return fmt.Errorf("error closing temporary file %s: %w", tmpname, err)
	}

	shaActual := hex.EncodeToString(hashing.Hasher.Sum(nil))
	if shaExpected != shaActual {
		return &common.ShaMismatchError{URL: urlToGet, ShaExpected: shaExpected, ShaActual: shaActual}
	}

	err = os.Rename(tmpname, destination)
	if err != nil {
		var linkErr *os.LinkError
		if errors.As(err, &linkErr) {
			fmt.Fprintf(os.Stderr, "Cross-device error trying to rename a file: %s -- will do a full copy\n", linkErr)
			var tempInput []byte
			tempInput, err = os.ReadFile(tmpname)
			if err != nil {
				return fmt.Errorf("error reading temporary file %s: %w",
					tmpname, err)
			}
			err = os.WriteFile(destination, tempInput, mode)
		}
	} else {
		err = os.Chmod(destination, mode)
	}
	return err
}
