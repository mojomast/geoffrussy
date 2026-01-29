package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mojomast/geoffrussy/internal/config"
	"github.com/spf13/cobra"
)

var configListProviders bool
var configSetKey bool
var configSetModel bool

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Geoffrey configuration",
	Long: `Manage Geoffrey configuration including API keys, provider selection,
 and default models for each pipeline stage.`,
	RunE: runConfig,
}

func init() {
	configCmd.Flags().BoolVar(&configListProviders, "list-providers", false, "List available providers and their models")
	configCmd.Flags().BoolVar(&configSetKey, "set-key", false, "Set API key interactively")
	configCmd.Flags().BoolVar(&configSetModel, "set-model", false, "Set default model for a stage")
}

func runConfig(cmd *cobra.Command, args []string) error {
	if configListProviders {
		return listProvidersAndModels()
	}

	if configSetKey {
		return setAPIKeyInteractive()
	}

	if configSetModel {
		return setDefaultModelInteractive()
	}

	return showConfigMenu()
}

func showConfigMenu() error {
	cfgMgr := config.NewManager()
	if err := cfgMgr.Load(nil); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	for {
		fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
		fmt.Println("║           Configuration Management                                ║")
		fmt.Println("╚════════════════════════════════════════════════════════════╝")
		fmt.Println()
		cfg := cfgMgr.GetConfig()
		displayCurrentConfig(cfg)
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  1) 🔑 Set/Update API Key")
		fmt.Println("  2) 🤖 Set Default Model for Stage")
		fmt.Println("  3) 📋 List Available Providers & Models")
		fmt.Println("  4) 💰 Set Budget Limit")
		fmt.Println("  5) 🔍 Toggle Verbose Logging")
		fmt.Println("  6) 💾 Save and Exit")
		fmt.Println("  q) Quit (Exit without Saving)")
		fmt.Println()
		fmt.Print("Select option: ")

		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			if err := setAPIKeyInteractive(); err != nil {
				fmt.Printf("⚠️  Error: %v\n", err)
			}
		case "2":
			if err := setDefaultModelInteractive(); err != nil {
				fmt.Printf("⚠️  Error: %v\n", err)
			}
		case "3":
			if err := listProvidersAndModels(); err != nil {
				fmt.Printf("⚠️  Error: %v\n", err)
			}
		case "4":
			if err := setBudgetLimitInteractive(cfgMgr); err != nil {
				fmt.Printf("⚠️  Error: %v\n", err)
			}
		case "5":
			cfg := cfgMgr.GetConfig()
			cfg.VerboseLogging = !cfg.VerboseLogging
			if cfg.VerboseLogging {
				fmt.Println("✅ Verbose logging enabled")
			} else {
				fmt.Println("✅ Verbose logging disabled")
			}
		case "6":
			if err := cfgMgr.Save(); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}
			fmt.Println("✅ Configuration saved!")
			return nil
		case "q", "Q":
			fmt.Println("❌ Exiting without saving...")
			return nil
		default:
			fmt.Println("⚠️  Invalid option")
		}
	}
}

