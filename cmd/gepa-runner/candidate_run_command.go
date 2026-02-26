package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
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

type CandidateRunCommand struct {
	*cmds.CommandDescription
}

var _ cmds.WriterCommand = (*CandidateRunCommand)(nil)

type CandidateRunSettings struct {
	ScriptPath     string `glazed:"script"`
	ConfigPath     string `glazed:"config"`
	InputFile      string `glazed:"input-file"`
	Record         bool   `glazed:"record"`
	RecordDB       string `glazed:"record-db"`
	CandidateID    string `glazed:"candidate-id"`
	ReflectionUsed string `glazed:"reflection-used"`
	Tags           string `glazed:"tags"`
	OutputFormat   string `glazed:"output-format"`
	OutResult      string `glazed:"out-result"`
	Debug          bool   `glazed:"debug"`
}

func NewCandidateRunCommand() (*CandidateRunCommand, error) {
	geppettoSections, err := geppettosections.CreateGeppettoSections()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create geppetto parameter layer")
	}

	description := cmds.NewCommandDescription(
		"run",
		cmds.WithShort("Run one candidate against one input using plugin run()"),
		cmds.WithLong(`Run a single candidate with a single input payload.

Required:
  --script PATH      JS optimizer plugin script
  --config PATH      Candidate config YAML (gepa.candidate-run/v2)
  --input-file PATH  Input payload file (JSON/YAML object)

Config YAML must not include script/input/output routing.
`),
		cmds.WithFlags(
			fields.New("script", fields.TypeString, fields.WithHelp("Path to JS optimizer plugin (descriptor)"), fields.WithRequired(true)),
			fields.New("config", fields.TypeString, fields.WithHelp("Path to candidate-run YAML config"), fields.WithRequired(true)),
			fields.New("input-file", fields.TypeString, fields.WithHelp("Path to run input file (JSON or YAML object)"), fields.WithRequired(true)),
			fields.New("record", fields.TypeBool, fields.WithHelp("Persist candidate run row to SQLite"), fields.WithDefault(false)),
			fields.New("record-db", fields.TypeString, fields.WithHelp("SQLite file path for candidate run records"), fields.WithDefault(".gepa-runner/runs.sqlite")),
			fields.New("candidate-id", fields.TypeString, fields.WithHelp("Override candidate_id from config metadata")),
			fields.New("reflection-used", fields.TypeString, fields.WithHelp("Override reflection_used from config metadata")),
			fields.New("tags", fields.TypeString, fields.WithHelp("Comma-separated tags key=value,key=value (merged with config tags)")),
			fields.New("output-format", fields.TypeString, fields.WithHelp("Output format: json|yaml|text|table"), fields.WithDefault("json")),
			fields.New("out-result", fields.TypeString, fields.WithHelp("Optional file path to write full run result JSON")),
			fields.New("debug", fields.TypeBool, fields.WithHelp("Debug mode - show parsed layers"), fields.WithDefault(false)),
		),
		cmds.WithSections(geppettoSections...),
	)

	return &CandidateRunCommand{CommandDescription: description}, nil
}

