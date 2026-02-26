package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	geppettosections "github.com/go-go-golems/geppetto/pkg/sections"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type DatasetGenerateCommand struct {
	*cmds.CommandDescription
}

var _ cmds.WriterCommand = (*DatasetGenerateCommand)(nil)

type DatasetGenerateSettings struct {
	ScriptPath     string `glazed:"script"`
	ConfigPath     string `glazed:"config"`
	Count          int    `glazed:"count"`
	Seed           int64  `glazed:"seed"`
	OutputDir      string `glazed:"output-dir"`
	OutputDB       string `glazed:"output-db"`
	OutputFileStem string `glazed:"output-file-stem"`
	DryRun         bool   `glazed:"dry-run"`
	Debug          bool   `glazed:"debug"`
}

func NewDatasetGenerateCommand() (*DatasetGenerateCommand, error) {
	geppettoSections, err := geppettosections.CreateGeppettoSections()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create geppetto parameter layer")
	}

	description := cmds.NewCommandDescription(
		"generate",
		cmds.WithShort("Generate synthetic datasets from a JS dataset-generator plugin"),
		cmds.WithLong(`Generate a dataset with a JS script and a generation config.

Required:
  --script PATH   Dataset generator script
  --config PATH   Dataset generation YAML config (gepa.dataset-generate/v2)

YAML config must not include script/output routing. Use CLI flags for outputs:
  --output-dir PATH
  --output-db PATH
`),
		cmds.WithFlags(
			fields.New("script", fields.TypeString, fields.WithHelp("Path to JS dataset generator plugin"), fields.WithRequired(true)),
			fields.New("config", fields.TypeString, fields.WithHelp("Path to dataset generation YAML config"), fields.WithRequired(true)),
			fields.New("count", fields.TypeInteger, fields.WithHelp("Override config count (>0)."), fields.WithDefault(0)),
			fields.New("seed", fields.TypeInteger, fields.WithHelp("Override config seed (>=0). Use -1 to keep config/default seed."), fields.WithDefault(-1)),
			fields.New("output-dir", fields.TypeString, fields.WithHelp("Directory to write JSONL dataset + metadata files")),
			fields.New("output-db", fields.TypeString, fields.WithHelp("SQLite file path for generated dataset records")),
			fields.New("output-file-stem", fields.TypeString, fields.WithHelp("File stem used for output files in --output-dir")),
			fields.New("dry-run", fields.TypeBool, fields.WithHelp("Validate and generate in-memory only; do not write files/db"), fields.WithDefault(false)),
			fields.New("debug", fields.TypeBool, fields.WithHelp("Debug mode - show parsed layers"), fields.WithDefault(false)),
		),
		cmds.WithSections(geppettoSections...),
	)

	return &DatasetGenerateCommand{CommandDescription: description}, nil
}

