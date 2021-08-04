package downloader

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/flavio/kuberlr/internal/common"
	"github.com/flavio/kuberlr/internal/osexec"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/schollz/progressbar/v3"
)

// KubectlStableURL URL of the text file used by kubernetes community
// to hold the latest stable version of kubernetes released
const KubectlStableURL = "https://storage.googleapis.com/kubernetes-release/release/stable.txt"

// Downloder is a helper class that is used to interact with the
// kubernetes infrastructure holding released binaries and release information
type Downloder struct {
}

func (d *Downloder) getContentsOfURL(url string) (string, error) {
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

	v, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", err
	}
	return string(v), nil
}

// UpstreamStableVersion returns the latest version of kubernetes that upstream
// considers stable
func (d *Downloder) UpstreamStableVersion() (semver.Version, error) {
	v, err := d.getContentsOfURL(KubectlStableURL)
	if err != nil {
		return semver.Version{}, err
	}
	return semver.ParseTolerant(v)
}

// GetKubectlBinary downloads the kubectl binary identified by the given version
// to the specified destination
func (d *Downloder) GetKubectlBinary(version semver.Version, destination string) error {
	var firstErr error
	const maxNumTries = 3
	const timeToSleepOnRetryPerIter = 10 // seconds

	for iter := 1; iter <= maxNumTries; iter++ {
		downloadURL, err := d.kubectlDownloadURL(version)
		if err != nil {
			return err
		}

		if _, err := os.Stat(filepath.Dir(destination)); err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(filepath.Dir(destination), os.ModePerm)
			}
			if err != nil {
				return err
			}
		}

		err = d.download(fmt.Sprintf("kubectl%s%s", version, osexec.Ext), downloadURL, destination, 0755)
		if err == nil {
			return nil
		}
		if iter == 0 {
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

func (d *Downloder) kubectlDownloadURL(v semver.Version) (string, error) {
	// Example: https://storage.googleapis.com/kubernetes-release/release/v1.18.0/bin/linux/amd64/kubectlI
	u, err := url.Parse(fmt.Sprintf(
		"https://storage.googleapis.com/kubernetes-release/release/v%d.%d.%d/bin/%s/%s/kubectl%s",
		v.Major,
		v.Minor,
		v.Patch,
		runtime.GOOS,
		runtime.GOARCH,
		osexec.Ext,
	))
	if err != nil {
		return "", err
	}

	return u.String(), nil
}

func (d *Downloder) download(desc, urlToGet, destination string, mode os.FileMode) error {
	shaURLToGet := urlToGet + ".sha256"
	shaExpected, err := d.getContentsOfURL(shaURLToGet)
	if err != nil {
		return fmt.Errorf("Error while trying to get contents of %s: %v", shaURLToGet, err)
	}
	shaExpected = strings.TrimRight(shaExpected, "\n")

	req, err := http.NewRequest("GET", urlToGet, nil)
	if err != nil {
		return fmt.Errorf(
			"Error while issuing GET request against %s: %v",
			urlToGet, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf(
			"Error while issuing GET request against %s: %v",
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
	temporaryDestinationFile, err := ioutil.TempFile(os.TempDir(), "kuberlr-kubectl-")
	if err != nil {
		return fmt.Errorf("Error trying to create temporary file in %s: %v", os.TempDir(), err)
	}
	defer temporaryDestinationFile.Close()
	defer os.Remove(temporaryDestinationFile.Name())

	// write progress to stderr, writing to stdout would
	// break bash/zsh/shell completion
	fmt.Fprintf(os.Stderr, "Downloading %s\n", urlToGet)
	bar := progressbar.NewOptions(
		int(resp.ContentLength),
		progressbar.OptionSetDescription(desc),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(10*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprintln(os.Stderr, " done.")
		}),
	)
	hasher := sha256.New()

	_, err = io.Copy(io.MultiWriter(temporaryDestinationFile, bar, hasher), resp.Body)
	if err != nil {
		return fmt.Errorf(
			"Error while downloading text of %s into file %s: %v",
			urlToGet, temporaryDestinationFile.Name(), err)
	}

	shaActual := hex.EncodeToString(hasher.Sum(nil))
	if shaExpected != shaActual {
		return &common.ShaMismatchError{URL: urlToGet, ShaExpected: shaExpected, ShaActual: shaActual}
	}

	err = os.Rename(temporaryDestinationFile.Name(), destination)
	if err != nil {
		linkErr, ok := err.(*os.LinkError)
		if ok {
			fmt.Fprintf(os.Stderr, "Cross-device error trying to rename a file: %s -- will do a full copy\n", linkErr)
			tempInput, err := ioutil.ReadFile(temporaryDestinationFile.Name())
			if err != nil {
				return fmt.Errorf("Error reading temporary file %s: %v",
					temporaryDestinationFile.Name(), err)
			}
			err = ioutil.WriteFile(destination, tempInput, mode)
		}
	} else {
		err = os.Chmod(destination, mode)
	}
	return err
}
