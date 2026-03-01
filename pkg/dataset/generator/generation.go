package generator

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type GenerateRowsOptions struct {
	Profile       string
	EngineOptions map[string]any
	Tags          map[string]any
	EventSink     EventSink
}

func GenerateRows(ctx context.Context, plugin *Plugin, cfg ResolvedConfig, options GenerateRowsOptions) ([]Row, int, error) {
	if plugin == nil {
		return nil, 0, fmt.Errorf("dataset generator plugin is nil")
	}
	rng := newSeededRNG(cfg.Seed)

	rows := make([]Row, 0, cfg.RequestedCount)
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
			row, metadata, err := plugin.GenerateOne(ctx, input, PluginGenerateOptions{
				Profile:       options.Profile,
				EngineOptions: options.EngineOptions,
				Tags:          options.Tags,
				Seed:          cfg.Seed,
				RNG:           rng,
				Config:        cfg.Config,
				EventSink:     options.EventSink,
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

			rows = append(rows, Row{RowIndex: rowIndex, Row: row, Metadata: metadata})
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

func GenerateDatasetID() string {
	return fmt.Sprintf("gepa-dataset-%d", time.Now().UnixNano())
}

func GenerateDefaultSeed() int64 {
	return time.Now().UnixNano()
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

type seededRNG struct {
	rng *rand.Rand
}

func newSeededRNG(seed int64) *seededRNG {
	return &seededRNG{rng: rand.New(rand.NewSource(seed))}
}

func (r *seededRNG) IntN(n int) int {
	if r == nil || r.rng == nil || n <= 0 {
		return 0
	}
	return r.rng.Intn(n)
}

func (r *seededRNG) Float64() float64 {
	if r == nil || r.rng == nil {
		return 0
	}
	return r.rng.Float64()
}

func (r *seededRNG) Choice(values []any) any {
	if r == nil || r.rng == nil || len(values) == 0 {
		return nil
	}
	return values[r.rng.Intn(len(values))]
}

func (r *seededRNG) Shuffle(values []any) {
	if r == nil || r.rng == nil || len(values) < 2 {
		return
	}
	r.rng.Shuffle(len(values), func(i, j int) {
		values[i], values[j] = values[j], values[i]
	})
}
