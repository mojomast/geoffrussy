package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mojomast/geoffrussy/internal/config"
	"github.com/mojomast/geoffrussy/internal/provider"
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
		fmt.Println("  7) ⭐ Manage Favorite Models")
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
		case "7":
			if err := manageFavoritesInteractive(cfgMgr); err != nil {
				fmt.Printf("⚠️  Error: %v\n", err)
			}
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

	cfgMgr := config.NewManager()
	if err := cfgMgr.Load(nil); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	cfg := cfgMgr.GetConfig()

	providerNames := provider.GetProviderNames()

	for _, name := range providerNames {
		fmt.Printf("\n📦 %s\n", strings.Title(name))

		// Create provider
		p, err := provider.CreateProvider(name)
		if err != nil {
			fmt.Printf("   Error: %v\n", err)
			continue
		}

		// Check authentication
		isAuthenticated := false
		var authErr error

		if name == "ollama" {
			// Ollama doesn't need API key
			if err := p.Authenticate(""); err == nil {
				isAuthenticated = true
			} else {
				authErr = err
			}
		} else {
			if key, ok := cfg.APIKeys[name]; ok && key != "" {
				if err := p.Authenticate(key); err == nil {
					isAuthenticated = true
				} else {
					authErr = err
				}
			}
		}

		if !isAuthenticated {
			if name == "ollama" {
				if authErr != nil {
					fmt.Printf("   ⚠️  Connection failed: %v\n", authErr)
				} else {
					fmt.Println("   ⚠️  Not connected (is Ollama running?)")
				}
			} else {
				if authErr != nil {
					fmt.Printf("   ⚠️  Authentication failed: %v\n", authErr)
				} else {
					fmt.Println("   ⚠️  Not configured (set API key to see models)")
				}
			}
			continue
		}

		// List models
		models, err := p.ListModels()
		if err != nil {
			fmt.Printf("   Error listing models: %v\n", err)
			continue
		}

		fmt.Println("   Models:")
		for _, model := range models {
			fmt.Printf("      • %s\n", model.Name)
		}
	}

	fmt.Println("\n💡 Tip: Use 'geoffrussy config --set-key' to configure providers")
	return nil
}

func setAPIKeyInteractive() error {
	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              Set API Key                                       ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	providerNames := provider.GetProviderNames()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Select a provider to configure (or type 'cancel'):")
	for i, name := range providerNames {
		fmt.Printf("  %d) %s\n", i+1, strings.Title(name))
	}
	fmt.Print("\nSelection: ")

	selection, _ := reader.ReadString('\n')
	selection = strings.TrimSpace(selection)

	if selection == "cancel" {
		fmt.Println("❌ Cancelled")
		return nil
	}

	index := 0
	if _, err := fmt.Sscanf(selection, "%d", &index); err != nil || index < 1 || index > len(providerNames) {
		return fmt.Errorf("invalid selection")
	}

	selectedName := providerNames[index-1]

	fmt.Printf("\nEnter API Key for %s (or press Enter to skip): ", strings.Title(selectedName))
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

	if err := cfgMgr.SetAPIKey(selectedName, apiKey); err != nil {
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

	cfgMgr := config.NewManager()
	if err := cfgMgr.Load(nil); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	cfg := cfgMgr.GetConfig()

	fmt.Printf("\nCurrent configured models:\n")
	if len(cfg.DefaultModels) == 0 {
		fmt.Println("   None configured")
	} else {
		for stage, model := range cfg.DefaultModels {
			fmt.Printf("   %s: %s\n", stage, model)
		}
	}

	fmt.Println("\nFetching available models...")

	bridge := provider.NewBridge()
	providerNames := provider.GetProviderNames()

	for _, name := range providerNames {
		if err := setupProvider(bridge, cfgMgr, name); err != nil {
			continue
		}
	}

	allModels, err := bridge.ListModels()
	if err != nil || len(allModels) == 0 {
		fmt.Println("⚠️  No models found. Configure providers first.")
		return nil
	}

	// Separate favorites
	var favorites []provider.Model
	var others []provider.Model

	for _, m := range allModels {
		if cfgMgr.IsFavoriteModel(m.Name) {
			favorites = append(favorites, m)
		} else {
			others = append(others, m)
		}
	}

	// Reconstruct sorted list
	sortedModels := append(favorites, others...)

	fmt.Println("\nAvailable Models:")
	fmt.Println("─────────────────────────────────────────────────────")
	for i, m := range sortedModels {
		prefix := "  "
		if cfgMgr.IsFavoriteModel(m.Name) {
			prefix = "⭐ "
		}
		fmt.Printf("  %d) %s%s (%s)\n", i+1, prefix, m.Name, strings.Title(m.Provider))
	}

	fmt.Printf("\nEnter model for %s stage (1-%d): ", selectedStage, len(sortedModels))
	modelInput, _ := reader.ReadString('\n')
	modelInput = strings.TrimSpace(modelInput)

	modelIndex := 0
	if _, err := fmt.Sscanf(modelInput, "%d", &modelIndex); err != nil || modelIndex < 1 || modelIndex > len(sortedModels) {
		return fmt.Errorf("invalid selection. Please enter a number between 1 and %d", len(sortedModels))
	}

	selectedModel := sortedModels[modelIndex-1]

	if err := cfgMgr.SetDefaultModel(selectedStage, selectedModel.Name); err != nil {
		return fmt.Errorf("failed to set default model: %w", err)
	}

	if err := cfgMgr.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✅ Default model for %s set to %s (%s)\n", selectedStage, selectedModel.Name, strings.Title(selectedModel.Provider))
	return nil
}

func manageFavoritesInteractive(cfgMgr *config.Manager) error {
	for {
		fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
		fmt.Println("║            Manage Favorite Models                             ║")
		fmt.Println("╚════════════════════════════════════════════════════════════╝")
		fmt.Println()

		favorites := cfgMgr.GetFavoriteModels()
		if len(favorites) == 0 {
			fmt.Println("No favorite models configured.")
		} else {
			fmt.Println("Current Favorites:")
			for _, fav := range favorites {
				fmt.Printf("  ⭐ %s\n", fav)
			}
		}

		fmt.Println("\nOptions:")
		fmt.Println("  1) ➕ Add Favorite")
		fmt.Println("  2) ➖ Remove Favorite")
		fmt.Println("  b) Back to Main Menu")
		fmt.Println()
		fmt.Print("Select option: ")

		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			if err := addFavoriteInteractive(cfgMgr); err != nil {
				fmt.Printf("⚠️  Error: %v\n", err)
			}
		case "2":
			if err := removeFavoriteInteractive(cfgMgr); err != nil {
				fmt.Printf("⚠️  Error: %v\n", err)
			}
		case "b", "B":
			return nil
		default:
			fmt.Println("⚠️  Invalid option")
		}
	}
}

