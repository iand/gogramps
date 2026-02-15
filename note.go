package gogramps

// Note represents a note in the Gramps database.
type Note struct {
	Class    string     `json:"_class"`
	Handle   string     `json:"handle"`
	GrampsID string     `json:"gramps_id"`
	Text     StyledText `json:"text"`
	Format   int        `json:"format"`
	Type     GrampsType `json:"type"`
	Change   int64      `json:"change"`
	TagList  []string   `json:"tag_list"`
	Private  bool       `json:"private"`
}

// Note format constants.
const (
	NoteFlowed    = 0
	NoteFormatted = 1
)
