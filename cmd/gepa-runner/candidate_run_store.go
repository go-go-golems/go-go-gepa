package main

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type candidateRunRecord struct {
	RunID                    string
	TimestampMS              int64
	PluginID                 string
	PluginName               string
	PluginRegistryIdentifier string
	Profile                  string
	CandidateID              string
	ReflectionUsed           string
	TagsJSON                 string
	CandidateJSON            string
	InputJSON                string
	OutputJSON               string
	MetadataJSON             string
	ConfigJSON               string
	Status                   string
	ErrorMessage             string
}

func writeCandidateRunRecord(dbPath string, record candidateRunRecord) error {
	dbPath = strings.TrimSpace(dbPath)
	if dbPath == "" {
		return fmt.Errorf("candidate run recorder: db path is empty")
	}
	if err := ensureParentDir(dbPath); err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()

	if err := ensureCandidateRunTable(db); err != nil {
		return err
	}

	_, err = db.Exec(`
INSERT INTO gepa_candidate_runs (
  run_id,
  timestamp_ms,
  plugin_id,
  plugin_name,
  plugin_registry_identifier,
  profile,
  candidate_id,
  reflection_used,
  tags_json,
  candidate_json,
  input_json,
  output_json,
  metadata_json,
  config_json,
  status,
  error
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`,
		record.RunID,
		record.TimestampMS,
		nullableString(record.PluginID),
		nullableString(record.PluginName),
		nullableString(record.PluginRegistryIdentifier),
		nullableString(record.Profile),
		nullableString(record.CandidateID),
		nullableString(record.ReflectionUsed),
		nullableString(record.TagsJSON),
		record.CandidateJSON,
		record.InputJSON,
		record.OutputJSON,
		nullableString(record.MetadataJSON),
		nullableString(record.ConfigJSON),
		record.Status,
		nullableString(record.ErrorMessage),
	)
	return err
}

func ensureCandidateRunTable(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("candidate run recorder: db is nil")
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS gepa_candidate_runs (
  run_id TEXT PRIMARY KEY,
  timestamp_ms INTEGER NOT NULL,
  plugin_id TEXT,
  plugin_name TEXT,
  plugin_registry_identifier TEXT,
  profile TEXT,
  candidate_id TEXT,
  reflection_used TEXT,
  tags_json TEXT,
  candidate_json TEXT NOT NULL,
  input_json TEXT NOT NULL,
  output_json TEXT NOT NULL,
  metadata_json TEXT,
  config_json TEXT,
  status TEXT NOT NULL,
  error TEXT
)`,
		`CREATE INDEX IF NOT EXISTS idx_gepa_candidate_runs_time ON gepa_candidate_runs (timestamp_ms DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_gepa_candidate_runs_candidate ON gepa_candidate_runs (candidate_id, timestamp_ms DESC)`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}
