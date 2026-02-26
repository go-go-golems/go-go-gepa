package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	gepaGeneratedDatasetsTable    = "gepa_generated_datasets"
	gepaGeneratedDatasetRowsTable = "gepa_generated_dataset_rows"
)

type generatedDatasetRow struct {
	RowIndex int
	Row      map[string]any
	Metadata map[string]any
}

type generatedDatasetRecord struct {
	DatasetID                string
	Name                     string
	RequestedCount           int
	GeneratedCount           int
	Seed                     int64
	PluginID                 string
	PluginName               string
	PluginRegistryIdentifier string
	ConfigAPIVersion         string
	ConfigJSON               string
	CreatedAtMS              int64
}

type generatedDatasetWriteResult struct {
	DatasetID      string
	RowsWritten    int
	OutputJSONL    string
	OutputMetadata string
	DBPath         string
}

func writeGeneratedDatasetFiles(outputDir, outputFileStem string, record generatedDatasetRecord, rows []generatedDatasetRow) (generatedDatasetWriteResult, error) {
	if strings.TrimSpace(outputDir) == "" {
		return generatedDatasetWriteResult{}, fmt.Errorf("output dir is empty")
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return generatedDatasetWriteResult{}, err
	}
	stem := strings.TrimSpace(outputFileStem)
	if stem == "" {
		stem = "dataset"
	}
	jsonlPath := filepath.Join(outputDir, stem+".jsonl")
	metaPath := filepath.Join(outputDir, stem+".metadata.json")

	f, err := os.Create(jsonlPath)
	if err != nil {
		return generatedDatasetWriteResult{}, err
	}
	defer func() {
		_ = f.Close()
	}()
	w := bufio.NewWriter(f)
	for _, row := range rows {
		blob, err := json.Marshal(row.Row)
		if err != nil {
			return generatedDatasetWriteResult{}, err
		}
		if _, err := w.Write(blob); err != nil {
			return generatedDatasetWriteResult{}, err
		}
		if err := w.WriteByte('\n'); err != nil {
			return generatedDatasetWriteResult{}, err
		}
	}
	if err := w.Flush(); err != nil {
		return generatedDatasetWriteResult{}, err
	}

	meta := map[string]any{
		"datasetId":                record.DatasetID,
		"name":                     record.Name,
		"requestedCount":           record.RequestedCount,
		"generatedCount":           record.GeneratedCount,
		"seed":                     record.Seed,
		"pluginId":                 record.PluginID,
		"pluginName":               record.PluginName,
		"pluginRegistryIdentifier": record.PluginRegistryIdentifier,
		"configApiVersion":         record.ConfigAPIVersion,
		"createdAtMs":              record.CreatedAtMS,
		"createdAt":                time.UnixMilli(record.CreatedAtMS).UTC().Format(time.RFC3339),
	}
	metaBlob, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return generatedDatasetWriteResult{}, err
	}
	if err := os.WriteFile(metaPath, metaBlob, 0o644); err != nil {
		return generatedDatasetWriteResult{}, err
	}

	return generatedDatasetWriteResult{
		DatasetID:      record.DatasetID,
		RowsWritten:    len(rows),
		OutputJSONL:    jsonlPath,
		OutputMetadata: metaPath,
	}, nil
}

func writeGeneratedDatasetToSQLite(dbPath string, record generatedDatasetRecord, rows []generatedDatasetRow) (generatedDatasetWriteResult, error) {
	dbPath = strings.TrimSpace(dbPath)
	if dbPath == "" {
		return generatedDatasetWriteResult{}, fmt.Errorf("output db is empty")
	}
	if err := ensureParentDir(dbPath); err != nil {
		return generatedDatasetWriteResult{}, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return generatedDatasetWriteResult{}, err
	}
	defer func() {
		_ = db.Close()
	}()

	if err := ensureGeneratedDatasetTables(db); err != nil {
		return generatedDatasetWriteResult{}, err
	}

	tx, err := db.Begin()
	if err != nil {
		return generatedDatasetWriteResult{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	_, err = tx.Exec(`
INSERT INTO gepa_generated_datasets (
  dataset_id,
  name,
  requested_count,
  generated_count,
  seed,
  plugin_id,
  plugin_name,
  plugin_registry_identifier,
  config_api_version,
  config_json,
  created_at_ms
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.DatasetID,
		nullableString(record.Name),
		record.RequestedCount,
		record.GeneratedCount,
		record.Seed,
		nullableString(record.PluginID),
		nullableString(record.PluginName),
		nullableString(record.PluginRegistryIdentifier),
		nullableString(record.ConfigAPIVersion),
		nullableString(record.ConfigJSON),
		record.CreatedAtMS,
	)
	if err != nil {
		return generatedDatasetWriteResult{}, err
	}

	rowStmt, err := tx.Prepare(`
INSERT INTO gepa_generated_dataset_rows (
  dataset_id,
  row_index,
  row_json,
  metadata_json
) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return generatedDatasetWriteResult{}, err
	}
	defer func() {
		_ = rowStmt.Close()
	}()

	for _, row := range rows {
		rowJSON, err := json.Marshal(row.Row)
		if err != nil {
			return generatedDatasetWriteResult{}, err
		}
		var metadataJSON []byte
		if row.Metadata != nil {
			metadataJSON, err = json.Marshal(row.Metadata)
			if err != nil {
				return generatedDatasetWriteResult{}, err
			}
		}
		if _, err := rowStmt.Exec(record.DatasetID, row.RowIndex, string(rowJSON), nullableString(string(metadataJSON))); err != nil {
			return generatedDatasetWriteResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return generatedDatasetWriteResult{}, err
	}

	return generatedDatasetWriteResult{
		DatasetID:   record.DatasetID,
		RowsWritten: len(rows),
		DBPath:      dbPath,
	}, nil
}

func ensureGeneratedDatasetTables(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("generated dataset recorder: db is nil")
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS gepa_generated_datasets (
  dataset_id TEXT PRIMARY KEY,
  name TEXT,
  requested_count INTEGER NOT NULL,
  generated_count INTEGER NOT NULL,
  seed INTEGER NOT NULL,
  plugin_id TEXT,
  plugin_name TEXT,
  plugin_registry_identifier TEXT,
  config_api_version TEXT,
  config_json TEXT,
  created_at_ms INTEGER NOT NULL
)`,
		`CREATE TABLE IF NOT EXISTS gepa_generated_dataset_rows (
  dataset_id TEXT NOT NULL,
  row_index INTEGER NOT NULL,
  row_json TEXT NOT NULL,
  metadata_json TEXT,
  PRIMARY KEY (dataset_id, row_index)
)`,
		`CREATE INDEX IF NOT EXISTS idx_gepa_generated_datasets_created ON gepa_generated_datasets (created_at_ms DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_gepa_generated_dataset_rows_dataset ON gepa_generated_dataset_rows (dataset_id, row_index)`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}
