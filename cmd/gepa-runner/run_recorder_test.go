package main

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	gepaopt "github.com/go-go-golems/go-go-gepa/pkg/optimizer/gepa"
)

func TestRunRecorderOptimizeWritesRunAndCandidates(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "runs.sqlite")
	rec, err := newRunRecorder(runRecorderConfig{
		DBPath:      dbPath,
		Mode:        "optimize",
		PluginID:    "example.plugin",
		PluginName:  "Example Plugin",
		Profile:     "default",
		DatasetSize: 3,
		Objective:   "maximize score",
		MaxEvals:    20,
		BatchSize:   4,
		SeedPrompt:  "seed",
	})
	if err != nil {
		t.Fatalf("newRunRecorder failed: %v", err)
	}

	candidateA := gepaopt.Candidate{"prompt": "A"}
	candidateB := gepaopt.Candidate{"prompt": "B"}
	bestHash := hashCandidate(candidateB)

	res := &gepaopt.Result{
		BestCandidate: candidateB,
		BestStats: gepaopt.CandidateStats{
			MeanScore:      0.9,
			MeanObjectives: gepaopt.ObjectiveScores{"score": 0.9},
			N:              3,
		},
		CallsUsed: 7,
		Candidates: []gepaopt.CandidateEntry{
			{
				ID:        0,
				ParentID:  -1,
				Hash:      hashCandidate(candidateA),
				Candidate: candidateA,
				GlobalStats: gepaopt.CandidateStats{
					MeanScore:      0.4,
					MeanObjectives: gepaopt.ObjectiveScores{"score": 0.4},
					N:              3,
				},
				EvalsCached: 3,
			},
			{
				ID:        1,
				ParentID:  0,
				Hash:      bestHash,
				Candidate: candidateB,
				GlobalStats: gepaopt.CandidateStats{
					MeanScore:      0.9,
					MeanObjectives: gepaopt.ObjectiveScores{"score": 0.9},
					N:              3,
				},
				EvalsCached: 3,
			},
		},
	}

	if err := rec.RecordOptimizeResult(res); err != nil {
		t.Fatalf("RecordOptimizeResult failed: %v", err)
	}
	if err := rec.Close(true, nil); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	var status string
	var callsUsed int
	var candidateCount int
	var bestCandidateHash string
	var pluginRegistryIdentifier string
	if err := db.QueryRow(
		`SELECT status, calls_used, candidate_count, best_candidate_hash, plugin_registry_identifier FROM gepa_runs LIMIT 1`,
	).Scan(&status, &callsUsed, &candidateCount, &bestCandidateHash, &pluginRegistryIdentifier); err != nil {
		t.Fatalf("query gepa_runs failed: %v", err)
	}
	if status != "completed" {
		t.Fatalf("expected status completed, got %q", status)
	}
	if callsUsed != 7 {
		t.Fatalf("expected calls_used 7, got %d", callsUsed)
	}
	if candidateCount != 2 {
		t.Fatalf("expected candidate_count 2, got %d", candidateCount)
	}
	if bestCandidateHash != bestHash {
		t.Fatalf("expected best hash %q, got %q", bestHash, bestCandidateHash)
	}
	if pluginRegistryIdentifier != defaultPluginRegistryIdentifier {
		t.Fatalf("expected plugin registry identifier %q, got %q", defaultPluginRegistryIdentifier, pluginRegistryIdentifier)
	}

	var candidateRows int
	if err := db.QueryRow(`SELECT COUNT(*) FROM gepa_candidate_metrics`).Scan(&candidateRows); err != nil {
		t.Fatalf("count candidate metrics failed: %v", err)
	}
	if candidateRows != 2 {
		t.Fatalf("expected 2 candidate rows, got %d", candidateRows)
	}
}

