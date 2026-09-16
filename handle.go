package gogramps

import (
	"fmt"
	"math/rand/v2"
	"time"
)

// NewHandle returns a fresh object handle in the database's storage form,
// following the algorithm Gramps uses: the current time in tenths of a
// millisecond and a random integer, each hex-encoded and concatenated.
func NewHandle() string {
	ticks := time.Now().UnixNano() / 100_000 // tenths of a millisecond since the Unix epoch
	return fmt.Sprintf("%08x%08x", ticks, rand.Int64())
}
