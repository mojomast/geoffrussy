package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mojomast/geoffrussy/internal/config"
	"github.com/mojomast/geoffrussy/internal/git"
	"github.com/mojomast/geoffrussy/internal/provider"
	"github.com/mojomast/geoffrussy/internal/state"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Geoffrussy in the current project",
	Long: `Initialize Geoffrussy in the current project by creating configuration
directory structure and prompting for API keys.`,
	RunE: runInit,
}

// flags for non-interactive configuration
var (
	flagAPIKeyOpenAI    string
	flagAPIKeyAnthropic string
	flagAPIKeyFirmware  string
	flagAPIKeyRequesty  string
	flagAPIKeyZAI       string
	flagAPIKeyKimi      string
	flagNonInteractive  bool
)

func init() {
	// Define flags for API keys
	initCmd.Flags().StringVar(&flagAPIKeyOpenAI, "api-key-openai", "", "OpenAI API key")
	initCmd.Flags().StringVar(&flagAPIKeyAnthropic, "api-key-anthropic", "", "Anthropic API key")
	initCmd.Flags().StringVar(&flagAPIKeyFirmware, "api-key-firmware", "", "Firmware.ai API key")
	initCmd.Flags().StringVar(&flagAPIKeyRequesty, "api-key-requesty", "", "Requesty.ai API key")
	initCmd.Flags().StringVar(&flagAPIKeyZAI, "api-key-zai", "", "Z.ai API key")
	initCmd.Flags().StringVar(&flagAPIKeyKimi, "api-key-kimi", "", "Kimi API key")
	initCmd.Flags().BoolVar(&flagNonInteractive, "non-interactive", false, "Run in non-interactive mode (no prompts)")

	// Mark flags as required
	initCmd.MarkFlagRequired("api-key-openai")
	initCmd.MarkFlagRequired("api-key-anthropic")
	initCmd.MarkFlagRequired("api-key-firmware")
	initCmd.MarkFlagRequired("api-key-requesty")
	initCmd.MarkFlagRequired("api-key-zai")
	initCmd.MarkFlagRequired("api-key-kimi")
	initCmd.MarkFlagRequired("non-interactive")
}

// validateConfiguration validates the configuration before running init
func validateConfiguration(cmd *cobra.Command, args []string) error {
	if flagNonInteractive {
		return validateNonInteractiveConfig()
	}
	return nil
}

