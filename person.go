package gogramps

// Person represents a person in the Gramps database.
type Person struct {
	Class            string            `json:"_class"`
	Handle           string            `json:"handle"`
	GrampsID         string            `json:"gramps_id"`
	Gender           int               `json:"gender"`
	PrimaryName      Name              `json:"primary_name"`
	AlternateNames   []Name            `json:"alternate_names"`
	DeathRefIndex    int               `json:"death_ref_index"`
	BirthRefIndex    int               `json:"birth_ref_index"`
	EventRefList     []EventRef        `json:"event_ref_list"`
	FamilyList       []string          `json:"family_list"`
	ParentFamilyList []string          `json:"parent_family_list"`
	MediaList        []MediaRef        `json:"media_list"`
	AddressList      []Address         `json:"address_list"`
	AttributeList    []Attribute       `json:"attribute_list"`
	URLs             []URL             `json:"urls"`
	LdsOrdList       []LdsOrd          `json:"lds_ord_list"`
	CitationList     []string          `json:"citation_list"`
	NoteList         []string          `json:"note_list"`
	Change           int64             `json:"change"`
	TagList          []string          `json:"tag_list"`
	Private          bool              `json:"private"`
	PersonRefList    []PersonRef       `json:"person_ref_list"`
	FamilySearchSync *FamilySearchSync `json:"familysearch_sync,omitempty"`
}

// FamilySearchSync holds FamilySearch synchronisation state for a person.
// Added in schema version 22. Absent (nil) in schema 21 databases.
type FamilySearchSync struct {
	Class             string  `json:"_class"`
	FSID              *string `json:"fsid"`
	IsRoot            bool    `json:"is_root"`
	StatusTS          *int64  `json:"status_ts"`
	ConfirmedTS       *int64  `json:"confirmed_ts"`
	GrampsModifiedTS  *int64  `json:"gramps_modified_ts"`
	FSModifiedTS      *int64  `json:"fs_modified_ts"`
	EssentialConflict bool    `json:"essential_conflict"`
	Conflict          bool    `json:"conflict"`
}

// Gender constants matching Gramps.
const (
	GenderFemale  = 0
	GenderMale    = 1
	GenderUnknown = 2
	GenderOther   = 3
)
