package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/flavio/kuberlr/internal/common"
)

// Cfg is used to retrieve the configuration of kuberlr
type Cfg struct {
	Paths []string
}

// NewCfg returns a new Cfg object that is pre-configured
// to look for the configuration files in the right set of
// directories
func NewCfg() *Cfg {
	return &Cfg{
		Paths: []string{
			"/usr/etc/",
			"/etc/",
			filepath.Join(common.HomeDir(), ".kuberlr"),
		},
	}
}

// Load reads the configuration files from disks and merges them
func (c *Cfg) Load() (*viper.Viper, error) {
	v := viper.New()
	v.SetDefault("AllowDownload", true)
	v.SetDefault("SystemPath", common.SystemPath)
	v.SetDefault("Timeout", 5)

	v.SetConfigType("toml")

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

func mergeConfig(v *viper.Viper, extraConfigPath string) error {
	cfgFile := filepath.Join(extraConfigPath, "kuberlr.conf")

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
