package gogramps

import "testing"

func addPersonID(t *testing.T, db *Database, grampsID string) {
	t.Helper()
	if err := db.AddPerson(&Person{Handle: NewHandle(), GrampsID: grampsID}); err != nil {
		t.Fatalf("AddPerson %q: unexpected error: %v", grampsID, err)
	}
}

func TestGrampsIDAllocatorEmptyDB(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	alloc, err := db.GrampsIDAllocator(ObjectPerson)
	if err != nil {
		t.Fatalf("GrampsIDAllocator: unexpected error: %v", err)
	}

	want := []string{"I00000", "I00001", "I00002"}
	for i, w := range want {
		if got := alloc.Next(); got != w {
			t.Errorf("Next() call %d = %q, want %q", i, got, w)
		}
	}
}

func TestGrampsIDAllocatorSeedsFromMax(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	addPersonID(t, db, "I00005")
	addPersonID(t, db, "I00042")
	addPersonID(t, db, "I00007")

	alloc, err := db.GrampsIDAllocator(ObjectPerson)
	if err != nil {
		t.Fatalf("GrampsIDAllocator: unexpected error: %v", err)
	}

	if got, want := alloc.Next(), "I00043"; got != want {
		t.Errorf("Next() = %q, want %q", got, want)
	}
	if got, want := alloc.Next(), "I00044"; got != want {
		t.Errorf("Next() = %q, want %q", got, want)
	}
}

func TestGrampsIDAllocatorNumericNotLexicalMax(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	// "I00099" sorts after "I000100" lexically but is the smaller number.
	addPersonID(t, db, "I00099")
	addPersonID(t, db, "I000100")

	alloc, err := db.GrampsIDAllocator(ObjectPerson)
	if err != nil {
		t.Fatalf("GrampsIDAllocator: unexpected error: %v", err)
	}

	if got, want := alloc.Next(), "I00101"; got != want {
		t.Errorf("Next() = %q, want %q", got, want)
	}
}

func TestGrampsIDAllocatorIgnoresNonMatching(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	addPersonID(t, db, "I00003")
	addPersonID(t, db, "PERSON-A") // wrong prefix, ignored
	addPersonID(t, db, "I00abc")   // non-numeric tail, ignored

	alloc, err := db.GrampsIDAllocator(ObjectPerson)
	if err != nil {
		t.Fatalf("GrampsIDAllocator: unexpected error: %v", err)
	}

	if got, want := alloc.Next(), "I00004"; got != want {
		t.Errorf("Next() = %q, want %q", got, want)
	}
}

func TestGrampsIDAllocatorPerTypeDefaults(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	cases := map[ObjectType]string{
		ObjectPerson:     "I00000",
		ObjectFamily:     "F00000",
		ObjectEvent:      "E00000",
		ObjectPlace:      "P00000",
		ObjectSource:     "S00000",
		ObjectCitation:   "C00000",
		ObjectRepository: "R00000",
		ObjectNote:       "N00000",
		ObjectMedia:      "O00000",
	}
	for typ, want := range cases {
		alloc, err := db.GrampsIDAllocator(typ)
		if err != nil {
			t.Fatalf("GrampsIDAllocator(%q): unexpected error: %v", typ, err)
		}
		if got := alloc.Next(); got != want {
			t.Errorf("Next() for %q = %q, want %q", typ, got, want)
		}
	}
}

func TestGrampsIDAllocatorIndependentPerType(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	// A person ID must not seed the event allocator.
	addPersonID(t, db, "I00050")

	alloc, err := db.GrampsIDAllocator(ObjectEvent)
	if err != nil {
		t.Fatalf("GrampsIDAllocator: unexpected error: %v", err)
	}
	if got, want := alloc.Next(), "E00000"; got != want {
		t.Errorf("Next() = %q, want %q", got, want)
	}
}

func TestGrampsIDAllocatorWithFormat(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	addPersonID(t, db, "I0003")

	alloc, err := db.GrampsIDAllocatorWithFormat(ObjectPerson, "I%04d")
	if err != nil {
		t.Fatalf("GrampsIDAllocatorWithFormat: unexpected error: %v", err)
	}
	if got, want := alloc.Next(), "I0004"; got != want {
		t.Errorf("Next() = %q, want %q", got, want)
	}
}

func TestGrampsIDAllocatorUnknownType(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	if _, err := db.GrampsIDAllocator(ObjectType("nonesuch")); err == nil {
		t.Fatal("GrampsIDAllocator: expected error for unknown object type, got nil")
	}
}

func TestHasHandle(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	handle := NewHandle()

	has, err := db.HasHandle(ObjectPerson, handle)
	if err != nil {
		t.Fatalf("HasHandle: unexpected error: %v", err)
	}
	if has {
		t.Errorf("HasHandle before insert = true, want false")
	}

	if err := db.AddPerson(&Person{Handle: handle, GrampsID: "I00001"}); err != nil {
		t.Fatalf("AddPerson: unexpected error: %v", err)
	}

	has, err = db.HasHandle(ObjectPerson, handle)
	if err != nil {
		t.Fatalf("HasHandle: unexpected error: %v", err)
	}
	if !has {
		t.Errorf("HasHandle after insert = false, want true")
	}

	// The check is scoped to the object type's table.
	has, err = db.HasHandle(ObjectEvent, handle)
	if err != nil {
		t.Fatalf("HasHandle: unexpected error: %v", err)
	}
	if has {
		t.Errorf("HasHandle(ObjectEvent) = true, want false for a person handle")
	}
}

func TestHasHandleUnknownType(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	if _, err := db.HasHandle(ObjectType("nonesuch"), NewHandle()); err == nil {
		t.Fatal("HasHandle: expected error for unknown object type, got nil")
	}
}

func TestNewUniqueHandle(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	h, err := db.NewUniqueHandle(ObjectPerson)
	if err != nil {
		t.Fatalf("NewUniqueHandle: unexpected error: %v", err)
	}
	has, err := db.HasHandle(ObjectPerson, h)
	if err != nil {
		t.Fatalf("HasHandle: unexpected error: %v", err)
	}
	if has {
		t.Errorf("NewUniqueHandle returned handle %q already present", h)
	}

	// A handle already in use is not handed out again.
	if err := db.AddPerson(&Person{Handle: h, GrampsID: "I00001"}); err != nil {
		t.Fatalf("AddPerson: unexpected error: %v", err)
	}
	h2, err := db.NewUniqueHandle(ObjectPerson)
	if err != nil {
		t.Fatalf("NewUniqueHandle: unexpected error: %v", err)
	}
	if h2 == h {
		t.Errorf("NewUniqueHandle returned the in-use handle %q", h2)
	}
}

func TestNewUniqueHandleUnknownType(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	if _, err := db.NewUniqueHandle(ObjectType("nonesuch")); err == nil {
		t.Fatal("NewUniqueHandle: expected error for unknown object type, got nil")
	}
}

func TestGrampsIDAllocatorInvalidFormat(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	if _, err := db.GrampsIDAllocatorWithFormat(ObjectPerson, "I000"); err == nil {
		t.Fatal("GrampsIDAllocatorWithFormat: expected error for format without a verb, got nil")
	}
}
