package main

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

func resolvePinocchioProfile(parsedValues *values.Values) (string, error) {
	var profileSettings struct {
		Profile string `glazed:"profile"`
	}
	if err := parsedValues.DecodeSectionInto(cli.ProfileSettingsSlug, &profileSettings); err != nil {
		// Older geppetto section bundles may omit profile-settings entirely.
		// Treat that as "no explicit profile selected" for backward compatibility.
		if strings.Contains(err.Error(), fmt.Sprintf("section %s not found", cli.ProfileSettingsSlug)) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(profileSettings.Profile), nil
}

func ensureProfileSettingsSection(sections []schema.Section) ([]schema.Section, error) {
	for _, section := range sections {
		if section != nil && section.GetSlug() == cli.ProfileSettingsSlug {
			return sections, nil
		}
	}
	profileSection, err := cli.NewProfileSettingsSection()
	if err != nil {
		return nil, err
	}
	return append(sections, profileSection), nil
}

func resolveEngineOptions(parsedValues *values.Values) (map[string]any, error) {
	opts := map[string]any{}

	var chatSettings struct {
		APIType           string `glazed:"ai-api-type"`
		Engine            string `glazed:"ai-engine"`
		MaxResponseTokens int    `glazed:"ai-max-response-tokens"`
	}
	if err := parsedValues.DecodeSectionInto("ai-chat", &chatSettings); err != nil {
		return nil, err
	}

	apiType := strings.TrimSpace(chatSettings.APIType)
	engine := strings.TrimSpace(chatSettings.Engine)
	if apiType != "" {
		opts["apiType"] = apiType
	}
	if engine != "" {
		opts["model"] = engine
	}
	if chatSettings.MaxResponseTokens > 0 {
		opts["maxTokens"] = chatSettings.MaxResponseTokens
	}

	switch strings.ToLower(apiType) {
	case "openai", "openai-responses", "anyscale", "fireworks":
		var providerSettings struct {
			APIKey  string `glazed:"openai-api-key"`
			BaseURL string `glazed:"openai-base-url"`
		}
		if err := parsedValues.DecodeSectionInto("openai-chat", &providerSettings); err == nil {
			if key := strings.TrimSpace(providerSettings.APIKey); key != "" {
				opts["apiKey"] = key
			}
			if baseURL := strings.TrimSpace(providerSettings.BaseURL); baseURL != "" {
				opts["baseURL"] = baseURL
			}
		}
	case "claude":
		var providerSettings struct {
			APIKey  string `glazed:"claude-api-key"`
			BaseURL string `glazed:"claude-base-url"`
		}
		if err := parsedValues.DecodeSectionInto("claude-chat", &providerSettings); err == nil {
			if key := strings.TrimSpace(providerSettings.APIKey); key != "" {
				opts["apiKey"] = key
			}
			if baseURL := strings.TrimSpace(providerSettings.BaseURL); baseURL != "" {
				opts["baseURL"] = baseURL
			}
		}
	case "gemini":
		var providerSettings struct {
			APIKey  string `glazed:"gemini-api-key"`
			BaseURL string `glazed:"gemini-base-url"`
		}
		if err := parsedValues.DecodeSectionInto("gemini-chat", &providerSettings); err == nil {
			if key := strings.TrimSpace(providerSettings.APIKey); key != "" {
				opts["apiKey"] = key
			}
			if baseURL := strings.TrimSpace(providerSettings.BaseURL); baseURL != "" {
				opts["baseURL"] = baseURL
			}
		}
	}

	return opts, nil
}
