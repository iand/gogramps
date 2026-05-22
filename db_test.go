package gogramps

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAndOpen(t *testing.T) {
	dir := t.TempDir()
	dbDir := filepath.Join(dir, "testdb")

	// Create a new database.
	db, err := Create(dbDir, "Test DB")
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}

	// Verify directory files.
	backend, err := os.ReadFile(filepath.Join(dbDir, backendFile))
	if err != nil {
		t.Fatalf("reading backend file: unexpected error: %v", err)
	}
	if string(backend) != backendString {
		t.Errorf("backend = %q, want %q", string(backend), backendString)
	}

	name, err := os.ReadFile(filepath.Join(dbDir, nameFile))
	if err != nil {
		t.Fatalf("reading name file: unexpected error: %v", err)
	}
	if string(name) != "Test DB" {
		t.Errorf("name = %q, want %q", string(name), "Test DB")
	}

	// Verify lock exists.
	_, err = os.Stat(filepath.Join(dbDir, lockFileName))
	if err != nil {
		t.Errorf("lock file should exist: %v", err)
	}

	// Close the database.
	if err := db.Close(); err != nil {
		t.Fatalf("Close: unexpected error: %v", err)
	}

	// Verify lock removed.
	_, err = os.Stat(filepath.Join(dbDir, lockFileName))
	if !os.IsNotExist(err) {
		t.Errorf("lock file should be removed after close")
	}

	// Re-open.
	db, err = Open(dbDir)
	if err != nil {
		t.Fatalf("Open: unexpected error: %v", err)
	}
	defer db.Close()
}

func TestOpenReadOnly(t *testing.T) {
	dir := t.TempDir()
	dbDir := filepath.Join(dir, "testdb")

	db, err := Create(dbDir, "Test")
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}
	db.Close()

	// OpenReadOnly should work even without lock concerns.
	ro, err := OpenReadOnly(dbDir)
	if err != nil {
		t.Fatalf("OpenReadOnly: unexpected error: %v", err)
	}
	defer ro.Close()

	// Writes should fail with ErrReadOnly.
	err = ro.AddPerson(&Person{Handle: NewHandle(), GrampsID: "I0001"})
	if !errors.Is(err, ErrReadOnly) {
		t.Errorf("AddPerson: got %v, want ErrReadOnly", err)
	}
	err = ro.UpdatePerson(&Person{Handle: NewHandle(), GrampsID: "I0001"})
	if !errors.Is(err, ErrReadOnly) {
		t.Errorf("UpdatePerson: got %v, want ErrReadOnly", err)
	}
	err = ro.DeletePerson(NewHandle())
	if !errors.Is(err, ErrReadOnly) {
		t.Errorf("DeletePerson: got %v, want ErrReadOnly", err)
	}
	err = ro.SetMetadata("key", "value")
	if !errors.Is(err, ErrReadOnly) {
		t.Errorf("SetMetadata: got %v, want ErrReadOnly", err)
	}
}

func TestLockContention(t *testing.T) {
	dir := t.TempDir()
	dbDir := filepath.Join(dir, "testdb")

	db, err := Create(dbDir, "Test")
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}
	defer db.Close()

	// Opening again should fail with ErrLocked.
	_, err = Open(dbDir)
	if err == nil {
		t.Fatal("expected lock error")
	}
	var lockErr *ErrLocked
	if !errors.As(err, &lockErr) {
		t.Errorf("expected ErrLocked, got %T: %v", err, err)
	}
}

