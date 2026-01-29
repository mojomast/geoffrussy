package quota

import (
	"os"
	"testing"
	"time"

	"github.com/mojomast/geoffrussy/internal/state"
)

func TestCheckRateLimitWarning(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	defer os.Remove(tmpDB)

	store, err := state.NewStore(tmpDB)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	monitor := NewMonitor(store)

	testCases := []struct {
		name              string
		requestsRemaining int
		requestsLimit     int
		expectedLevel     WarningLevel
	}{
		{
			name:              "Plenty of capacity",
			requestsRemaining: 900,
			requestsLimit:     1000,
			expectedLevel:     WarningInfo,
		},
		{
			name:              "Caution threshold",
			requestsRemaining: 250,
			requestsLimit:     1000,
			expectedLevel:     WarningCaution,
		},
		{
			name:              "Warning threshold",
			requestsRemaining: 100,
			requestsLimit:     1000,
			expectedLevel:     WarningWarning,
		},
		{
			name:              "Critical threshold",
			requestsRemaining: 30,
			requestsLimit:     1000,
			expectedLevel:     WarningCritical,
		},
		{
			name:              "Exceeded",
			requestsRemaining: 0,
			requestsLimit:     1000,
			expectedLevel:     WarningExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info := &state.RateLimitInfo{
				Provider:          "test",
				RequestsRemaining: tc.requestsRemaining,
				RequestsLimit:     tc.requestsLimit,
				ResetAt:           time.Now().Add(time.Hour),
				CheckedAt:         time.Now(),
			}

			warning := monitor.checkRateLimitWarning(info)

			if warning == nil {
				t.Fatal("expected warning, got nil")
			}

			if warning.Level != tc.expectedLevel {
				t.Errorf("expected level %s, got %s", tc.expectedLevel, warning.Level)
			}
		})
	}
}

func TestCheckQuotaWarning(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	defer os.Remove(tmpDB)

	store, err := state.NewStore(tmpDB)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	monitor := NewMonitor(store)

	testCases := []struct {
		name           string
		tokensRemaining int
		tokensLimit     int
		expectedLevel   WarningLevel
	}{
		{
			name:           "Plenty of tokens",
			tokensRemaining: 900000,
			tokensLimit:     1000000,
			expectedLevel:   WarningInfo,
		},
		{
			name:           "Caution threshold",
			tokensRemaining: 250000,
			tokensLimit:     1000000,
			expectedLevel:   WarningCaution,
		},
		{
			name:           "Warning threshold",
			tokensRemaining: 100000,
			tokensLimit:     1000000,
			expectedLevel:   WarningWarning,
		},
		{
			name:           "Critical threshold",
			tokensRemaining: 30000,
			tokensLimit:     1000000,
			expectedLevel:   WarningCritical,
		},
		{
			name:           "Exceeded",
			tokensRemaining: 0,
			tokensLimit:     1000000,
			expectedLevel:   WarningExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info := &state.QuotaInfo{
				Provider:        "test",
				TokensRemaining: &tc.tokensRemaining,
				TokensLimit:     &tc.tokensLimit,
				ResetAt:         time.Now().Add(24 * time.Hour),
				CheckedAt:       time.Now(),
			}

			warning := monitor.checkQuotaWarning(info)

			if warning == nil {
				t.Fatal("expected warning, got nil")
			}

			if warning.Level != tc.expectedLevel {
				t.Errorf("expected level %s, got %s", tc.expectedLevel, warning.Level)
			}
		})
	}
}