func (c *DatasetGenerateCommand) RunIntoWriter(ctx context.Context, parsedValues *values.Values, w io.Writer) error {
	s := &DatasetGenerateSettings{}
	if err := parsedValues.DecodeSectionInto(values.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	if s.Debug {
		b, err := yaml.Marshal(parsedValues)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, "=== Parsed Layers Debug ===")
		fmt.Fprintln(w, string(b))
		fmt.Fprintln(w, "==========================")
		return nil
	}

	if strings.TrimSpace(s.ScriptPath) == "" {
		return fmt.Errorf("--script is required")
	}
	if strings.TrimSpace(s.ConfigPath) == "" {
		return fmt.Errorf("--config is required")
	}
	if !s.DryRun && strings.TrimSpace(s.OutputDir) == "" && strings.TrimSpace(s.OutputDB) == "" {
		return fmt.Errorf("no output target configured (set --output-dir and/or --output-db, or use --dry-run)")
	}

	cfg, rawYAML, err := loadDatasetGenerateConfig(s.ConfigPath)
	if err != nil {
		return err
	}
	resolvedCfg, err := resolveDatasetGenerateConfig(cfg, rawYAML, s)
	if err != nil {
		return err
	}

	profile, err := resolvePinocchioProfile(parsedValues)
	if err != nil {
		return errors.Wrap(err, "failed to resolve pinocchio profile")
	}
	if err := applyProfileEnvironment(profile, parsedValues); err != nil {
		return errors.Wrap(err, "failed to apply profile environment")
	}
	engineOptions, err := resolveEngineOptions(parsedValues)
	if err != nil {
		return errors.Wrap(err, "failed to resolve engine options from parsed settings")
	}

	absScript, err := filepath.Abs(s.ScriptPath)
	if err != nil {
		return err
	}
	absConfig, err := filepath.Abs(s.ConfigPath)
	if err != nil {
		return err
	}
	scriptRoot := filepath.Dir(absScript)

	jsrt, err := newJSRuntime(scriptRoot)
	if err != nil {
		return err
	}
	defer jsrt.Close()

	hostContext := map[string]any{
		"app":           "gepa-runner",
		"scriptPath":    filepath.ToSlash(absScript),
		"scriptRoot":    filepath.ToSlash(scriptRoot),
		"profile":       profile,
		"engineOptions": engineOptions,
		"configPath":    filepath.ToSlash(absConfig),
		"configName":    resolvedCfg.Config.Name,
	}

	plugin, meta, err := loadDatasetGeneratorPlugin(jsrt, absScript, hostContext)
	if err != nil {
		return err
	}
	log.Info().
		Str("plugin_id", meta.ID).
		Str("plugin_name", meta.Name).
		Str("plugin_registry_identifier", meta.RegistryIdentifier).
		Msg("Loaded dataset generator plugin")

	pluginTags := map[string]any{
		"plugin_id":                  meta.ID,
		"plugin_name":                meta.Name,
		"plugin_registry_identifier": meta.RegistryIdentifier,
		"command":                    "dataset_generate",
	}

	rows, skippedInvalid, err := generateDatasetRows(plugin, resolvedCfg, profile, engineOptions, pluginTags)
	if err != nil {
		return err
	}

	record := generatedDatasetRecord{
		DatasetID:                generateDatasetID(),
		Name:                     resolvedCfg.Config.Name,
		RequestedCount:           resolvedCfg.RequestedCount,
		GeneratedCount:           len(rows),
		Seed:                     resolvedCfg.Seed,
		PluginID:                 meta.ID,
		PluginName:               meta.Name,
		PluginRegistryIdentifier: meta.RegistryIdentifier,
		ConfigAPIVersion:         resolvedCfg.Config.APIVersion,
		ConfigJSON:               resolvedCfg.ConfigCanonical,
		CreatedAtMS:              time.Now().UnixMilli(),
	}

	fileWrite := generatedDatasetWriteResult{DatasetID: record.DatasetID}
	dbWrite := generatedDatasetWriteResult{DatasetID: record.DatasetID}
	if !resolvedCfg.DryRun {
		if resolvedCfg.OutputDir != "" {
			fileWrite, err = writeGeneratedDatasetFiles(resolvedCfg.OutputDir, resolvedCfg.OutputFileStem, record, rows)
			if err != nil {
				return errors.Wrap(err, "failed to write generated dataset files")
			}
		}
		if resolvedCfg.OutputDB != "" {
			dbWrite, err = writeGeneratedDatasetToSQLite(resolvedCfg.OutputDB, record, rows)
			if err != nil {
				return errors.Wrap(err, "failed to write generated dataset to sqlite")
			}
		}
	}

	fmt.Fprintf(w, "Plugin: %s (%s) [registry=%s]\n", meta.Name, meta.ID, meta.RegistryIdentifier)
	fmt.Fprintf(w, "Config: %s\n", absConfig)
	fmt.Fprintf(w, "Dataset ID: %s\n", record.DatasetID)
	fmt.Fprintf(w, "Requested rows: %d\n", record.RequestedCount)
	fmt.Fprintf(w, "Generated rows: %d\n", record.GeneratedCount)
	fmt.Fprintf(w, "Seed: %d\n", record.Seed)
	if skippedInvalid > 0 {
		fmt.Fprintf(w, "Skipped invalid attempts: %d\n", skippedInvalid)
	}
	if resolvedCfg.DryRun {
		fmt.Fprintln(w, "Dry-run: no output files or sqlite rows were written")
	}
	if fileWrite.OutputJSONL != "" {
		fmt.Fprintf(w, "Wrote JSONL: %s\n", fileWrite.OutputJSONL)
		fmt.Fprintf(w, "Wrote metadata: %s\n", fileWrite.OutputMetadata)
	}
	if dbWrite.DBPath != "" {
		fmt.Fprintf(w, "Wrote SQLite rows: %s\n", dbWrite.DBPath)
	}
	if len(rows) > 0 {
		firstBlob, _ := json.Marshal(rows[0].Row)
		fmt.Fprintf(w, "First row: %s\n", string(firstBlob))
	}

	return nil
}