func addFavoriteInteractive(cfgMgr *config.Manager) error {
	fmt.Println("\nFetching available models...")

	bridge := provider.NewBridge()
	providerNames := provider.GetProviderNames()

	for _, name := range providerNames {
		if err := setupProvider(bridge, cfgMgr, name); err != nil {
			continue
		}
	}

	allModels, err := bridge.ListModels()
	if err != nil || len(allModels) == 0 {
		return fmt.Errorf("no models found. Configure providers first")
	}

	fmt.Println("\nSelect model to favorite:")
	fmt.Println("─────────────────────────────────────────────────────")
	for i, m := range allModels {
		prefix := "   "
		if cfgMgr.IsFavoriteModel(m.Name) {
			prefix = "⭐ "
		}
		fmt.Printf("  %d) %s%s (%s)\n", i+1, prefix, m.Name, strings.Title(m.Provider))
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nEnter number (1-%d): ", len(allModels))

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	index := 0
	if _, err := fmt.Sscanf(input, "%d", &index); err != nil || index < 1 || index > len(allModels) {
		return fmt.Errorf("invalid selection")
	}

	selected := allModels[index-1]
	if err := cfgMgr.AddFavoriteModel(selected.Name); err != nil {
		return err
	}

	if err := cfgMgr.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Added %s to favorites\n", selected.Name)
	return nil
}

func removeFavoriteInteractive(cfgMgr *config.Manager) error {
	favorites := cfgMgr.GetFavoriteModels()
	if len(favorites) == 0 {
		return fmt.Errorf("no favorites to remove")
	}

	fmt.Println("\nSelect favorite to remove:")
	fmt.Println("─────────────────────────────────────────────────────")
	for i, fav := range favorites {
		fmt.Printf("  %d) %s\n", i+1, fav)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nEnter number (1-%d): ", len(favorites))

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	index := 0
	if _, err := fmt.Sscanf(input, "%d", &index); err != nil || index < 1 || index > len(favorites) {
		return fmt.Errorf("invalid selection")
	}

	selected := favorites[index-1]
	if err := cfgMgr.RemoveFavoriteModel(selected); err != nil {
		return err
	}

	if err := cfgMgr.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Removed %s from favorites\n", selected)
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
