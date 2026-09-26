package gogramps

// Citation confidence constants matching Gramps.
const (
	ConfidenceVeryLow  = 0
	ConfidenceLow      = 1
	ConfidenceNormal   = 2
	ConfidenceHigh     = 3
	ConfidenceVeryHigh = 4
)

// NoteType values used by the constructors.
const (
	NoteTypeGeneral = 1
)

// NewDate returns an empty date with every member set to the Gramps default.
func NewDate() *Date {
	return &Date{
		Class:   "Date",
		Dateval: []int{0, 0, 0, 0},
	}
}

// NewStyledText returns empty styled text with every member set to the Gramps
// default.
func NewStyledText() StyledText {
	return StyledText{
		Class: "StyledText",
		Tags:  []StyledTextTag{},
	}
}

// NewNote returns a general note with every member set to the Gramps default.
func NewNote() *Note {
	return &Note{
		Class:   "Note",
		Text:    NewStyledText(),
		Format:  NoteFlowed,
		Type:    GrampsType{Class: "NoteType", Value: NoteTypeGeneral},
		TagList: []string{},
	}
}

// NewSource returns a source with every member set to the Gramps default.
func NewSource() *Source {
	return &Source{
		Class:         "Source",
		NoteList:      []string{},
		MediaList:     []MediaRef{},
		AttributeList: []SrcAttribute{},
		RepoRefList:   []RepoRef{},
		TagList:       []string{},
	}
}

// NewCitation returns a citation with every member set to the Gramps default.
func NewCitation() *Citation {
	return &Citation{
		Class:         "Citation",
		Date:          NewDate(),
		Confidence:    ConfidenceNormal,
		NoteList:      []string{},
		MediaList:     []MediaRef{},
		AttributeList: []SrcAttribute{},
		TagList:       []string{},
	}
}