func (c *CandidateRunCommand) RunIntoWriter(ctx context.Context, parsedValues *values.Values, w io.Writer) error {
	s := &CandidateRunSettings{}
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
	if strings.TrimSpace(s.InputFile) == "" {
		return fmt.Errorf("--input-file is required")
	}

	cfg, rawYAML, err := loadCandidateRunConfig(s.ConfigPath)
	if err != nil {
		return err
	}
	resolvedCfg, err := resolveCandidateRunConfig(cfg, rawYAML, candidateRunResolveOptions{
		ConfigPath:     s.ConfigPath,
		CandidateID:    s.CandidateID,
		ReflectionUsed: s.ReflectionUsed,
		Tags:           s.Tags,
	})
	if err != nil {
		return err
	}

	inputPayload, err := loadCandidateRunInputFile(s.InputFile)
	if err != nil {
		return err
	}

	profile, err := resolvePinocchioProfile(parsedValues)
	if err != nil {
		return errors.Wrap(err, "failed to resolve pinocchio profile")
	}
	engineOptions, err := resolveEngineOptions(parsedValues)
	if err != nil {
		return errors.Wrap(err, "failed to resolve engine options from parsed settings")
	}
	effectiveProfile := strings.TrimSpace(profile)
	if effectiveProfile == "" {
		effectiveProfile = resolvedCfg.RuntimeProfile
	}
	effectiveEngineOptions := mergeStringAnyMaps(engineOptions, resolvedCfg.EngineOverrides)

	runTimestamp := time.Now().UnixMilli()
	runRecord := candidateRunRecord{
		RunID:          generateRunID("candidate-run"),
		TimestampMS:    runTimestamp,
		Profile:        effectiveProfile,
		CandidateID:    strings.TrimSpace(resolvedCfg.CandidateID),
		ReflectionUsed: strings.TrimSpace(resolvedCfg.ReflectionUsed),
		Status:         "failed",
		ConfigJSON:     resolvedCfg.ConfigCanonical,
	}

	candidateJSON, err := marshalJSONString(resolvedCfg.Candidate)
	if err != nil {
		return err
	}
	runRecord.CandidateJSON = candidateJSON
	inputJSON, err := marshalJSONString(inputPayload)
	if err != nil {
		return err
	}
	runRecord.InputJSON = inputJSON
	tagsJSON, err := marshalJSONString(resolvedCfg.Tags)
	if err != nil {
		return err
	}
	runRecord.TagsJSON = tagsJSON

	outputResult := any(nil)
	outputPayload := any(nil)
	metadataPayload := any(nil)

	finalize := func(runErr error) error {
		if runErr != nil {
			runRecord.Status = "failed"
			runRecord.ErrorMessage = truncateString(runErr.Error(), 4000)
		} else {
			runRecord.Status = "completed"
			runRecord.ErrorMessage = ""
		}

		if outputPayload != nil {
			if outputJSON, err := marshalJSONString(outputPayload); err == nil {
				runRecord.OutputJSON = outputJSON
			}
		}
		if runRecord.OutputJSON == "" {
			runRecord.OutputJSON = "null"
		}
		if metadataPayload != nil {
			if metadataJSON, err := marshalJSONString(metadataPayload); err == nil {
				runRecord.MetadataJSON = metadataJSON
			}
		}

		if s.Record {
			recordDB := strings.TrimSpace(s.RecordDB)
			if recordDB == "" {
				recordDB = ".gepa-runner/runs.sqlite"
			}
			if writeErr := writeCandidateRunRecord(recordDB, runRecord); writeErr != nil {
				if runErr != nil {
					return runErr
				}
				return writeErr
			}
		}
		return runErr
	}

	absScript, err := filepath.Abs(s.ScriptPath)
	if err != nil {
		return finalize(err)
	}
	absConfig, err := filepath.Abs(s.ConfigPath)
	if err != nil {
		return finalize(err)
	}
	absInput, err := filepath.Abs(s.InputFile)
	if err != nil {
		return finalize(err)
	}

	jsrt, err := newJSRuntime(filepath.Dir(absScript))
	if err != nil {
		return finalize(err)
	}
	defer jsrt.Close()

	hostContext := map[string]any{
		"app":           "gepa-runner",
		"scriptPath":    filepath.ToSlash(absScript),
		"scriptRoot":    filepath.ToSlash(filepath.Dir(absScript)),
		"profile":       effectiveProfile,
		"engineOptions": effectiveEngineOptions,
		"configPath":    filepath.ToSlash(absConfig),
		"inputPath":     filepath.ToSlash(absInput),
	}
	plugin, meta, err := loadOptimizerPlugin(jsrt, absScript, hostContext)
	if err != nil {
		return finalize(err)
	}
	if !plugin.HasRun() {
		return finalize(fmt.Errorf("plugin run() is required for candidate run mode"))
	}

	runRecord.PluginID = meta.ID
	runRecord.PluginName = meta.Name
	runRecord.PluginRegistryIdentifier = meta.RegistryIdentifier

	pluginTags := map[string]any{
		"plugin_id":                  meta.ID,
		"plugin_name":                meta.Name,
		"plugin_registry_identifier": meta.RegistryIdentifier,
		"command":                    "candidate_run",
	}
	for k, v := range resolvedCfg.Tags {
		pluginTags[fmt.Sprintf("tag_%s", k)] = v
	}

	outputResult, err = plugin.Run(inputPayload, resolvedCfg.Candidate, pluginEvaluateOptions{
		Profile:       effectiveProfile,
		EngineOptions: effectiveEngineOptions,
		Tags:          pluginTags,
	})
	if err != nil {
		return finalize(err)
	}

	outputPayload, metadataPayload = splitCandidateRunOutput(outputResult)

	result := map[string]any{
		"runId": runRecord.RunID,
		"plugin": map[string]any{
			"id":                 meta.ID,
			"name":               meta.Name,
			"apiVersion":         meta.APIVersion,
			"kind":               meta.Kind,
			"registryIdentifier": meta.RegistryIdentifier,
		},
		"profile":        effectiveProfile,
		"candidateId":    resolvedCfg.CandidateID,
		"reflectionUsed": resolvedCfg.ReflectionUsed,
		"tags":           resolvedCfg.Tags,
		"candidate":      resolvedCfg.Candidate,
		"input":          inputPayload,
		"output":         outputPayload,
	}
	if metadataPayload != nil {
		result["metadata"] = metadataPayload
	}

	log.Info().
		Str("plugin_id", meta.ID).
		Str("plugin_name", meta.Name).
		Str("plugin_registry_identifier", meta.RegistryIdentifier).
		Msg("Loaded candidate-run plugin")

	if strings.TrimSpace(s.OutResult) != "" {
		blob, _ := json.MarshalIndent(result, "", "  ")
		if err := os.WriteFile(s.OutResult, blob, 0o644); err != nil {
			return finalize(errors.Wrap(err, "failed to write out result"))
		}
		fmt.Fprintf(w, "Wrote result to: %s\n", s.OutResult)
	}

	if err := renderCandidateRunOutput(w, strings.TrimSpace(s.OutputFormat), result); err != nil {
		return finalize(err)
	}

	if err := finalize(nil); err != nil {
		return err
	}
	if s.Record {
		recordDB := strings.TrimSpace(s.RecordDB)
		if recordDB == "" {
			recordDB = ".gepa-runner/runs.sqlite"
		}
		fmt.Fprintf(w, "\nRecorded run in SQLite: %s\n", recordDB)
	}

	return nil
}

