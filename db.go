package gogramps

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// ErrReadOnly is returned when a write operation is attempted on a read-only database.
var ErrReadOnly = errors.New("database is read-only")

// ErrUnsupportedSchema is returned when opening a database with an unsupported schema version.
type ErrUnsupportedSchema struct {
	Version int
}

func (e *ErrUnsupportedSchema) Error() string {
	return fmt.Sprintf("unsupported schema version %d (supported: %d-%d)", e.Version, minSupportedSchemaVersion, maxSupportedSchemaVersion)
}

const minSupportedSchemaVersion = 21

const (
	sqliteDBFile  = "sqlite.db"
	backendFile   = "database.txt"
	nameFile      = "name.txt"
	backendString = "sqlite"
)

// Database represents an open Gramps database.
type Database struct {
	db       *sql.DB
	dir      string
	locked   bool
	readonly bool
	version  int
}

// SchemaVersion returns the schema version number of the database.
func (d *Database) SchemaVersion() int { return d.version }

// Open opens an existing Gramps database directory and acquires a lock.
func Open(path string) (*Database, error) {
	if err := checkBackend(path); err != nil {
		return nil, err
	}

	if err := writeLockFile(path); err != nil {
		return nil, err
	}

	db, err := openSQLite(path)
	if err != nil {
		removeLockFile(path)
		return nil, err
	}

	exists, err := schemaExists(db)
	if err != nil {
		db.Close()
		removeLockFile(path)
		return nil, fmt.Errorf("checking schema: %w", err)
	}
	if !exists {
		db.Close()
		removeLockFile(path)
		return nil, fmt.Errorf("database at %s has no schema", path)
	}

	version, err := readSchemaVersion(db)
	if err != nil {
		db.Close()
		removeLockFile(path)
		return nil, err
	}
	if err := validateSchemaVersion(version); err != nil {
		db.Close()
		removeLockFile(path)
		return nil, err
	}

	return &Database{db: db, dir: path, locked: true, version: version}, nil
}

// OpenReadOnly opens an existing Gramps database directory without locking.
func OpenReadOnly(path string) (*Database, error) {
	if err := checkBackend(path); err != nil {
		return nil, err
	}

	db, err := openSQLite(path)
	if err != nil {
		return nil, err
	}

	exists, err := schemaExists(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("checking schema: %w", err)
	}
	if !exists {
		db.Close()
		return nil, fmt.Errorf("database at %s has no schema", path)
	}

	version, err := readSchemaVersion(db)
	if err != nil {
		db.Close()
		return nil, err
	}
	if err := validateSchemaVersion(version); err != nil {
		db.Close()
		return nil, err
	}

	return &Database{db: db, dir: path, readonly: true, version: version}, nil
}

// Create creates a new Gramps database directory with the given name.
func Create(path string, name string) (*Database, error) {
	if err := os.MkdirAll(path, 0o777); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	// Write backend identifier.
	if err := os.WriteFile(filepath.Join(path, backendFile), []byte(backendString), 0o666); err != nil {
		return nil, fmt.Errorf("writing backend file: %w", err)
	}

	// Write name file.
	if err := os.WriteFile(filepath.Join(path, nameFile), []byte(name), 0o666); err != nil {
		return nil, fmt.Errorf("writing name file: %w", err)
	}

	// Acquire lock.
	if err := writeLockFile(path); err != nil {
		return nil, err
	}

	db, err := openSQLite(path)
	if err != nil {
		removeLockFile(path)
		return nil, err
	}

	// Create schema.
	if err := createSchema(db); err != nil {
		db.Close()
		removeLockFile(path)
		return nil, fmt.Errorf("creating schema: %w", err)
	}

	// Set schema version.
	if err := setMetadata(db, "version", fmt.Sprintf("%d", maxSupportedSchemaVersion)); err != nil {
		db.Close()
		removeLockFile(path)
		return nil, fmt.Errorf("setting schema version: %w", err)
	}

	return &Database{db: db, dir: path, locked: true, version: maxSupportedSchemaVersion}, nil
}

// Close closes the database and releases the lock if held.
func (d *Database) Close() error {
	var errs []error
	if d.db != nil {
		errs = append(errs, d.db.Close())
	}
	if d.locked {
		errs = append(errs, removeLockFile(d.dir))
	}
	return errors.Join(errs...)
}

// Dir returns the database directory path.
func (d *Database) Dir() string { return d.dir }