// validateNonInteractiveConfig validates the non-interactive configuration
func validateNonInteractiveConfig() error {
	// Check if required API keys are provided via flags
	requiredProviders := []string{"openai", "anthropic", "firmware", "requesty", "zai", "kimi"}

	// Load existing config
	cfgManager := config.NewManager()
	if err := cfgManager.Load(nil); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := cfgManager.GetConfig()

	for _, provider := range requiredProviders {
		// Priority: flag > environment variable > config file
		key, err := getAPIKey(provider)
		if err != nil {
			return fmt.Errorf("API key not found for %s. Use --api-key-%s flag or set %s environment variable", provider, provider, strings.ToUpper(provider)+"_API_KEY")
		}
		cfg.APIKeys[provider] = key
	}

	// Save the configuration
	if err := cfgManager.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// getAPIKey retrieves the API key for a provider, checking in order: flag, environment variable, config
func getAPIKey(provider string) (string, error) {
	// Check flag first
	switch provider {
	case "openai":
		if flagAPIKeyOpenAI != "" {
			return flagAPIKeyOpenAI, nil
		}
	case "anthropic":
		if flagAPIKeyAnthropic != "" {
			return flagAPIKeyAnthropic, nil
		}
	case "firmware":
		if flagAPIKeyFirmware != "" {
			return flagAPIKeyFirmware, nil
		}
	case "requesty":
		if flagAPIKeyRequesty != "" {
			return flagAPIKeyRequesty, nil
		}
	case "zai":
		if flagAPIKeyZAI != "" {
			return flagAPIKeyZAI, nil
		}
	case "kimi":
		if flagAPIKeyKimi != "" {
			return flagAPIKeyKimi, nil
		}
	}

	// Check environment variable
	envVar := "GEOFFRUSSY_" + strings.ToUpper(provider) + "_API_KEY"
	if apiKey := os.Getenv(envVar); apiKey != "" {
		return apiKey, nil
	}

	// Check config file
	cfgManager := config.NewManager()
	if err := cfgManager.Load(nil); err != nil {
		return "", fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := cfgManager.GetConfig()
	if apiKey, ok := cfg.APIKeys[provider]; ok {
		return apiKey, nil
	}

	return "", fmt.Errorf("no API key found for %s", provider)
}

func runInit(cmd *cobra.Command, args []string) error {
	if flagNonInteractive {
		return runInitNonInteractive(cmd, args)
	}
	return runInitInteractive(cmd, args)
}

func runInitInteractive(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Initializing Geoffrussy...")

	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create configuration directory
	configDir := filepath.Join(os.Getenv("HOME"), ".geoffrussy")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	fmt.Printf("✓ Created configuration directory: %s\n", configDir)

	// Initialize configuration manager and load existing config
	cfgManager := config.NewManager()
	if err := cfgManager.Load(nil); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if config already has API keys
	cfg := cfgManager.GetConfig()
	if len(cfg.APIKeys) > 0 {
		fmt.Println("⚠️  Configuration file already exists with API keys")
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Do you want to reconfigure? (y/N): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Skipping configuration...")
		} else {
			if err := promptForAPIKeys(cfgManager); err != nil {
				return err
			}
		}
	} else {
		if err := promptForAPIKeys(cfgManager); err != nil {
			return err
		}
	}

	// Save configuration
	if err := cfgManager.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}
	fmt.Println("✓ Configuration saved")

	// Initialize database
	dbPath := filepath.Join(cwd, ".geoffrussy", "state.db")
	store, err := state.NewStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer store.Close()
	fmt.Printf("✓ Initialized database: %s\n", dbPath)

	// Create or update project in state store
	projectID := filepath.Base(cwd)
	project := &state.Project{
		ID:           projectID,
		Name:         projectID,
		CreatedAt:    time.Now(),
		CurrentStage: state.StageInit,
		CurrentPhase: "",
	}

	// Check if project exists
	existingProject, err := store.GetProject(projectID)
	if err != nil {
		// Project doesn't exist, create it
		if err := store.CreateProject(project); err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}
		fmt.Printf("✓ Created project: %s\n", projectID)
	} else {
		// Project exists, update it
		existingProject.CurrentStage = state.StageInit
		existingProject.Name = projectID
		if err := store.UpdateProject(existingProject); err != nil {
			return fmt.Errorf("failed to update project: %w", err)
		}
		fmt.Printf("✓ Updated project: %s\n", projectID)
	}

	// Initialize Git repository if needed
	gitManager := git.NewManager(cwd)
	isRepo, err := gitManager.IsRepository()
	if err != nil {
		return fmt.Errorf("failed to check git repository: %w", err)
	}

	if !isRepo {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Initialize Git repository? (Y/n): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response == "" || response == "y" || response == "yes" {
			if err := gitManager.Initialize(); err != nil {
				return fmt.Errorf("failed to initialize git repository: %w", err)
			}
			fmt.Println("✓ Initialized Git repository")
		}
	} else {
		fmt.Println("✓ Git repository already initialized")
	}

	fmt.Println("\n✨ Geoffrussy initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Run 'geoffrussy interview' to start the project interview")
	fmt.Println("  2. Run 'geoffrussy design' to generate architecture")
	fmt.Println("  3. Run 'geoffrussy plan' to create development plan")
	fmt.Println("  4. Run 'geoffrussy develop' to start implementation")

	return nil
}

