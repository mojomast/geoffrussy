package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/mojomast/geoffrussy/internal/config"
	"github.com/mojomast/geoffrussy/internal/provider"
)

// formatDuration formats a duration in a human-readable way
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

// formatTimeSince formats time since as a human-readable duration
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

	if overrideModel != "" {
		// User specified model override, find provider for it
		provider := guessProviderFromModel(overrideModel)
		if provider == "" {
			// Could not guess, use first available provider
			for p := range cfg.APIKeys {
				return p, overrideModel, nil
			}
			return "", "", fmt.Errorf("no provider configured for model override")
		}
		if _, ok := cfg.APIKeys[provider]; ok {
			return provider, overrideModel, nil
		}
		return provider, overrideModel, nil
	}

	// Try to get default model for the stage
	model, err := cfgMgr.GetDefaultModel(stage)
	if err == nil && model != "" {
		// Model configured, guess provider from model name
		provider := guessProviderFromModel(model)
		if provider != "" {
			if _, ok := cfg.APIKeys[provider]; ok {
				return provider, model, nil
			}
			return provider, model, nil
		}
	}

	// No default model for stage, or provider not configured, use first available provider
	for provider := range cfg.APIKeys {
		// Check if this provider has a default model
		if defaultModel, ok := cfg.DefaultModels[provider]; ok && defaultModel != "" {
			return provider, defaultModel, nil
		}
		// Otherwise use the first provider with a key
		if _, ok := cfg.APIKeys[provider]; ok {
			return provider, "gpt-3.5-turbo", nil
		}
	}

	return "", "", fmt.Errorf("no API keys configured. Run 'geoffrussy config' to set up providers")
}

// guessProviderFromModel attempts to guess the provider from a model name
func guessProviderFromModel(model string) string {
	lowerModel := strings.ToLower(model)

	// OpenAI models
	if strings.Contains(lowerModel, "gpt") {
		return "openai"
	}

	// Anthropic models
	if strings.Contains(lowerModel, "claude") {
		return "anthropic"
	}

	// Kimi models
	if strings.Contains(lowerModel, "moonshot") || strings.Contains(lowerModel, "kimi") {
		return "kimi"
	}

	// Z.ai models
	if strings.Contains(lowerModel, "zai") {
		return "zai"
	}

	// OpenCode models
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

	// Special handling for Ollama which doesn't require an API key
	if providerName == "ollama" {
		if err := p.Authenticate(""); err != nil {
			return fmt.Errorf("failed to authenticate/connect to %s: %w", providerName, err)
		}
		return bridge.RegisterProvider(p)
	}

	// For all other providers, we expect an API key
	apiKey, err := cfgMgr.GetAPIKey(providerName)
	if err != nil {
		return err
	}

	if err := p.Authenticate(apiKey); err != nil {
		return fmt.Errorf("failed to authenticate %s: %w", providerName, err)
	}

	return bridge.RegisterProvider(p)
}
