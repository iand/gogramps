//go:build gramps_schema23

package gogramps

import "iter"

const maxSupportedSchemaVersion = 23

// DNAProviderType values.
const (
	DNAProviderUnknown    = -1
	DNAProviderCustom     = 0
	DNAProviderAncestry   = 1
	DNAProvider23AndMe    = 2
	DNAProviderMyHeritage = 3
	DNAProviderFTDNA      = 4
	DNAProviderGEDmatch   = 5
	DNAProviderLivingDNA  = 6
)

// DNATestType values.
const (
	DNATestUnknown   = -1
	DNATestCustom    = 0
	DNATestAutosomal = 1
	DNATestYDNA12    = 2
	DNATestYDNA37    = 3
	DNATestYDNA67    = 4
	DNATestYDNA111   = 5
	DNATestBigY      = 6
	DNATestMtDNAHVR1 = 7
	DNATestMtDNAFull = 8
)

// DNAGenomeBuildType values.
const (
	DNAGenomeBuildUnknown = -1
	DNAGenomeBuildCustom  = 0
	DNAGenomeBuildGRCh37  = 1
	DNAGenomeBuildGRCh38  = 2
)

// DNASegment phase values.
const (
	DNAPhaseUnassigned = 0
	DNAPhaseUnknown    = 1
	DNAPhaseMaternal   = 2
	DNAPhasePaternal   = 3
)

// DNASegment IBD state values.
const (
	DNAIBDUnknown              = 0
	DNAIBDHalfIdenticalRegion  = 1
	DNAIBDFullyIdenticalRegion = 2
)

// SharedAncestor confidence values.
const (
	SharedAncestorPossible  = 0
	SharedAncestorProbable  = 1
	SharedAncestorConfirmed = 2
	SharedAncestorRejected  = 3
)

// DNAAttribute is a typed key/value attribute on a DNATest or DNAMatch.
type DNAAttribute struct {
	Class   string     `json:"_class"`
	Private bool       `json:"private"`
	Type    GrampsType `json:"type"`
	Value   string     `json:"value"`
}

// DNASegment is a single shared chromosomal segment within a DNAMatch.
type DNASegment struct {
	Class      string  `json:"_class"`
	Chromosome string  `json:"chromosome"`
	StartBP    int     `json:"start_bp"`
	EndBP      int     `json:"end_bp"`
	StartRSID  *string `json:"start_rsid"`
	EndRSID    *string `json:"end_rsid"`
	SharedCM   float64 `json:"shared_cm"`
	SNPCount   int     `json:"snp_count"`
	Phase      int     `json:"phase"`
	IBDState   int     `json:"ibd_state"`
}

// SharedAncestor records a hypothesis or confirmed MRCA for a DNAMatch.
type SharedAncestor struct {
	Class        string   `json:"_class"`
	PersonHandle *string  `json:"person_handle"`
	Description  string   `json:"description"`
	Confidence   int      `json:"confidence"`
	CitationList []string `json:"citation_list"`
	NoteList     []string `json:"note_list"`
}

// DNATest represents one DNA kit for one person at one provider.
type DNATest struct {
	Class         string         `json:"_class"`
	Handle        string         `json:"handle"`
	GrampsID      string         `json:"gramps_id"`
	PersonHandle  *string        `json:"person_handle"`
	AccountName   string         `json:"account_name"`
	Provider      GrampsType     `json:"provider"`
	KitID         string         `json:"kit_id"`
	TestType      GrampsType     `json:"test_type"`
	GenomeBuild   GrampsType     `json:"genome_build"`
	Date          *Date          `json:"date"`
	Haplogroup    string         `json:"haplogroup"`
	CitationList  []string       `json:"citation_list"`
	NoteList      []string       `json:"note_list"`
	MediaList     []MediaRef     `json:"media_list"`
	AttributeList []DNAAttribute `json:"attribute_list"`
	Change        int            `json:"change"`
	TagList       []string       `json:"tag_list"`
	Private       bool           `json:"private"`
}

// DNAMatch represents a pairwise DNA match between two kits.
type DNAMatch struct {
	Class                 string           `json:"_class"`
	Handle                string           `json:"handle"`
	GrampsID              string           `json:"gramps_id"`
	SubjectTestHandle     *string          `json:"subject_test_handle"`
	MatchTestHandle       *string          `json:"match_test_handle"`
	SharedCM              float64          `json:"shared_cm"`
	PercentShared         float64          `json:"percent_shared"`
	SegmentCount          int              `json:"segment_count"`
	LargestSegmentCM      float64          `json:"largest_segment_cm"`
	PredictedRelationship string           `json:"predicted_relationship"`
	PredictedGenerations  *float64         `json:"predicted_generations"`
	SharedAncestorList    []SharedAncestor `json:"shared_ancestor_list"`
	SegmentList           []DNASegment     `json:"segment_list"`
	CitationList          []string         `json:"citation_list"`
	NoteList              []string         `json:"note_list"`
	MediaList             []MediaRef       `json:"media_list"`
	AttributeList         []DNAAttribute   `json:"attribute_list"`
	Change                int              `json:"change"`
	TagList               []string         `json:"tag_list"`
	Private               bool             `json:"private"`
}

