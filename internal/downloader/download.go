package downloader

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/schollz/progressbar/v3"
)

func KubectlDownloadURL(version string) (string, error) {
	// Example: https://storage.googleapis.com/kubernetes-release/release/v1.18.0/bin/linux/amd64/kubectlI
	u, err := url.Parse(fmt.Sprintf(
		"https://storage.googleapis.com/kubernetes-release/release/v%s/bin/%s/%s/kubectl",
		version,
		runtime.GOOS,
		runtime.GOARCH,
	))
	if err != nil {
		return "", err
	}

	return u.String(), nil
}

func Download(urlToGet, destination string, mode os.FileMode) error {
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