func TestRunRecorderEvalWritesExampleRowsAndFailureStatus(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "runs.sqlite")
	rec, err := newRunRecorder(runRecorderConfig{
		DBPath:      dbPath,
		Mode:        "eval",
		PluginID:    "example.eval",
		PluginName:  "Example Eval",
		Profile:     "default",
		DatasetSize: 2,
		SeedPrompt:  "seed",
	})
	if err != nil {
		t.Fatalf("newRunRecorder failed: %v", err)
	}

	candidate := gepaopt.Candidate{"prompt": "seed"}
	stats := gepaopt.CandidateStats{
		MeanScore:      0.5,
		MeanObjectives: gepaopt.ObjectiveScores{"score": 0.5},
		N:              2,
	}
	evals := []gepaopt.ExampleEval{
		{
			ExampleIndex: 0,
			Result: gepaopt.EvalResult{
				Score:      1.0,
				Objectives: gepaopt.ObjectiveScores{"score": 1.0},
				Feedback:   "ok",
				Output:     map[string]any{"text": "a"},
				Trace:      map[string]any{"id": "t1"},
				Raw:        map[string]any{"raw": true},
			},
		},
		{
			ExampleIndex: 1,
			Result: gepaopt.EvalResult{
				Score:      0.0,
				Objectives: gepaopt.ObjectiveScores{"score": 0.0},
				Feedback:   "bad",
				Output:     map[string]any{"text": "b"},
				Trace:      map[string]any{"id": "t2"},
				Raw:        map[string]any{"raw": false},
			},
		},
	}

	if err := rec.RecordEvalResult(candidate, stats, evals); err != nil {
		t.Fatalf("RecordEvalResult failed: %v", err)
	}
	if err := rec.Close(false, errors.New("run failed")); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	var status, errMsg string
	var pluginRegistryIdentifier string
	if err := db.QueryRow(`SELECT status, COALESCE(error, ''), plugin_registry_identifier FROM gepa_runs LIMIT 1`).Scan(&status, &errMsg, &pluginRegistryIdentifier); err != nil {
		t.Fatalf("query gepa_runs failed: %v", err)
	}
	if status != "failed" {
		t.Fatalf("expected failed status, got %q", status)
	}
	if !strings.Contains(errMsg, "run failed") {
		t.Fatalf("expected error message to contain run failure, got %q", errMsg)
	}
	if pluginRegistryIdentifier != defaultPluginRegistryIdentifier {
		t.Fatalf("expected plugin registry identifier %q, got %q", defaultPluginRegistryIdentifier, pluginRegistryIdentifier)
	}

	var evalRows int
	if err := db.QueryRow(`SELECT COUNT(*) FROM gepa_eval_examples`).Scan(&evalRows); err != nil {
		t.Fatalf("count eval rows failed: %v", err)
	}
	if evalRows != 2 {
		t.Fatalf("expected 2 eval rows, got %d", evalRows)
	}
}

func TestEnsureRecorderTablesAddsPluginRegistryIdentifierToLegacySchema(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "legacy.sqlite")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	legacySchema := `
CREATE TABLE gepa_runs (
  run_id TEXT PRIMARY KEY,
  mode TEXT NOT NULL,
  status TEXT NOT NULL,
  started_at_ms INTEGER NOT NULL,
  finished_at_ms INTEGER NOT NULL,
  duration_ms INTEGER NOT NULL,
  plugin_id TEXT,
  plugin_name TEXT,
  profile TEXT,
  dataset_size INTEGER NOT NULL DEFAULT 0,
  objective TEXT,
  max_evals INTEGER,
  batch_size INTEGER,
  calls_used INTEGER,
  best_mean_score REAL,
  best_n INTEGER,
  mean_score REAL,
  mean_n INTEGER,
  candidate_count INTEGER,
  best_candidate_hash TEXT,
  seed_prompt_sha256 TEXT,
  error TEXT,
  created_at_ms INTEGER NOT NULL
)`
	if _, err := db.Exec(legacySchema); err != nil {
		t.Fatalf("create legacy gepa_runs schema failed: %v", err)
	}

	if err := ensureRecorderTables(db); err != nil {
		t.Fatalf("ensureRecorderTables failed: %v", err)
	}

	exists, err := hasColumn(db, gepaRunsTable, "plugin_registry_identifier")
	if err != nil {
		t.Fatalf("hasColumn failed: %v", err)
	}
	if !exists {
		t.Fatalf("expected plugin_registry_identifier column to be present after migration")
	}
}

func hasColumn(db *sql.DB, table, column string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return false, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var (
		cid       int
		name      string
		typ       string
		notnull   int
		dfltValue sql.NullString
		pk        int
	)
	for rows.Next() {
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}