func splitCandidateRunOutput(result any) (any, any) {
	m, ok := result.(map[string]any)
	if !ok {
		return result, nil
	}
	output, hasOutput := m["output"]
	metadata, hasMetadata := m["metadata"]
	if hasOutput {
		if hasMetadata {
			return output, metadata
		}
		return output, nil
	}
	if hasMetadata {
		return result, metadata
	}
	return result, nil
}

func renderCandidateRunOutput(w io.Writer, outputFormat string, result map[string]any) error {
	format := strings.ToLower(strings.TrimSpace(outputFormat))
	if format == "" {
		format = "json"
	}

	switch format {
	case "json":
		blob, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w, string(blob))
		return err
	case "yaml", "yml":
		blob, err := yaml.Marshal(result)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w, string(blob))
		return err
	case "text":
		if output, ok := result["output"]; ok {
			s, ok := output.(string)
			if ok {
				_, err := fmt.Fprintln(w, s)
				return err
			}
			blob, err := json.Marshal(output)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(w, string(blob))
			return err
		}
		_, err := fmt.Fprintln(w, "")
		return err
	case "table":
		pluginName := ""
		pluginID := ""
		if pluginRaw, ok := result["plugin"]; ok {
			if pluginMap, ok := pluginRaw.(map[string]any); ok {
				pluginName = stringOrEmpty(pluginMap["name"])
				pluginID = stringOrEmpty(pluginMap["id"])
			}
		}
		fmt.Fprintf(w, "Run ID          %s\n", stringOrEmpty(result["runId"]))
		fmt.Fprintf(w, "Plugin          %s (%s)\n", pluginName, pluginID)
		fmt.Fprintf(w, "Profile         %s\n", stringOrEmpty(result["profile"]))
		fmt.Fprintf(w, "Candidate ID    %s\n", stringOrEmpty(result["candidateId"]))
		fmt.Fprintf(w, "Reflection Used %s\n", stringOrEmpty(result["reflectionUsed"]))
		if output, ok := result["output"]; ok {
			blob, err := json.Marshal(output)
			if err != nil {
				return err
			}
			fmt.Fprintf(w, "Output          %s\n", truncateString(string(blob), 400))
		}
		return nil
	default:
		return fmt.Errorf("unsupported --output-format %q (expected json|yaml|text|table)", outputFormat)
	}
}

func mergeStringAnyMaps(base map[string]any, override map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		out[k] = v
	}
	return out
}
