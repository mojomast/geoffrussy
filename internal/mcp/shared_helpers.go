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
	stage = normalizeModelStage(stage)

	modelName := overrideModel
	if modelName == "" {
		var err error
		modelName, err = cfgMgr.ResolveDefaultModel(stage)
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
		} else if _, ok := cfg.APIKeys["openrouter"]; ok {
			providerName = "openrouter"
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

func normalizeModelStage(stage string) string {
	if stage == "plan" || stage == "plan.generate" {
		return "devplan.generate"
	}
	return stage
}

func guessProviderFromModel(model string) string {
	lowerModel := strings.ToLower(model)
	switch {
	case strings.Contains(lowerModel, "gpt"):
		return "openai"
	case strings.Contains(lowerModel, "codex"):
		return "openai-codex"
	case strings.Contains(lowerModel, "claude"):
		return "anthropic"
	case strings.Contains(lowerModel, "moonshot"), strings.Contains(lowerModel, "kimi"):
		return "kimi"
	case strings.Contains(lowerModel, "glm"), strings.Contains(lowerModel, "zai"):
		return "zai"
	case strings.Contains(lowerModel, "opencode"):
		return "opencode"
	case strings.Contains(lowerModel, "llama"), strings.Contains(lowerModel, "mixtral"):
		return "groq"
	case strings.Contains(lowerModel, "mistral"), strings.Contains(lowerModel, "pixtral"):
		return "mistral"
	case strings.Contains(lowerModel, "sonar"), strings.Contains(lowerModel, "perplexity"):
		return "perplexity"
	case strings.Contains(lowerModel, "fireworks"):
		return "fireworks"
	case strings.Contains(lowerModel, "deepinfra"):
		return "deepinfra"
	case strings.Contains(lowerModel, "qwen"), strings.Contains(lowerModel, "deepseek"):
		return "together"
	default:
		return ""
	}
}
