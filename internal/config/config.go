package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/flavio/kuberlr/internal/common"
)

type Cfg struct {
	Paths []string
}

func NewCfg() *Cfg {
	return &Cfg{
		Paths: []string{
			"/usr/etc/",
			"/etc/",
			filepath.Join(common.HomeDir(), ".kuberlr"),
		},
	}
}

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
