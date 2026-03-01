package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func mustScriptPath(t *testing.T, name string) string {
	t.Helper()

	candidates := []string{}

	if _, file, _, ok := runtime.Caller(0); ok {
		candidates = append(candidates, filepath.Join(filepath.Dir(file), "scripts", name))
	}

	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, "scripts", name),
			filepath.Join(wd, "cmd", "gepa-runner", "scripts", name),
		)
	}

	for _, c := range candidates {
		abs, err := filepath.Abs(c)
		if err != nil {
			continue
		}
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}

	t.Fatalf("failed to resolve script path for %s (tried %v)", name, candidates)
	return ""
}

func TestExampleScriptsLoadAndExposeExpectedHooks(t *testing.T) {
	tests := []struct {
		name                 string
		script               string
		expectMerge          bool
		expectInitial        bool
		expectSelect         bool
		expectComponentSI    bool
		expectDatasetMinimum int
	}{
		{
			name:                 "smoke noop",
			script:               "smoke_noop_optimizer.js",
			expectMerge:          false,
			expectInitial:        false,
			expectSelect:         false,
			expectComponentSI:    false,
			expectDatasetMinimum: 2,
		},
		{
			name:                 "toy math",
			script:               "toy_math_optimizer.js",
			expectMerge:          true,
			expectInitial:        true,
			expectSelect:         true,
			expectComponentSI:    true,
			expectDatasetMinimum: 4,
		},
		{
			name:                 "multi param",
			script:               "multi_param_math_optimizer.js",
			expectMerge:          true,
			expectInitial:        true,
			expectSelect:         true,
			expectComponentSI:    true,
			expectDatasetMinimum: 4,
		},
		{
			name:                 "seedless heuristic merge",
			script:               "seedless_heuristic_merge_optimizer.js",
			expectMerge:          true,
			expectInitial:        true,
			expectSelect:         true,
			expectComponentSI:    false,
			expectDatasetMinimum: 4,
		},
		{
			name:                 "optimize-anything style",
			script:               "optimize_anything_style_optimizer.js",
			expectMerge:          true,
			expectInitial:        true,
			expectSelect:         true,
			expectComponentSI:    true,
			expectDatasetMinimum: 4,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			scriptPath := mustScriptPath(t, tc.script)
			rt, err := newJSRuntime(filepath.Dir(scriptPath))
			if err != nil {
				t.Fatalf("newJSRuntime failed: %v", err)
			}
			defer rt.Close()

			plugin, meta, err := loadOptimizerPlugin(rt, scriptPath, map[string]any{
				"app":        "test",
				"scriptPath": scriptPath,
			})
			if err != nil {
				t.Fatalf("loadOptimizerPlugin failed: %v", err)
			}

			if meta.ID == "" || meta.Name == "" {
				t.Fatalf("expected non-empty plugin metadata, got id=%q name=%q", meta.ID, meta.Name)
			}

			if got := plugin.HasMerge(); got != tc.expectMerge {
				t.Fatalf("HasMerge mismatch: got=%v want=%v", got, tc.expectMerge)
			}
			if got := plugin.HasInitialCandidate(); got != tc.expectInitial {
				t.Fatalf("HasInitialCandidate mismatch: got=%v want=%v", got, tc.expectInitial)
			}
			if got := plugin.HasSelectComponents(); got != tc.expectSelect {
				t.Fatalf("HasSelectComponents mismatch: got=%v want=%v", got, tc.expectSelect)
			}
			if got := plugin.HasComponentSideInfo(); got != tc.expectComponentSI {
				t.Fatalf("HasComponentSideInfo mismatch: got=%v want=%v", got, tc.expectComponentSI)
			}

			dataset, err := plugin.Dataset(context.Background())
			if err != nil {
				t.Fatalf("Dataset() failed: %v", err)
			}
			if len(dataset) < tc.expectDatasetMinimum {
				t.Fatalf("unexpected dataset size: got=%d want>=%d", len(dataset), tc.expectDatasetMinimum)
			}
		})
	}
}
