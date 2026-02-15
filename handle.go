package gogramps

import (
	"crypto/rand"
	"encoding/hex"
)

// NewHandle generates a new unique handle string matching the Gramps format.
// Gramps handles are hex-encoded random bytes, typically 25-26 characters.
func NewHandle() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("gogramps: failed to generate random handle: " + err.Error())
	}
	return hex.EncodeToString(b)
}
