package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func runGepaRunnerCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	allArgs := append([]string{"run", "./cmd/gepa-runner"}, args...)
	cmd := exec.Command("go", allArgs...)
	cmd.Dir = repoRoot(t)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func TestCandidateRunStreamCLIOutput(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "candidate-plugin.js")
	configPath := filepath.Join(tmp, "candidate-config.yaml")
	inputPath := filepath.Join(tmp, "candidate-input.json")

	script := `
const { defineOptimizerPlugin, OPTIMIZER_PLUGIN_API_VERSION } = require("gepa/plugins");

module.exports = defineOptimizerPlugin({
  apiVersion: OPTIMIZER_PLUGIN_API_VERSION,
  kind: "optimizer",
  id: "example.stream.candidate",
  name: "Example Stream Candidate",
  create() {
    return {
      run(input, options) {
        return Promise.resolve().then(() => {
          if (options && typeof options.emitEvent === "function") {
            options.emitEvent({ type: "candidate-start", data: { input: input } });
          }
          return {
            output: { answer: "ok" },
            metadata: { mode: "async" }
          };
        });
      }
    };
  }
});
`
	config := `
apiVersion: gepa.candidate-run/v2
candidate:
  prompt: hello
metadata:
  candidate_id: c-1
`
	input := `{"question":"How are you?"}`

	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(inputPath, []byte(input), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	out, err := runGepaRunnerCLI(t,
		"candidate", "run",
		"--script", scriptPath,
		"--config", configPath,
		"--input-file", inputPath,
		"--stream",
		"--output-format", "json",
	)
	if err != nil {
		t.Fatalf("candidate run failed: %v\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "stream-event ") {
		t.Fatalf("expected stream-event output\n%s", out)
	}
	if !strings.Contains(out, "\"type\":\"candidate-start\"") {
		t.Fatalf("expected candidate-start stream type\n%s", out)
	}
	if !strings.Contains(out, "\"runId\"") {
		t.Fatalf("expected final candidate run payload\n%s", out)
	}
}

func TestDatasetGenerateStreamCLIOutput(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "dataset-plugin.js")
	configPath := filepath.Join(tmp, "dataset-config.yaml")

	script := `
const { defineDatasetGenerator, DATASET_GENERATOR_API_VERSION } = require("gepa/plugins");

module.exports = defineDatasetGenerator({
  apiVersion: DATASET_GENERATOR_API_VERSION,
  kind: "dataset-generator",
  id: "example.stream.dataset",
  name: "Example Stream Dataset",
  create() {
    return {
      generateOne(input, options) {
        return Promise.resolve().then(() => {
          if (options && options.events && typeof options.events.emit === "function") {
            options.events.emit({ type: "row-start", data: { index: input.index } });
          }
          return {
            row: { value: "ok", idx: input.index },
            metadata: { mode: "async" }
          };
        });
      }
    };
  }
});
`
	config := `
apiVersion: gepa.dataset-generate/v2
name: stream-dataset
count: 1
prompting:
  user_template: "unused"
validation:
  required_fields:
    - value
  max_retries: 0
  drop_invalid: false
`

	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	out, err := runGepaRunnerCLI(t,
		"dataset", "generate",
		"--script", scriptPath,
		"--config", configPath,
		"--dry-run",
		"--stream",
	)
	if err != nil {
		t.Fatalf("dataset generate failed: %v\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "stream-event ") {
		t.Fatalf("expected stream-event output\n%s", out)
	}
	if !strings.Contains(out, "\"type\":\"row-start\"") {
		t.Fatalf("expected row-start stream type\n%s", out)
	}
	if !strings.Contains(out, "Dry-run: no output files or sqlite rows were written") {
		t.Fatalf("expected dry-run summary\n%s", out)
	}
}
