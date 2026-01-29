package provider

import (
	"fmt"
	"math"
	"time"
)

// Provider is the interface that all AI model providers must implement
type Provider interface {
	Name() string
	Authenticate(apiKey string) error
	IsAuthenticated() bool
	ListModels() ([]Model, error)
	DiscoverModels() ([]Model, error) // For dynamic discovery (OpenCode)
	Call(model string, prompt string) (*Response, error)
	Stream(model string, prompt string) (<-chan string, error)
	GetRateLimitInfo() (*RateLimitInfo, error)
	GetQuotaInfo() (*QuotaInfo, error)
	SupportsCodingPlan() bool // For Z.ai and Kimi
}

// Response represents a response from an AI model provider
type Response struct {
	Content            string
	TokensInput        int
	TokensOutput       int
	Model              string
	Provider           string
	Timestamp          time.Time
	RateLimitRemaining int
	QuotaRemaining     int
}

// RateLimitInfo contains rate limiting information from a provider
type RateLimitInfo struct {
	RequestsRemaining int
	RequestsLimit     int
	ResetAt           time.Time
	RetryAfter        time.Duration
}

// QuotaInfo contains quota information from a provider
type QuotaInfo struct {
	TokensRemaining int
	TokensLimit     int
	CostRemaining   float64
	CostLimit       float64
	ResetAt         time.Time
}

// Model represents an AI model
type Model struct {
	Provider     string
	Name         string
	DisplayName  string
	Capabilities []string
	PriceInput   float64 // per 1K tokens
	PriceOutput  float64 // per 1K tokens
}

// BaseProvider provides common functionality for all providers
type BaseProvider struct {
	name          string
	apiKey        string
	authenticated bool
	maxRetries    int
	baseDelay     time.Duration
}

// NewBaseProvider creates a new base provider
func NewBaseProvider(name string) *BaseProvider {
	return &BaseProvider{
		name:       name,
		maxRetries: 3,
		baseDelay:  time.Second,
	}
}

// Name returns the provider name
func (b *BaseProvider) Name() string {
	return b.name
}

// Authenticate stores the API key
func (b *BaseProvider) Authenticate(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}
	b.apiKey = apiKey
	b.authenticated = true
	return nil
}

// IsAuthenticated returns whether the provider is authenticated
func (b *BaseProvider) IsAuthenticated() bool {
	return b.authenticated
}

// GetAPIKey returns the stored API key
func (b *BaseProvider) GetAPIKey() string {
	return b.apiKey
}

// SetMaxRetries sets the maximum number of retries
func (b *BaseProvider) SetMaxRetries(maxRetries int) {
	b.maxRetries = maxRetries
}

// SetBaseDelay sets the base delay for exponential backoff
func (b *BaseProvider) SetBaseDelay(delay time.Duration) {
	b.baseDelay = delay
}

// RetryWithBackoff executes a function with exponential backoff retry logic
func (b *BaseProvider) RetryWithBackoff(fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= b.maxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on last attempt
		if attempt == b.maxRetries {
			break
		}

		// Calculate exponential backoff delay
		delay := b.baseDelay * time.Duration(math.Pow(2, float64(attempt)))
		time.Sleep(delay)
	}

	return fmt.Errorf("failed after %d retries: %w", b.maxRetries, lastErr)
}

// DiscoverModels is a default implementation that returns an error
// Providers that support dynamic discovery should override this
func (b *BaseProvider) DiscoverModels() ([]Model, error) {
	return nil, fmt.Errorf("provider %s does not support dynamic model discovery", b.name)
}

// GetRateLimitInfo is a default implementation that returns nil
// Providers that support rate limiting should override this
func (b *BaseProvider) GetRateLimitInfo() (*RateLimitInfo, error) {
	return nil, nil
}

// GetQuotaInfo is a default implementation that returns nil
// Providers that support quotas should override this
func (b *BaseProvider) GetQuotaInfo() (*QuotaInfo, error) {
	return nil, nil
}

// SupportsCodingPlan is a default implementation that returns false
// Providers that support coding plans should override this
func (b *BaseProvider) SupportsCodingPlan() bool {
	return false
}