func generateDatasetRows(plugin *datasetGeneratorPlugin, cfg datasetGenerateResolvedConfig, profile string, engineOptions map[string]any, tags map[string]any) ([]generatedDatasetRow, int, error) {
	if plugin == nil {
		return nil, 0, fmt.Errorf("dataset generator plugin is nil")
	}
	rng := rand.New(rand.NewSource(cfg.Seed))
	jsRandom := &jsRNG{rng: rng}

	rows := make([]generatedDatasetRow, 0, cfg.RequestedCount)
	skippedInvalid := 0
	retries := max(0, cfg.MaxRetries)
	maxAttempts := cfg.RequestedCount * max(1, retries+1) * 20
	if maxAttempts < cfg.RequestedCount {
		maxAttempts = cfg.RequestedCount
	}

	for attempts := 0; len(rows) < cfg.RequestedCount && attempts < maxAttempts; attempts++ {
		rowIndex := len(rows)
		accepted := false
		var lastErr error

		for try := 0; try <= retries; try++ {
			input := map[string]any{
				"index":     rowIndex,
				"attempt":   try,
				"seed":      cfg.Seed,
				"name":      cfg.Config.Name,
				"variables": cfg.Config.Prompting.Variables,
				"promptSpec": map[string]any{
					"system":        cfg.Config.Prompting.System,
					"user_template": cfg.Config.Prompting.UserTemplate,
				},
			}
			row, metadata, err := plugin.GenerateOne(input, datasetGeneratorGenerateOptions{
				Profile:       profile,
				EngineOptions: engineOptions,
				Tags:          tags,
				Seed:          cfg.Seed,
				RNG:           jsRandom,
				Config:        cfg.Config,
			})
			if err != nil {
				lastErr = err
				continue
			}

			missing := missingRequiredFields(row, cfg.RequiredFields)
			if len(missing) > 0 {
				lastErr = fmt.Errorf("generated row missing required fields: %s", strings.Join(missing, ", "))
				continue
			}

			rows = append(rows, generatedDatasetRow{
				RowIndex: rowIndex,
				Row:      row,
				Metadata: metadata,
			})
			accepted = true
			break
		}

		if accepted {
			continue
		}
		if cfg.DropInvalid {
			skippedInvalid++
			continue
		}
		if lastErr == nil {
			lastErr = fmt.Errorf("failed to generate row %d", rowIndex)
		}
		return nil, skippedInvalid, lastErr
	}

	if len(rows) < cfg.RequestedCount {
		return nil, skippedInvalid, fmt.Errorf("failed to generate requested row count: generated=%d requested=%d", len(rows), cfg.RequestedCount)
	}
	return rows, skippedInvalid, nil
}

func missingRequiredFields(row map[string]any, requiredFields []string) []string {
	if len(requiredFields) == 0 {
		return nil
	}
	missing := make([]string, 0, len(requiredFields))
	for _, key := range requiredFields {
		value, ok := row[key]
		if !ok || !hasNonEmptyValue(value) {
			missing = append(missing, key)
		}
	}
	return missing
}

func hasNonEmptyValue(v any) bool {
	if v == nil {
		return false
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s) != ""
	}
	return true
}

type jsRNG struct {
	rng *rand.Rand
}

func (r *jsRNG) IntN(max int) int {
	if r == nil || r.rng == nil || max <= 0 {
		return 0
	}
	return r.rng.Intn(max)
}

func (r *jsRNG) Float64() float64 {
	if r == nil || r.rng == nil {
		return 0
	}
	return r.rng.Float64()
}

func (r *jsRNG) Choice(values []any) any {
	if r == nil || r.rng == nil || len(values) == 0 {
		return nil
	}
	return values[r.rng.Intn(len(values))]
}

func (r *jsRNG) Shuffle(values []any) {
	if r == nil || r.rng == nil || len(values) < 2 {
		return
	}
	r.rng.Shuffle(len(values), func(i, j int) {
		values[i], values[j] = values[j], values[i]
	})
}

func generateDatasetID() string {
	return fmt.Sprintf("gepa-dataset-%d", time.Now().UnixNano())
}

func generateDefaultSeed() int64 {
	return time.Now().UnixNano()
}
