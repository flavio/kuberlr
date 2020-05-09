package main

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/blang/semver"
	"github.com/spf13/cobra"

	"github.com/flavio/kuberlr/cmd/kuberlr/flags"
	"github.com/flavio/kuberlr/internal/kubectl_versioner"
	"github.com/flavio/kuberlr/internal/kubehelper"
	"k8s.io/klog"
)

func main() {
	klog.InitFlags(nil)

	if strings.HasSuffix(os.Args[0], "kubectl") {
		kubectlWrapperMode()
	}
	nativeMode()
}

func nativeMode() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		// grab the base filename if the binary file is link
		Use: filepath.Base(os.Args[0]),
	}

	cmd.AddCommand(
		NewVersionCmd(),
	)

	flags.RegisterVerboseFlag(cmd.PersistentFlags())

	return cmd
}

func kubectlWrapperMode() {
	version, err := kubectlVersionToUse()
	if err != nil {
		klog.Fatal(err)
	}

	kubectlBin, err := kubectl_versioner.EnsureKubectlIsAvailable(version)
	if err != nil {
		klog.Fatal(err)
	}

	childArgs := append([]string{kubectlBin}, os.Args[1:]...)
	err = syscall.Exec(kubectlBin, childArgs, os.Environ())
	klog.Fatal(err)
}

func kubectlVersionToUse() (semver.Version, error) {
	version, err := kubehelper.ApiVersion()
	if err != nil && isTimeout(err) {
		// the remote server is unreachable, let's get
		// the latest version of kubectl that is available on the system
		klog.Info("Remote kubernetes server unreachable")
		version, err = kubectl_versioner.MostRecentKubectlDownloaded()
		if err != nil && isNoVersionFound(err) {
			klog.Info("No local kubectl binary found, fetching latest stable release version")
			version, err = kubectl_versioner.UpstreamStableVersion()
		}
	}
	return version, err
}

func isTimeout(err error) bool {
	urlError, ok := err.(*url.Error)
	return ok && urlError.Timeout()
}

func isNoVersionFound(err error) bool {
	nvError, ok := err.(*kubectl_versioner.NoVersionFoundError)
	return ok && nvError.NoVersionFound()
}
