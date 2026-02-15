package gogramps

import (
	"encoding/json"
	"fmt"
)

// GrampsType represents a Gramps enumerated type (e.g., EventType, NameType).
// In JSON it serializes as {"_class": "EventType", "value": 12, "string": ""}.
// The value is the integer type code, and string is only non-empty for custom types.
type GrampsType struct {
	Class  string `json:"_class"`
	Value  int    `json:"value"`
	String string `json:"string"`
}

// Date represents a Gramps date object.
// dateval is stored as an array of integers representing day/month/year values.
// For simple dates: [day, month, year, slash_flag]
// For compound dates (range/span): [day1, month1, year1, slash1, day2, month2, year2, slash2]
type Date struct {
	Class    string `json:"_class"`
	Calendar int    `json:"calendar"`
	Modifier int    `json:"modifier"`
	Quality  int    `json:"quality"`
	Dateval  []int  `json:"dateval"`
	Text     string `json:"text"`
	Sortval  int    `json:"sortval"`
	Newyear  int    `json:"newyear"`
	Format   *int   `json:"format"` // typically nil
}

// UnmarshalJSON implements custom JSON unmarshalling for Date.
// Gramps Python serializes dateval as e.g. [29, 8, 1987, false] where the
// 4th (and 8th for compound dates) element is a Python bool (JSON true/false).
// This method converts booleans to int (false→0, true→1) so that Dateval
// can remain []int.
func (d *Date) UnmarshalJSON(data []byte) error {
	type dateAlias Date
	var raw struct {
		dateAlias
		RawDateval []json.RawMessage `json:"dateval"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*d = Date(raw.dateAlias)
	d.Dateval = make([]int, len(raw.RawDateval))
	for i, elem := range raw.RawDateval {
		var f float64
		if err := json.Unmarshal(elem, &f); err == nil {
			d.Dateval[i] = int(f)
			continue
		}
		var b bool
		if err := json.Unmarshal(elem, &b); err == nil {
			if b {
				d.Dateval[i] = 1
			} else {
				d.Dateval[i] = 0
			}
			continue
		}
		return fmt.Errorf("dateval[%d]: unsupported type: %s", i, string(elem))
	}
	return nil
}

// StyledTextTag represents a formatting tag within styled text.
type StyledTextTag struct {
	Class  string     `json:"_class"`
	Name   GrampsType `json:"name"`
	Value  string     `json:"value"`
	Ranges [][]int    `json:"ranges"`
}

// UnmarshalJSON implements custom JSON unmarshalling for StyledTextTag.
// Gramps Python serializes the Value field as either a string or an integer
// (e.g., font size). This method accepts both.
func (s *StyledTextTag) UnmarshalJSON(data []byte) error {
	type alias StyledTextTag
	var raw struct {
		alias
		RawValue json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*s = StyledTextTag(raw.alias)
	if len(raw.RawValue) > 0 {
		// Try string first.
		var str string
		if err := json.Unmarshal(raw.RawValue, &str); err == nil {
			s.Value = str
		} else {
			// Try number, convert to string.
			var num json.Number
			if err := json.Unmarshal(raw.RawValue, &num); err == nil {
				s.Value = num.String()
			}
		}
	}
	return nil
}

// StyledText represents formatted text with optional markup tags.
type StyledText struct {
	Class  string          `json:"_class"`
	String string          `json:"string"`
	Tags   []StyledTextTag `json:"tags"`
}

// Surname represents one surname component of a Name.
type Surname struct {
	Class      string     `json:"_class"`
	Surname    string     `json:"surname"`
	Prefix     string     `json:"prefix"`
	Primary    bool       `json:"primary"`
	Origintype GrampsType `json:"origintype"`
	Connector  string     `json:"connector"`
}

// Name represents a person's name.
type Name struct {
	Class        string     `json:"_class"`
	Private      bool       `json:"private"`
	CitationList []string   `json:"citation_list"`
	NoteList     []string   `json:"note_list"`
	Date         *Date      `json:"date"`
	FirstName    string     `json:"first_name"`
	SurnameList  []Surname  `json:"surname_list"`
	Suffix       string     `json:"suffix"`
	Title        string     `json:"title"`
	Type         GrampsType `json:"type"`
	GroupAs      string     `json:"group_as"`
	SortAs       int        `json:"sort_as"`
	DisplayAs    int        `json:"display_as"`
	Call         string     `json:"call"`
	Nick         string     `json:"nick"`
	Famnick      string     `json:"famnick"`
}

// Attribute represents a person/family/event attribute.
type Attribute struct {
	Class        string     `json:"_class"`
	Private      bool       `json:"private"`
	CitationList []string   `json:"citation_list"`
	NoteList     []string   `json:"note_list"`
	Type         GrampsType `json:"type"`
	Value        string     `json:"value"`
}

// SrcAttribute represents a source attribute (different from Attribute).
type SrcAttribute struct {
	Class   string     `json:"_class"`
	Private bool       `json:"private"`
	Type    GrampsType `json:"type"`
	Value   string     `json:"value"`
}

// Address represents a physical address.
type Address struct {
	Class        string   `json:"_class"`
	Private      bool     `json:"private"`
	CitationList []string `json:"citation_list"`
	NoteList     []string `json:"note_list"`
	Date         *Date    `json:"date"`
	Street       string   `json:"street"`
	Locality     string   `json:"locality"`
	City         string   `json:"city"`
	County       string   `json:"county"`
	State        string   `json:"state"`
	Country      string   `json:"country"`
	Postal       string   `json:"postal"`
	Phone        string   `json:"phone"`
}

// URL represents a web URL or other URI.
type URL struct {
	Class   string     `json:"_class"`
	Private bool       `json:"private"`
	Path    string     `json:"path"`
	Desc    string     `json:"desc"`
	Type    GrampsType `json:"type"`
}

// Location represents an alternate location for a place.
type Location struct {
	Class    string `json:"_class"`
	Street   string `json:"street"`
	Locality string `json:"locality"`
	City     string `json:"city"`
	County   string `json:"county"`
	State    string `json:"state"`
	Country  string `json:"country"`
	Postal   string `json:"postal"`
	Phone    string `json:"phone"`
	Parish   string `json:"parish"`
}

// PlaceName represents a name of a place, possibly with a date range.
type PlaceName struct {
	Class string `json:"_class"`
	Value string `json:"value"`
	Date  *Date  `json:"date"`
	Lang  string `json:"lang"`
}
