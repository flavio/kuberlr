package main

import (
	"github.com/flavio/kuberlr/internal/osexec"
	"github.com/spf13/cobra"
	"k8s.io/klog"
	"os"
	"path/filepath"
	"strings"

	"github.com/flavio/kuberlr/cmd/kuberlr/flags"
	"github.com/flavio/kuberlr/internal/config"
	"github.com/flavio/kuberlr/internal/finder"
)

func main() {
	klog.InitFlags(nil)

	binary := osexec.TrimExt(filepath.Base(os.Args[0]))
	if strings.HasSuffix(binary, "kubectl") {
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
		NewBinsCmd(),
	)

	flags.RegisterVerboseFlag(cmd.PersistentFlags())

	return cmd
}

func kubectlWrapperMode() {
	cfg := config.NewCfg()
	v, err := cfg.Load()
	if err != nil {
		klog.Fatal(err)
	}

	kFinder := finder.NewKubectlFinder("", v.GetString("SystemPath"))
	versioner := finder.NewVersioner(kFinder)
	version, err := versioner.KubectlVersionToUse(v.GetInt64("Timeout"))
	if err != nil {
		klog.Fatal(err)
	}

	kubectlBin, err := versioner.EnsureCompatibleKubectlAvailable(
		version,
		v.GetBool("AllowDownload"))
	if err != nil {
		klog.Fatal(err)
	}

	childArgs := append([]string{kubectlBin}, os.Args[1:]...)
	err = osexec.Exec(kubectlBin, childArgs, os.Environ())
	klog.Fatal(err)
}
