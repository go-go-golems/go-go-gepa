package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteGeneratedDatasetToSQLite(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "generated.sqlite")
	record := generatedDatasetRecord{
		DatasetID:                "dataset-1",
		Name:                     "arith",
		RequestedCount:           2,
		GeneratedCount:           2,
		Seed:                     42,
		PluginID:                 "example.generator",
		PluginName:               "Example Generator",
		PluginRegistryIdentifier: defaultPluginRegistryIdentifier,
		ConfigAPIVersion:         datasetGenerateConfigAPIVersion,
		ConfigJSON:               `{"apiVersion":"gepa.dataset-generate/v2","count":2}`,
		CreatedAtMS:              123,
	}
	rows := []generatedDatasetRow{
		{RowIndex: 0, Row: map[string]any{"question": "2+2", "answer": "4"}, Metadata: map[string]any{"difficulty": "easy"}},
		{RowIndex: 1, Row: map[string]any{"question": "3+3", "answer": "6"}, Metadata: map[string]any{"difficulty": "easy"}},
	}

	out, err := writeGeneratedDatasetToSQLite(dbPath, record, rows)
	if err != nil {
		t.Fatalf("writeGeneratedDatasetToSQLite failed: %v", err)
	}
	if out.RowsWritten != 2 {
		t.Fatalf("expected 2 rows written, got %d", out.RowsWritten)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	var datasetCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM gepa_generated_datasets`).Scan(&datasetCount); err != nil {
		t.Fatalf("count dataset rows: %v", err)
	}
	if datasetCount != 1 {
		t.Fatalf("expected one dataset row, got %d", datasetCount)
	}

	var rowCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM gepa_generated_dataset_rows`).Scan(&rowCount); err != nil {
		t.Fatalf("count dataset row rows: %v", err)
	}
	if rowCount != 2 {
		t.Fatalf("expected two dataset rows, got %d", rowCount)
	}
}

func TestWriteGeneratedDatasetFiles(t *testing.T) {
	outputDir := t.TempDir()
	record := generatedDatasetRecord{
		DatasetID:                "dataset-2",
		Name:                     "arith",
		RequestedCount:           1,
		GeneratedCount:           1,
		Seed:                     42,
		PluginID:                 "example.generator",
		PluginName:               "Example Generator",
		PluginRegistryIdentifier: defaultPluginRegistryIdentifier,
		ConfigAPIVersion:         datasetGenerateConfigAPIVersion,
		CreatedAtMS:              123,
	}
	rows := []generatedDatasetRow{
		{RowIndex: 0, Row: map[string]any{"question": "2+2", "answer": "4"}, Metadata: map[string]any{"difficulty": "easy"}},
	}

	out, err := writeGeneratedDatasetFiles(outputDir, "generated", record, rows)
	if err != nil {
		t.Fatalf("writeGeneratedDatasetFiles failed: %v", err)
	}
	if out.OutputJSONL == "" || out.OutputMetadata == "" {
		t.Fatalf("expected non-empty output paths: %#v", out)
	}
	if _, err := os.Stat(out.OutputJSONL); err != nil {
		t.Fatalf("jsonl file missing: %v", err)
	}
	if _, err := os.Stat(out.OutputMetadata); err != nil {
		t.Fatalf("metadata file missing: %v", err)
	}
}