// readSchemaVersion reads the schema version number from the metadata table.
// It handles the legacy blob format (schema < 21) and the JSON format (schema >= 21).
func readSchemaVersion(db *sql.DB) (int, error) {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('metadata') WHERE name='json_data'").Scan(&count); err != nil {
		return 0, fmt.Errorf("reading metadata schema: %w", err)
	}

	if count == 0 {
		// Legacy schema: version is pickle-encoded in the value BLOB column.
		var blob []byte
		if err := db.QueryRow("SELECT value FROM metadata WHERE setting = 'version'").Scan(&blob); err != nil {
			return 0, fmt.Errorf("reading schema version: %w", err)
		}
		return extractVersionFromPickle(blob), nil
	}

	v, err := getMetadata(db, "version")
	if err != nil {
		return 0, fmt.Errorf("reading schema version: %w", err)
	}
	if v == nil {
		return minSupportedSchemaVersion, nil
	}
	vs, ok := v.(string)
	if !ok {
		return 0, fmt.Errorf("unexpected schema version type: %T", v)
	}
	var version int
	if _, err := fmt.Sscanf(vs, "%d", &version); err != nil {
		return 0, fmt.Errorf("parsing schema version %q: %w", vs, err)
	}
	return version, nil
}

// validateSchemaVersion returns ErrUnsupportedSchema if v is outside the supported range.
func validateSchemaVersion(v int) error {
	if v < minSupportedSchemaVersion || v > maxSupportedSchemaVersion {
		return &ErrUnsupportedSchema{Version: v}
	}
	return nil
}

// extractVersionFromPickle extracts a version integer from a pickle-encoded blob.
// Gramps serialises the version as a Python string; the digits appear as ASCII near the end.
func extractVersionFromPickle(blob []byte) int {
	for i := len(blob) - 1; i >= 0; i-- {
		if blob[i] >= '0' && blob[i] <= '9' {
			j := i
			for j > 0 && blob[j-1] >= '0' && blob[j-1] <= '9' {
				j--
			}
			var version int
			fmt.Sscanf(string(blob[j:i+1]), "%d", &version)
			return version
		}
	}
	return 0
}

func checkBackend(path string) error {
	data, err := os.ReadFile(filepath.Join(path, backendFile))
	if err != nil {
		return fmt.Errorf("reading backend file: %w", err)
	}
	if string(data) != backendString {
		return fmt.Errorf("unsupported backend: %q", string(data))
	}
	return nil
}

func openSQLite(dir string) (*sql.DB, error) {
	dbPath := filepath.Join(dir, sqliteDBFile)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening SQLite database: %w", err)
	}
	// Enable WAL mode for better concurrency.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting WAL mode: %w", err)
	}
	return db, nil
}

// GetMetadata retrieves a metadata value by key.
func (d *Database) GetMetadata(key string) (any, error) {
	return getMetadata(d.db, key)
}

// SetMetadata sets a metadata value.
func (d *Database) SetMetadata(key string, value any) error {
	if d.readonly {
		return ErrReadOnly
	}
	return setMetadata(d.db, key, value)
}

// get retrieves a single object from the specified table by handle.
func get[T any](d *Database, table, handle string) (*T, error) {
	var jsonData string
	err := d.db.QueryRow(
		fmt.Sprintf("SELECT json_data FROM %s WHERE handle = ?", table),
		handle,
	).Scan(&jsonData)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var obj T
	if err := json.Unmarshal([]byte(jsonData), &obj); err != nil {
		return nil, fmt.Errorf("unmarshalling %s: %w", table, err)
	}
	return &obj, nil
}

// getByGrampsID retrieves a single object by its gramps_id secondary column.
func getByGrampsID[T any](d *Database, table, grampsID string) (*T, error) {
	var jsonData string
	err := d.db.QueryRow(
		fmt.Sprintf("SELECT json_data FROM %s WHERE gramps_id = ?", table),
		grampsID,
	).Scan(&jsonData)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var obj T
	if err := json.Unmarshal([]byte(jsonData), &obj); err != nil {
		return nil, fmt.Errorf("unmarshalling %s: %w", table, err)
	}
	return &obj, nil
}

