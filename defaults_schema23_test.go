//go:build gramps_dbschema23

package gogramps

// init adds the DNA constructors to TestConstructorDefaults.
func init() {
	defaultCases = append(defaultCases,
		defaultCase{"DNATest", func() any { return NewDNATest() }},
		defaultCase{"DNAMatch", func() any { return NewDNAMatch() }},
		defaultCase{"DNAAttribute", func() any { return NewDNAAttribute() }},
		defaultCase{"DNASegment", func() any { return NewDNASegment() }},
		defaultCase{"PredictedRelationship", func() any { return NewPredictedRelationship() }},
		defaultCase{"SharedAncestor", func() any { return NewSharedAncestor() }},
	)
}
