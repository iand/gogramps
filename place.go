package gogramps

// Place represents a place in the Gramps database.
type Place struct {
	Class        string      `json:"_class"`
	Handle       string      `json:"handle"`
	GrampsID     string      `json:"gramps_id"`
	Title        string      `json:"title"`
	Long         string      `json:"long"`
	Lat          string      `json:"lat"`
	PlaceRefList []PlaceRef  `json:"placeref_list"`
	Name         PlaceName   `json:"name"`
	AltNames     []PlaceName `json:"alt_names"`
	PlaceType    GrampsType  `json:"place_type"`
	Code         string      `json:"code"`
	AltLoc       []Location  `json:"alt_loc"`
	URLs         []URL       `json:"urls"`
	MediaList    []MediaRef  `json:"media_list"`
	CitationList []string    `json:"citation_list"`
	NoteList     []string    `json:"note_list"`
	Change       int64       `json:"change"`
	TagList      []string    `json:"tag_list"`
	Private      bool        `json:"private"`
}
