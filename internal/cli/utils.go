package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/mojomast/geoffrussy/internal/config"
	"github.com/mojomast/geoffrussy/internal/provider"
)

func formatDuration(d time.Duration) string {
	if d < 0 {
		return "expired"
	}

	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	} else if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}

func formatTimeSince(t time.Time) string {
	duration := time.Since(t)
	if duration < time.Minute {
		return fmt.Sprintf("%.0fs", duration.Seconds())
	} else if duration < time.Hour {
		return fmt.Sprintf("%.0fm", duration.Minutes())
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}

func getProviderAndModel(cfgMgr *config.Manager, stage, overrideModel string) (string, string, error) {
	cfg := cfgMgr.GetConfig()

	modelName := overrideModel
	if modelName == "" {
		var err error
		modelName, err = cfgMgr.GetDefaultModel(stage)
		if err != nil || modelName == "" {
			for provider := range cfg.APIKeys {
				if defaultModel, ok := cfg.DefaultModels[provider]; ok && defaultModel != "" {
					return provider, defaultModel, nil
				}
				if _, ok := cfg.APIKeys[provider]; ok {
					if provider == "requesty" {
						return provider, "openai/gpt-4", nil
					}
					return provider, "gpt-3.5-turbo", nil
				}
			}
			return "", "", fmt.Errorf("no API keys configured. Run 'geoffrussy config' to set up providers")
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
				if _, ok := cfg.APIKeys[p]; ok {
					providerName = p
					break
				}
			}
		}
	}

	if providerName == "" {
		for p := range cfg.APIKeys {
			return p, modelName, nil
		}
		return "", "", fmt.Errorf("no provider configured for model: %s", modelName)
	}

	if _, ok := cfg.APIKeys[providerName]; !ok {
		return "", "", fmt.Errorf("no API key configured for provider '%s'. Run 'geoffrussy config --set-key'", providerName)
	}

	return providerName, modelName, nil
}

func guessProviderFromModel(model string) string {
	lowerModel := strings.ToLower(model)

	if strings.Contains(lowerModel, "gpt") {
		return "openai"
	}

	if strings.Contains(lowerModel, "claude") {
		return "anthropic"
	}

	if strings.Contains(lowerModel, "moonshot") || strings.Contains(lowerModel, "kimi") {
		return "kimi"
	}

	if strings.Contains(lowerModel, "zai") {
		return "zai"
	}

	if strings.Contains(lowerModel, "opencode") {
		return "opencode"
	}

	return ""
}

func setupProvider(bridge *provider.Bridge, cfgMgr *config.Manager, providerName string) error {
	p, err := provider.CreateProvider(providerName)
	if err != nil {
		return err
	}

	if providerName == "ollama" {
		if err := p.Authenticate(""); err != nil {
			return fmt.Errorf("failed to authenticate/connect to %s: %w", providerName, err)
		}
		return bridge.RegisterProvider(p)
	}

	apiKey, err := cfgMgr.GetAPIKey(providerName)
	if err != nil {
		return err
	}

	if err := p.Authenticate(apiKey); err != nil {
		return fmt.Errorf("failed to authenticate %s: %w", providerName, err)
	}

	return bridge.RegisterProvider(p)
}
