package downloader

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpstreamVersionLookup(t *testing.T) {
	tests := []struct {
		name string

		// fake mirror behaviour
		status int
		body   string

		// method under test
		call func(*Downloder) (semver.Version, error)

		// expectations
		expectedPath    string
		expectedVersion string
		expectError     bool
	}{
		{
			name:   "stable version",
			status: http.StatusOK,
			body:   "v1.30.4\n",
			call: func(d *Downloder) (semver.Version, error) {
				return d.UpstreamStableVersion()
			},
			expectedPath:    "/release/stable.txt",
			expectedVersion: "1.30.4",
		},
		{
			name:   "stable version for minor",
			status: http.StatusOK,
			body:   "v1.30.7\n",
			call: func(d *Downloder) (semver.Version, error) {
				return d.UpstreamStableVersionForMinor(1, 30)
			},
			expectedPath:    "/release/stable-1.30.txt",
			expectedVersion: "1.30.7",
		},
		{
			name:   "unknown minor returns error",
			status: http.StatusNotFound,
			call: func(d *Downloder) (semver.Version, error) {
				return d.UpstreamStableVersionForMinor(1, 999)
			},
			expectedPath: "/release/stable-1.999.txt",
			expectError:  true,
		},
		{
			name:   "unparsable body returns error",
			status: http.StatusOK,
			body:   "not-a-version",
			call: func(d *Downloder) (semver.Version, error) {
				return d.UpstreamStableVersionForMinor(1, 30)
			},
			expectedPath: "/release/stable-1.30.txt",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The mirror URL is configured through an environment variable,
			// which is process-global state, so this test can't run in
			// parallel with its siblings.
			var requestedPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestedPath = r.URL.Path
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			t.Cleanup(srv.Close)
			t.Setenv("KUBERLR_KUBEMIRRORURL", srv.URL)

			v, err := tt.call(&Downloder{})

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, semver.MustParse(tt.expectedVersion), v)
			}
			assert.Equal(t, tt.expectedPath, requestedPath)
		})
	}
}
