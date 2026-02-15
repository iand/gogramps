package gogramps

// EventRef is a reference from a Person or Family to an Event.
type EventRef struct {
	Class         string      `json:"_class"`
	Private       bool        `json:"private"`
	CitationList  []string    `json:"citation_list"`
	NoteList      []string    `json:"note_list"`
	AttributeList []Attribute `json:"attribute_list"`
	Ref           string      `json:"ref"`
	Role          GrampsType  `json:"role"`
}

// MediaRef is a reference to a Media object, optionally cropped.
type MediaRef struct {
	Class         string      `json:"_class"`
	Private       bool        `json:"private"`
	CitationList  []string    `json:"citation_list"`
	NoteList      []string    `json:"note_list"`
	AttributeList []Attribute `json:"attribute_list"`
	Ref           string      `json:"ref"`
	Rect          []int       `json:"rect"`
}

// ChildRef is a reference from a Family to a child Person.
type ChildRef struct {
	Class        string     `json:"_class"`
	Private      bool       `json:"private"`
	CitationList []string   `json:"citation_list"`
	NoteList     []string   `json:"note_list"`
	Ref          string     `json:"ref"`
	Frel         GrampsType `json:"frel"`
	Mrel         GrampsType `json:"mrel"`
}

// PersonRef is a reference from one Person to another.
type PersonRef struct {
	Class        string   `json:"_class"`
	Private      bool     `json:"private"`
	CitationList []string `json:"citation_list"`
	NoteList     []string `json:"note_list"`
	Ref          string   `json:"ref"`
	Rel          string   `json:"rel"`
}

// RepoRef is a reference from a Source to a Repository.
type RepoRef struct {
	Class      string     `json:"_class"`
	Private    bool       `json:"private"`
	NoteList   []string   `json:"note_list"`
	Ref        string     `json:"ref"`
	CallNumber string     `json:"call_number"`
	MediaType  GrampsType `json:"media_type"`
}

// PlaceRef is a reference from one Place to another (hierarchy).
type PlaceRef struct {
	Class string `json:"_class"`
	Ref   string `json:"ref"`
	Date  *Date  `json:"date"`
}

// LdsOrd represents an LDS ordinance.
type LdsOrd struct {
	Class        string   `json:"_class"`
	Private      bool     `json:"private"`
	CitationList []string `json:"citation_list"`
	NoteList     []string `json:"note_list"`
	Date         *Date    `json:"date"`
	Type         int      `json:"type"`
	Place        string   `json:"place"`
	Famc         string   `json:"famc"`
	Temple       string   `json:"temple"`
	Status       int      `json:"status"`
}
