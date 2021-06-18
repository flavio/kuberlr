package common_test

import (
	"github.com/flavio/kuberlr/internal/common"
	"testing"
)

func TestShaError(t *testing.T) {
	err := &common.ShaMismatchError{Url: "https://example.com/resource-1.2.3", ShaExpected: "abc", ShaActual: "def"}
	if !common.IsShaMismatch(err) {
		t.Errorf("Expected error %v to be a ShaMismatchError", err)
	}
	if err.Error() != "SHA mismatch for URL https://example.com/resource-1.2.3: expected 'abc', got 'def'" {
		t.Errorf("Expected error %v to have mismatch details ShaMismatchError", err)
	}
}
