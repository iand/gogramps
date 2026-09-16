package gogramps

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// ObjectType identifies a primary Gramps object type that carries a Gramps ID.
type ObjectType string

// The object types that carry a Gramps ID. Tags are excluded because they are
// keyed by name rather than by a Gramps ID.
const (
	ObjectPerson     ObjectType = "person"
	ObjectFamily     ObjectType = "family"
	ObjectEvent      ObjectType = "event"
	ObjectPlace      ObjectType = "place"
	ObjectSource     ObjectType = "source"
	ObjectCitation   ObjectType = "citation"
	ObjectRepository ObjectType = "repository"
	ObjectNote       ObjectType = "note"
	ObjectMedia      ObjectType = "media"
)

// defaultIDFormat maps each object type to the Gramps default ID format.
var defaultIDFormat = map[ObjectType]string{
	ObjectPerson:     "I%05d",
	ObjectFamily:     "F%05d",
	ObjectEvent:      "E%05d",
	ObjectPlace:      "P%05d",
	ObjectSource:     "S%05d",
	ObjectCitation:   "C%05d",
	ObjectRepository: "R%05d",
	ObjectNote:       "N%05d",
	ObjectMedia:      "O%05d",
}

// IDAllocator hands out consecutive Gramps IDs for a single object type.
type IDAllocator struct {
	format string // format holding one integer verb, such as "I%05d"
	next   int
}

// Next returns the next Gramps ID and advances the allocator.
func (a *IDAllocator) Next() string {
	id := fmt.Sprintf(a.format, a.next)
	a.next++
	return id
}

// GrampsIDAllocator returns an allocator for objType seeded from the current
// maximum Gramps ID in the database, using that type's default ID format such
// as "I%05d" for persons.
func (d *Database) GrampsIDAllocator(objType ObjectType) (*IDAllocator, error) {
	format, ok := defaultIDFormat[objType]
	if !ok {
		return nil, fmt.Errorf("gogramps: no default Gramps ID format for object type %q", objType)
	}
	return d.GrampsIDAllocatorWithFormat(objType, format)
}

// GrampsIDAllocatorWithFormat is like GrampsIDAllocator but uses a custom
// format, a string holding a single integer verb such as "I%04d".
func (d *Database) GrampsIDAllocatorWithFormat(objType ObjectType, format string) (*IDAllocator, error) {
	if _, ok := defaultIDFormat[objType]; !ok {
		return nil, fmt.Errorf("gogramps: object type %q has no Gramps ID", objType)
	}
	literal, err := idFormatLiteral(format)
	if err != nil {
		return nil, err
	}
	highest, err := d.maxGrampsIDIndex(string(objType), literal)
	if err != nil {
		return nil, err
	}
	next := 0
	if highest >= 0 {
		next = highest + 1
	}
	return &IDAllocator{format: format, next: next}, nil
}

// idFormatLiteral returns the fixed text that precedes the integer verb in a
// Gramps ID format, for example "I" for "I%05d". It rejects a format that does
// not format an integer into that fixed text followed by digits.
func idFormatLiteral(format string) (string, error) {
	literal, _, ok := strings.Cut(format, "%")
	if !ok {
		return "", fmt.Errorf("gogramps: Gramps ID format %q has no format verb", format)
	}
	tail, ok := strings.CutPrefix(fmt.Sprintf(format, 7), literal)
	if !ok || !allDigits(tail) {
		return "", fmt.Errorf("gogramps: Gramps ID format %q is not a valid integer format", format)
	}
	return literal, nil
}

// HasHandle reports whether the object type's table already contains handle. A
// handle is unique when this returns false, so a new object of that type can
// safely use it. The lookup is a single indexed primary-key probe.
func (d *Database) HasHandle(objType ObjectType, handle string) (bool, error) {
	if _, ok := defaultIDFormat[objType]; !ok {
		return false, fmt.Errorf("gogramps: unknown object type %q", objType)
	}
	var one int
	err := d.db.QueryRow(
		fmt.Sprintf("SELECT 1 FROM %s WHERE handle = ?", string(objType)),
		handle,
	).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// NewUniqueHandle returns a fresh handle that the object type's table does not
// already contain, regenerating until NewHandle yields an unused value.
func (d *Database) NewUniqueHandle(objType ObjectType) (string, error) {
	for {
		h := NewHandle()
		has, err := d.HasHandle(objType, h)
		if err != nil {
			return "", err
		}
		if !has {
			return h, nil
		}
	}
}

// maxGrampsIDIndex returns the highest integer index among the Gramps IDs in
// table that start with literal, or -1 when none match.
func (d *Database) maxGrampsIDIndex(table, literal string) (int, error) {
	rows, err := d.db.Query(fmt.Sprintf("SELECT gramps_id FROM %s", table))
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	highest := -1
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		tail, ok := strings.CutPrefix(id, literal)
		if !ok || !allDigits(tail) {
			continue
		}
		n, err := strconv.Atoi(tail)
		if err != nil {
			continue
		}
		if n > highest {
			highest = n
		}
	}
	return highest, rows.Err()
}

// allDigits reports whether s is non-empty and contains only ASCII digits.
func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
