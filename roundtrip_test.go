package gogramps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"testing"
)

// roundTripTable pairs a table name with a function that decodes and
// re-encodes one json_data value through the table's Go type.
type roundTripTable struct {
	table  string
	encode func([]byte) ([]byte, error)
}

// reencode decodes data into a value of type T and encodes it again.
func reencode[T any](data []byte) ([]byte, error) {
	var obj T
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}
	return json.Marshal(&obj)
}

// roundTripTables lists the tables checked by TestRoundTrip.
var roundTripTables = []roundTripTable{
	{"person", reencode[Person]},
	{"family", reencode[Family]},
	{"event", reencode[Event]},
	{"place", reencode[Place]},
	{"source", reencode[Source]},
	{"citation", reencode[Citation]},
	{"repository", reencode[Repository]},
	{"media", reencode[Media]},
	{"note", reencode[Note]},
	{"tag", reencode[Tag]},
}

// TestRoundTrip checks that every stored object decodes and re-encodes to
// JSON equal to its json_data. It runs against the schema 21 test data, and
// also against the database directory named by GOGRAMPS_ROUNDTRIP_DB.
func TestRoundTrip(t *testing.T) {
	dirs := []string{"testdata/schema21"}
	if dir := os.Getenv("GOGRAMPS_ROUNDTRIP_DB"); dir != "" {
		dirs = append(dirs, dir)
	}
	for _, dir := range dirs {
		t.Run(dir, func(t *testing.T) {
			db, err := OpenReadOnly(dir)
			if err != nil {
				t.Fatalf("OpenReadOnly: unexpected error: %v", err)
			}
			defer db.Close()
			for _, rt := range roundTripTables {
				t.Run(rt.table, func(t *testing.T) {
					checkRoundTrip(t, db, rt)
				})
			}
		})
	}
}

// checkRoundTrip reports each object of one table whose re-encoded JSON
// differs from its json_data, up to a limit.
func checkRoundTrip(t *testing.T, db *Database, rt roundTripTable) {
	t.Helper()
	var tables int
	if err := db.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", rt.table).Scan(&tables); err != nil {
		t.Fatalf("checking table %s: unexpected error: %v", rt.table, err)
	}
	if tables == 0 {
		t.Skipf("database has no %s table", rt.table)
	}
	rows, err := db.db.Query(fmt.Sprintf("SELECT handle, json_data FROM %s", rt.table))
	if err != nil {
		t.Fatalf("query %s: unexpected error: %v", rt.table, err)
	}
	defer rows.Close()

	const maxReports = 5
	count, failed := 0, 0
	for rows.Next() {
		var handle, data string
		if err := rows.Scan(&handle, &data); err != nil {
			t.Fatalf("scan: unexpected error: %v", err)
		}
		count++
		got, err := rt.encode([]byte(data))
		if err != nil {
			t.Fatalf("%s %s: re-encode: unexpected error: %v", rt.table, handle, err)
		}
		path, diff, err := jsonDiff([]byte(data), got)
		if err != nil {
			t.Fatalf("%s %s: compare: unexpected error: %v", rt.table, handle, err)
		}
		if path == "" && diff == "" {
			continue
		}
		failed++
		if failed <= maxReports {
			t.Errorf("%s %s: differs at %s: %s", rt.table, handle, path, diff)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: unexpected error: %v", err)
	}
	if failed > maxReports {
		t.Errorf("%s: %d of %d objects differ", rt.table, failed, count)
	}
}

// jsonDiff compares two JSON documents as RFC 6902 section 4.6 defines
// equality, and returns the JSON Pointer and a description of the first
// difference, or empty strings when the documents are equal.
func jsonDiff(a, b []byte) (string, string, error) {
	va, err := decodeJSON(a)
	if err != nil {
		return "", "", err
	}
	vb, err := decodeJSON(b)
	if err != nil {
		return "", "", err
	}
	path, diff := diffValues("", va, vb)
	return path, diff, nil
}

// decodeJSON decodes data with numbers kept as json.Number.
func decodeJSON(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}

// diffValues returns the path and description of the first difference
// between a and b.
func diffValues(path string, a, b any) (string, string) {
	switch av := a.(type) {
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok {
			return path, fmt.Sprintf("stored %s, encoded %s", describe(a), describe(b))
		}
		keys := make([]string, 0, len(av)+len(bv))
		for k := range av {
			keys = append(keys, k)
		}
		for k := range bv {
			if _, ok := av[k]; !ok {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		for _, k := range keys {
			ae, aok := av[k]
			be, bok := bv[k]
			switch {
			case !aok:
				return path + "/" + k, "member absent from stored"
			case !bok:
				return path + "/" + k, "member absent from encoded"
			}
			if p, d := diffValues(path+"/"+k, ae, be); d != "" {
				return p, d
			}
		}
		return "", ""
	case []any:
		bv, ok := b.([]any)
		if !ok {
			return path, fmt.Sprintf("stored %s, encoded %s", describe(a), describe(b))
		}
		if len(av) != len(bv) {
			return path, fmt.Sprintf("stored length %d, encoded length %d", len(av), len(bv))
		}
		for i := range av {
			if p, d := diffValues(path+"/"+strconv.Itoa(i), av[i], bv[i]); d != "" {
				return p, d
			}
		}
		return "", ""
	case json.Number:
		bv, ok := b.(json.Number)
		if !ok {
			return path, fmt.Sprintf("stored %s, encoded %s", describe(a), describe(b))
		}
		af, aerr := av.Float64()
		bf, berr := bv.Float64()
		if aerr != nil || berr != nil || af != bf {
			return path, fmt.Sprintf("stored %s, encoded %s", av, bv)
		}
		return "", ""
	default:
		if a != b {
			return path, fmt.Sprintf("stored %s, encoded %s", describe(a), describe(b))
		}
		return "", ""
	}
}

// describe returns a short form of a decoded JSON value for messages.
func describe(v any) string {
	switch v := v.(type) {
	case nil:
		return "null"
	case map[string]any:
		return "object"
	case []any:
		return "array"
	case string:
		return strconv.Quote(v)
	default:
		return fmt.Sprint(v)
	}
}
