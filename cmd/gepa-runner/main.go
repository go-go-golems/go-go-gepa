package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/geppetto/pkg/inference/engine/factory"
	geppettosections "github.com/go-go-golems/geppetto/pkg/sections"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	gepaopt "github.com/go-go-golems/go-go-gepa/pkg/optimizer/gepa"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var rootCmd = &cobra.Command{
	Use:   "gepa-runner",
	Short: "GEPA-style prompt optimization on top of Geppetto + JS evaluators",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return logging.InitLoggerFromCobra(cmd)
	},
}

type OptimizeCommand struct {
	*cmds.CommandDescription
}

var _ cmds.WriterCommand = (*OptimizeCommand)(nil)

type OptimizeSettings struct {
	ScriptPath        string  `glazed:"script"`
	DatasetPath       string  `glazed:"dataset"`
	Seed              string  `glazed:"seed"`
	SeedFile          string  `glazed:"seed-file"`
	SeedCandidate     string  `glazed:"seed-candidate"`
	Objective         string  `glazed:"objective"`
	MaxEvalCalls      int     `glazed:"max-evals"`
	BatchSize         int     `glazed:"batch-size"`
	MergeProb         float64 `glazed:"merge-prob"`
	MaxSideInfoChars  int     `glazed:"max-side-info-chars"`
	OptimizableKeys   string  `glazed:"optimizable-keys"`
	ComponentSelector string  `glazed:"component-selector"`
	OutPrompt         string  `glazed:"out-prompt"`
	OutReport         string  `glazed:"out-report"`
	Record            bool    `glazed:"record"`
	RecordDB          string  `glazed:"record-db"`
	Debug             bool    `glazed:"debug"`
}

func NewOptimizeCommand() (*OptimizeCommand, error) {
	geppettoSections, err := geppettosections.CreateGeppettoSections()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create geppetto parameter layer")
	}

	description := cmds.NewCommandDescription(
		"optimize",
		cmds.WithShort("Run a GEPA-style reflective prompt evolution loop"),
		cmds.WithFlags(
			fields.New("script", fields.TypeString, fields.WithHelp("Path to JS optimizer plugin (descriptor)"), fields.WithRequired(true)),
			fields.New("dataset", fields.TypeString, fields.WithHelp("Path to dataset (.json or .jsonl). Optional if plugin provides dataset().")),
			fields.New("seed", fields.TypeString, fields.WithHelp("Seed prompt text (overrides --seed-file)")),
			fields.New("seed-file", fields.TypeString, fields.WithHelp("Path to seed prompt file")),
			fields.New("seed-candidate", fields.TypeString, fields.WithHelp("Path to seed candidate file (JSON or YAML) containing a map of parameter names to strings. If set, overrides --seed/--seed-file and enables multi-parameter optimization.")),
			fields.New("objective", fields.TypeString, fields.WithHelp("Natural-language optimization objective (used in reflection prompt)")),
			fields.New("max-evals", fields.TypeInteger, fields.WithHelp("Max evaluator calls (each example eval counts as 1)"), fields.WithDefault(200)),
			fields.New("batch-size", fields.TypeInteger, fields.WithHelp("Minibatch size per iteration"), fields.WithDefault(8)),
			fields.New("merge-prob", fields.TypeFloat, fields.WithHelp("Probability of attempting a merge (crossover) step between two prompts (0 disables)"), fields.WithDefault(0.0)),
			fields.New("max-side-info-chars", fields.TypeInteger, fields.WithHelp("Cap formatted side-info chars passed to reflection LLM (0 = uncapped)"), fields.WithDefault(8000)),
			fields.New("optimizable-keys", fields.TypeString, fields.WithHelp("Comma-separated list of candidate keys to optimize (defaults to all keys in the seed candidate).")),
			fields.New("component-selector", fields.TypeString, fields.WithHelp("Which parameter(s) to update per iteration: round_robin (default) or all."), fields.WithDefault("round_robin")),
			fields.New("out-prompt", fields.TypeString, fields.WithHelp("Write best prompt to this file (optional)")),
			fields.New("out-report", fields.TypeString, fields.WithHelp("Write JSON optimization report to this file (optional)")),
			fields.New("record", fields.TypeBool, fields.WithHelp("Persist run/candidate/eval metrics to SQLite"), fields.WithDefault(false)),
			fields.New("record-db", fields.TypeString, fields.WithHelp("SQLite file path for GEPA run metrics"), fields.WithDefault(".gepa-runner/runs.sqlite")),
			fields.New("debug", fields.TypeBool, fields.WithHelp("Debug mode - show parsed layers"), fields.WithDefault(false)),
		),
		cmds.WithSections(geppettoSections...),
	)

	return &OptimizeCommand{CommandDescription: description}, nil
}

