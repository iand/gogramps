package gogramps

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// metadataDoc is the typed wrapper used by Gramps JSON serializer for metadata values.
// Format: {"type": "str", "value": "21"} or {"type": "int", "value": 42}
type metadataDoc struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

func getMetadata(db *sql.DB, key string) (any, error) {
	var jsonData sql.NullString
	err := db.QueryRow("SELECT json_data FROM metadata WHERE setting = ?", key).Scan(&jsonData)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !jsonData.Valid || jsonData.String == "" {
		return nil, nil
	}
	var doc metadataDoc
	if err := json.Unmarshal([]byte(jsonData.String), &doc); err != nil {
		return nil, err
	}
	return doc.Value, nil
}

func setMetadata(db *sql.DB, key string, value any) error {
	doc := metadataDoc{
		Type:  fmt.Sprintf("%T", value),
		Value: value,
	}
	// Use Go type name to approximate Python type names for common cases.
	switch value.(type) {
	case string:
		doc.Type = "str"
	case int, int64:
		doc.Type = "int"
	case float64:
		doc.Type = "float"
	case bool:
		doc.Type = "bool"
	case []any:
		doc.Type = "list"
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	var exists int
	err = db.QueryRow("SELECT 1 FROM metadata WHERE setting = ?", key).Scan(&exists)
	if err == sql.ErrNoRows {
		_, err = db.Exec("INSERT INTO metadata (setting, json_data) VALUES (?, ?)", key, string(data))
		return err
	}
	if err != nil {
		return err
	}
	_, err = db.Exec("UPDATE metadata SET json_data = ? WHERE setting = ?", string(data), key)
	return err
}