func displayCurrentConfig(cfg *config.Config) {
	fmt.Println("Current Configuration:")
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Println("\n🔑 API Keys:")
	if len(cfg.APIKeys) == 0 {
		fmt.Println("   None configured")
	} else {
		for provider := range cfg.APIKeys {
			masked := maskAPIKey(cfg.APIKeys[provider])
			fmt.Printf("   %s: %s\n", provider, masked)
		}
	}

	fmt.Println("\n🤖 Default Models:")
	if len(cfg.DefaultModels) == 0 {
		fmt.Println("   None configured")
	} else {
		for stage, model := range cfg.DefaultModels {
			fmt.Printf("   %s: %s\n", stage, model)
		}
	}

	fmt.Printf("\n💰 Budget Limit: $%.2f\n", cfg.BudgetLimit)
	if cfg.VerboseLogging {
		fmt.Println("🔍 Verbose Logging: ✅ Enabled")
	} else {
		fmt.Println("🔍 Verbose Logging: ❌ Disabled")
	}
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

func listProvidersAndModels() error {
	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║          Available Providers & Models                         ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	providers := []struct {
		name    string
		display string
		models  []string
	}{
		{"openai", "OpenAI", []string{"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"}},
		{"anthropic", "Anthropic", []string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku", "claude-2.1"}},
		{"ollama", "Ollama (Local)", []string{"llama2", "mistral", "neural-chat", "codellama"}},
		{"firmware", "Firmware.ai", []string{"claude-3-opus", "claude-3-sonnet", "gpt-4"}},
		{"requesty", "Requesty.ai", []string{"claude-3-opus", "claude-3-sonnet", "gpt-4"}},
		{"zai", "Z.ai", []string{"zai-c3", "zai-c3-turbo"}},
		{"kimi", "Kimi", []string{"moonshot-v1-32k", "moonshot-v1-128k"}},
		{"opencode", "OpenCode", []string{"opencode-1", "opencode-2"}},
	}

	for _, p := range providers {
		fmt.Printf("\n📦 %s\n", p.display)
		fmt.Println("   Models:")
		for _, model := range p.models {
			fmt.Printf("      • %s\n", model)
		}
	}

	fmt.Println("\n💡 Tip: Use provider names with 'geoffrussy config --set-key'")
	fmt.Println("   Available providers: openai, anthropic, ollama, firmware, requesty, zai, kimi, opencode")
	return nil
}

func setAPIKeyInteractive() error {
	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              Set API Key                                       ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	providers := []struct {
		name   string
		prompt string
	}{
		{"openai", "OpenAI"},
		{"anthropic", "Anthropic"},
		{"ollama", "Ollama"},
		{"firmware", "Firmware.ai"},
		{"requesty", "Requesty.ai"},
		{"zai", "Z.ai"},
		{"kimi", "Kimi"},
		{"opencode", "OpenCode"},
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Select a provider to configure (or type 'cancel'):")
	for i, p := range providers {
		fmt.Printf("  %d) %s\n", i+1, p.prompt)
	}
	fmt.Print("\nSelection: ")

	selection, _ := reader.ReadString('\n')
	selection = strings.TrimSpace(selection)

	if selection == "cancel" {
		fmt.Println("❌ Cancelled")
		return nil
	}

	index := 0
	if _, err := fmt.Sscanf(selection, "%d", &index); err != nil || index < 1 || index > len(providers) {
		return fmt.Errorf("invalid selection")
	}

	selected := providers[index-1]
	fmt.Printf("\nEnter API Key for %s (or press Enter to skip): ", selected.prompt)
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		fmt.Println("⏭️  Skipped")
		return nil
	}

	cfgMgr := config.NewManager()
	if err := cfgMgr.Load(nil); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := cfgMgr.SetAPIKey(selected.name, apiKey); err != nil {
		return fmt.Errorf("failed to set API key: %w", err)
	}

	if err := cfgMgr.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("✅ API key configured!")
	return nil
}

func setDefaultModelInteractive() error {
	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║            Set Default Model for Stage                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	stages := []string{"interview", "design", "devplan", "review", "develop"}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Select a stage:")
	for i, stage := range stages {
		fmt.Printf("  %d) %s\n", i+1, stage)
	}
	fmt.Print("\nSelection: ")

	selection, _ := reader.ReadString('\n')
	selection = strings.TrimSpace(selection)

	index := 0
	if _, err := fmt.Sscanf(selection, "%d", &index); err != nil || index < 1 || index > len(stages) {
		return fmt.Errorf("invalid selection")
	}

	selectedStage := stages[index-1]

	fmt.Printf("\nCurrent configured models:\n")
	cfgMgr := config.NewManager()
	if err := cfgMgr.Load(nil); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	cfg := cfgMgr.GetConfig()

	if len(cfg.DefaultModels) == 0 {
		fmt.Println("   None configured")
	} else {
		for stage, model := range cfg.DefaultModels {
			fmt.Printf("   %s: %s\n", stage, model)
		}
	}

	fmt.Printf("\nEnter model for %s stage: ", selectedStage)
	model, _ := reader.ReadString('\n')
	model = strings.TrimSpace(model)

	if model == "" {
		fmt.Println("⏭️  Skipped")
		return nil
	}

	if err := cfgMgr.SetDefaultModel(selectedStage, model); err != nil {
		return fmt.Errorf("failed to set default model: %w", err)
	}

	if err := cfgMgr.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✅ Default model for %s set to %s\n", selectedStage, model)
	return nil
}

func setBudgetLimitInteractive(cfgMgr *config.Manager) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter budget limit in USD (or 0 for unlimited): ")
	limitStr, _ := reader.ReadString('\n')
	limitStr = strings.TrimSpace(limitStr)

	var limit float64
	if _, err := fmt.Sscanf(limitStr, "%f", &limit); err != nil {
		return fmt.Errorf("invalid budget limit: %w", err)
	}

	cfg := cfgMgr.GetConfig()
	cfg.BudgetLimit = limit

	if err := cfgMgr.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	if limit > 0 {
		fmt.Printf("✅ Budget limit set to $%.2f\n", limit)
	} else {
		fmt.Println("✅ Budget limit removed (unlimited)")
	}
	return nil
}
