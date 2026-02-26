package main

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestWriteCandidateRunRecord(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "candidate-runs.sqlite")
	record := candidateRunRecord{
		RunID:                    "candidate-run-1",
		TimestampMS:              123,
		PluginID:                 "example.optimizer",
		PluginName:               "Example Optimizer",
		PluginRegistryIdentifier: "local.example",
		Profile:                  "default",
		CandidateID:              "cand-1",
		ReflectionUsed:           "merge-a-b",
		TagsJSON:                 `{"suite":"smoke"}`,
		CandidateJSON:            `{"prompt":"Solve"}`,
		InputJSON:                `{"question":"2+2"}`,
		OutputJSON:               `{"answer":"4"}`,
		MetadataJSON:             `{"latency_ms":12}`,
		ConfigJSON:               `{"apiVersion":"gepa.candidate-run/v2"}`,
		Status:                   "completed",
	}

	if err := writeCandidateRunRecord(dbPath, record); err != nil {
		t.Fatalf("writeCandidateRunRecord failed: %v", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM gepa_candidate_runs`).Scan(&count); err != nil {
		t.Fatalf("count candidate runs failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 candidate run row, got %d", count)
	}

	var status string
	if err := db.QueryRow(`SELECT status FROM gepa_candidate_runs WHERE run_id = ?`, record.RunID).Scan(&status); err != nil {
		t.Fatalf("query candidate run failed: %v", err)
	}
	if status != "completed" {
		t.Fatalf("unexpected status: %q", status)
	}
}
