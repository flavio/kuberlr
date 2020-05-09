package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/flavio/kuberlr/cmd/kuberlr/flags"
	"github.com/flavio/kuberlr/internal/kubectl_versioner"
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
	version, err := kubectl_versioner.KubectlVersionToUse()
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
