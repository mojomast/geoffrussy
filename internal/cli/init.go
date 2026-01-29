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
	"github.com/mojomast/geoffrussy/internal/state"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Geoffrey in the current project",
	Long: `Initialize Geoffrey in the current project by creating configuration
directory structure and prompting for API keys.`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Initializing Geoffrey...")

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

	// Initialize configuration manager
	cfgManager := config.NewManager()

	// Check if config already exists
	configPath := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("⚠️  Configuration file already exists")
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

	fmt.Println("\n✨ Geoffrey initialized successfully!")
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

	providerModels := map[string][]string{
		"openai":    {"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"},
		"anthropic": {"claude-3-opus", "claude-3-sonnet", "claude-3-haiku", "claude-2.1"},
		"ollama":    {"llama2", "mistral", "neural-chat", "codellama"},
		"firmware":  {"claude-3-opus", "claude-3-sonnet", "gpt-4"},
		"requesty":  {"claude-3-opus", "claude-3-sonnet", "gpt-4"},
		"zai":       {"zai-c3", "zai-c3-turbo"},
		"kimi":      {"moonshot-v1-32k", "moonshot-v1-128k"},
		"opencode":  {"opencode-1", "opencode-2"},
	}

	providerNames := map[string]string{
		"openai":    "OpenAI",
		"anthropic": "Anthropic",
		"ollama":    "Ollama (Local)",
		"firmware":  "Firmware.ai",
		"requesty":  "Requesty.ai",
		"zai":       "Z.ai",
		"kimi":      "Kimi",
		"opencode":  "OpenCode",
	}

	if len(cfg.APIKeys) == 0 {
		fmt.Println("⚠️  No API keys configured. Run 'geoffrussy config' to add keys.")
		return
	}

	for provider := range cfg.APIKeys {
		models, ok := providerModels[provider]
		if !ok {
			continue
		}
		fmt.Printf("\n📦 %s:\n", providerNames[provider])
		for _, model := range models {
			fmt.Printf("   • %s\n", model)
		}
	}
}
