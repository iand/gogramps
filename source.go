package gogramps

// Source represents a source of information in the Gramps database.
type Source struct {
	Class         string         `json:"_class"`
	Handle        string         `json:"handle"`
	GrampsID      string         `json:"gramps_id"`
	Title         string         `json:"title"`
	Author        string         `json:"author"`
	Pubinfo       string         `json:"pubinfo"`
	Abbrev        string         `json:"abbrev"`
	NoteList      []string       `json:"note_list"`
	MediaList     []MediaRef     `json:"media_list"`
	AttributeList []SrcAttribute `json:"attribute_list"`
	RepoRefList   []RepoRef      `json:"reporef_list"`
	Change        int64          `json:"change"`
	TagList       []string       `json:"tag_list"`
	Private       bool           `json:"private"`
}
