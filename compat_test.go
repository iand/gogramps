package gogramps_test

import (
	"compress/gzip"
	"encoding/xml"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/iand/gogramps"
	"github.com/iand/grampsxml"
)

// TestCompat opens a Gramps schema21 SQLite database and its XML export,
// then verifies that key fields match for every primary object type.
func TestCompat(t *testing.T) {
	const (
		dbDir   = "testdata/schema21"
		xmlFile = "testdata/schema21/example.gramps"
	)

	// Parse the XML export.
	xdb := parseXML(t, xmlFile)

	// Open the SQLite database read-only.
	db, err := gogramps.OpenReadOnly(dbDir)
	if err != nil {
		t.Fatalf("OpenReadOnly: unexpected error: %v", err)
	}
	defer db.Close()

	t.Run("Tag", func(t *testing.T) {
		xmlTags := make(map[string]grampsxml.Tag)
		if xdb.Tags != nil {
			for _, tag := range xdb.Tags.Tag {
				xmlTags[stripHandlePrefix(tag.Handle)] = tag
			}
		}

		count := 0
		for tag, err := range db.Tags() {
			if err != nil {
				t.Fatalf("Tags iterator: unexpected error: %v", err)
			}
			count++
			xt, ok := xmlTags[tag.Handle]
			if !ok {
				t.Errorf("tag %q: not found in XML", tag.Handle)
				continue
			}
			if tag.Name != xt.Name {
				t.Errorf("tag %q Name: got %q, want %q", tag.Handle, tag.Name, xt.Name)
			}
			if tag.Color != xt.Color {
				t.Errorf("tag %q Color: got %q, want %q", tag.Handle, tag.Color, xt.Color)
			}
			wantPriority, _ := strconv.Atoi(xt.Priority)
			if tag.Priority != wantPriority {
				t.Errorf("tag %q Priority: got %d, want %d", tag.Handle, tag.Priority, wantPriority)
			}
		}
		if count != len(xmlTags) {
			t.Errorf("Tag count: got %d, want %d", count, len(xmlTags))
		}
	})

	t.Run("Event", func(t *testing.T) {
		xmlEvents := make(map[string]grampsxml.Event)
		if xdb.Events != nil {
			for _, ev := range xdb.Events.Event {
				xmlEvents[stripHandlePrefix(ev.Handle)] = ev
			}
		}

		count := 0
		for ev, err := range db.Events() {
			if err != nil {
				t.Fatalf("Events iterator: unexpected error: %v", err)
			}
			count++
			xev, ok := xmlEvents[ev.Handle]
			if !ok {
				t.Errorf("event %q: not found in XML", ev.Handle)
				continue
			}
			wantID := ptrStr(xev.ID)
			if ev.GrampsID != wantID {
				t.Errorf("event %q GrampsID: got %q, want %q", ev.Handle, ev.GrampsID, wantID)
			}
			wantDesc := ptrStr(xev.Description)
			if ev.Description != wantDesc {
				t.Errorf("event %q Description: got %q, want %q", ev.Handle, ev.Description, wantDesc)
			}
			wantPlace := ""
			if xev.Place != nil {
				wantPlace = stripHandlePrefix(xev.Place.Hlink)
			}
			if ev.Place != wantPlace {
				t.Errorf("event %q Place: got %q, want %q", ev.Handle, ev.Place, wantPlace)
			}
			wantPrivate := ptrBool(xev.Priv)
			if ev.Private != wantPrivate {
				t.Errorf("event %q Private: got %v, want %v", ev.Handle, ev.Private, wantPrivate)
			}
		}
		if count != len(xmlEvents) {
			t.Errorf("Event count: got %d, want %d", count, len(xmlEvents))
		}
	})

	t.Run("Person", func(t *testing.T) {
		xmlPersons := make(map[string]grampsxml.Person)
		if xdb.People != nil {
			for _, p := range xdb.People.Person {
				xmlPersons[stripHandlePrefix(p.Handle)] = p
			}
		}

		count := 0
		for p, err := range db.People() {
			if err != nil {
				t.Fatalf("People iterator: unexpected error: %v", err)
			}
			count++
			xp, ok := xmlPersons[p.Handle]
			if !ok {
				t.Errorf("person %q: not found in XML", p.Handle)
				continue
			}
			wantID := ptrStr(xp.ID)
			if p.GrampsID != wantID {
				t.Errorf("person %q GrampsID: got %q, want %q", p.Handle, p.GrampsID, wantID)
			}

			wantGender := xmlGenderToInt(xp.Gender)
			if p.Gender != wantGender {
				t.Errorf("person %q Gender: got %d, want %d", p.Handle, p.Gender, wantGender)
			}

			// Find primary name in XML (Alt is nil or false).
			var xPrimaryName *grampsxml.Name
			for i := range xp.Name {
				if !ptrBool(xp.Name[i].Alt) {
					xPrimaryName = &xp.Name[i]
					break
				}
			}
			if xPrimaryName != nil {
				wantFirst := ptrStr(xPrimaryName.First)
				if p.PrimaryName.FirstName != wantFirst {
					t.Errorf("person %q FirstName: got %q, want %q", p.Handle, p.PrimaryName.FirstName, wantFirst)
				}
				wantSurname := xmlPrimarySurname(xPrimaryName.Surname)
				gotSurname := dbPrimarySurname(p.PrimaryName.SurnameList)
				if gotSurname != wantSurname {
					t.Errorf("person %q Surname: got %q, want %q", p.Handle, gotSurname, wantSurname)
				}
			}

			// Family handles (Parentin in XML = FamilyList in DB).
			wantFamilies := make(map[string]bool)
			for _, pi := range xp.Parentin {
				wantFamilies[stripHandlePrefix(pi.Hlink)] = true
			}
			gotFamilies := make(map[string]bool)
			for _, fh := range p.FamilyList {
				gotFamilies[fh] = true
			}
			if !mapsEqual(gotFamilies, wantFamilies) {
				t.Errorf("person %q FamilyList: got %v, want %v", p.Handle, p.FamilyList, keys(wantFamilies))
			}

			// Parent family handles (Childof in XML = ParentFamilyList in DB).
			wantParentFamilies := make(map[string]bool)
			for _, co := range xp.Childof {
				wantParentFamilies[stripHandlePrefix(co.Hlink)] = true
			}
			gotParentFamilies := make(map[string]bool)
			for _, fh := range p.ParentFamilyList {
				gotParentFamilies[fh] = true
			}
			if !mapsEqual(gotParentFamilies, wantParentFamilies) {
				t.Errorf("person %q ParentFamilyList: got %v, want %v", p.Handle, p.ParentFamilyList, keys(wantParentFamilies))
			}

			wantPrivate := ptrBool(xp.Priv)
			if p.Private != wantPrivate {
				t.Errorf("person %q Private: got %v, want %v", p.Handle, p.Private, wantPrivate)
			}
		}
		if count != len(xmlPersons) {
			t.Errorf("Person count: got %d, want %d", count, len(xmlPersons))
		}
	})

	t.Run("Family", func(t *testing.T) {
		xmlFamilies := make(map[string]grampsxml.Family)
		if xdb.Families != nil {
			for _, f := range xdb.Families.Family {
				xmlFamilies[stripHandlePrefix(f.Handle)] = f
			}
		}

		count := 0
		for f, err := range db.Families() {
			if err != nil {
				t.Fatalf("Families iterator: unexpected error: %v", err)
			}
			count++
			xf, ok := xmlFamilies[f.Handle]
			if !ok {
				t.Errorf("family %q: not found in XML", f.Handle)
				continue
			}
			wantID := ptrStr(xf.ID)
			if f.GrampsID != wantID {
				t.Errorf("family %q GrampsID: got %q, want %q", f.Handle, f.GrampsID, wantID)
			}

			wantFather := ""
			if xf.Father != nil {
				wantFather = stripHandlePrefix(xf.Father.Hlink)
			}
			gotFather := ptrStr(f.FatherHandle)
			if gotFather != wantFather {
				t.Errorf("family %q FatherHandle: got %q, want %q", f.Handle, gotFather, wantFather)
			}

			wantMother := ""
			if xf.Mother != nil {
				wantMother = stripHandlePrefix(xf.Mother.Hlink)
			}
			gotMother := ptrStr(f.MotherHandle)
			if gotMother != wantMother {
				t.Errorf("family %q MotherHandle: got %q, want %q", f.Handle, gotMother, wantMother)
			}

			// Child ref handles.
			wantChildren := make(map[string]bool)
			for _, cr := range xf.Childref {
				wantChildren[stripHandlePrefix(cr.Hlink)] = true
			}
			gotChildren := make(map[string]bool)
			for _, cr := range f.ChildRefList {
				gotChildren[cr.Ref] = true
			}
			if !mapsEqual(gotChildren, wantChildren) {
				t.Errorf("family %q children: got %v, want %v", f.Handle, keys(gotChildren), keys(wantChildren))
			}
		}
		if count != len(xmlFamilies) {
			t.Errorf("Family count: got %d, want %d", count, len(xmlFamilies))
		}
	})

	t.Run("Source", func(t *testing.T) {
		xmlSources := make(map[string]grampsxml.Source)
		if xdb.Sources != nil {
			for _, s := range xdb.Sources.Source {
				xmlSources[stripHandlePrefix(s.Handle)] = s
			}
		}

		count := 0
		for s, err := range db.Sources() {
			if err != nil {
				t.Fatalf("Sources iterator: unexpected error: %v", err)
			}
			count++
			xs, ok := xmlSources[s.Handle]
			if !ok {
				t.Errorf("source %q: not found in XML", s.Handle)
				continue
			}
			wantID := ptrStr(xs.ID)
			if s.GrampsID != wantID {
				t.Errorf("source %q GrampsID: got %q, want %q", s.Handle, s.GrampsID, wantID)
			}
			if s.Title != ptrStr(xs.Stitle) {
				t.Errorf("source %q Title: got %q, want %q", s.Handle, s.Title, ptrStr(xs.Stitle))
			}
			if s.Author != ptrStr(xs.Sauthor) {
				t.Errorf("source %q Author: got %q, want %q", s.Handle, s.Author, ptrStr(xs.Sauthor))
			}
			if s.Pubinfo != ptrStr(xs.Spubinfo) {
				t.Errorf("source %q Pubinfo: got %q, want %q", s.Handle, s.Pubinfo, ptrStr(xs.Spubinfo))
			}
			if s.Abbrev != ptrStr(xs.Sabbrev) {
				t.Errorf("source %q Abbrev: got %q, want %q", s.Handle, s.Abbrev, ptrStr(xs.Sabbrev))
			}
		}
		if count != len(xmlSources) {
			t.Errorf("Source count: got %d, want %d", count, len(xmlSources))
		}
	})

	t.Run("Citation", func(t *testing.T) {
		xmlCitations := make(map[string]grampsxml.Citation)
		if xdb.Citations != nil {
			for _, c := range xdb.Citations.Citation {
				xmlCitations[stripHandlePrefix(c.Handle)] = c
			}
		}

		count := 0
		for c, err := range db.Citations() {
			if err != nil {
				t.Fatalf("Citations iterator: unexpected error: %v", err)
			}
			count++
			xc, ok := xmlCitations[c.Handle]
			if !ok {
				t.Errorf("citation %q: not found in XML", c.Handle)
				continue
			}
			wantID := ptrStr(xc.ID)
			if c.GrampsID != wantID {
				t.Errorf("citation %q GrampsID: got %q, want %q", c.Handle, c.GrampsID, wantID)
			}
			if c.Page != ptrStr(xc.Page) {
				t.Errorf("citation %q Page: got %q, want %q", c.Handle, c.Page, ptrStr(xc.Page))
			}
			wantConf, _ := strconv.Atoi(xc.Confidence)
			if c.Confidence != wantConf {
				t.Errorf("citation %q Confidence: got %d, want %d", c.Handle, c.Confidence, wantConf)
			}
			wantSource := ""
			if xc.Sourceref != nil {
				wantSource = stripHandlePrefix(xc.Sourceref.Hlink)
			}
			gotSource := ptrStr(c.SourceHandle)
			if gotSource != wantSource {
				t.Errorf("citation %q SourceHandle: got %q, want %q", c.Handle, gotSource, wantSource)
			}
		}
		if count != len(xmlCitations) {
			t.Errorf("Citation count: got %d, want %d", count, len(xmlCitations))
		}
	})

	t.Run("Repository", func(t *testing.T) {
		xmlRepos := make(map[string]grampsxml.Repository)
		if xdb.Repositories != nil {
			for _, r := range xdb.Repositories.Repository {
				xmlRepos[stripHandlePrefix(r.Handle)] = r
			}
		}

		count := 0
		for r, err := range db.Repositories() {
			if err != nil {
				t.Fatalf("Repositories iterator: unexpected error: %v", err)
			}
			count++
			xr, ok := xmlRepos[r.Handle]
			if !ok {
				t.Errorf("repository %q: not found in XML", r.Handle)
				continue
			}
			wantID := ptrStr(xr.ID)
			if r.GrampsID != wantID {
				t.Errorf("repository %q GrampsID: got %q, want %q", r.Handle, r.GrampsID, wantID)
			}
			if r.Name != xr.Rname {
				t.Errorf("repository %q Name: got %q, want %q", r.Handle, r.Name, xr.Rname)
			}
		}
		if count != len(xmlRepos) {
			t.Errorf("Repository count: got %d, want %d", count, len(xmlRepos))
		}
	})

	t.Run("Note", func(t *testing.T) {
		xmlNotes := make(map[string]grampsxml.Note)
		if xdb.Notes != nil {
			for _, n := range xdb.Notes.Note {
				xmlNotes[stripHandlePrefix(n.Handle)] = n
			}
		}

		count := 0
		for n, err := range db.Notes() {
			if err != nil {
				t.Fatalf("Notes iterator: unexpected error: %v", err)
			}
			count++
			xn, ok := xmlNotes[n.Handle]
			if !ok {
				t.Errorf("note %q: not found in XML", n.Handle)
				continue
			}
			wantID := ptrStr(xn.ID)
			if n.GrampsID != wantID {
				t.Errorf("note %q GrampsID: got %q, want %q", n.Handle, n.GrampsID, wantID)
			}
			if n.Text.String != xn.Text {
				t.Errorf("note %q Text: got %q, want %q", n.Handle, n.Text.String, xn.Text)
			}
		}
		if count != len(xmlNotes) {
			t.Errorf("Note count: got %d, want %d", count, len(xmlNotes))
		}
	})

	t.Run("Media", func(t *testing.T) {
		xmlMedia := make(map[string]grampsxml.Object)
		if xdb.Objects != nil {
			for _, o := range xdb.Objects.Object {
				xmlMedia[stripHandlePrefix(o.Handle)] = o
			}
		}

		count := 0
		for m, err := range db.MediaObjects() {
			if err != nil {
				t.Fatalf("MediaObjects iterator: unexpected error: %v", err)
			}
			count++
			xm, ok := xmlMedia[m.Handle]
			if !ok {
				t.Errorf("media %q: not found in XML", m.Handle)
				continue
			}
			wantID := ptrStr(xm.ID)
			if m.GrampsID != wantID {
				t.Errorf("media %q GrampsID: got %q, want %q", m.Handle, m.GrampsID, wantID)
			}
			if m.Path != xm.File.Src {
				t.Errorf("media %q Path: got %q, want %q", m.Handle, m.Path, xm.File.Src)
			}
			if m.Mime != xm.File.Mime {
				t.Errorf("media %q Mime: got %q, want %q", m.Handle, m.Mime, xm.File.Mime)
			}
			if m.Desc != xm.File.Description {
				t.Errorf("media %q Desc: got %q, want %q", m.Handle, m.Desc, xm.File.Description)
			}
		}
		if count != len(xmlMedia) {
			t.Errorf("Media count: got %d, want %d", count, len(xmlMedia))
		}
	})

	t.Run("Place", func(t *testing.T) {
		xmlPlaces := make(map[string]grampsxml.Placeobj)
		if xdb.Places != nil {
			for _, p := range xdb.Places.Place {
				xmlPlaces[stripHandlePrefix(p.Handle)] = p
			}
		}

		count := 0
		for p, err := range db.Places() {
			if err != nil {
				t.Fatalf("Places iterator: unexpected error: %v", err)
			}
			count++
			xp, ok := xmlPlaces[p.Handle]
			if !ok {
				t.Errorf("place %q: not found in XML", p.Handle)
				continue
			}
			wantID := ptrStr(xp.ID)
			if p.GrampsID != wantID {
				t.Errorf("place %q GrampsID: got %q, want %q", p.Handle, p.GrampsID, wantID)
			}
			if p.Title != ptrStr(xp.Ptitle) {
				t.Errorf("place %q Title: got %q, want %q", p.Handle, p.Title, ptrStr(xp.Ptitle))
			}

			// Coordinates.
			wantLong := ""
			wantLat := ""
			if xp.Coord != nil {
				wantLong = xp.Coord.Long
				wantLat = xp.Coord.Lat
			}
			if p.Long != wantLong {
				t.Errorf("place %q Long: got %q, want %q", p.Handle, p.Long, wantLong)
			}
			if p.Lat != wantLat {
				t.Errorf("place %q Lat: got %q, want %q", p.Handle, p.Lat, wantLat)
			}

			// Primary name: first Pname in XML = Name.Value in DB.
			if len(xp.Pname) > 0 {
				wantName := xp.Pname[0].Value
				if p.Name.Value != wantName {
					t.Errorf("place %q Name: got %q, want %q", p.Handle, p.Name.Value, wantName)
				}
			}
		}
		if count != len(xmlPlaces) {
			t.Errorf("Place count: got %d, want %d", count, len(xmlPlaces))
		}
	})
}

