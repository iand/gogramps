package gogramps

// Family represents a family unit in the Gramps database.
type Family struct {
	Class         string      `json:"_class"`
	Handle        string      `json:"handle"`
	GrampsID      string      `json:"gramps_id"`
	FatherHandle  *string     `json:"father_handle"`
	MotherHandle  *string     `json:"mother_handle"`
	ChildRefList  []ChildRef  `json:"child_ref_list"`
	Type          GrampsType  `json:"type"`
	EventRefList  []EventRef  `json:"event_ref_list"`
	MediaList     []MediaRef  `json:"media_list"`
	AttributeList []Attribute `json:"attribute_list"`
	LdsOrdList    []LdsOrd    `json:"lds_ord_list"`
	CitationList  []string    `json:"citation_list"`
	NoteList      []string    `json:"note_list"`
	Change        int64       `json:"change"`
	TagList       []string    `json:"tag_list"`
	Private       bool        `json:"private"`
	Complete      int         `json:"complete"`
}
