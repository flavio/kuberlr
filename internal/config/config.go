package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"

	"github.com/flavio/kuberlr/internal/common"
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

// Load reads the configuration files from disks and merges them.
func (c *Cfg) Load() (*viper.Viper, error) {
	v := viper.New()
	v.SetDefault("AllowDownload", true)
	v.SetDefault("SystemPath", common.SystemPath)
	v.SetDefault("Timeout", DefaultTimeout)
	v.SetDefault("KubeMirrorUrl", "https://dl.k8s.io")
	v.SetDefault("UseLatestIfNoCompatible", false)

	v.SetConfigType("toml")

	// read environment variables, they take precedence
	v.AutomaticEnv()
	v.SetEnvPrefix("KUBERLR")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if len(c.Paths) == 0 {
		return v, nil
	}

	for _, path := range c.Paths {
		if err := mergeConfig(v, path); err != nil {
			return viper.New(), err
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
