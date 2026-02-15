package gogramps

// Media represents a media object in the Gramps database.
type Media struct {
	Class         string      `json:"_class"`
	Handle        string      `json:"handle"`
	GrampsID      string      `json:"gramps_id"`
	Path          string      `json:"path"`
	Mime          string      `json:"mime"`
	Desc          string      `json:"desc"`
	Checksum      string      `json:"checksum"`
	Thumb         *string     `json:"thumb"`
	Date          *Date       `json:"date"`
	AttributeList []Attribute `json:"attribute_list"`
	CitationList  []string    `json:"citation_list"`
	NoteList      []string    `json:"note_list"`
	Change        int64       `json:"change"`
	TagList       []string    `json:"tag_list"`
	Private       bool        `json:"private"`
}
