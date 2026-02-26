package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	geppettosections "github.com/go-go-golems/geppetto/pkg/sections"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	datasetgen "github.com/go-go-golems/go-go-gepa/pkg/dataset/generator"
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
	Stream         bool   `glazed:"stream"`
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
			fields.New("stream", fields.TypeBool, fields.WithHelp("Stream plugin-emitted events as they arrive"), fields.WithDefault(false)),
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

	profile, err := resolvePinocchioProfile(parsedValues)
	if err != nil {
		return errors.Wrap(err, "failed to resolve pinocchio profile")
	}
	engineOptions, err := resolveEngineOptions(parsedValues)
	if err != nil {
		return errors.Wrap(err, "failed to resolve engine options from parsed settings")
	}

	absScript, err := filepath.Abs(strings.TrimSpace(s.ScriptPath))
	if err != nil {
		return err
	}
	jsrt, err := newJSRuntime(filepath.Dir(absScript))
	if err != nil {
		return err
	}
	defer jsrt.Close()

	result, err := datasetgen.RunWithRuntime(jsrt.vm, jsrt.runner, jsrt.reqMod, datasetgen.RunInput{
		ScriptPath:    s.ScriptPath,
		ConfigPath:    s.ConfigPath,
		Profile:       profile,
		EngineOptions: engineOptions,
		EventSink:     newCommandEventSink(w, s.Stream, "dataset_generate"),
		ResolveOptions: datasetgen.ResolveOptions{
			ConfigPath:     s.ConfigPath,
			Count:          s.Count,
			Seed:           s.Seed,
			OutputDir:      s.OutputDir,
			OutputDB:       s.OutputDB,
			OutputFileStem: s.OutputFileStem,
			DryRun:         s.DryRun,
		},
		AppName: "gepa-runner",
	})
	if err != nil {
		return err
	}

	meta := result.PluginMeta
	log.Info().
		Str("plugin_id", meta.ID).
		Str("plugin_name", meta.Name).
		Str("plugin_registry_identifier", meta.RegistryIdentifier).
		Msg("Loaded dataset generator plugin")

	record := result.Record
	fmt.Fprintf(w, "Plugin: %s (%s) [registry=%s]\n", meta.Name, meta.ID, meta.RegistryIdentifier)
	fmt.Fprintf(w, "Config: %s\n", result.AbsConfigPath)
	fmt.Fprintf(w, "Dataset ID: %s\n", record.DatasetID)
	fmt.Fprintf(w, "Requested rows: %d\n", record.RequestedCount)
	fmt.Fprintf(w, "Generated rows: %d\n", record.GeneratedCount)
	fmt.Fprintf(w, "Seed: %d\n", record.Seed)
	if result.SkippedInvalid > 0 {
		fmt.Fprintf(w, "Skipped invalid attempts: %d\n", result.SkippedInvalid)
	}
	if result.ResolvedConfig.DryRun {
		fmt.Fprintln(w, "Dry-run: no output files or sqlite rows were written")
	}
	if result.FileWrite.OutputJSONL != "" {
		fmt.Fprintf(w, "Wrote JSONL: %s\n", result.FileWrite.OutputJSONL)
		fmt.Fprintf(w, "Wrote metadata: %s\n", result.FileWrite.OutputMetadata)
	}
	if result.DBWrite.DBPath != "" {
		fmt.Fprintf(w, "Wrote SQLite rows: %s\n", result.DBWrite.DBPath)
	}
	if len(result.Rows) > 0 {
		firstBlob, _ := json.Marshal(result.Rows[0].Row)
		fmt.Fprintf(w, "First row: %s\n", string(firstBlob))
	}

	return nil
}
