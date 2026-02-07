package mcp

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mojomast/geoffrussy/internal/config"
	"github.com/mojomast/geoffrussy/internal/state"
)

func openStateStore(projectPath string) (*state.Store, error) {
	dbPath := filepath.Join(projectPath, ".geoffrussy", "state.db")
	return state.NewStore(dbPath)
}

func getProviderAndModel(cfgMgr *config.Manager, stage, overrideModel string) (string, string, error) {
	cfg := cfgMgr.GetConfig()

	modelName := overrideModel
	if modelName == "" {
		var err error
		modelName, err = cfgMgr.GetDefaultModel(stage)
		if err != nil || modelName == "" {
			for providerName := range cfg.APIKeys {
				if defaultModel, ok := cfg.DefaultModels[providerName]; ok && defaultModel != "" {
					return providerName, defaultModel, nil
				}
				if providerName == "requesty" {
					return providerName, "openai/gpt-4", nil
				}
				return providerName, "gpt-3.5-turbo", nil
			}
			return "", "", fmt.Errorf("no API keys configured")
		}
	}

	providerName := ""
	if strings.Contains(modelName, "/") {
		if _, ok := cfg.APIKeys["requesty"]; ok {
			providerName = "requesty"
		} else {
			providerName = guessProviderFromModel(modelName)
		}
	} else {
		providerName = guessProviderFromModel(modelName)
		if providerName == "" {
			for p := range cfg.APIKeys {
				providerName = p
				break
			}
		}
	}

	if providerName == "" {
		for p := range cfg.APIKeys {
			return p, modelName, nil
		}
		return "", "", fmt.Errorf("no provider configured for model: %s", modelName)
	}

	return providerName, modelName, nil
}

func guessProviderFromModel(model string) string {
	lowerModel := strings.ToLower(model)
	switch {
	case strings.Contains(lowerModel, "gpt"):
		return "openai"
	case strings.Contains(lowerModel, "claude"):
		return "anthropic"
	case strings.Contains(lowerModel, "moonshot"), strings.Contains(lowerModel, "kimi"):
		return "kimi"
	case strings.Contains(lowerModel, "glm"), strings.Contains(lowerModel, "zai"):
		return "zai"
	case strings.Contains(lowerModel, "opencode"):
		return "opencode"
	default:
		return ""
	}
}