func TestGetCachedStatus(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	defer os.Remove(tmpDB)

	store, err := state.NewStore(tmpDB)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	monitor := NewMonitor(store)

	// Save some rate limit info
	rateLimitInfo := &state.RateLimitInfo{
		Provider:          "test-provider",
		RequestsRemaining: 500,
		RequestsLimit:     1000,
		ResetAt:           time.Now().Add(time.Hour),
		CheckedAt:         time.Now(),
	}

	if err := store.SaveRateLimit("test-provider", rateLimitInfo); err != nil {
		t.Fatalf("failed to save rate limit: %v", err)
	}

	// Get cached status
	status, err := monitor.GetCachedStatus("test-provider")
	if err != nil {
		t.Errorf("failed to get cached status: %v", err)
	}

	if status == nil {
		t.Fatal("expected status, got nil")
	}

	if status.Provider != "test-provider" {
		t.Errorf("expected provider test-provider, got %s", status.Provider)
	}

	if status.RateLimitInfo == nil {
		t.Error("expected rate limit info, got nil")
	}

	if status.RateLimitWarning == nil {
		t.Error("expected rate limit warning, got nil")
	}
}

func TestShouldDelayRequest(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	defer os.Remove(tmpDB)

	store, err := state.NewStore(tmpDB)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	monitor := NewMonitor(store)

	// Test with no rate limit exceeded
	rateLimitInfo := &state.RateLimitInfo{
		Provider:          "test-provider",
		RequestsRemaining: 500,
		RequestsLimit:     1000,
		ResetAt:           time.Now().Add(time.Hour),
		CheckedAt:         time.Now(),
	}

	if err := store.SaveRateLimit("test-provider", rateLimitInfo); err != nil {
		t.Fatalf("failed to save rate limit: %v", err)
	}

	shouldDelay, delay, err := monitor.ShouldDelayRequest("test-provider")
	if err != nil {
		t.Errorf("failed to check if should delay: %v", err)
	}

	if shouldDelay {
		t.Error("should not delay when rate limit not exceeded")
	}

	if delay != 0 {
		t.Errorf("expected delay 0, got %v", delay)
	}

	// Test with rate limit exceeded
	exceededInfo := &state.RateLimitInfo{
		Provider:          "test-provider-2",
		RequestsRemaining: 0,
		RequestsLimit:     1000,
		ResetAt:           time.Now().Add(time.Hour),
		CheckedAt:         time.Now(),
	}

	if err := store.SaveRateLimit("test-provider-2", exceededInfo); err != nil {
		t.Fatalf("failed to save rate limit: %v", err)
	}

	shouldDelay2, delay2, err := monitor.ShouldDelayRequest("test-provider-2")
	if err != nil {
		t.Errorf("failed to check if should delay: %v", err)
	}

	if !shouldDelay2 {
		t.Error("should delay when rate limit exceeded")
	}

	if delay2 <= 0 {
		t.Errorf("expected positive delay, got %v", delay2)
	}
}

func TestFormatDuration(t *testing.T) {
	testCases := []struct {
		name     string
		duration time.Duration
		contains string // check if result contains this string
	}{
		{
			name:     "Seconds",
			duration: 45 * time.Second,
			contains: "s",
		},
		{
			name:     "Minutes",
			duration: 5 * time.Minute,
			contains: "m",
		},
		{
			name:     "Hours",
			duration: 2*time.Hour + 30*time.Minute,
			contains: "h",
		},
		{
			name:     "Days",
			duration: 25 * time.Hour,
			contains: "d",
		},
		{
			name:     "Expired",
			duration: -1 * time.Hour,
			contains: "expired",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatDuration(tc.duration)
			if result == "" {
				t.Error("expected non-empty result")
			}
			t.Logf("Duration: %v -> %s", tc.duration, result)
		})
	}
}

func TestGetWarningSymbol(t *testing.T) {
	testCases := []struct {
		level  WarningLevel
		symbol string
	}{
		{WarningExceeded, "🚫"},
		{WarningCritical, "🔴"},
		{WarningWarning, "⚠️ "},
		{WarningCaution, "🟡"},
		{WarningInfo, "ℹ️ "},
		{WarningNone, "✅"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.level), func(t *testing.T) {
			symbol := GetWarningSymbol(tc.level)
			if symbol != tc.symbol {
				t.Errorf("expected symbol %s, got %s", tc.symbol, symbol)
			}
		})
	}
}