func runInitNonInteractive(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Initializing Geoffrussy (non-interactive)...")

	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create configuration directory
	configDir := filepath.Join(os.Getenv("HOME"), ".geoffrussy")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	fmt.Printf("✓ Created configuration directory: %s\n", configDir)

	// Initialize configuration manager and load existing config
	cfgManager := config.NewManager()
	if err := cfgManager.Load(nil); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// In non-interactive mode, configuration should already be set by validateNonInteractiveConfig
	// Just verify it's set
	cfg := cfgManager.GetConfig()
	if len(cfg.APIKeys) == 0 {
		return fmt.Errorf("no API keys configured. Please set them using environment variables or flags")
	}
	fmt.Println("✓ Configuration loaded")

	// Initialize database
	dbPath := filepath.Join(cwd, ".geoffrussy", "state.db")
	store, err := state.NewStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer store.Close()
	fmt.Printf("✓ Initialized database: %s\n", dbPath)

	// Create or update project in state store
	projectID := filepath.Base(cwd)
	project := &state.Project{
		ID:           projectID,
		Name:         projectID,
		CreatedAt:    time.Now(),
		CurrentStage: state.StageInit,
		CurrentPhase: "",
	}

	// Check if project exists
	existingProject, err := store.GetProject(projectID)
	if err != nil {
		// Project doesn't exist, create it
		if err := store.CreateProject(project); err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}
		fmt.Printf("✓ Created project: %s\n", projectID)
	} else {
		// Project exists, update it
		existingProject.CurrentStage = state.StageInit
		existingProject.Name = projectID
		if err := store.UpdateProject(existingProject); err != nil {
			return fmt.Errorf("failed to update project: %w", err)
		}
		fmt.Printf("✓ Updated project: %s\n", projectID)
	}

	fmt.Println("\n✨ Geoffrussy initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Run 'geoffrussy interview' to start the project interview")
	fmt.Println("  2. Run 'geoffrussy design' to generate architecture")
	fmt.Println("  3. Run 'geoffrussy plan' to create development plan")
	fmt.Println("  4. Run 'geoffrussy develop' to start implementation")

	return nil
}

func promptForAPIKeys(cfgManager *config.Manager) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n📝 API Key Configuration")
	fmt.Println("Enter API keys for the providers you want to use (press Enter to skip):")

	providers := []struct {
		name string
		key  string
	}{
		{"OpenAI", "openai"},
		{"Anthropic", "anthropic"},
		{"Firmware.ai", "firmware"},
		{"Requesty.ai", "requesty"},
		{"Z.ai", "zai"},
		{"Kimi", "kimi"},
	}

	for _, provider := range providers {
		fmt.Printf("\n%s API Key: ", provider.name)
		apiKey, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(apiKey)
		if apiKey != "" {
			cfgManager.SetAPIKey(provider.key, apiKey)
			fmt.Printf("✓ %s API key configured\n", provider.name)
		}
	}

	// Prompt for default model
	fmt.Println("\n📦 Available Models:")
	fmt.Println("─────────────────────────────────────────────────────")
	displayConfiguredModels(cfgManager)

	fmt.Print("\nDefault model for interview stage (select from above): ")
	defaultModel, _ := reader.ReadString('\n')
	defaultModel = strings.TrimSpace(defaultModel)
	if defaultModel != "" {
		cfgManager.SetDefaultModel("interview", defaultModel)
		fmt.Printf("✓ Default interview model set to: %s\n", defaultModel)
	}

	return nil
}

func displayConfiguredModels(cfgMgr *config.Manager) {
	cfg := cfgMgr.GetConfig()

	if len(cfg.APIKeys) == 0 {
		fmt.Println("⚠️  No API keys configured. Run 'geoffrussy config' to add keys.")
		return
	}

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
		return
	}

	modelsByProvider := make(map[string][]string)
	for _, m := range allModels {
		modelsByProvider[m.Provider] = append(modelsByProvider[m.Provider], m.Name)
	}

	providerDisplayNames := map[string]string{
		"openai":    "OpenAI",
		"anthropic": "Anthropic",
		"ollama":    "Ollama (Local)",
		"firmware":  "Firmware.ai",
		"requesty":  "Requesty.ai",
		"zai":       "Z.ai",
		"kimi":      "Kimi",
		"opencode":  "OpenCode",
	}

	for provider := range cfg.APIKeys {
		models, ok := modelsByProvider[provider]
		if !ok {
			continue
		}
		displayName := providerDisplayNames[provider]
		if displayName == "" {
			displayName = strings.Title(provider)
		}
		fmt.Printf("\n📦 %s:\n", displayName)
		for _, model := range models {
			fmt.Printf("   • %s\n", model)
		}
	}
}
