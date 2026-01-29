package provider

import (
	"fmt"
	"sort"
)

// ProviderFactory is a function that creates a new Provider
type ProviderFactory func() Provider

// Registry maintains a list of available providers
var Registry = map[string]ProviderFactory{
	"anthropic": func() Provider { return NewAnthropicProvider() },
	"firmware":  func() Provider { return NewFirmwareProvider() },
	"kimi":      func() Provider { return NewKimiProvider() },
	"ollama":    func() Provider { return NewOllamaProvider("") }, // Default URL
	"openai":    func() Provider { return NewOpenAIProvider() },
	"opencode":  func() Provider { return NewOpenCodeProvider() },
	"requesty":  func() Provider { return NewRequestyProvider() },
	"zai":       func() Provider { return NewZAIProvider() },
}

// GetProviderNames returns a list of all registered provider names sorted alphabetically
func GetProviderNames() []string {
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// CreateProvider creates a new instance of the specified provider
func CreateProvider(name string) (Provider, error) {
	factory, ok := Registry[name]
	if !ok {
		return nil, fmt.Errorf("provider factory not found for: %s", name)
	}
	return factory(), nil
}
