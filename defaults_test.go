package gogramps

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// defaultCase pairs a fixture name with a constructor whose output is
// compared with it.
type defaultCase struct {
	name string
	make func() any
}

// defaultCases lists the constructors checked by TestConstructorDefaults.
var defaultCases = []defaultCase{
	{"Date", func() any { return NewDate() }},
	{"StyledText", func() any { return NewStyledText() }},
	{"Note", func() any { return NewNote() }},
	{"Source", func() any { return NewSource() }},
	{"Citation", func() any { return NewCitation() }},
}

// TestConstructorDefaults compares each constructor's encoding with the dict
// of the default instance of the same Gramps class. The fixtures in
// testdata/defaults are object_to_dict output from dna-core commit
// b7aa0f5291; their null handle and gramps_id are replaced by the empty
// strings a Go string field encodes.
func TestConstructorDefaults(t *testing.T) {
	for _, tc := range defaultCases {
		t.Run(tc.name, func(t *testing.T) {
			want := loadDefaultFixture(t, tc.name)
			got, err := json.Marshal(tc.make())
			if err != nil {
				t.Fatalf("Marshal: unexpected error: %v", err)
			}
			path, diff, err := jsonDiff(want, got)
			if err != nil {
				t.Fatalf("compare: unexpected error: %v", err)
			}
			if diff != "" {
				t.Errorf("differs at %s: %s", path, diff)
			}
		})
	}
}

// loadDefaultFixture reads a default-object fixture, replacing a null handle
// or gramps_id with an empty string.
func loadDefaultFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "defaults", name+".json"))
	if err != nil {
		t.Fatalf("reading fixture: unexpected error: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("decoding fixture: unexpected error: %v", err)
	}
	for _, k := range []string{"handle", "gramps_id"} {
		if v, ok := m[k]; ok && v == nil {
			m[k] = ""
		}
	}
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("encoding fixture: unexpected error: %v", err)
	}
	return out
}

func TestDateMarshalSlash(t *testing.T) {
	tests := []struct {
		name string
		in   string // in is the stored JSON, or empty to encode a Date literal
		date Date
		want string
	}{
		{
			name: "literal writes booleans",
			date: Date{Class: "Date", Dateval: []int{30, 3, 1851, 0, 1, 4, 1852, 1}},
			want: `[30,3,1851,false,1,4,1852,true]`,
		},
		{
			name: "decoded booleans kept",
			in:   `{"_class":"Date","dateval":[30,3,1851,true]}`,
			want: `[30,3,1851,true]`,
		},
		{
			name: "decoded numbers kept",
			in:   `{"_class":"Date","dateval":[0,0,0,0]}`,
			want: `[0,0,0,0]`,
		},
		{
			name: "nil dateval written as null",
			date: Date{Class: "Date"},
			want: `null`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := tc.date
			if tc.in != "" {
				if err := json.Unmarshal([]byte(tc.in), &d); err != nil {
					t.Fatalf("Unmarshal: unexpected error: %v", err)
				}
			}
			data, err := json.Marshal(&d)
			if err != nil {
				t.Fatalf("Marshal: unexpected error: %v", err)
			}
			var got struct {
				Dateval json.RawMessage `json:"dateval"`
			}
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("decoding output: unexpected error: %v", err)
			}
			if string(got.Dateval) != tc.want {
				t.Errorf("dateval = %s, want %s", got.Dateval, tc.want)
			}
		})
	}
}

func TestStyledTextTagValue(t *testing.T) {
	tests := []struct {
		name  string
		in    string // in is the stored JSON, or empty to encode a literal
		tag   StyledTextTag
		value string // value is the new Value set before encoding, when non-empty
		want  string
	}{
		{name: "literal string", tag: StyledTextTag{Value: "bold"}, want: `"bold"`},
		{name: "literal empty", tag: StyledTextTag{}, want: `""`},
		{name: "decoded string", in: `{"value":"x"}`, want: `"x"`},
		{name: "decoded null", in: `{"value":null}`, want: `null`},
		{name: "decoded number", in: `{"value":12}`, want: `12`},
		{name: "decoded null then set", in: `{"value":null}`, value: "x", want: `"x"`},
		{name: "decoded number then set to text", in: `{"value":12}`, value: "big", want: `"big"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tag := tc.tag
			if tc.in != "" {
				if err := json.Unmarshal([]byte(tc.in), &tag); err != nil {
					t.Fatalf("Unmarshal: unexpected error: %v", err)
				}
			}
			if tc.value != "" {
				tag.Value = tc.value
			}
			data, err := json.Marshal(&tag)
			if err != nil {
				t.Fatalf("Marshal: unexpected error: %v", err)
			}
			var got struct {
				Value json.RawMessage `json:"value"`
			}
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("decoding output: unexpected error: %v", err)
			}
			if string(got.Value) != tc.want {
				t.Errorf("value = %s, want %s", got.Value, tc.want)
			}
		})
	}
}

func TestGetJSON(t *testing.T) {
	db, err := OpenReadOnly("testdata/schema21")
	if err != nil {
		t.Fatalf("OpenReadOnly: unexpected error: %v", err)
	}
	defer db.Close()

	const handle = "a5af0eb667015e355db"
	var want string
	if err := db.db.QueryRow("SELECT json_data FROM event WHERE handle = ?", handle).Scan(&want); err != nil {
		t.Fatalf("query: unexpected error: %v", err)
	}

	got, err := db.GetJSON(ObjectEvent, handle)
	if err != nil {
		t.Fatalf("GetJSON: unexpected error: %v", err)
	}
	if string(got) != want {
		t.Errorf("GetJSON = %s, want %s", got, want)
	}

	got, err = db.GetJSON(ObjectEvent, "no-such-handle")
	if err != nil {
		t.Fatalf("GetJSON missing: unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("GetJSON missing = %s, want nil", got)
	}

	if _, err := db.GetJSON(ObjectType("nosuchtable"), handle); err == nil {
		t.Errorf("GetJSON unknown type: expected error, got nil")
	}
}
