package gogramps

import "database/sql"

// createSchema creates the Gramps SQLite schema for JSON-based databases.
// This matches the schema created by Gramps Python's DBAPI._create_schema(json_data=True).
func createSchema(db *sql.DB) error {
	stmts := []string{
		// Primary object tables
		`CREATE TABLE person (
			handle VARCHAR(50) PRIMARY KEY NOT NULL,
			given_name TEXT,
			surname TEXT,
			json_data TEXT
		)`,
		`CREATE TABLE family (
			handle VARCHAR(50) PRIMARY KEY NOT NULL,
			json_data TEXT
		)`,
		`CREATE TABLE source (
			handle VARCHAR(50) PRIMARY KEY NOT NULL,
			json_data TEXT
		)`,
		`CREATE TABLE citation (
			handle VARCHAR(50) PRIMARY KEY NOT NULL,
			json_data TEXT
		)`,
		`CREATE TABLE event (
			handle VARCHAR(50) PRIMARY KEY NOT NULL,
			json_data TEXT
		)`,
		`CREATE TABLE media (
			handle VARCHAR(50) PRIMARY KEY NOT NULL,
			json_data TEXT
		)`,
		`CREATE TABLE place (
			handle VARCHAR(50) PRIMARY KEY NOT NULL,
			enclosed_by VARCHAR(50),
			json_data TEXT
		)`,
		`CREATE TABLE repository (
			handle VARCHAR(50) PRIMARY KEY NOT NULL,
			json_data TEXT
		)`,
		`CREATE TABLE note (
			handle VARCHAR(50) PRIMARY KEY NOT NULL,
			json_data TEXT
		)`,
		`CREATE TABLE tag (
			handle VARCHAR(50) PRIMARY KEY NOT NULL,
			json_data TEXT
		)`,

		// Secondary tables
		`CREATE TABLE reference (
			obj_handle VARCHAR(50),
			obj_class TEXT,
			ref_handle VARCHAR(50),
			ref_class TEXT
		)`,
		`CREATE TABLE name_group (
			name VARCHAR(50) PRIMARY KEY NOT NULL,
			grouping TEXT
		)`,
		`CREATE TABLE metadata (
			setting VARCHAR(50) PRIMARY KEY NOT NULL,
			json_data TEXT,
			value BLOB
		)`,
		`CREATE TABLE gender_stats (
			given_name TEXT,
			female INTEGER,
			male INTEGER,
			unknown INTEGER
		)`,

		// Secondary columns (added by _create_secondary_columns)
		`ALTER TABLE person ADD COLUMN gramps_id TEXT`,
		`ALTER TABLE person ADD COLUMN change INTEGER`,
		`ALTER TABLE person ADD COLUMN private INTEGER`,
		`ALTER TABLE family ADD COLUMN gramps_id TEXT`,
		`ALTER TABLE family ADD COLUMN change INTEGER`,
		`ALTER TABLE family ADD COLUMN private INTEGER`,
		`ALTER TABLE family ADD COLUMN father_handle VARCHAR(50)`,
		`ALTER TABLE family ADD COLUMN mother_handle VARCHAR(50)`,
		`ALTER TABLE event ADD COLUMN gramps_id TEXT`,
		`ALTER TABLE event ADD COLUMN change INTEGER`,
		`ALTER TABLE event ADD COLUMN private INTEGER`,
		`ALTER TABLE place ADD COLUMN gramps_id TEXT`,
		`ALTER TABLE place ADD COLUMN title TEXT`,
		`ALTER TABLE place ADD COLUMN change INTEGER`,
		`ALTER TABLE place ADD COLUMN private INTEGER`,
		`ALTER TABLE repository ADD COLUMN gramps_id TEXT`,
		`ALTER TABLE repository ADD COLUMN name TEXT`,
		`ALTER TABLE repository ADD COLUMN change INTEGER`,
		`ALTER TABLE repository ADD COLUMN private INTEGER`,
		`ALTER TABLE source ADD COLUMN gramps_id TEXT`,
		`ALTER TABLE source ADD COLUMN title TEXT`,
		`ALTER TABLE source ADD COLUMN change INTEGER`,
		`ALTER TABLE source ADD COLUMN private INTEGER`,
		`ALTER TABLE citation ADD COLUMN gramps_id TEXT`,
		`ALTER TABLE citation ADD COLUMN page TEXT`,
		`ALTER TABLE citation ADD COLUMN change INTEGER`,
		`ALTER TABLE citation ADD COLUMN private INTEGER`,
		`ALTER TABLE media ADD COLUMN gramps_id TEXT`,
		`ALTER TABLE media ADD COLUMN desc TEXT`,
		`ALTER TABLE media ADD COLUMN change INTEGER`,
		`ALTER TABLE media ADD COLUMN private INTEGER`,
		`ALTER TABLE note ADD COLUMN gramps_id TEXT`,
		`ALTER TABLE note ADD COLUMN change INTEGER`,
		`ALTER TABLE note ADD COLUMN private INTEGER`,
		`ALTER TABLE tag ADD COLUMN name TEXT`,
		`ALTER TABLE tag ADD COLUMN change INTEGER`,
		`ALTER TABLE tag ADD COLUMN priority INTEGER`,

		// Indices
		`CREATE INDEX person_gramps_id ON person(gramps_id)`,
		`CREATE INDEX person_surname ON person(surname)`,
		`CREATE INDEX person_given_name ON person(given_name)`,
		`CREATE INDEX source_title ON source(title)`,
		`CREATE INDEX source_gramps_id ON source(gramps_id)`,
		`CREATE INDEX citation_page ON citation(page)`,
		`CREATE INDEX citation_gramps_id ON citation(gramps_id)`,
		`CREATE INDEX media_desc ON media(desc)`,
		`CREATE INDEX media_gramps_id ON media(gramps_id)`,
		`CREATE INDEX place_title ON place(title)`,
		`CREATE INDEX place_enclosed_by ON place(enclosed_by)`,
		`CREATE INDEX place_gramps_id ON place(gramps_id)`,
		`CREATE INDEX tag_name ON tag(name)`,
		`CREATE INDEX reference_ref_handle ON reference(ref_handle)`,
		`CREATE INDEX family_gramps_id ON family(gramps_id)`,
		`CREATE INDEX event_gramps_id ON event(gramps_id)`,
		`CREATE INDEX repository_gramps_id ON repository(gramps_id)`,
		`CREATE INDEX note_gramps_id ON note(gramps_id)`,
		`CREATE INDEX reference_obj_handle ON reference(obj_handle)`,
	}

	stmts = append(stmts, schema23Tables()...)

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// schemaExists checks whether the database already has a schema.
func schemaExists(db *sql.DB) (bool, error) {
	var count int
	err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='person'").Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
