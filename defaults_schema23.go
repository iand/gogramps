//go:build gramps_dbschema23

package gogramps

// NewDNATest returns a DNA test with every member set to the Gramps default.
func NewDNATest() *DNATest {
	return &DNATest{
		Class:         "DNATest",
		Provider:      GrampsType{Class: "DNAProviderType", Value: DNAProviderUnknown},
		TestType:      GrampsType{Class: "DNATestType", Value: DNATestUnknown},
		GenomeBuild:   GrampsType{Class: "DNAGenomeBuildType", Value: DNAGenomeBuildUnknown},
		Date:          NewDate(),
		CitationList:  []string{},
		NoteList:      []string{},
		MediaList:     []MediaRef{},
		AttributeList: []DNAAttribute{},
		TagList:       []string{},
	}
}

// NewDNAMatch returns a DNA match with every member set to the Gramps default.
func NewDNAMatch() *DNAMatch {
	return &DNAMatch{
		Class:                     "DNAMatch",
		Provider:                  GrampsType{Class: "DNAProviderType", Value: DNAProviderUnknown},
		PredictedRelationshipList: []PredictedRelationship{},
		SharedAncestorList:        []SharedAncestor{},
		SegmentList:               []DNASegment{},
		CitationList:              []string{},
		NoteList:                  []string{},
		MediaList:                 []MediaRef{},
		AttributeList:             []DNAAttribute{},
		TagList:                   []string{},
	}
}

// NewDNAAttribute returns a DNA attribute with every member set to the Gramps
// default.
func NewDNAAttribute() DNAAttribute {
	return DNAAttribute{
		Class:        "DNAAttribute",
		Type:         GrampsType{Class: "DNAAttributeType", Value: DNAAttributeUnknown},
		CitationList: []string{},
		NoteList:     []string{},
	}
}

// NewDNASegment returns a DNA segment with every member set to the Gramps
// default.
func NewDNASegment() DNASegment {
	return DNASegment{
		Class:       "DNASegment",
		GenomeBuild: GrampsType{Class: "DNAGenomeBuildType", Value: DNAGenomeBuildUnknown},
	}
}

// NewPredictedRelationship returns a predicted relationship with every member
// set to the Gramps default.
func NewPredictedRelationship() PredictedRelationship {
	return PredictedRelationship{
		Class:        "PredictedRelationship",
		CitationList: []string{},
		NoteList:     []string{},
	}
}

// NewSharedAncestor returns a shared ancestor with every member set to the
// Gramps default.
func NewSharedAncestor() SharedAncestor {
	return SharedAncestor{
		Class:        "SharedAncestor",
		CitationList: []string{},
		NoteList:     []string{},
	}
}
