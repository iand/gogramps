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
		YHaplogroup:  "R-M269",
		MtHaplogroup: "H1a",
		Change:       1700000000,
	}

	if err := db.AddDNATest(dt); err != nil {
		t.Fatalf("AddDNATest: unexpected error: %v", err)
	}

	// Verify secondary columns are populated.
	var secGrampsID, secPersonHandle, secAccountName, secKitID, secYHaplogroup, secMtHaplogroup string
	var secChange, secPrivate int
	err := db.db.QueryRow(
		"SELECT gramps_id, person_handle, account_name, kit_id, y_haplogroup, mt_haplogroup, change, private FROM dnatest WHERE handle = ?",
		handle,
	).Scan(&secGrampsID, &secPersonHandle, &secAccountName, &secKitID, &secYHaplogroup, &secMtHaplogroup, &secChange, &secPrivate)
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
	if secYHaplogroup != "R-M269" {
		t.Errorf("secondary y_haplogroup = %q, want %q", secYHaplogroup, "R-M269")
	}
	if secMtHaplogroup != "H1a" {
		t.Errorf("secondary mt_haplogroup = %q, want %q", secMtHaplogroup, "H1a")
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
	if got.YHaplogroup != "R-M269" {
		t.Errorf("YHaplogroup = %q, want %q", got.YHaplogroup, "R-M269")
	}
	if got.MtHaplogroup != "H1a" {
		t.Errorf("MtHaplogroup = %q, want %q", got.MtHaplogroup, "H1a")
	}

	got2, err := db.GetDNATestByGrampsID("D0001")
	if err != nil {
		t.Fatalf("GetDNATestByGrampsID: unexpected error: %v", err)
	}
	if got2 == nil || got2.Handle != handle {
		t.Errorf("GetDNATestByGrampsID: wrong result")
	}

	dt.YHaplogroup = "I-M253"
	if err := db.UpdateDNATest(dt); err != nil {
		t.Fatalf("UpdateDNATest: unexpected error: %v", err)
	}
	if err := db.db.QueryRow("SELECT y_haplogroup FROM dnatest WHERE handle = ?", handle).Scan(&secYHaplogroup); err != nil {
		t.Fatalf("secondary y_haplogroup after update: unexpected error: %v", err)
	}
	if secYHaplogroup != "I-M253" {
		t.Errorf("secondary y_haplogroup after update = %q, want %q", secYHaplogroup, "I-M253")
	}
	got, err = db.GetDNATest(handle)
	if err != nil {
		t.Fatalf("GetDNATest after update: unexpected error: %v", err)
	}
	if got.YHaplogroup != "I-M253" {
		t.Errorf("YHaplogroup after update = %q, want %q", got.YHaplogroup, "I-M253")
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
	dm := &DNAMatch{
		Handle:                   handle,
		GrampsID:                 "M0001",
		SubjectTestHandle:        &subjectHandle,
		MatchTestHandle:          &matchHandle,
		SharedCM:                 187.3,
		SharedCMWeighted:         182.1,
		PercentShared:            2.8,
		SegmentCount:             8,
		LargestSegmentCM:         54.2,
		LargestSegmentCMWeighted: 52.0,
		PredictedRelationshipList: []PredictedRelationship{
			{
				Class:           "PredictedRelationship",
				Description:     "1st Cousin",
				SubjectMRCAGens: 2,
				SubjectSide:     PredictedRelSideMaternal,
				MatchMRCAGens:   2,
				MatchSide:       PredictedRelSidePaternal,
				FullOrHalf:      PredictedRelFOHFull,
				Probability:     0.85,
				CitationList:    []string{"citation_handle_1"},
				NoteList:        []string{"note_handle_1"},
			},
		},
		SegmentList: []DNASegment{
			{
				Class:            "DNASegment",
				Chromosome:       "7",
				StartBP:          12000000,
				EndBP:            30000000,
				SharedCM:         54.2,
				SharedCMWeighted: 52.0,
				SNPCount:         678,
				Origin:           DNAOriginMaternal,
				IBDState:         DNAIBDHalfIdenticalRegion,
				GenomeBuild:      GrampsType{Class: "DNAGenomeBuildType", Value: DNAGenomeBuildGRCh37},
			},
		},
		Change: 1700000000,
	}

	if err := db.AddDNAMatch(dm); err != nil {
		t.Fatalf("AddDNAMatch: unexpected error: %v", err)
	}

	// Verify secondary columns are populated.
	var secGrampsID, secSubjectHandle, secMatchHandle string
	var secSharedCM, secSharedCMWeighted, secPercentShared, secLargestCM, secLargestCMWeighted float64
	var secSegCount, secChange, secPrivate int
	err := db.db.QueryRow(
		"SELECT gramps_id, subject_test_handle, match_test_handle, shared_cm, shared_cm_weighted, percent_shared, segment_count, largest_segment_cm, largest_segment_cm_weighted, change, private FROM dnamatch WHERE handle = ?",
		handle,
	).Scan(&secGrampsID, &secSubjectHandle, &secMatchHandle, &secSharedCM, &secSharedCMWeighted, &secPercentShared, &secSegCount, &secLargestCM, &secLargestCMWeighted, &secChange, &secPrivate)
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
	if secSharedCMWeighted != 182.1 {
		t.Errorf("secondary shared_cm_weighted = %v, want 182.1", secSharedCMWeighted)
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
	if secLargestCMWeighted != 52.0 {
		t.Errorf("secondary largest_segment_cm_weighted = %v, want 52.0", secLargestCMWeighted)
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
	if got.SharedCMWeighted != 182.1 {
		t.Errorf("SharedCMWeighted = %v, want 182.1", got.SharedCMWeighted)
	}
	if len(got.PredictedRelationshipList) != 1 {
		t.Fatalf("PredictedRelationshipList len = %d, want 1", len(got.PredictedRelationshipList))
	}
	pr := got.PredictedRelationshipList[0]
	if pr.Description != "1st Cousin" {
		t.Errorf("PredictedRelationship description = %q, want %q", pr.Description, "1st Cousin")
	}
	if pr.SubjectSide != PredictedRelSideMaternal || pr.MatchSide != PredictedRelSidePaternal {
		t.Errorf("PredictedRelationship sides = %d/%d, want %d/%d", pr.SubjectSide, pr.MatchSide, PredictedRelSideMaternal, PredictedRelSidePaternal)
	}
	if pr.Probability != 0.85 {
		t.Errorf("PredictedRelationship probability = %v, want 0.85", pr.Probability)
	}
	if len(got.SegmentList) != 1 {
		t.Fatalf("SegmentList len = %d, want 1", len(got.SegmentList))
	}
	seg := got.SegmentList[0]
	if seg.Origin != DNAOriginMaternal {
		t.Errorf("segment origin = %d, want %d", seg.Origin, DNAOriginMaternal)
	}
	if seg.SharedCMWeighted != 52.0 {
		t.Errorf("segment shared_cm_weighted = %v, want 52.0", seg.SharedCMWeighted)
	}
	if seg.GenomeBuild.Value != DNAGenomeBuildGRCh37 {
		t.Errorf("segment genome_build = %d, want %d", seg.GenomeBuild.Value, DNAGenomeBuildGRCh37)
	}

	dm.SharedCMWeighted = 180.0
	if err := db.UpdateDNAMatch(dm); err != nil {
		t.Fatalf("UpdateDNAMatch: unexpected error: %v", err)
	}
	if err := db.db.QueryRow("SELECT shared_cm_weighted FROM dnamatch WHERE handle = ?", handle).Scan(&secSharedCMWeighted); err != nil {
		t.Fatalf("secondary shared_cm_weighted after update: unexpected error: %v", err)
	}
	if secSharedCMWeighted != 180.0 {
		t.Errorf("secondary shared_cm_weighted after update = %v, want 180.0", secSharedCMWeighted)
	}
	got, err = db.GetDNAMatch(handle)
	if err != nil {
		t.Fatalf("GetDNAMatch after update: unexpected error: %v", err)
	}
	if got.SharedCMWeighted != 180.0 {
		t.Errorf("SharedCMWeighted after update = %v, want 180.0", got.SharedCMWeighted)
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
