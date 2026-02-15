package gogramps

// Tag represents a tag in the Gramps database.
type Tag struct {
	Class    string `json:"_class"`
	Handle   string `json:"handle"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Priority int    `json:"priority"`
	Change   int64  `json:"change"`
}
