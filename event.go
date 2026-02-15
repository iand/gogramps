package gogramps

// Event represents an event in the Gramps database.
type Event struct {
	Class         string      `json:"_class"`
	Handle        string      `json:"handle"`
	GrampsID      string      `json:"gramps_id"`
	Type          GrampsType  `json:"type"`
	Date          *Date       `json:"date"`
	Description   string      `json:"description"`
	Place         string      `json:"place"`
	CitationList  []string    `json:"citation_list"`
	NoteList      []string    `json:"note_list"`
	MediaList     []MediaRef  `json:"media_list"`
	AttributeList []Attribute `json:"attribute_list"`
	Change        int64       `json:"change"`
	TagList       []string    `json:"tag_list"`
	Private       bool        `json:"private"`
}
