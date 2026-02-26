package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

const datasetGenerateConfigAPIVersion = "gepa.dataset-generate/v2"

type datasetGeneratePromptingConfig struct {
	System       string         `yaml:"system" json:"system"`
	UserTemplate string         `yaml:"user_template" json:"user_template"`
	Variables    map[string]any `yaml:"variables" json:"variables"`
}

type datasetGenerateValidationConfig struct {
	RequiredFields []string `yaml:"required_fields" json:"required_fields"`
	MaxRetries     int      `yaml:"max_retries" json:"max_retries"`
	DropInvalid    bool     `yaml:"drop_invalid" json:"drop_invalid"`
}

type datasetGenerateConfig struct {
	APIVersion string                          `yaml:"apiVersion" json:"apiVersion"`
	Name       string                          `yaml:"name" json:"name"`
	Count      int                             `yaml:"count" json:"count"`
	Seed       *int64                          `yaml:"seed,omitempty" json:"seed,omitempty"`
	Prompting  datasetGeneratePromptingConfig  `yaml:"prompting" json:"prompting"`
	Validation datasetGenerateValidationConfig `yaml:"validation" json:"validation"`
}

type datasetGenerateResolvedConfig struct {
	Config           datasetGenerateConfig
	RequestedCount   int
	Seed             int64
	SeedFromCLI      bool
	CountFromCLI     bool
	ConfigPath       string
	ConfigRawYAML    string
	ConfigCanonical  string
	OutputDir        string
	OutputDB         string
	OutputFileStem   string
	DryRun           bool
	RequiredFields   []string
	MaxRetries       int
	DropInvalid      bool
	PromptingContext map[string]any
}

var forbiddenDatasetConfigKeys = []string{
	"script",
	"output",
	"output-db",
	"output_db",
	"output-dir",
	"output_dir",
	"output-file-stem",
	"output_file_stem",
	"output-format",
	"output_format",
}

func loadDatasetGenerateConfig(path string) (datasetGenerateConfig, string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return datasetGenerateConfig{}, "", fmt.Errorf("dataset config path is empty")
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		return datasetGenerateConfig{}, "", err
	}

	var raw map[string]any
	if err := yaml.Unmarshal(blob, &raw); err != nil {
		return datasetGenerateConfig{}, "", errorsWrapYAML(path, err)
	}
	if err := validateDatasetGenerateTopLevelKeys(raw); err != nil {
		return datasetGenerateConfig{}, "", err
	}

	cfg := datasetGenerateConfig{}
	if err := yaml.Unmarshal(blob, &cfg); err != nil {
		return datasetGenerateConfig{}, "", errorsWrapYAML(path, err)
	}

	cfg.APIVersion = strings.TrimSpace(cfg.APIVersion)
	cfg.Name = strings.TrimSpace(cfg.Name)
	if cfg.APIVersion == "" {
		return datasetGenerateConfig{}, "", fmt.Errorf("dataset config missing apiVersion")
	}
	if cfg.APIVersion != datasetGenerateConfigAPIVersion {
		return datasetGenerateConfig{}, "", fmt.Errorf("unsupported dataset config apiVersion %q (expected %q)", cfg.APIVersion, datasetGenerateConfigAPIVersion)
	}
	if cfg.Count < 0 {
		return datasetGenerateConfig{}, "", fmt.Errorf("dataset config count must be >= 0")
	}
	if cfg.Validation.MaxRetries < 0 {
		return datasetGenerateConfig{}, "", fmt.Errorf("dataset config validation.max_retries must be >= 0")
	}

	rawYAML := string(blob)
	return cfg, rawYAML, nil
}

func validateDatasetGenerateTopLevelKeys(raw map[string]any) error {
	if len(raw) == 0 {
		return fmt.Errorf("dataset config is empty")
	}
	keys := make([]string, 0, len(raw))
	for k := range raw {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, key := range keys {
		if slices.Contains(forbiddenDatasetConfigKeys, key) {
			return fmt.Errorf("dataset config key %q is not allowed (script/output routing must come from CLI flags)", key)
		}
	}
	return nil
}

func resolveDatasetGenerateConfig(cfg datasetGenerateConfig, rawYAML string, settings *DatasetGenerateSettings) (datasetGenerateResolvedConfig, error) {
	if settings == nil {
		return datasetGenerateResolvedConfig{}, fmt.Errorf("dataset settings are nil")
	}

	count := cfg.Count
	countFromCLI := false
	if settings.Count > 0 {
		count = settings.Count
		countFromCLI = true
	}
	if count <= 0 {
		return datasetGenerateResolvedConfig{}, fmt.Errorf("dataset generation count must be > 0 (set count in config or --count)")
	}

	var seed int64
	seedFromCLI := false
	if settings.Seed >= 0 {
		seed = settings.Seed
		seedFromCLI = true
	} else if cfg.Seed != nil {
		seed = *cfg.Seed
	} else {
		seed = generateDefaultSeed()
	}

	requiredFields := sanitizeRequiredFields(cfg.Validation.RequiredFields)
	maxRetries := cfg.Validation.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	promptingContext := map[string]any{
		"system":        cfg.Prompting.System,
		"user_template": cfg.Prompting.UserTemplate,
		"variables":     cfg.Prompting.Variables,
	}

	outputStem := strings.TrimSpace(settings.OutputFileStem)
	if outputStem == "" {
		outputStem = defaultDatasetOutputStem(cfg.Name)
	}

	canonicalConfigJSON, err := json.Marshal(cfg)
	if err != nil {
		return datasetGenerateResolvedConfig{}, err
	}

	return datasetGenerateResolvedConfig{
		Config:           cfg,
		RequestedCount:   count,
		Seed:             seed,
		SeedFromCLI:      seedFromCLI,
		CountFromCLI:     countFromCLI,
		ConfigPath:       strings.TrimSpace(settings.ConfigPath),
		ConfigRawYAML:    rawYAML,
		ConfigCanonical:  string(canonicalConfigJSON),
		OutputDir:        strings.TrimSpace(settings.OutputDir),
		OutputDB:         strings.TrimSpace(settings.OutputDB),
		OutputFileStem:   outputStem,
		DryRun:           settings.DryRun,
		RequiredFields:   requiredFields,
		MaxRetries:       maxRetries,
		DropInvalid:      cfg.Validation.DropInvalid,
		PromptingContext: promptingContext,
	}, nil
}

func defaultDatasetOutputStem(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "dataset"
	}
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.Trim(name, "-")
	if name == "" {
		return "dataset"
	}
	return name
}

func sanitizeRequiredFields(fields []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		k := strings.TrimSpace(field)
		if k == "" {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, k)
	}
	return out
}

func errorsWrapYAML(path string, err error) error {
	if strings.TrimSpace(path) == "" {
		return err
	}
	return fmt.Errorf("failed to parse YAML config %q: %w", filepath.Clean(path), err)
}
