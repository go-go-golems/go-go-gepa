package generator

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/pkg/runtimeowner"
)

type RunInput struct {
	ScriptPath     string
	ConfigPath     string
	Profile        string
	EngineOptions  map[string]any
	ResolveOptions ResolveOptions
	HostContext    map[string]any
	AppName        string
	EventSink      EventSink
}

type RunResult struct {
	AbsScriptPath  string
	AbsConfigPath  string
	ResolvedConfig ResolvedConfig
	PluginMeta     PluginMeta
	Record         Record
	Rows           []Row
	SkippedInvalid int
	FileWrite      WriteResult
	DBWrite        WriteResult
}

func RunWithRuntime(vm *goja.Runtime, runner runtimeowner.Runner, req *require.RequireModule, input RunInput) (*RunResult, error) {
	scriptPath := strings.TrimSpace(input.ScriptPath)
	configPath := strings.TrimSpace(input.ConfigPath)
	if scriptPath == "" {
		return nil, fmt.Errorf("--script is required")
	}
	if configPath == "" {
		return nil, fmt.Errorf("--config is required")
	}

	resolveOptions := input.ResolveOptions
	if strings.TrimSpace(resolveOptions.ConfigPath) == "" {
		resolveOptions.ConfigPath = configPath
	}
	if !resolveOptions.DryRun && strings.TrimSpace(resolveOptions.OutputDir) == "" && strings.TrimSpace(resolveOptions.OutputDB) == "" {
		return nil, fmt.Errorf("no output target configured (set --output-dir and/or --output-db, or use --dry-run)")
	}

	cfg, rawYAML, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	resolvedCfg, err := ResolveConfig(cfg, rawYAML, resolveOptions)
	if err != nil {
		return nil, err
	}

	absScript, err := filepath.Abs(scriptPath)
	if err != nil {
		return nil, err
	}
	absConfig, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}
	scriptRoot := filepath.Dir(absScript)

	hostContext := cloneMap(input.HostContext)
	if hostContext == nil {
		hostContext = map[string]any{}
	}
	appName := strings.TrimSpace(input.AppName)
	if appName == "" {
		appName = "gepa-runner"
	}
	hostContext["app"] = appName
	hostContext["scriptPath"] = filepath.ToSlash(absScript)
	hostContext["scriptRoot"] = filepath.ToSlash(scriptRoot)
	hostContext["profile"] = strings.TrimSpace(input.Profile)
	hostContext["engineOptions"] = input.EngineOptions
	hostContext["configPath"] = filepath.ToSlash(absConfig)
	hostContext["configName"] = resolvedCfg.Config.Name

	plugin, meta, err := LoadPlugin(vm, runner, req, absScript, hostContext)
	if err != nil {
		return nil, err
	}

	pluginTags := map[string]any{
		"plugin_id":                  meta.ID,
		"plugin_name":                meta.Name,
		"plugin_registry_identifier": meta.RegistryIdentifier,
		"command":                    "dataset_generate",
	}

	rows, skippedInvalid, err := GenerateRows(plugin, resolvedCfg, GenerateRowsOptions{
		Profile:       input.Profile,
		EngineOptions: input.EngineOptions,
		Tags:          pluginTags,
		EventSink:     input.EventSink,
	})
	if err != nil {
		return nil, err
	}

	record := Record{
		DatasetID:                GenerateDatasetID(),
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

	fileWrite := WriteResult{DatasetID: record.DatasetID}
	dbWrite := WriteResult{DatasetID: record.DatasetID}
	if !resolvedCfg.DryRun {
		if resolvedCfg.OutputDir != "" {
			fileWrite, err = WriteFiles(resolvedCfg.OutputDir, resolvedCfg.OutputFileStem, record, rows)
			if err != nil {
				return nil, fmt.Errorf("failed to write generated dataset files: %w", err)
			}
		}
		if resolvedCfg.OutputDB != "" {
			dbWrite, err = WriteSQLite(resolvedCfg.OutputDB, record, rows)
			if err != nil {
				return nil, fmt.Errorf("failed to write generated dataset to sqlite: %w", err)
			}
		}
	}

	return &RunResult{
		AbsScriptPath:  absScript,
		AbsConfigPath:  absConfig,
		ResolvedConfig: resolvedCfg,
		PluginMeta:     meta,
		Record:         record,
		Rows:           rows,
		SkippedInvalid: skippedInvalid,
		FileWrite:      fileWrite,
		DBWrite:        dbWrite,
	}, nil
}

func cloneMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
