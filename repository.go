package gogramps

// Repository represents a repository in the Gramps database.
type Repository struct {
	Class       string     `json:"_class"`
	Handle      string     `json:"handle"`
	GrampsID    string     `json:"gramps_id"`
	Type        GrampsType `json:"type"`
	Name        string     `json:"name"`
	NoteList    []string   `json:"note_list"`
	AddressList []Address  `json:"address_list"`
	URLs        []URL      `json:"urls"`
	Change      int64      `json:"change"`
	TagList     []string   `json:"tag_list"`
	Private     bool       `json:"private"`
}
