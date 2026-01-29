package provider

import (
	"errors"
	"testing"
	"time"
)

func TestNewBaseProvider(t *testing.T) {
	bp := NewBaseProvider("test-provider")
	
	if bp == nil {
		t.Fatal("NewBaseProvider returned nil")
	}
	if bp.Name() != "test-provider" {
		t.Errorf("Expected name 'test-provider', got '%s'", bp.Name())
	}
	if bp.IsAuthenticated() {
		t.Error("Expected provider to not be authenticated initially")
	}
	if bp.maxRetries != 3 {
		t.Errorf("Expected maxRetries to be 3, got %d", bp.maxRetries)
	}
	if bp.baseDelay != time.Second {
		t.Errorf("Expected baseDelay to be 1s, got %v", bp.baseDelay)
	}
}

func TestBaseProvider_Authenticate(t *testing.T) {
	bp := NewBaseProvider("test-provider")
	
	// Test successful authentication
	err := bp.Authenticate("test-api-key")
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}
	if !bp.IsAuthenticated() {
		t.Error("Expected provider to be authenticated")
	}
	if bp.GetAPIKey() != "test-api-key" {
		t.Errorf("Expected API key 'test-api-key', got '%s'", bp.GetAPIKey())
	}
	
	// Test empty API key
	bp2 := NewBaseProvider("test-provider-2")
	err = bp2.Authenticate("")
	if err == nil {
		t.Error("Expected error for empty API key")
	}
	if bp2.IsAuthenticated() {
		t.Error("Expected provider to not be authenticated with empty key")
	}
}

func TestBaseProvider_SetMaxRetries(t *testing.T) {
	bp := NewBaseProvider("test-provider")
	
	bp.SetMaxRetries(5)
	if bp.maxRetries != 5 {
		t.Errorf("Expected maxRetries to be 5, got %d", bp.maxRetries)
	}
}

func TestBaseProvider_SetBaseDelay(t *testing.T) {
	bp := NewBaseProvider("test-provider")
	
	bp.SetBaseDelay(2 * time.Second)
	if bp.baseDelay != 2*time.Second {
		t.Errorf("Expected baseDelay to be 2s, got %v", bp.baseDelay)
	}
}

func TestBaseProvider_RetryWithBackoff_Success(t *testing.T) {
	bp := NewBaseProvider("test-provider")
	bp.SetBaseDelay(10 * time.Millisecond) // Short delay for testing
	
	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary error")
		}
		return nil
	}
	
	err := bp.RetryWithBackoff(fn)
	if err != nil {
		t.Fatalf("RetryWithBackoff failed: %v", err)
	}
	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
}

func TestBaseProvider_RetryWithBackoff_ImmediateSuccess(t *testing.T) {
	bp := NewBaseProvider("test-provider")
	
	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}
	
	err := bp.RetryWithBackoff(fn)
	if err != nil {
		t.Fatalf("RetryWithBackoff failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestBaseProvider_RetryWithBackoff_AllFail(t *testing.T) {
	bp := NewBaseProvider("test-provider")
	bp.SetMaxRetries(2)
	bp.SetBaseDelay(10 * time.Millisecond) // Short delay for testing
	
	callCount := 0
	fn := func() error {
		callCount++
		return errors.New("persistent error")
	}
	
	err := bp.RetryWithBackoff(fn)
	if err == nil {
		t.Error("Expected error after all retries failed")
	}
	// Should be called maxRetries + 1 times (initial + retries)
	if callCount != 3 {
		t.Errorf("Expected 3 calls (1 initial + 2 retries), got %d", callCount)
	}
}

func TestBaseProvider_RetryWithBackoff_ExponentialDelay(t *testing.T) {
	bp := NewBaseProvider("test-provider")
	bp.SetMaxRetries(3)
	bp.SetBaseDelay(50 * time.Millisecond)
	
	callTimes := []time.Time{}
	fn := func() error {
		callTimes = append(callTimes, time.Now())
		return errors.New("error")
	}
	
	_ = bp.RetryWithBackoff(fn)
	
	// Verify exponential backoff delays
	if len(callTimes) != 4 {
		t.Fatalf("Expected 4 calls, got %d", len(callTimes))
	}
	
	// Check delays between calls (approximately)
	// Delay 1: ~50ms (2^0 * 50ms)
	// Delay 2: ~100ms (2^1 * 50ms)
	// Delay 3: ~200ms (2^2 * 50ms)
	
	delay1 := callTimes[1].Sub(callTimes[0])
	delay2 := callTimes[2].Sub(callTimes[1])
	delay3 := callTimes[3].Sub(callTimes[2])
	
	// Allow some tolerance for timing
	tolerance := 30 * time.Millisecond
	
	if delay1 < 50*time.Millisecond-tolerance || delay1 > 50*time.Millisecond+tolerance {
		t.Errorf("Expected first delay ~50ms, got %v", delay1)
	}
	if delay2 < 100*time.Millisecond-tolerance || delay2 > 100*time.Millisecond+tolerance {
		t.Errorf("Expected second delay ~100ms, got %v", delay2)
	}
	if delay3 < 200*time.Millisecond-tolerance || delay3 > 200*time.Millisecond+tolerance {
		t.Errorf("Expected third delay ~200ms, got %v", delay3)
	}
}

func TestBaseProvider_DiscoverModels(t *testing.T) {
	bp := NewBaseProvider("test-provider")
	
	models, err := bp.DiscoverModels()
	if err == nil {
		t.Error("Expected error for unsupported dynamic discovery")
	}
	if models != nil {
		t.Error("Expected nil models for unsupported dynamic discovery")
	}
}

func TestBaseProvider_GetRateLimitInfo(t *testing.T) {
	bp := NewBaseProvider("test-provider")
	
	info, err := bp.GetRateLimitInfo()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if info != nil {
		t.Error("Expected nil rate limit info for default implementation")
	}
}

func TestBaseProvider_GetQuotaInfo(t *testing.T) {
	bp := NewBaseProvider("test-provider")
	
	info, err := bp.GetQuotaInfo()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if info != nil {
		t.Error("Expected nil quota info for default implementation")
	}
}

func TestBaseProvider_SupportsCodingPlan(t *testing.T) {
	bp := NewBaseProvider("test-provider")
	
	if bp.SupportsCodingPlan() {
		t.Error("Expected false for default implementation")
	}
}
