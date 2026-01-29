package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mojomast/geoffrussy/internal/config"
	"github.com/mojomast/geoffrussy/internal/quota"
	"github.com/mojomast/geoffrussy/internal/state"
	"github.com/spf13/cobra"
)

var (
	quotaRefresh bool
)

var quotaCmd = &cobra.Command{
	Use:   "quota",
	Short: "Check rate limits and quotas",
	Long: `Check rate limits and quotas for all configured providers.
Displays warnings if approaching limits.

Use --refresh to force a refresh from providers (requires providers to be configured).`,
	RunE: runQuota,
}

func init() {
	quotaCmd.Flags().BoolVar(&quotaRefresh, "refresh", false, "Force refresh from providers")
}

func runQuota(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfgMgr := config.NewManager()
	if err := cfgMgr.Load(nil); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	cfg := cfgMgr.GetConfig()

	// Initialize state store
	configDir := filepath.Dir(cfg.ConfigPath)
	dbPath := filepath.Join(configDir, "geoffrussy.db")
	store, err := state.NewStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize state store: %w", err)
	}
	defer store.Close()

	// Create quota monitor
	monitor := quota.NewMonitor(store)

	// Get list of configured providers
	providers := getConfiguredProviders(cfg)

	if len(providers) == 0 {
		fmt.Println("ℹ️  No providers configured yet.")
		fmt.Println("   Run 'geoffrussy init' to configure API providers.")
		return nil
	}

	fmt.Println("📊 Rate Limits & Quotas")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	if quotaRefresh {
		fmt.Println("⚠️  Refresh flag specified, but provider instances not available.")
		fmt.Println("   Showing cached data instead.")
		fmt.Println()
	}

	// Display status for each provider
	hasAnyData := false
	for _, providerName := range providers {
		status, err := monitor.GetCachedStatus(providerName)
		if err != nil {
			fmt.Printf("❌ %s: Error retrieving status - %v\n\n", providerName, err)
			continue
		}

		// Skip if no data available
		if status.RateLimitInfo == nil && status.QuotaInfo == nil {
			continue
		}

		hasAnyData = true
		displayProviderStatus(providerName, status)
	}

	if !hasAnyData {
		fmt.Println("ℹ️  No quota data available yet.")
		fmt.Println("   Quota data is collected automatically during API calls.")
		fmt.Println()
	}

	return nil
}

// getConfiguredProviders returns list of providers with API keys
func getConfiguredProviders(cfg *config.Config) []string {
	var providers []string
	for provider := range cfg.APIKeys {
		providers = append(providers, provider)
	}
	return providers
}

// displayProviderStatus displays the status for a single provider
func displayProviderStatus(providerName string, status *quota.ProviderStatus) {
	fmt.Printf("🔌 %s\n", strings.ToUpper(providerName))
	fmt.Println(strings.Repeat("─", 60))

	// Display rate limit info
	if status.RateLimitInfo != nil {
		fmt.Println("\n  Rate Limits:")
		info := status.RateLimitInfo

		if status.RateLimitWarning != nil {
			symbol := quota.GetWarningSymbol(status.RateLimitWarning.Level)
			fmt.Printf("  %s %s\n", symbol, status.RateLimitWarning.Message)
			fmt.Printf("     Resets: %s\n", info.ResetAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("     Checked: %s ago\n", formatTimeSince(info.CheckedAt))
		} else {
			fmt.Printf("  ✅ %d / %d requests remaining\n", info.RequestsRemaining, info.RequestsLimit)
			fmt.Printf("     Resets: %s\n", info.ResetAt.Format("2006-01-02 15:04:05"))
		}
	}

	// Display quota info
	if status.QuotaInfo != nil {
		fmt.Println("\n  Quotas:")
		info := status.QuotaInfo

		// Tokens quota
		if info.TokensLimit != nil && *info.TokensLimit > 0 {
			remaining := 0
			if info.TokensRemaining != nil {
				remaining = *info.TokensRemaining
			}

			if status.QuotaWarning != nil {
				symbol := quota.GetWarningSymbol(status.QuotaWarning.Level)
				fmt.Printf("  %s Tokens: %s\n", symbol, status.QuotaWarning.Message)
			} else {
				fmt.Printf("  ✅ Tokens: %d / %d remaining\n", remaining, *info.TokensLimit)
			}
		}

		// Cost quota
		if info.CostLimit != nil && *info.CostLimit > 0 {
			remaining := 0.0
			if info.CostRemaining != nil {
				remaining = *info.CostRemaining
			}

			fmt.Printf("     Cost: $%.2f / $%.2f remaining\n", remaining, *info.CostLimit)
		}

		fmt.Printf("     Resets: %s\n", info.ResetAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("     Checked: %s ago\n", formatTimeSince(info.CheckedAt))
	}

	// Health status
	fmt.Println()
	if status.IsHealthy {
		fmt.Println("  ✅ Healthy - OK to make requests")
	} else {
		fmt.Println("  ⚠️  Unhealthy - Approaching or exceeded limits")
	}

	if status.ShouldDelay {
		fmt.Printf("  ⏸️  Recommended delay: %s\n", formatDuration(status.RecommendedDelay))
	}

	fmt.Println()
}


