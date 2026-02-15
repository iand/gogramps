package gogramps

// Citation represents a citation of a source in the Gramps database.
type Citation struct {
	Class         string         `json:"_class"`
	Handle        string         `json:"handle"`
	GrampsID      string         `json:"gramps_id"`
	Date          *Date          `json:"date"`
	Page          string         `json:"page"`
	Confidence    int            `json:"confidence"`
	SourceHandle  *string        `json:"source_handle"`
	NoteList      []string       `json:"note_list"`
	MediaList     []MediaRef     `json:"media_list"`
	AttributeList []SrcAttribute `json:"attribute_list"`
	Change        int64          `json:"change"`
	TagList       []string       `json:"tag_list"`
	Private       bool           `json:"private"`
}