func (c *OptimizeCommand) RunIntoWriter(ctx context.Context, parsedValues *values.Values, w io.Writer) error {
	s := &OptimizeSettings{}
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

	var seedCandidate gepaopt.Candidate
	seedTextForRecord := ""
	if strings.TrimSpace(s.SeedCandidate) != "" {
		cand, err := loadSeedCandidateFile(s.SeedCandidate)
		if err != nil {
			return err
		}
		if len(cand) == 0 {
			return fmt.Errorf("seed candidate is empty (check --seed-candidate)")
		}
		seedCandidate = cand
		if p, ok := cand["prompt"]; ok {
			seedTextForRecord = strings.TrimSpace(p)
		}
		if seedTextForRecord == "" {
			blob, _ := json.Marshal(cand)
			seedTextForRecord = string(blob)
		}
	} else {
		seedText, err := resolveSeedText(s.Seed, s.SeedFile)
		if err != nil {
			return err
		}
		if strings.TrimSpace(seedText) == "" {
			return fmt.Errorf("seed prompt is empty (use --seed, --seed-file, or --seed-candidate)")
		}
		seedCandidate = gepaopt.Candidate{"prompt": seedText}
		seedTextForRecord = seedText
	}

	// Ensure JS-side engine creation resolves the same profile by default.
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

	// Reflection LLM engine (Go side).
	engine, err := factory.NewEngineFromParsedValues(parsedValues)
	if err != nil {
		return errors.Wrap(err, "failed to create reflection engine from parsed values")
	}

	absScript, err := filepath.Abs(s.ScriptPath)
	if err != nil {
		return err
	}
	scriptRoot := filepath.Dir(absScript)

	// Load JS plugin.
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
	}
	plugin, meta, err := loadOptimizerPlugin(jsrt, absScript, hostContext)
	if err != nil {
		return err
	}
	log.Info().Str("plugin_id", meta.ID).Str("plugin_name", meta.Name).Msg("Loaded optimizer plugin")

	// Load dataset.
	var examples []any
	if strings.TrimSpace(s.DatasetPath) != "" {
		examples, err = loadDataset(s.DatasetPath)
		if err != nil {
			return err
		}
	} else {
		examples, err = plugin.Dataset()
		if err != nil {
			return err
		}
	}
	if len(examples) == 0 {
		return fmt.Errorf("dataset is empty")
	}

	evalFn := func(ctx context.Context, cand gepaopt.Candidate, exampleIndex int, example any) (gepaopt.EvalResult, error) {
		return plugin.Evaluate(cand, exampleIndex, example, pluginEvaluateOptions{
			Profile:       profile,
			EngineOptions: engineOptions,
		})
	}

	cfg := gepaopt.Config{
		MaxEvalCalls:     s.MaxEvalCalls,
		BatchSize:        s.BatchSize,
		MergeProbability: s.MergeProb,
		Objective:        s.Objective,
		MaxSideInfoChars: s.MaxSideInfoChars,
	}
	if strings.TrimSpace(s.OptimizableKeys) != "" {
		parts := strings.Split(s.OptimizableKeys, ",")
		for _, p := range parts {
			k := strings.TrimSpace(p)
			if k != "" {
				cfg.OptimizableKeys = append(cfg.OptimizableKeys, k)
			}
		}
	}
	cfg.ComponentSelector = strings.TrimSpace(s.ComponentSelector)

	reflector := &gepaopt.Reflector{
		Engine:        engine,
		Objective:     cfg.Objective,
		System:        cfg.ReflectionSystemPrompt,
		Template:      cfg.ReflectionPromptTemplate,
		MergeSystem:   cfg.MergeSystemPrompt,
		MergeTemplate: cfg.MergePromptTemplate,
	}

	opt := gepaopt.NewOptimizer(cfg, evalFn, reflector)
	if plugin.HasMerge() {
		opt.SetMergeFunc(func(ctx context.Context, in gepaopt.MergeInput) (string, string, error) {
			return plugin.Merge(in, pluginEvaluateOptions{
				Profile:       profile,
				EngineOptions: engineOptions,
			})
		})
	}

	var recorder *runRecorder
	if s.Record {
		recordDB := strings.TrimSpace(s.RecordDB)
		if recordDB == "" {
			recordDB = ".gepa-runner/runs.sqlite"
		}
		recorder, err = newRunRecorder(runRecorderConfig{
			DBPath:      recordDB,
			Mode:        "optimize",
			PluginID:    meta.ID,
			PluginName:  meta.Name,
			Profile:     profile,
			DatasetSize: len(examples),
			Objective:   s.Objective,
			MaxEvals:    s.MaxEvalCalls,
			BatchSize:   s.BatchSize,
			SeedPrompt:  seedTextForRecord,
		})
		if err != nil {
			return errors.Wrap(err, "failed to create run recorder")
		}
	}
	finalizeRun := func(runErr error) error {
		if recorder == nil {
			return runErr
		}
		closeErr := recorder.Close(runErr == nil, runErr)
		if runErr != nil {
			return runErr
		}
		return closeErr
	}

	res, err := opt.Optimize(ctx, seedCandidate, examples)
	if err != nil {
		return finalizeRun(err)
	}
	if recorder != nil {
		if err := recorder.RecordOptimizeResult(res); err != nil {
			return finalizeRun(err)
		}
	}

	// Output summary.
	fmt.Fprintf(w, "Plugin: %s (%s)\n", meta.Name, meta.ID)
	fmt.Fprintf(w, "Dataset: %d examples\n", len(examples))
	fmt.Fprintf(w, "Calls used: %d / %d\n", res.CallsUsed, s.MaxEvalCalls)
	fmt.Fprintf(w, "Best mean score (over cached evals): %.6f (n=%d)\n", res.BestStats.MeanScore, res.BestStats.N)

	bestPrompt, hasPrompt := res.BestCandidate["prompt"]
	if strings.TrimSpace(s.OutPrompt) != "" {
		if !hasPrompt {
			return finalizeRun(fmt.Errorf("best candidate does not contain a 'prompt' key (use --out-report to capture the full candidate)"))
		}
		if err := os.WriteFile(s.OutPrompt, []byte(bestPrompt), 0o644); err != nil {
			return finalizeRun(errors.Wrap(err, "failed to write out prompt"))
		}
		fmt.Fprintf(w, "Wrote best prompt to: %s\n", s.OutPrompt)
	} else {
		fmt.Fprintln(w, "\n=== Best Prompt ===")
		if hasPrompt {
			fmt.Fprintln(w, bestPrompt)
		} else {
			blob, _ := json.MarshalIndent(res.BestCandidate, "", "  ")
			fmt.Fprintln(w, string(blob))
		}
	}

	if strings.TrimSpace(s.OutReport) != "" {
		blob, _ := json.MarshalIndent(res, "", "  ")
		if err := os.WriteFile(s.OutReport, blob, 0o644); err != nil {
			return finalizeRun(errors.Wrap(err, "failed to write out report"))
		}
		fmt.Fprintf(w, "Wrote report to: %s\n", s.OutReport)
	}

	return finalizeRun(nil)
}

func main() {
	err := clay.InitGlazed("pinocchio", rootCmd)
	cobra.CheckErr(err)

	helpSystem := help.NewHelpSystem()
	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)
	cobra.CheckErr(err)

	optCmd, err := NewOptimizeCommand()
	cobra.CheckErr(err)

	command, err := cli.BuildCobraCommand(optCmd,
		cli.WithCobraMiddlewaresFunc(geppettosections.GetCobraCommandGeppettoMiddlewares),
		cli.WithProfileSettingsSection(),
	)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)

	evalCmd, err := NewEvalCommand()
	cobra.CheckErr(err)
	command2, err := cli.BuildCobraCommand(evalCmd,
		cli.WithCobraMiddlewaresFunc(geppettosections.GetCobraCommandGeppettoMiddlewares),
		cli.WithProfileSettingsSection(),
	)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command2)
	rootCmd.AddCommand(newEvalReportCommand())

	cobra.CheckErr(rootCmd.Execute())
}
