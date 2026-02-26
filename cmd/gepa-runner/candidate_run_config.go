package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	gepaopt "github.com/go-go-golems/go-go-gepa/pkg/optimizer/gepa"
	"gopkg.in/yaml.v3"
)

const candidateRunConfigAPIVersion = "gepa.candidate-run/v2"

type candidateRunMetadataConfig struct {
	CandidateID    string            `yaml:"candidate_id" json:"candidate_id"`
	ReflectionUsed string            `yaml:"reflection_used" json:"reflection_used"`
	Tags           map[string]string `yaml:"tags" json:"tags"`
}

type candidateRunRuntimeConfig struct {
	Profile         string         `yaml:"profile" json:"profile"`
	EngineOverrides map[string]any `yaml:"engine_overrides" json:"engine_overrides"`
}

type candidateRunConfig struct {
	APIVersion string                     `yaml:"apiVersion" json:"apiVersion"`
	Candidate  map[string]any             `yaml:"candidate" json:"candidate"`
	Metadata   candidateRunMetadataConfig `yaml:"metadata" json:"metadata"`
	Runtime    candidateRunRuntimeConfig  `yaml:"runtime" json:"runtime"`
}

type candidateRunResolveOptions struct {
	ConfigPath     string
	CandidateID    string
	ReflectionUsed string
	Tags           string
}

type candidateRunResolvedConfig struct {
	Config          candidateRunConfig
	ConfigPath      string
	ConfigRawYAML   string
	ConfigCanonical string
	Candidate       gepaopt.Candidate
	CandidateID     string
	ReflectionUsed  string
	Tags            map[string]string
	RuntimeProfile  string
	EngineOverrides map[string]any
}

var forbiddenCandidateRunConfigKeys = []string{
	"script",
	"input",
	"input-file",
	"input_file",
	"output",
	"output-db",
	"output_db",
	"output-dir",
	"output_dir",
	"output-file",
	"output_file",
	"record",
	"record-db",
	"record_db",
}

func loadCandidateRunConfig(path string) (candidateRunConfig, string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return candidateRunConfig{}, "", fmt.Errorf("candidate run config path is empty")
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		return candidateRunConfig{}, "", err
	}

	var raw map[string]any
	if err := yaml.Unmarshal(blob, &raw); err != nil {
		return candidateRunConfig{}, "", fmt.Errorf("failed to parse YAML config %q: %w", filepath.Clean(path), err)
	}
	if err := validateCandidateRunTopLevelKeys(raw); err != nil {
		return candidateRunConfig{}, "", err
	}

	cfg := candidateRunConfig{}
	if err := yaml.Unmarshal(blob, &cfg); err != nil {
		return candidateRunConfig{}, "", fmt.Errorf("failed to parse YAML config %q: %w", filepath.Clean(path), err)
	}

	cfg.APIVersion = strings.TrimSpace(cfg.APIVersion)
	if cfg.APIVersion == "" {
		return candidateRunConfig{}, "", fmt.Errorf("candidate run config missing apiVersion")
	}
	if cfg.APIVersion != candidateRunConfigAPIVersion {
		return candidateRunConfig{}, "", fmt.Errorf("unsupported candidate run config apiVersion %q (expected %q)", cfg.APIVersion, candidateRunConfigAPIVersion)
	}

	return cfg, string(blob), nil
}

func resolveCandidateRunConfig(cfg candidateRunConfig, rawYAML string, options candidateRunResolveOptions) (candidateRunResolvedConfig, error) {
	candidate := gepaopt.Candidate{}
	for k, v := range cfg.Candidate {
		key := strings.TrimSpace(k)
		if key == "" {
			continue
		}
		candidate[key] = coerceToString(v)
	}
	if len(candidate) == 0 {
		return candidateRunResolvedConfig{}, fmt.Errorf("candidate run config must include non-empty candidate map")
	}

	candidateID := strings.TrimSpace(cfg.Metadata.CandidateID)
	if overrideID := strings.TrimSpace(options.CandidateID); overrideID != "" {
		candidateID = overrideID
	}

	reflectionUsed := strings.TrimSpace(cfg.Metadata.ReflectionUsed)
	if overrideReflection := strings.TrimSpace(options.ReflectionUsed); overrideReflection != "" {
		reflectionUsed = overrideReflection
	}

	tags := map[string]string{}
	for k, v := range cfg.Metadata.Tags {
		key := strings.TrimSpace(k)
		if key == "" {
			continue
		}
		tags[key] = strings.TrimSpace(v)
	}
	cliTags, err := parseCandidateRunTags(options.Tags)
	if err != nil {
		return candidateRunResolvedConfig{}, err
	}
	for k, v := range cliTags {
		tags[k] = v
	}

	canonicalConfigJSON, err := json.Marshal(cfg)
	if err != nil {
		return candidateRunResolvedConfig{}, err
	}

	return candidateRunResolvedConfig{
		Config:          cfg,
		ConfigPath:      strings.TrimSpace(options.ConfigPath),
		ConfigRawYAML:   rawYAML,
		ConfigCanonical: string(canonicalConfigJSON),
		Candidate:       candidate,
		CandidateID:     candidateID,
		ReflectionUsed:  reflectionUsed,
		Tags:            tags,
		RuntimeProfile:  strings.TrimSpace(cfg.Runtime.Profile),
		EngineOverrides: cloneStringAnyMap(cfg.Runtime.EngineOverrides),
	}, nil
}

func validateCandidateRunTopLevelKeys(raw map[string]any) error {
	if len(raw) == 0 {
		return fmt.Errorf("candidate run config is empty")
	}
	keys := make([]string, 0, len(raw))
	for k := range raw {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, key := range keys {
		if slices.Contains(forbiddenCandidateRunConfigKeys, key) {
			return fmt.Errorf("candidate run config key %q is not allowed (script/input/output/storage routing must come from CLI flags)", key)
		}
	}
	return nil
}

func parseCandidateRunTags(raw string) (map[string]string, error) {
	out := map[string]string{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return out, nil
	}
	parts := strings.Split(raw, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid --tags entry %q (expected key=value)", part)
		}
		key := strings.TrimSpace(kv[0])
		if key == "" {
			return nil, fmt.Errorf("invalid --tags entry %q (empty key)", part)
		}
		out[key] = strings.TrimSpace(kv[1])
	}
	return out, nil
}

func loadCandidateRunInputFile(path string) (map[string]any, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("candidate run input path is empty")
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var v any
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(blob, &v)
	default:
		err = json.Unmarshal(blob, &v)
	}
	if err != nil {
		var v2 any
		if ext == ".yaml" || ext == ".yml" {
			if err2 := json.Unmarshal(blob, &v2); err2 == nil {
				v = v2
				err = nil
			}
		} else {
			if err2 := yaml.Unmarshal(blob, &v2); err2 == nil {
				v = v2
				err = nil
			}
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse input file %q: %w", filepath.Clean(path), err)
	}

	if m, ok := v.(map[string]any); ok {
		if len(m) == 0 {
			return nil, fmt.Errorf("candidate run input object is empty")
		}
		return m, nil
	}
	if m2, ok := v.(map[any]any); ok {
		out := map[string]any{}
		for k, vv := range m2 {
			out[fmt.Sprintf("%v", k)] = vv
		}
		if len(out) == 0 {
			return nil, fmt.Errorf("candidate run input object is empty")
		}
		return out, nil
	}
	return nil, fmt.Errorf("candidate run input file must contain an object/map, got %T", v)
}

func cloneStringAnyMap(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
