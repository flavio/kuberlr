package downloader

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/blang/semver"
	"github.com/schollz/progressbar/v3"
)

const KUBECTL_STABLE_URL = "https://storage.googleapis.com/kubernetes-release/release/stable.txt"

type Downloder struct {
}

func (d *Downloder) UpstreamStableVersion() (semver.Version, error) {
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

func (d *Downloder) GetKubectlBinary(version semver.Version, destination string) error {
	downloadUrl, err := d.kubectlDownloadURL(version)
	if err != nil {
		return err
	}

	return d.download(downloadUrl, destination, 0755)
}

func (d *Downloder) kubectlDownloadURL(v semver.Version) (string, error) {
	// Example: https://storage.googleapis.com/kubernetes-release/release/v1.18.0/bin/linux/amd64/kubectlI
	u, err := url.Parse(fmt.Sprintf(
		"https://storage.googleapis.com/kubernetes-release/release/v%d.%d.%d/bin/%s/%s/kubectl",
		v.Major,
		v.Minor,
		v.Patch,
		runtime.GOOS,
		runtime.GOARCH,
	))
	if err != nil {
		return "", err
	}

	return u.String(), nil
}

func (d *Downloder) download(urlToGet, destination string, mode os.FileMode) error {
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

	f, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY, mode)
	if err != nil {
		return fmt.Errorf(
			"Error while downloading %s to %s: %v",
			urlToGet, destination, err)
	}
	defer f.Close()

	// write progress to stderr, writing to stdout would
	// break bash/zsh/shell completion
	bar := progressbar.NewOptions(
		int(resp.ContentLength),
		progressbar.OptionSetDescription(urlToGet),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionThrottle(10*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprintln(os.Stderr, " done.")
		}),
	)

	_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
	return err
}