// DNATest operations.

func (d *Database) GetDNATest(handle string) (*DNATest, error) {
	return get[DNATest](d, "dnatest", handle)
}

func (d *Database) GetDNATestByGrampsID(id string) (*DNATest, error) {
	return getByGrampsID[DNATest](d, "dnatest", id)
}

func (d *Database) DNATests() iter.Seq2[*DNATest, error] {
	return iterAll[DNATest](d, "dnatest")
}

func (d *Database) AddDNATest(t *DNATest) error {
	t.Class = "DNATest"
	return d.addObject("dnatest", t.Handle, t, dnaTestSecondary(t))
}

func (d *Database) UpdateDNATest(t *DNATest) error {
	t.Class = "DNATest"
	return d.updateObject("dnatest", t.Handle, t, dnaTestSecondary(t))
}

func (d *Database) DeleteDNATest(handle string) error {
	return d.deleteObject("dnatest", handle)
}

// DNAMatch operations.

func (d *Database) GetDNAMatch(handle string) (*DNAMatch, error) {
	return get[DNAMatch](d, "dnamatch", handle)
}

func (d *Database) GetDNAMatchByGrampsID(id string) (*DNAMatch, error) {
	return getByGrampsID[DNAMatch](d, "dnamatch", id)
}

func (d *Database) DNAMatches() iter.Seq2[*DNAMatch, error] {
	return iterAll[DNAMatch](d, "dnamatch")
}

func (d *Database) AddDNAMatch(m *DNAMatch) error {
	m.Class = "DNAMatch"
	return d.addObject("dnamatch", m.Handle, m, dnaMatchSecondary(m))
}

func (d *Database) UpdateDNAMatch(m *DNAMatch) error {
	m.Class = "DNAMatch"
	return d.updateObject("dnamatch", m.Handle, m, dnaMatchSecondary(m))
}

func (d *Database) DeleteDNAMatch(handle string) error {
	return d.deleteObject("dnamatch", handle)
}

func dnaTestSecondary(t *DNATest) secondaryValues {
	return secondaryValues{
		sets:   []string{"gramps_id", "person_handle", "account_name", "kit_id", "haplogroup", "change", "private"},
		values: []any{t.GrampsID, t.PersonHandle, t.AccountName, t.KitID, t.Haplogroup, t.Change, boolToInt(t.Private)},
	}
}

func dnaMatchSecondary(m *DNAMatch) secondaryValues {
	return secondaryValues{
		sets:   []string{"gramps_id", "subject_test_handle", "match_test_handle", "shared_cm", "percent_shared", "segment_count", "largest_segment_cm", "predicted_relationship", "predicted_generations", "change", "private"},
		values: []any{m.GrampsID, m.SubjectTestHandle, m.MatchTestHandle, m.SharedCM, m.PercentShared, m.SegmentCount, m.LargestSegmentCM, m.PredictedRelationship, m.PredictedGenerations, m.Change, boolToInt(m.Private)},
	}
}

func schema23Tables() []string {
	return []string{
		`CREATE TABLE dnatest (
			handle VARCHAR(50) PRIMARY KEY NOT NULL,
			json_data TEXT
		)`,
		`CREATE TABLE dnamatch (
			handle VARCHAR(50) PRIMARY KEY NOT NULL,
			json_data TEXT
		)`,
		`ALTER TABLE dnatest ADD COLUMN gramps_id TEXT`,
		`ALTER TABLE dnatest ADD COLUMN person_handle VARCHAR(50)`,
		`ALTER TABLE dnatest ADD COLUMN account_name TEXT`,
		`ALTER TABLE dnatest ADD COLUMN kit_id TEXT`,
		`ALTER TABLE dnatest ADD COLUMN haplogroup TEXT`,
		`ALTER TABLE dnatest ADD COLUMN change INTEGER`,
		`ALTER TABLE dnatest ADD COLUMN private INTEGER`,
		`ALTER TABLE dnamatch ADD COLUMN gramps_id TEXT`,
		`ALTER TABLE dnamatch ADD COLUMN subject_test_handle VARCHAR(50)`,
		`ALTER TABLE dnamatch ADD COLUMN match_test_handle VARCHAR(50)`,
		`ALTER TABLE dnamatch ADD COLUMN shared_cm REAL`,
		`ALTER TABLE dnamatch ADD COLUMN percent_shared REAL`,
		`ALTER TABLE dnamatch ADD COLUMN segment_count INTEGER`,
		`ALTER TABLE dnamatch ADD COLUMN largest_segment_cm REAL`,
		`ALTER TABLE dnamatch ADD COLUMN predicted_relationship TEXT`,
		`ALTER TABLE dnamatch ADD COLUMN predicted_generations REAL`,
		`ALTER TABLE dnamatch ADD COLUMN change INTEGER`,
		`ALTER TABLE dnamatch ADD COLUMN private INTEGER`,
		`CREATE INDEX dnatest_gramps_id ON dnatest(gramps_id)`,
		`CREATE INDEX dnamatch_gramps_id ON dnamatch(gramps_id)`,
	}
}
