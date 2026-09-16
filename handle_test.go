package gogramps

import (
	"strings"
	"testing"
)

func TestNewHandle(t *testing.T) {
	const hexDigits = "0123456789abcdef"

	h := NewHandle()

	// The handle is two hex-encoded fields, each at least 8 digits wide.
	if len(h) < 16 {
		t.Errorf("NewHandle() = %q, want at least 16 characters", h)
	}
	for i, r := range h {
		if !strings.ContainsRune(hexDigits, r) {
			t.Errorf("NewHandle() = %q has non-hex byte %q at %d", h, r, i)
		}
	}
}

func TestNewHandleUnique(t *testing.T) {
	seen := make(map[string]bool)
	for i := range 10000 {
		h := NewHandle()
		if seen[h] {
			t.Fatalf("NewHandle() returned duplicate %q after %d calls", h, i)
		}
		seen[h] = true
	}
}
