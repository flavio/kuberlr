package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/flavio/kuberlr/internal/osexec"

	"github.com/blang/semver/v4"

	"github.com/flavio/kuberlr/cmd/kuberlr/flags"
	"github.com/flavio/kuberlr/internal/config"
	"github.com/flavio/kuberlr/internal/finder"
	"github.com/flavio/kuberlr/internal/logger"
)

func main() {
	binary := osexec.TrimExt(filepath.Base(os.Args[0]))
	if strings.HasSuffix(binary, "kubectl") {
		kubectlWrapperMode(os.Args[1:])
	}
	nativeMode()
}

func nativeMode() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func NewKubectlWrapperCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "kubectl",
		Short:              "Wrap and exec a suitable version kubectl command",
		DisableFlagParsing: true,
		Run: func(_ *cobra.Command, args []string) {
			kubectlWrapperMode(args)
		},
	}
}

func newRootCmd() *cobra.Command {
	var output flags.Output

	cmd := &cobra.Command{
		// grab the base filename if the binary file is link
		Use: filepath.Base(os.Args[0]),
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			log := setupOutput(cmd, &output)
			cmd.SetContext(logger.NewContext(cmd.Context(), log))
		},
	}

	cmd.AddCommand(
		NewVersionCmd(),
		NewBinsCmd(),
		NewGetCmd(),
		NewUpdateCmd(),
		NewRmCmd(),
		NewKubectlWrapperCmd(),
	)

	output.Register(cmd.PersistentFlags())

	return cmd
}

// setupOutput configures the terminal output for the native subcommands and
// returns the logger to use. Command line flags take precedence over the
// configuration file and the environment. Problems are reported as warnings,
// they never prevent the subcommand from running.
func setupOutput(cmd *cobra.Command, output *flags.Output) *logger.Logger {
	v, loadErr := config.NewCfg().Load()
	if loadErr != nil {
		v = viper.New()
	}

	opts, colorErr := config.LoggerOptions(v)
	opts, flagErr := output.Apply(cmd.Flags(), opts)

	log := logger.Setup(opts)

	if loadErr != nil {
		log.Warn("cannot load configuration, using defaults", "error", loadErr)
	}
	if colorErr != nil {
		log.Warn("ignoring invalid Color setting", "error", colorErr)
	}
	if flagErr != nil {
		log.Warn("ignoring invalid --color flag", "error", flagErr)
	}

	return log
}

func kubectlWrapperMode(args []string) {
	cfg := config.NewCfg()
	v, err := cfg.Load()
	if err != nil {
		// the configuration is unusable: set up the default output just to
		// be able to report the failure
		log := logger.Setup(logger.Options{})
		log.Fatalf("cannot load configuration: %v", err)
	}

	opts, colorErr := config.LoggerOptions(v)
	log := logger.Setup(opts)
	if colorErr != nil {
		log.Warn("ignoring invalid Color setting", "error", colorErr)
	}

	kubectlFinder := finder.NewKubectlFinder("", v.GetString("SystemPath"))
	versioner := finder.NewVersioner(kubectlFinder, log)

	var version semver.Version
	if finder.IsLocalOnlyCommand(args) {
		log.Trace("kubectl command does not need to talk to the API server, skipping remote version check")
		version, err = versioner.MostRecentKubectlVersionAvailableOrLatestFromUpstream()
	} else {
		version, err = versioner.KubectlVersionToUse(v.GetInt64("Timeout"))
	}
	if err != nil {
		log.Fatalf("cannot determine which kubectl version to use: %v", err)
	}

	kubectlBin, err := versioner.EnsureCompatibleKubectlAvailable(
		version,
		v.GetBool("AllowDownload"),
		v.GetBool("UseLatestIfNoCompatible"),
	)
	if err != nil {
		log.Fatalf("cannot find a compatible kubectl: %v", err)
	}

	childArgs := append([]string{kubectlBin}, args...)
	err = osexec.Exec(kubectlBin, childArgs, os.Environ())
	log.Fatalf("cannot execute kubectl binary located at %s: %v", kubectlBin, err)
}
