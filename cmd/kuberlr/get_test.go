package main

import (
	"errors"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMinorVersionResolver struct {
	called      bool
	gotMajor    uint64
	gotMinor    uint64
	returnVer   semver.Version
	returnError error
}

func (f *fakeMinorVersionResolver) UpstreamStableVersionForMinor(major, minor uint64) (semver.Version, error) {
	f.called = true
	f.gotMajor = major
	f.gotMinor = minor
	return f.returnVer, f.returnError
}

func TestResolveRequestedVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string

		// behaviour of the fake resolver
		resolverVersion string
		resolverErr     error

		// expectations
		expectedVersion    string
		expectError        bool
		expectResolverCall bool
		expectedMajor      uint64
		expectedMinor      uint64
	}{
		{
			name:            "full version is used verbatim",
			arg:             "v1.19.1",
			expectedVersion: "1.19.1",
		},
		{
			name:            "pre-release version is used verbatim",
			arg:             "1.36.0-beta.1",
			expectedVersion: "1.36.0-beta.1",
		},
		{
			name:               "major.minor is resolved to latest patch",
			arg:                "1.36",
			resolverVersion:    "1.36.4",
			expectedVersion:    "1.36.4",
			expectResolverCall: true,
			expectedMajor:      1,
			expectedMinor:      36,
		},
		{
			name:               "major.minor falls back to patch 0 on resolver error",
			arg:                "1.99",
			resolverErr:        errors.New("boom"),
			expectedVersion:    "1.99.0",
			expectResolverCall: true,
			expectedMajor:      1,
			expectedMinor:      99,
		},
		{
			name:        "invalid input is rejected",
			arg:         "not-a-version",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resolver := &fakeMinorVersionResolver{returnError: tt.resolverErr}
			if tt.resolverVersion != "" {
				resolver.returnVer = semver.MustParse(tt.resolverVersion)
			}

			v, err := resolveRequestedVersion(tt.arg, resolver)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, semver.MustParse(tt.expectedVersion), v)
			}

			assert.Equal(t, tt.expectResolverCall, resolver.called)
			if tt.expectResolverCall {
				assert.Equal(t, tt.expectedMajor, resolver.gotMajor)
				assert.Equal(t, tt.expectedMinor, resolver.gotMinor)
			}
		})
	}
}
