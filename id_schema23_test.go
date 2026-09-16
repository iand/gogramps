//go:build gramps_dbschema23

package gogramps

import "testing"

func TestGrampsIDAllocatorDNATest(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	if err := db.AddDNATest(&DNATest{Handle: NewHandle(), GrampsID: "T00004"}); err != nil {
		t.Fatalf("AddDNATest: unexpected error: %v", err)
	}

	alloc, err := db.GrampsIDAllocator(ObjectDNATest)
	if err != nil {
		t.Fatalf("GrampsIDAllocator: unexpected error: %v", err)
	}
	if got, want := alloc.Next(), "T00005"; got != want {
		t.Errorf("Next() = %q, want %q", got, want)
	}
}

func TestGrampsIDAllocatorDNAMatch(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	alloc, err := db.GrampsIDAllocator(ObjectDNAMatch)
	if err != nil {
		t.Fatalf("GrampsIDAllocator: unexpected error: %v", err)
	}
	if got, want := alloc.Next(), "M00000"; got != want {
		t.Errorf("Next() = %q, want %q", got, want)
	}
}
