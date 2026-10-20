package common

import (
	"fmt"
	"path/filepath"
	"runtime"
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
		// the Windows form, identical to the previous case on Unix
		{"~" + string(filepath.Separator) + "bins", filepath.Join(home, "bins")},
		{"~other/bins", "~other/bins"},
		{"/opt/~/bins", "/opt/~/bins"},
		{"", ""},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, ExpandHome(tt.path), tt.path)
	}
}

func TestLocalDownloadDir(t *testing.T) {
	home := filepath.Join("fake", "home")
	t.Setenv(HomeDirEnvKey(), home)

	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	tests := []struct {
		name     string
		root     string
		expected string
	}{
		{
			name:     "empty root uses the default directory",
			root:     "",
			expected: filepath.Join(home, ".kuberlr", platform),
		},
		{
			name:     "custom root",
			root:     filepath.Join("opt", "kubectl"),
			expected: filepath.Join("opt", "kubectl", platform),
		},
		{
			name:     "tilde is not expanded, the caller does that",
			root:     "~/bins",
			expected: filepath.Join("~", "bins", platform),
		},
	}
	for _, tt := range tests { //nolint:paralleltest // uses t.Setenv
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, LocalDownloadDir(tt.root))
		})
	}
}