// parseXML reads and parses a Gramps XML file (may be gzip-compressed).
func parseXML(t *testing.T, path string) grampsxml.Database {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("opening XML file: unexpected error: %v", err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		// Not gzip — reopen as plain XML.
		f.Close()
		f, err = os.Open(path)
		if err != nil {
			t.Fatalf("reopening XML file: unexpected error: %v", err)
		}
		defer f.Close()
		var db grampsxml.Database
		if err := xml.NewDecoder(f).Decode(&db); err != nil {
			t.Fatalf("decoding XML: unexpected error: %v", err)
		}
		return db
	}
	defer gr.Close()

	var db grampsxml.Database
	if err := xml.NewDecoder(gr).Decode(&db); err != nil {
		t.Fatalf("decoding gzipped XML: unexpected error: %v", err)
	}
	return db
}

// stripHandlePrefix removes the leading underscore that Gramps XML export
// adds to handle values. The SQLite database stores handles without this prefix.
func stripHandlePrefix(h string) string {
	return strings.TrimPrefix(h, "_")
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func xmlGenderToInt(g string) int {
	switch g {
	case "M":
		return gogramps.GenderMale
	case "F":
		return gogramps.GenderFemale
	case "U":
		return gogramps.GenderUnknown
	default:
		return gogramps.GenderOther
	}
}

func xmlPrimarySurname(surnames []grampsxml.Surname) string {
	for _, sn := range surnames {
		if sn.Prim == nil || *sn.Prim {
			return sn.Surname
		}
	}
	if len(surnames) > 0 {
		return surnames[0].Surname
	}
	return ""
}

func dbPrimarySurname(surnames []gogramps.Surname) string {
	for _, sn := range surnames {
		if sn.Primary {
			return sn.Surname
		}
	}
	if len(surnames) > 0 {
		return surnames[0].Surname
	}
	return ""
}

func mapsEqual(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

func keys(m map[string]bool) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}
