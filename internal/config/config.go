package config

import (
	"context"
	"os"
	"strings"

	"github.com/spf13/viper"

	"github.com/flavio/kuberlr/internal/common"
	"github.com/flavio/kuberlr/internal/logger"
)

const DefaultTimeout = 5

// Cfg is used to retrieve the configuration of kuberlr.
type Cfg struct {
	Paths []string
}

// NewCfg returns a new Cfg object that is pre-configured
// to look for the configuration files in the right set of
// directories.
func NewCfg() *Cfg {
	return &Cfg{
		Paths: configFiles,
	}
}

// Defaults returns a configuration that holds only the built-in default
// values. It does not read the configuration files or the environment.
func Defaults() *viper.Viper {
	v := viper.New()
	v.SetDefault("AllowDownload", true)
	v.SetDefault("SystemPath", common.SystemPath)
	v.SetDefault("DownloadPath", common.LocalRootDir())
	v.SetDefault("Timeout", DefaultTimeout)
	v.SetDefault("KubeMirrorUrl", "https://dl.k8s.io")
	v.SetDefault("UseLatestIfNoCompatible", false)
	v.SetDefault("Verbosity", logger.VerbosityDefault)
	v.SetDefault("Quiet", false)
	v.SetDefault("Color", string(logger.ColorAuto))

	return v
}

// Load reads the configuration files from disk and merges them on top of
// the built-in defaults. When a file cannot be read, Load returns the
// error together with the built-in defaults, so the caller always gets a
// usable configuration.
func (c *Cfg) Load() (*viper.Viper, error) {
	v := Defaults()

	v.SetConfigType("toml")

	// read environment variables, they take precedence
	v.AutomaticEnv()
	v.SetEnvPrefix("KUBERLR")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	for _, path := range c.Paths {
		if err := mergeConfig(v, path); err != nil {
			return Defaults(), err
		}
	}

	return v, nil
}

// GetKubeMirrorURL returns the URL of the kubernetes mirror.
func (c *Cfg) GetKubeMirrorURL() (string, error) {
	v, err := c.Load()
	if err != nil {
		return "", err
	}

	return v.GetString("KubeMirrorUrl"), nil
}

// LoggerOptions builds the terminal output options from the loaded
// configuration. An invalid "Color" value is reported as an error; the
// returned options are still usable and fall back to automatic color detection.
func LoggerOptions(v *viper.Viper) (logger.Options, error) {
	color, err := logger.ParseColorMode(v.GetString("Color"))

	return logger.Options{
		Verbosity: v.GetInt("Verbosity"),
		Quiet:     v.GetBool("Quiet"),
		Color:     color,
	}, err
}

// configKey is the context key under which the configuration is stored.
type configKey struct{}

// NewContext returns a copy of ctx carrying the given configuration.
func NewContext(ctx context.Context, v *viper.Viper) context.Context {
	return context.WithValue(ctx, configKey{}, v)
}

// FromContext returns the configuration stored in ctx by NewContext. When ctx
// carries no configuration, FromContext returns the built-in defaults.
func FromContext(ctx context.Context) *viper.Viper {
	if ctx != nil {
		if v, ok := ctx.Value(configKey{}).(*viper.Viper); ok && v != nil {
			return v
		}
	}
	return Defaults()
}

func mergeConfig(v *viper.Viper, cfgFile string) error {
	_, err := os.Stat(cfgFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	v.SetConfigFile(cfgFile)

	return v.MergeInConfig()
}
