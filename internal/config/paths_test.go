package config

import (
	"path/filepath"
	"testing"
)

func TestXDGConfigFile(t *testing.T) {
	xdg := t.TempDir()

	tests := []struct {
		name      string
		xdgConfig string
		home      string
		want      string
	}{
		{
			name:      "XDG_CONFIG_HOME set",
			xdgConfig: xdg,
			home:      "/home/user",
			want:      filepath.Join(xdg, "kuberlr", "kuberlr.conf"),
		},
		{
			name: "XDG_CONFIG_HOME unset",
			home: "/home/user",
			want: filepath.Join("/home/user", ".config", "kuberlr", "kuberlr.conf"),
		},
		{
			name:      "XDG_CONFIG_HOME relative",
			xdgConfig: "relative",
			home:      "/home/user",
			want:      filepath.Join("/home/user", ".config", "kuberlr", "kuberlr.conf"),
		},
		{
			name: "no XDG_CONFIG_HOME and no HOME",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", tt.xdgConfig)
			t.Setenv("HOME", tt.home)

			if got := xdgConfigFile(); got != tt.want {
				t.Errorf("got %q instead of %q", got, tt.want)
			}
		})
	}
}
