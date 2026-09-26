//go:build gramps_dbschema23

package gogramps

// init adds the DNA tables to the round-trip check.
func init() {
	roundTripTables = append(roundTripTables,
		roundTripTable{"dnatest", reencode[DNATest]},
		roundTripTable{"dnamatch", reencode[DNAMatch]},
	)
}
