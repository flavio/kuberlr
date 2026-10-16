package common

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandHome(t *testing.T) {
	home := filepath.Join("fake", "home")
	t.Setenv(HomeDirEnvKey(), home)

	tests := []struct {
		path     string
		expected string
	}{
		{"~", home},
		{"~/bins", filepath.Join(home, "bins")},
		{"~" + string(filepath.Separator) + "bins", filepath.Join(home, "bins")},
		{"~other/bins", "~other/bins"},
		{"/opt/~/bins", "/opt/~/bins"},
		{"", ""},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, ExpandHome(tt.path), tt.path)
	}
}
