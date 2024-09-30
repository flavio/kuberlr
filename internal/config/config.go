package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/flavio/kuberlr/internal/common"
)

const DefaultTimeout = 5

func ThisExecutableDir() string {
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return ""
	}

	return filepath.Dir(execPath)
}

// Cfg is used to retrieve the configuration of kuberlr.
type Cfg struct {
	Paths []string
}

// NewCfg returns a new Cfg object that is pre-configured
// to look for the configuration files in the right set of
// directories.
func NewCfg() *Cfg {
	return &Cfg{
		Paths: configPaths,
	}
}

// Load reads the configuration files from disks and merges them.
func (c *Cfg) Load() (*viper.Viper, error) {
	v := viper.New()
	v.SetDefault("AllowDownload", true)
	v.SetDefault("SystemPath", common.SystemPath)
	v.SetDefault("Timeout", DefaultTimeout)
	v.SetDefault("KubeMirrorUrl", "https://dl.k8s.io")

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

// GetKubeMirrorURL returns the URL of the kubernetes mirror.
func (c *Cfg) GetKubeMirrorURL() (string, error) {
	v, err := c.Load()
	if err != nil {
		return "", err
	}

	return v.GetString("KubeMirrorUrl"), nil
}

func mergeConfig(v *viper.Viper, extraConfigPath string) error {
	if len(extraConfigPath) == 0 {
		return nil
	}
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
