package gepa

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-go-golems/geppetto/pkg/inference/engine"
	"github.com/go-go-golems/geppetto/pkg/turns"
)

// Reflector runs the natural-language reflection step that proposes prompt mutations.
type Reflector struct {
	Engine        engine.Engine
	System        string
	Template      string
	MergeSystem   string
	MergeTemplate string
	Objective     string
}

func (r *Reflector) Propose(ctx context.Context, currentInstruction string, sideInfo string) (string, string, error) {
	if r == nil || r.Engine == nil {
		return "", "", fmt.Errorf("reflector: engine is nil")
	}
	sys := strings.TrimSpace(r.System)
	if sys == "" {
		sys = "You are an expert prompt engineer."
	}
	tmpl := r.Template
	if strings.TrimSpace(tmpl) == "" {
		tmpl = DefaultReflectionPromptTemplate
	}
	if !strings.Contains(tmpl, "<curr_param>") || !strings.Contains(tmpl, "<side_info>") {
		return "", "", fmt.Errorf("reflector: template must include <curr_param> and <side_info>")
	}

	user := strings.ReplaceAll(tmpl, "<curr_param>", currentInstruction)
	user = strings.ReplaceAll(user, "<side_info>", sideInfo)
	if strings.TrimSpace(r.Objective) != "" {
		user = fmt.Sprintf("Objective:\n%s\n\n%s", strings.TrimSpace(r.Objective), user)
	}

	turn := turns.NewTurnBuilder().
		WithSystemPrompt(sys).
		WithUserPrompt(user).
		Build()

	out, err := r.Engine.RunInference(ctx, turn)
	if err != nil {
		return "", "", fmt.Errorf("reflector: inference failed: %w", err)
	}

	raw := ExtractAssistantText(out)
	proposed := extractTripleBacktickBlock(raw)
	if strings.TrimSpace(proposed) == "" {
		proposed = strings.TrimSpace(raw)
	}
	return proposed, raw, nil
}

func (r *Reflector) Merge(ctx context.Context, instructionA string, instructionB string, sideInfoA string, sideInfoB string) (string, string, error) {
	if r == nil || r.Engine == nil {
		return "", "", fmt.Errorf("reflector: engine is nil")
	}

	sys := strings.TrimSpace(r.MergeSystem)
	if sys == "" {
		sys = strings.TrimSpace(r.System)
	}
	if sys == "" {
		sys = "You are an expert prompt engineer."
	}

	tmpl := r.MergeTemplate
	if strings.TrimSpace(tmpl) == "" {
		tmpl = DefaultMergePromptTemplate
	}
	if !strings.Contains(tmpl, "<param_a>") || !strings.Contains(tmpl, "<param_b>") || !strings.Contains(tmpl, "<side_info_a>") || !strings.Contains(tmpl, "<side_info_b>") {
		return "", "", fmt.Errorf("reflector: merge template must include <param_a>, <param_b>, <side_info_a>, and <side_info_b>")
	}

	user := strings.ReplaceAll(tmpl, "<param_a>", instructionA)
	user = strings.ReplaceAll(user, "<param_b>", instructionB)
	user = strings.ReplaceAll(user, "<side_info_a>", sideInfoA)
	user = strings.ReplaceAll(user, "<side_info_b>", sideInfoB)
	if strings.TrimSpace(r.Objective) != "" {
		user = fmt.Sprintf("Objective:\n%s\n\n%s", strings.TrimSpace(r.Objective), user)
	}

	turn := turns.NewTurnBuilder().
		WithSystemPrompt(sys).
		WithUserPrompt(user).
		Build()

	out, err := r.Engine.RunInference(ctx, turn)
	if err != nil {
		return "", "", fmt.Errorf("reflector: inference failed: %w", err)
	}

	raw := ExtractAssistantText(out)
	proposed := extractTripleBacktickBlock(raw)
	if strings.TrimSpace(proposed) == "" {
		proposed = strings.TrimSpace(raw)
	}
	return proposed, raw, nil
}

func ExtractAssistantText(t *turns.Turn) string {
	if t == nil {
		return ""
	}
	var parts []string
	for _, b := range t.Blocks {
		if b.Kind == turns.BlockKindLLMText || b.Role == turns.RoleAssistant {
			if b.Payload != nil {
				if s, ok := b.Payload[turns.PayloadKeyText].(string); ok {
					s = strings.TrimSpace(s)
					if s != "" {
						parts = append(parts, s)
					}
				}
			}
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

var tripleBacktickRe = regexp.MustCompile("(?s)```(?:[a-zA-Z0-9_-]+\\n)?\\s*(.*?)\\s*```")

func extractTripleBacktickBlock(s string) string {
	m := tripleBacktickRe.FindStringSubmatch(s)
	if len(m) >= 2 {
		return strings.TrimSpace(m[1])
	}
	return ""
}
