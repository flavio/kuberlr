package downloader

import (
	"crypto/sha1" //nolint:gosec // sha1 is needed by old releases of kubectl
	"crypto/sha512"
	"hash"

	"github.com/blang/semver/v4"
)

// Hashing contains the hashing details for the downloader
type Hashing struct {
	// Suffix of the file containing the hash
	Suffix string

	// Hasher is the hash calculator to use
	Hasher hash.Hash
}

// NewHashing returns the hashing details for the downloader
//
//nolint:gosec // sha1 is needed by old releases of kubectl
func NewHashing(version semver.Version) (*Hashing, error) {
	// - sha1 is available in range [1.0.0, 1.18)
	// - sha256 is available from v1.16.0
	// - sha512 is available from 1.12.0

	rangeConstraint, parseErr := semver.ParseRange(">=1.12.0")
	if parseErr != nil {
		return nil, parseErr
	}
	if rangeConstraint(version) {
		return &Hashing{
			Suffix: ".sha512",
			Hasher: sha512.New(),
		}, nil
	}

	// we have to resort to sha1
	return &Hashing{
		Suffix: ".sha1",
		Hasher: sha1.New(),
	}, nil
}