// iterAll returns an iterator over all objects in a table.
func iterAll[T any](d *Database, table string) iter.Seq2[*T, error] {
	return func(yield func(*T, error) bool) {
		rows, err := d.db.Query(
			fmt.Sprintf("SELECT json_data FROM %s", table),
		)
		if err != nil {
			yield(nil, err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var jsonData string
			if err := rows.Scan(&jsonData); err != nil {
				if !yield(nil, err) {
					return
				}
				continue
			}
			var obj T
			if err := json.Unmarshal([]byte(jsonData), &obj); err != nil {
				if !yield(nil, fmt.Errorf("unmarshalling %s: %w", table, err)) {
					return
				}
				continue
			}
			if !yield(&obj, nil) {
				return
			}
		}
		if err := rows.Err(); err != nil {
			yield(nil, err)
		}
	}
}

// Person operations

func (d *Database) GetPerson(handle string) (*Person, error) {
	return get[Person](d, "person", handle)
}

func (d *Database) GetPersonByGrampsID(id string) (*Person, error) {
	return getByGrampsID[Person](d, "person", id)
}

func (d *Database) People() iter.Seq2[*Person, error] {
	return iterAll[Person](d, "person")
}

func (d *Database) AddPerson(p *Person) error {
	p.Class = "Person"
	return d.addObject("person", p.Handle, p, personSecondary(p))
}

func (d *Database) UpdatePerson(p *Person) error {
	p.Class = "Person"
	return d.updateObject("person", p.Handle, p, personSecondary(p))
}

func (d *Database) DeletePerson(handle string) error {
	return d.deleteObject("person", handle)
}

// Family operations

func (d *Database) GetFamily(handle string) (*Family, error) {
	return get[Family](d, "family", handle)
}

func (d *Database) GetFamilyByGrampsID(id string) (*Family, error) {
	return getByGrampsID[Family](d, "family", id)
}

func (d *Database) Families() iter.Seq2[*Family, error] {
	return iterAll[Family](d, "family")
}

func (d *Database) AddFamily(f *Family) error {
	f.Class = "Family"
	return d.addObject("family", f.Handle, f, familySecondary(f))
}

func (d *Database) UpdateFamily(f *Family) error {
	f.Class = "Family"
	return d.updateObject("family", f.Handle, f, familySecondary(f))
}

func (d *Database) DeleteFamily(handle string) error {
	return d.deleteObject("family", handle)
}

// Event operations

func (d *Database) GetEvent(handle string) (*Event, error) {
	return get[Event](d, "event", handle)
}

func (d *Database) GetEventByGrampsID(id string) (*Event, error) {
	return getByGrampsID[Event](d, "event", id)
}

func (d *Database) Events() iter.Seq2[*Event, error] {
	return iterAll[Event](d, "event")
}

func (d *Database) AddEvent(e *Event) error {
	e.Class = "Event"
	return d.addObject("event", e.Handle, e, eventSecondary(e))
}

func (d *Database) UpdateEvent(e *Event) error {
	e.Class = "Event"
	return d.updateObject("event", e.Handle, e, eventSecondary(e))
}

func (d *Database) DeleteEvent(handle string) error {
	return d.deleteObject("event", handle)
}

// Place operations

func (d *Database) GetPlace(handle string) (*Place, error) {
	return get[Place](d, "place", handle)
}

func (d *Database) GetPlaceByGrampsID(id string) (*Place, error) {
	return getByGrampsID[Place](d, "place", id)
}

func (d *Database) Places() iter.Seq2[*Place, error] {
	return iterAll[Place](d, "place")
}

func (d *Database) AddPlace(p *Place) error {
	p.Class = "Place"
	return d.addObject("place", p.Handle, p, placeSecondary(p))
}

func (d *Database) UpdatePlace(p *Place) error {
	p.Class = "Place"
	return d.updateObject("place", p.Handle, p, placeSecondary(p))
}

func (d *Database) DeletePlace(handle string) error {
	return d.deleteObject("place", handle)
}

// Source operations

func (d *Database) GetSource(handle string) (*Source, error) {
	return get[Source](d, "source", handle)
}

func (d *Database) GetSourceByGrampsID(id string) (*Source, error) {
	return getByGrampsID[Source](d, "source", id)
}

func (d *Database) Sources() iter.Seq2[*Source, error] {
	return iterAll[Source](d, "source")
}

func (d *Database) AddSource(s *Source) error {
	s.Class = "Source"
	return d.addObject("source", s.Handle, s, sourceSecondary(s))
}

func (d *Database) UpdateSource(s *Source) error {
	s.Class = "Source"
	return d.updateObject("source", s.Handle, s, sourceSecondary(s))
}

func (d *Database) DeleteSource(handle string) error {
	return d.deleteObject("source", handle)
}

// Citation operations

func (d *Database) GetCitation(handle string) (*Citation, error) {
	return get[Citation](d, "citation", handle)
}

func (d *Database) GetCitationByGrampsID(id string) (*Citation, error) {
	return getByGrampsID[Citation](d, "citation", id)
}

func (d *Database) Citations() iter.Seq2[*Citation, error] {
	return iterAll[Citation](d, "citation")
}

func (d *Database) AddCitation(c *Citation) error {
	c.Class = "Citation"
	return d.addObject("citation", c.Handle, c, citationSecondary(c))
}

func (d *Database) UpdateCitation(c *Citation) error {
	c.Class = "Citation"
	return d.updateObject("citation", c.Handle, c, citationSecondary(c))
}

func (d *Database) DeleteCitation(handle string) error {
	return d.deleteObject("citation", handle)
}

// Repository operations

func (d *Database) GetRepository(handle string) (*Repository, error) {
	return get[Repository](d, "repository", handle)
}

func (d *Database) GetRepositoryByGrampsID(id string) (*Repository, error) {
	return getByGrampsID[Repository](d, "repository", id)
}

func (d *Database) Repositories() iter.Seq2[*Repository, error] {
	return iterAll[Repository](d, "repository")
}

func (d *Database) AddRepository(r *Repository) error {
	r.Class = "Repository"
	return d.addObject("repository", r.Handle, r, repositorySecondary(r))
}

func (d *Database) UpdateRepository(r *Repository) error {
	r.Class = "Repository"
	return d.updateObject("repository", r.Handle, r, repositorySecondary(r))
}

func (d *Database) DeleteRepository(handle string) error {
	return d.deleteObject("repository", handle)
}

// Note operations

func (d *Database) GetNote(handle string) (*Note, error) {
	return get[Note](d, "note", handle)
}

func (d *Database) GetNoteByGrampsID(id string) (*Note, error) {
	return getByGrampsID[Note](d, "note", id)
}

func (d *Database) Notes() iter.Seq2[*Note, error] {
	return iterAll[Note](d, "note")
}

func (d *Database) AddNote(n *Note) error {
	n.Class = "Note"
	return d.addObject("note", n.Handle, n, noteSecondary(n))
}

func (d *Database) UpdateNote(n *Note) error {
	n.Class = "Note"
	return d.updateObject("note", n.Handle, n, noteSecondary(n))
}

func (d *Database) DeleteNote(handle string) error {
	return d.deleteObject("note", handle)
}

// Media operations

func (d *Database) GetMedia(handle string) (*Media, error) {
	return get[Media](d, "media", handle)
}

func (d *Database) GetMediaByGrampsID(id string) (*Media, error) {
	return getByGrampsID[Media](d, "media", id)
}

func (d *Database) MediaObjects() iter.Seq2[*Media, error] {
	return iterAll[Media](d, "media")
}

func (d *Database) AddMedia(m *Media) error {
	m.Class = "Media"
	return d.addObject("media", m.Handle, m, mediaSecondary(m))
}

func (d *Database) UpdateMedia(m *Media) error {
	m.Class = "Media"
	return d.updateObject("media", m.Handle, m, mediaSecondary(m))
}

func (d *Database) DeleteMedia(handle string) error {
	return d.deleteObject("media", handle)
}

// Tag operations

func (d *Database) GetTag(handle string) (*Tag, error) {
	return get[Tag](d, "tag", handle)
}

func (d *Database) GetTagByName(name string) (*Tag, error) {
	var jsonData string
	err := d.db.QueryRow("SELECT json_data FROM tag WHERE name = ?", name).Scan(&jsonData)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var tag Tag
	if err := json.Unmarshal([]byte(jsonData), &tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

func (d *Database) Tags() iter.Seq2[*Tag, error] {
	return iterAll[Tag](d, "tag")
}

func (d *Database) AddTag(t *Tag) error {
	t.Class = "Tag"
	return d.addObject("tag", t.Handle, t, tagSecondary(t))
}

func (d *Database) UpdateTag(t *Tag) error {
	t.Class = "Tag"
	return d.updateObject("tag", t.Handle, t, tagSecondary(t))
}

func (d *Database) DeleteTag(handle string) error {
	return d.deleteObject("tag", handle)
}

// addObject inserts a new object into the specified table.
func (d *Database) addObject(table, handle string, obj any, secondary secondaryValues) error {
	if d.readonly {
		return ErrReadOnly
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("marshalling %s: %w", table, err)
	}
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		fmt.Sprintf("INSERT INTO %s (handle, json_data) VALUES (?, ?)", table),
		handle, string(data),
	)
	if err != nil {
		return fmt.Errorf("inserting %s: %w", table, err)
	}

	if err := secondary.update(tx, table, handle); err != nil {
		return err
	}

	return tx.Commit()
}

// updateObject updates an existing object in the specified table.
func (d *Database) updateObject(table, handle string, obj any, secondary secondaryValues) error {
	if d.readonly {
		return ErrReadOnly
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("marshalling %s: %w", table, err)
	}
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		fmt.Sprintf("UPDATE %s SET json_data = ? WHERE handle = ?", table),
		string(data), handle,
	)
	if err != nil {
		return fmt.Errorf("updating %s: %w", table, err)
	}

	if err := secondary.update(tx, table, handle); err != nil {
		return err
	}

	return tx.Commit()
}

// deleteObject removes an object and its backlinks from the specified table.
func (d *Database) deleteObject(table, handle string) error {
	if d.readonly {
		return ErrReadOnly
	}
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE handle = ?", table), handle)
	if err != nil {
		return err
	}

	// Remove backlinks.
	_, err = tx.Exec("DELETE FROM reference WHERE obj_handle = ?", handle)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// secondaryValues holds the secondary column values to be updated alongside the json_data.
type secondaryValues struct {
	sets   []string
	values []any
}

func (s secondaryValues) update(tx *sql.Tx, table, handle string) error {
	if len(s.sets) == 0 {
		return nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "UPDATE %s SET ", table)
	for i, set := range s.sets {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(set)
		b.WriteString(" = ?")
	}
	b.WriteString(" WHERE handle = ?")
	args := append(s.values, handle)
	_, err := tx.Exec(b.String(), args...)
	return err
}

func personSecondary(p *Person) secondaryValues {
	givenName := p.PrimaryName.FirstName
	surname := ""
	for _, sn := range p.PrimaryName.SurnameList {
		if sn.Primary {
			surname = sn.Surname
			break
		}
	}
	return secondaryValues{
		sets:   []string{"gramps_id", "given_name", "surname", "change", "private"},
		values: []any{p.GrampsID, givenName, surname, p.Change, boolToInt(p.Private)},
	}
}

func familySecondary(f *Family) secondaryValues {
	var fatherHandle, motherHandle any
	if f.FatherHandle != nil {
		fatherHandle = *f.FatherHandle
	}
	if f.MotherHandle != nil {
		motherHandle = *f.MotherHandle
	}
	return secondaryValues{
		sets:   []string{"gramps_id", "father_handle", "mother_handle", "change", "private"},
		values: []any{f.GrampsID, fatherHandle, motherHandle, f.Change, boolToInt(f.Private)},
	}
}

func eventSecondary(e *Event) secondaryValues {
	return secondaryValues{
		sets:   []string{"gramps_id", "change", "private"},
		values: []any{e.GrampsID, e.Change, boolToInt(e.Private)},
	}
}

func placeSecondary(p *Place) secondaryValues {
	enclosedBy := ""
	if len(p.PlaceRefList) > 0 {
		enclosedBy = p.PlaceRefList[0].Ref
	}
	return secondaryValues{
		sets:   []string{"gramps_id", "title", "enclosed_by", "change", "private"},
		values: []any{p.GrampsID, p.Title, enclosedBy, p.Change, boolToInt(p.Private)},
	}
}

func sourceSecondary(s *Source) secondaryValues {
	return secondaryValues{
		sets:   []string{"gramps_id", "title", "change", "private"},
		values: []any{s.GrampsID, s.Title, s.Change, boolToInt(s.Private)},
	}
}

func citationSecondary(c *Citation) secondaryValues {
	return secondaryValues{
		sets:   []string{"gramps_id", "page", "change", "private"},
		values: []any{c.GrampsID, c.Page, c.Change, boolToInt(c.Private)},
	}
}

func repositorySecondary(r *Repository) secondaryValues {
	return secondaryValues{
		sets:   []string{"gramps_id", "name", "change", "private"},
		values: []any{r.GrampsID, r.Name, r.Change, boolToInt(r.Private)},
	}
}

func noteSecondary(n *Note) secondaryValues {
	return secondaryValues{
		sets:   []string{"gramps_id", "change", "private"},
		values: []any{n.GrampsID, n.Change, boolToInt(n.Private)},
	}
}

func mediaSecondary(m *Media) secondaryValues {
	return secondaryValues{
		sets:   []string{"gramps_id", "desc", "change", "private"},
		values: []any{m.GrampsID, m.Desc, m.Change, boolToInt(m.Private)},
	}
}

func tagSecondary(t *Tag) secondaryValues {
	return secondaryValues{
		sets:   []string{"name", "change", "priority"},
		values: []any{t.Name, t.Change, t.Priority},
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
