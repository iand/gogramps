//go:build gramps_schema23

package gogramps

import "testing"

func TestDNATestCRUD(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	handle := NewHandle()
	personHandle := "person_handle_123"
	dt := &DNATest{
		Handle:       handle,
		GrampsID:     "D0001",
		PersonHandle: &personHandle,
		AccountName:  "john.doe",
		Provider:     GrampsType{Class: "DNAProviderType", Value: DNAProviderAncestry},
		KitID:        "KIT12345",
		TestType:     GrampsType{Class: "DNATestType", Value: DNATestAutosomal},
		GenomeBuild:  GrampsType{Class: "DNAGenomeBuildType", Value: DNAGenomeBuildGRCh37},
		Haplogroup:   "R-M269",
		Change:       1700000000,
	}

	if err := db.AddDNATest(dt); err != nil {
		t.Fatalf("AddDNATest: unexpected error: %v", err)
	}

	// Verify secondary columns are populated.
	var secGrampsID, secPersonHandle, secAccountName, secKitID, secHaplogroup string
	var secChange, secPrivate int
	err := db.db.QueryRow(
		"SELECT gramps_id, person_handle, account_name, kit_id, haplogroup, change, private FROM dnatest WHERE handle = ?",
		handle,
	).Scan(&secGrampsID, &secPersonHandle, &secAccountName, &secKitID, &secHaplogroup, &secChange, &secPrivate)
	if err != nil {
		t.Fatalf("secondary columns query: unexpected error: %v", err)
	}
	if secGrampsID != "D0001" {
		t.Errorf("secondary gramps_id = %q, want %q", secGrampsID, "D0001")
	}
	if secPersonHandle != personHandle {
		t.Errorf("secondary person_handle = %q, want %q", secPersonHandle, personHandle)
	}
	if secAccountName != "john.doe" {
		t.Errorf("secondary account_name = %q, want %q", secAccountName, "john.doe")
	}
	if secKitID != "KIT12345" {
		t.Errorf("secondary kit_id = %q, want %q", secKitID, "KIT12345")
	}
	if secHaplogroup != "R-M269" {
		t.Errorf("secondary haplogroup = %q, want %q", secHaplogroup, "R-M269")
	}
	if secChange != 1700000000 {
		t.Errorf("secondary change = %d, want %d", secChange, 1700000000)
	}
	if secPrivate != 0 {
		t.Errorf("secondary private = %d, want 0", secPrivate)
	}

	got, err := db.GetDNATest(handle)
	if err != nil {
		t.Fatalf("GetDNATest: unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("GetDNATest returned nil")
	}
	if got.GrampsID != "D0001" {
		t.Errorf("GrampsID = %q, want %q", got.GrampsID, "D0001")
	}
	if got.PersonHandle == nil || *got.PersonHandle != personHandle {
		t.Errorf("PersonHandle mismatch")
	}
	if got.KitID != "KIT12345" {
		t.Errorf("KitID = %q, want %q", got.KitID, "KIT12345")
	}
	if got.Haplogroup != "R-M269" {
		t.Errorf("Haplogroup = %q, want %q", got.Haplogroup, "R-M269")
	}

	got2, err := db.GetDNATestByGrampsID("D0001")
	if err != nil {
		t.Fatalf("GetDNATestByGrampsID: unexpected error: %v", err)
	}
	if got2 == nil || got2.Handle != handle {
		t.Errorf("GetDNATestByGrampsID: wrong result")
	}

	dt.Haplogroup = "I-M253"
	if err := db.UpdateDNATest(dt); err != nil {
		t.Fatalf("UpdateDNATest: unexpected error: %v", err)
	}
	if err := db.db.QueryRow("SELECT haplogroup FROM dnatest WHERE handle = ?", handle).Scan(&secHaplogroup); err != nil {
		t.Fatalf("secondary haplogroup after update: unexpected error: %v", err)
	}
	if secHaplogroup != "I-M253" {
		t.Errorf("secondary haplogroup after update = %q, want %q", secHaplogroup, "I-M253")
	}
	got, err = db.GetDNATest(handle)
	if err != nil {
		t.Fatalf("GetDNATest after update: unexpected error: %v", err)
	}
	if got.Haplogroup != "I-M253" {
		t.Errorf("Haplogroup after update = %q, want %q", got.Haplogroup, "I-M253")
	}

	count := 0
	for _, err := range db.DNATests() {
		if err != nil {
			t.Fatalf("DNATests: unexpected error: %v", err)
		}
		count++
	}
	if count != 1 {
		t.Errorf("DNATests count = %d, want 1", count)
	}

	if err := db.DeleteDNATest(handle); err != nil {
		t.Fatalf("DeleteDNATest: unexpected error: %v", err)
	}
	got, err = db.GetDNATest(handle)
	if err != nil {
		t.Fatalf("GetDNATest after delete: unexpected error: %v", err)
	}
	if got != nil {
		t.Error("expected nil after delete")
	}
}

func TestDNAMatchCRUD(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	handle := NewHandle()
	subjectHandle := "dnatest_subject_123"
	matchHandle := "dnatest_match_456"
	generations := 2.5
	dm := &DNAMatch{
		Handle:                handle,
		GrampsID:              "M0001",
		SubjectTestHandle:     &subjectHandle,
		MatchTestHandle:       &matchHandle,
		SharedCM:              187.3,
		PercentShared:         2.8,
		SegmentCount:          8,
		LargestSegmentCM:      54.2,
		PredictedRelationship: "1st Cousin",
		PredictedGenerations:  &generations,
		Change:                1700000000,
	}

	if err := db.AddDNAMatch(dm); err != nil {
		t.Fatalf("AddDNAMatch: unexpected error: %v", err)
	}

	// Verify secondary columns are populated.
	var secGrampsID, secSubjectHandle, secMatchHandle, secPredRel string
	var secSharedCM, secPercentShared, secLargestCM, secPredGen float64
	var secSegCount, secChange, secPrivate int
	err := db.db.QueryRow(
		"SELECT gramps_id, subject_test_handle, match_test_handle, shared_cm, percent_shared, segment_count, largest_segment_cm, predicted_relationship, predicted_generations, change, private FROM dnamatch WHERE handle = ?",
		handle,
	).Scan(&secGrampsID, &secSubjectHandle, &secMatchHandle, &secSharedCM, &secPercentShared, &secSegCount, &secLargestCM, &secPredRel, &secPredGen, &secChange, &secPrivate)
	if err != nil {
		t.Fatalf("secondary columns query: unexpected error: %v", err)
	}
	if secGrampsID != "M0001" {
		t.Errorf("secondary gramps_id = %q, want %q", secGrampsID, "M0001")
	}
	if secSubjectHandle != subjectHandle {
		t.Errorf("secondary subject_test_handle = %q, want %q", secSubjectHandle, subjectHandle)
	}
	if secMatchHandle != matchHandle {
		t.Errorf("secondary match_test_handle = %q, want %q", secMatchHandle, matchHandle)
	}
	if secSharedCM != 187.3 {
		t.Errorf("secondary shared_cm = %v, want 187.3", secSharedCM)
	}
	if secPercentShared != 2.8 {
		t.Errorf("secondary percent_shared = %v, want 2.8", secPercentShared)
	}
	if secSegCount != 8 {
		t.Errorf("secondary segment_count = %d, want 8", secSegCount)
	}
	if secLargestCM != 54.2 {
		t.Errorf("secondary largest_segment_cm = %v, want 54.2", secLargestCM)
	}
	if secPredRel != "1st Cousin" {
		t.Errorf("secondary predicted_relationship = %q, want %q", secPredRel, "1st Cousin")
	}
	if secPredGen != 2.5 {
		t.Errorf("secondary predicted_generations = %v, want 2.5", secPredGen)
	}
	if secChange != 1700000000 {
		t.Errorf("secondary change = %d, want %d", secChange, 1700000000)
	}
	if secPrivate != 0 {
		t.Errorf("secondary private = %d, want 0", secPrivate)
	}

	got, err := db.GetDNAMatch(handle)
	if err != nil {
		t.Fatalf("GetDNAMatch: unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("GetDNAMatch returned nil")
	}
	if got.GrampsID != "M0001" {
		t.Errorf("GrampsID = %q, want %q", got.GrampsID, "M0001")
	}
	if got.SharedCM != 187.3 {
		t.Errorf("SharedCM = %v, want 187.3", got.SharedCM)
	}
	if got.PredictedRelationship != "1st Cousin" {
		t.Errorf("PredictedRelationship = %q, want %q", got.PredictedRelationship, "1st Cousin")
	}
	if got.PredictedGenerations == nil || *got.PredictedGenerations != 2.5 {
		t.Errorf("PredictedGenerations mismatch")
	}

	dm.PredictedRelationship = "2nd Cousin"
	if err := db.UpdateDNAMatch(dm); err != nil {
		t.Fatalf("UpdateDNAMatch: unexpected error: %v", err)
	}
	if err := db.db.QueryRow("SELECT predicted_relationship FROM dnamatch WHERE handle = ?", handle).Scan(&secPredRel); err != nil {
		t.Fatalf("secondary predicted_relationship after update: unexpected error: %v", err)
	}
	if secPredRel != "2nd Cousin" {
		t.Errorf("secondary predicted_relationship after update = %q, want %q", secPredRel, "2nd Cousin")
	}
	got, err = db.GetDNAMatch(handle)
	if err != nil {
		t.Fatalf("GetDNAMatch after update: unexpected error: %v", err)
	}
	if got.PredictedRelationship != "2nd Cousin" {
		t.Errorf("PredictedRelationship after update = %q, want %q", got.PredictedRelationship, "2nd Cousin")
	}

	count := 0
	for _, err := range db.DNAMatches() {
		if err != nil {
			t.Fatalf("DNAMatches: unexpected error: %v", err)
		}
		count++
	}
	if count != 1 {
		t.Errorf("DNAMatches count = %d, want 1", count)
	}

	if err := db.DeleteDNAMatch(handle); err != nil {
		t.Fatalf("DeleteDNAMatch: unexpected error: %v", err)
	}
	got, err = db.GetDNAMatch(handle)
	if err != nil {
		t.Fatalf("GetDNAMatch after delete: unexpected error: %v", err)
	}
	if got != nil {
		t.Error("expected nil after delete")
	}
}