func TestPersonCRUD(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	handle := NewHandle()
	p := &Person{
		Handle:   handle,
		GrampsID: "I0001",
		Gender:   GenderMale,
		PrimaryName: Name{
			Class:     "Name",
			FirstName: "John",
			SurnameList: []Surname{
				{Class: "Surname", Surname: "Doe", Primary: true, Origintype: GrampsType{Class: "NameOriginType"}},
			},
			Type: GrampsType{Class: "NameType"},
		},
		Change: 1700000000,
	}

	// Add
	if err := db.AddPerson(p); err != nil {
		t.Fatalf("AddPerson: unexpected error: %v", err)
	}

	// Get by handle
	got, err := db.GetPerson(handle)
	if err != nil {
		t.Fatalf("GetPerson: unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("GetPerson returned nil")
	}
	if got.GrampsID != "I0001" {
		t.Errorf("GrampsID = %q, want %q", got.GrampsID, "I0001")
	}
	if got.Gender != GenderMale {
		t.Errorf("Gender = %d, want %d", got.Gender, GenderMale)
	}
	if got.PrimaryName.FirstName != "John" {
		t.Errorf("FirstName = %q, want %q", got.PrimaryName.FirstName, "John")
	}

	// Get by GrampsID
	got, err = db.GetPersonByGrampsID("I0001")
	if err != nil {
		t.Fatalf("GetPersonByGrampsID: unexpected error: %v", err)
	}
	if got == nil || got.Handle != handle {
		t.Errorf("GetPersonByGrampsID: got wrong person")
	}

	// Update
	p.PrimaryName.FirstName = "Jane"
	p.Gender = GenderFemale
	if err := db.UpdatePerson(p); err != nil {
		t.Fatalf("UpdatePerson: unexpected error: %v", err)
	}
	got, err = db.GetPerson(handle)
	if err != nil {
		t.Fatalf("GetPerson after update: unexpected error: %v", err)
	}
	if got.PrimaryName.FirstName != "Jane" {
		t.Errorf("FirstName after update = %q, want %q", got.PrimaryName.FirstName, "Jane")
	}

	// Iterate
	count := 0
	for _, err := range db.People() {
		if err != nil {
			t.Fatalf("People iterator: unexpected error: %v", err)
		}
		count++
	}
	if count != 1 {
		t.Errorf("People count = %d, want 1", count)
	}

	// Delete
	if err := db.DeletePerson(handle); err != nil {
		t.Fatalf("DeletePerson: unexpected error: %v", err)
	}
	got, err = db.GetPerson(handle)
	if err != nil {
		t.Fatalf("GetPerson after delete: unexpected error: %v", err)
	}
	if got != nil {
		t.Error("expected nil after delete")
	}
}

func TestFamilyCRUD(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	handle := NewHandle()
	fatherHandle := "father_handle_123"
	motherHandle := "mother_handle_456"
	f := &Family{
		Handle:       handle,
		GrampsID:     "F0001",
		FatherHandle: &fatherHandle,
		MotherHandle: &motherHandle,
		Type:         GrampsType{Class: "FamilyRelType", Value: 1},
		Change:       1700000000,
	}

	if err := db.AddFamily(f); err != nil {
		t.Fatalf("AddFamily: unexpected error: %v", err)
	}

	got, err := db.GetFamily(handle)
	if err != nil {
		t.Fatalf("GetFamily: unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("GetFamily returned nil")
	}
	if got.GrampsID != "F0001" {
		t.Errorf("GrampsID = %q, want %q", got.GrampsID, "F0001")
	}
	if got.FatherHandle == nil || *got.FatherHandle != fatherHandle {
		t.Errorf("FatherHandle mismatch")
	}
}

func TestEventCRUD(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	handle := NewHandle()
	e := &Event{
		Handle:      handle,
		GrampsID:    "E0001",
		Type:        GrampsType{Class: "EventType", Value: 12},
		Description: "Birth of John",
		Change:      1700000000,
	}

	if err := db.AddEvent(e); err != nil {
		t.Fatalf("AddEvent: unexpected error: %v", err)
	}

	got, err := db.GetEvent(handle)
	if err != nil {
		t.Fatalf("GetEvent: unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("GetEvent returned nil")
	}
	if got.Description != "Birth of John" {
		t.Errorf("Description = %q, want %q", got.Description, "Birth of John")
	}
}

func TestTagCRUD(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	handle := NewHandle()
	tag := &Tag{
		Handle:   handle,
		Name:     "Important",
		Color:    "#ff0000000000",
		Priority: 1,
		Change:   1700000000,
	}

	if err := db.AddTag(tag); err != nil {
		t.Fatalf("AddTag: unexpected error: %v", err)
	}

	got, err := db.GetTagByName("Important")
	if err != nil {
		t.Fatalf("GetTagByName: unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("GetTagByName returned nil")
	}
	if got.Color != "#ff0000000000" {
		t.Errorf("Color = %q, want %q", got.Color, "#ff0000000000")
	}
}

func TestNoteCRUD(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	handle := NewHandle()
	n := &Note{
		Handle:   handle,
		GrampsID: "N0001",
		Text:     StyledText{Class: "StyledText", String: "Hello world"},
		Format:   NoteFlowed,
		Type:     GrampsType{Class: "NoteType", Value: 1},
		Change:   1700000000,
	}

	if err := db.AddNote(n); err != nil {
		t.Fatalf("AddNote: unexpected error: %v", err)
	}

	got, err := db.GetNote(handle)
	if err != nil {
		t.Fatalf("GetNote: unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("GetNote returned nil")
	}
	if got.Text.String != "Hello world" {
		t.Errorf("Text = %q, want %q", got.Text.String, "Hello world")
	}
}

func TestMetadata(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	if err := db.SetMetadata("version", "5.2.0"); err != nil {
		t.Fatalf("SetMetadata: unexpected error: %v", err)
	}

	got, err := db.GetMetadata("version")
	if err != nil {
		t.Fatalf("GetMetadata: unexpected error: %v", err)
	}
	if got != "5.2.0" {
		t.Errorf("metadata = %v, want %q", got, "5.2.0")
	}

	// Update existing.
	if err := db.SetMetadata("version", "5.3.0"); err != nil {
		t.Fatalf("SetMetadata update: unexpected error: %v", err)
	}
	got, err = db.GetMetadata("version")
	if err != nil {
		t.Fatalf("GetMetadata after update: unexpected error: %v", err)
	}
	if got != "5.3.0" {
		t.Errorf("metadata = %v, want %q", got, "5.3.0")
	}

	// Non-existent key.
	got, err = db.GetMetadata("nonexistent")
	if err != nil {
		t.Fatalf("GetMetadata nonexistent: unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for nonexistent key, got %v", got)
	}
}

func TestPersonJSONRoundTrip(t *testing.T) {
	p := &Person{
		Class:    "Person",
		Handle:   "abc123def456",
		GrampsID: "I0001",
		Gender:   GenderMale,
		PrimaryName: Name{
			Class:     "Name",
			FirstName: "John",
			SurnameList: []Surname{
				{
					Class:      "Surname",
					Surname:    "Doe",
					Prefix:     "",
					Primary:    true,
					Origintype: GrampsType{Class: "NameOriginType", Value: 0},
					Connector:  "",
				},
			},
			Type:      GrampsType{Class: "NameType", Value: 0},
			SortAs:    0,
			DisplayAs: 0,
		},
		DeathRefIndex:    -1,
		BirthRefIndex:    0,
		EventRefList:     []EventRef{},
		FamilyList:       []string{},
		ParentFamilyList: []string{},
		MediaList:        []MediaRef{},
		AddressList:      []Address{},
		AttributeList:    []Attribute{},
		URLs:             []URL{},
		LdsOrdList:       []LdsOrd{},
		CitationList:     []string{},
		NoteList:         []string{},
		Change:           1700000000,
		TagList:          []string{},
		PersonRefList:    []PersonRef{},
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal: unexpected error: %v", err)
	}

	var got Person
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: unexpected error: %v", err)
	}

	if got.Class != "Person" {
		t.Errorf("Class = %q, want %q", got.Class, "Person")
	}
	if got.Handle != p.Handle {
		t.Errorf("Handle = %q, want %q", got.Handle, p.Handle)
	}
	if got.Gender != GenderMale {
		t.Errorf("Gender = %d, want %d", got.Gender, GenderMale)
	}
	if got.PrimaryName.FirstName != "John" {
		t.Errorf("FirstName = %q, want %q", got.PrimaryName.FirstName, "John")
	}
	if len(got.PrimaryName.SurnameList) != 1 || got.PrimaryName.SurnameList[0].Surname != "Doe" {
		t.Errorf("Surname mismatch")
	}
	if got.DeathRefIndex != -1 {
		t.Errorf("DeathRefIndex = %d, want -1", got.DeathRefIndex)
	}
	if got.BirthRefIndex != 0 {
		t.Errorf("BirthRefIndex = %d, want 0", got.BirthRefIndex)
	}
}

func TestAllObjectTypesRoundTrip(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	// Source
	srcHandle := NewHandle()
	if err := db.AddSource(&Source{
		Handle:   srcHandle,
		GrampsID: "S0001",
		Title:    "Census 1900",
		Author:   "US Government",
		Change:   1700000000,
	}); err != nil {
		t.Fatalf("AddSource: unexpected error: %v", err)
	}
	src, err := db.GetSource(srcHandle)
	if err != nil {
		t.Fatalf("GetSource: unexpected error: %v", err)
	}
	if src.Title != "Census 1900" {
		t.Errorf("Source.Title = %q, want %q", src.Title, "Census 1900")
	}

	// Citation
	citHandle := NewHandle()
	srcHandleStr := srcHandle
	if err := db.AddCitation(&Citation{
		Handle:       citHandle,
		GrampsID:     "C0001",
		Page:         "Page 42",
		Confidence:   2,
		SourceHandle: &srcHandleStr,
		Change:       1700000000,
	}); err != nil {
		t.Fatalf("AddCitation: unexpected error: %v", err)
	}
	cit, err := db.GetCitation(citHandle)
	if err != nil {
		t.Fatalf("GetCitation: unexpected error: %v", err)
	}
	if cit.Page != "Page 42" {
		t.Errorf("Citation.Page = %q, want %q", cit.Page, "Page 42")
	}

	// Place
	placeHandle := NewHandle()
	if err := db.AddPlace(&Place{
		Handle:    placeHandle,
		GrampsID:  "P0001",
		Title:     "New York",
		Name:      PlaceName{Class: "PlaceName", Value: "New York"},
		PlaceType: GrampsType{Class: "PlaceType", Value: 7},
		Change:    1700000000,
	}); err != nil {
		t.Fatalf("AddPlace: unexpected error: %v", err)
	}
	place, err := db.GetPlace(placeHandle)
	if err != nil {
		t.Fatalf("GetPlace: unexpected error: %v", err)
	}
	if place.Title != "New York" {
		t.Errorf("Place.Title = %q, want %q", place.Title, "New York")
	}

	// Repository
	repoHandle := NewHandle()
	if err := db.AddRepository(&Repository{
		Handle:   repoHandle,
		GrampsID: "R0001",
		Type:     GrampsType{Class: "RepositoryType", Value: 1},
		Name:     "National Archives",
		Change:   1700000000,
	}); err != nil {
		t.Fatalf("AddRepository: unexpected error: %v", err)
	}
	repo, err := db.GetRepository(repoHandle)
	if err != nil {
		t.Fatalf("GetRepository: unexpected error: %v", err)
	}
	if repo.Name != "National Archives" {
		t.Errorf("Repository.Name = %q, want %q", repo.Name, "National Archives")
	}

	// Media
	mediaHandle := NewHandle()
	if err := db.AddMedia(&Media{
		Handle:   mediaHandle,
		GrampsID: "O0001",
		Path:     "/photos/wedding.jpg",
		Mime:     "image/jpeg",
		Desc:     "Wedding photo",
		Change:   1700000000,
	}); err != nil {
		t.Fatalf("AddMedia: unexpected error: %v", err)
	}
	media, err := db.GetMedia(mediaHandle)
	if err != nil {
		t.Fatalf("GetMedia: unexpected error: %v", err)
	}
	if media.Desc != "Wedding photo" {
		t.Errorf("Media.Desc = %q, want %q", media.Desc, "Wedding photo")
	}
}

func TestGetNonExistent(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	got, err := db.GetPerson("nonexistent")
	if err != nil {
		t.Fatalf("GetPerson: unexpected error: %v", err)
	}
	if got != nil {
		t.Error("expected nil for nonexistent handle")
	}
}

func TestMultiplePeopleIteration(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	for i := range 5 {
		p := &Person{
			Handle:   NewHandle(),
			GrampsID: fmt.Sprintf("I%04d", i+1),
			Gender:   GenderMale,
			PrimaryName: Name{
				Class:     "Name",
				FirstName: fmt.Sprintf("Person%d", i),
				SurnameList: []Surname{
					{Class: "Surname", Surname: "Test", Primary: true, Origintype: GrampsType{Class: "NameOriginType"}},
				},
				Type: GrampsType{Class: "NameType"},
			},
		}
		if err := db.AddPerson(p); err != nil {
			t.Fatalf("AddPerson %d: unexpected error: %v", i, err)
		}
	}

	count := 0
	for _, err := range db.People() {
		if err != nil {
			t.Fatalf("People: unexpected error: %v", err)
		}
		count++
	}
	if count != 5 {
		t.Errorf("People count = %d, want 5", count)
	}
}

func TestUnsupportedSchemaVersion(t *testing.T) {
	// testdata/schema18 contains a Gramps database with schema version 18.
	dbDir := filepath.Join("testdata", "schema18")

	_, err := Open(dbDir)
	if err == nil {
		t.Fatal("expected error opening schema18 database")
	}
	var schemaErr *ErrUnsupportedSchema
	if !errors.As(err, &schemaErr) {
		t.Fatalf("expected ErrUnsupportedSchema, got %T: %v", err, err)
	}
	if schemaErr.Version != 18 {
		t.Errorf("schema version = %d, want 18", schemaErr.Version)
	}

	// OpenReadOnly should also fail.
	_, err = OpenReadOnly(dbDir)
	if err == nil {
		t.Fatal("expected error opening schema18 database read-only")
	}
	if !errors.As(err, &schemaErr) {
		t.Fatalf("expected ErrUnsupportedSchema from OpenReadOnly, got %T: %v", err, err)
	}
}

func TestSupportedSchemaVersion(t *testing.T) {
	// testdata/schema21 contains a Gramps database with schema version 21.
	dbDir := filepath.Join("testdata", "schema21")

	db, err := OpenReadOnly(dbDir)
	if err != nil {
		t.Fatalf("OpenReadOnly: unexpected error: %v", err)
	}
	defer db.Close()

	// Verify schema version can be read.
	v, err := db.GetMetadata("version")
	if err != nil {
		t.Fatalf("GetMetadata version: unexpected error: %v", err)
	}
	if v != "21" {
		t.Errorf("version = %v, want %q", v, "21")
	}
	if db.SchemaVersion() != 21 {
		t.Errorf("SchemaVersion() = %d, want 21", db.SchemaVersion())
	}
}

func TestSchemaVersion(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(t *testing.T) *Database
		want  int
	}{
		{
			name:  "newly_created",
			setup: func(t *testing.T) *Database { return createTestDB(t) },
			want:  maxSupportedSchemaVersion,
		},
		{
			name: "schema21",
			setup: func(t *testing.T) *Database {
				db, err := OpenReadOnly(filepath.Join("testdata", "schema21"))
				if err != nil {
					t.Fatalf("OpenReadOnly: %v", err)
				}
				return db
			},
			want: 21,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setup(t)
			defer db.Close()
			if got := db.SchemaVersion(); got != tc.want {
				t.Errorf("SchemaVersion() = %d, want %d", got, tc.want)
			}
		})
	}
}

func createTestDB(t *testing.T) *Database {
	t.Helper()
	dir := t.TempDir()
	dbDir := filepath.Join(dir, "testdb")
	db, err := Create(dbDir, "Test DB")
	if err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}
	return db
}
