package downloader

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha512"
	"testing"

	"github.com/blang/semver/v4"
)

func TestHashing(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		expectedSuffix string
		inputData      []byte
		expectedHash   []byte
	}{
		{
			name:           "sha1",
			version:        "1.11.0",
			expectedSuffix: ".sha1",
			inputData:      []byte("hello"),
			expectedHash:   sha1.New().Sum([]byte("hello")),
		},
		{
			name:           "sha512",
			version:        "1.12.0",
			expectedSuffix: ".sha512",
			inputData:      []byte("hello"),
			expectedHash:   sha512.New().Sum([]byte("hello")),
		},
	}

	for _, test := range tests {
		tableTest := test // ensure tt is correctly scoped when used in function literal
		t.Run(tableTest.name, func(t *testing.T) {
			version, err := semver.Parse(tableTest.version)
			if err != nil {
				t.Fatalf("failed to parse version %s: %v", tableTest.version, err)
			}

			hashing, err := NewHashing(version)
			if err != nil {
				t.Fatalf("failed to create hashing: %v", err)
			}

			if hashing.Suffix != tableTest.expectedSuffix {
				t.Errorf("expected suffix %s, got %s", tableTest.expectedSuffix, hashing.Suffix)
			}

			hash := hashing.Hasher.Sum(tableTest.inputData)
			if !bytes.Equal(hash, tableTest.expectedHash) {
				t.Errorf("expected hash %v, got %v", tableTest.expectedHash, hash)
			}
		})
	}
}
