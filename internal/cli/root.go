package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mojomast/geoffrussy/internal/checkpoint"
	"github.com/mojomast/geoffrussy/internal/config"
	"github.com/mojomast/geoffrussy/internal/git"
	"github.com/mojomast/geoffrussy/internal/resume"
	"github.com/mojomast/geoffrussy/internal/state"
	"github.com/spf13/cobra"
)

var (
	version string
	cfgFile string
	verbose bool
	rootCmd *cobra.Command
)

// Execute runs the root command
func Execute(ver string) error {
	version = ver
	return rootCmd.Execute()
}

func init() {
	rootCmd = &cobra.Command{
		Use:   "geoffrussy",
		Short: "Geoffrey - AI-powered development orchestration platform",
		Long: Banner() + `

Geoffrey is a next-generation AI-powered development orchestration platform
that reimagines human-AI collaboration on software projects.

The system prioritizes deep project understanding through a multi-stage iterative
pipeline: Interview → Architecture Design → DevPlan Generation → Phase Review.`,
		Version: version,
		RunE:    runRootWithResumeCheck,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Don't print banner for help commands
			if !argsContains(args, "--help") && !argsContains(args, "-h") {
				fmt.Print(Banner())
				fmt.Println()
			}
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.geoffrussy/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(interviewCmd)
	rootCmd.AddCommand(designCmd)
	rootCmd.AddCommand(planCmd)
	rootCmd.AddCommand(reviewCmd)
	rootCmd.AddCommand(developCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(quotaCmd)
	rootCmd.AddCommand(checkpointCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(resumeCmd)
	rootCmd.AddCommand(navigateCmd)
}

func argsContains(args []string, s string) bool {
	for _, arg := range args {
		if arg == s {
			return true
		}
	}
	return false
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Geoffrey version %s\n", version)
	},
}

// runRootWithResumeCheck runs when geoffrussy is invoked without a subcommand
// It checks for incomplete work and offers to resume
func runRootWithResumeCheck(cmd *cobra.Command, args []string) error {
	// Try to load configuration
	cfgMgr := config.NewManager()
	if err := cfgMgr.Load(nil); err != nil {
		// If config doesn't exist, show help instead
		return cmd.Help()
	}
	cfg := cfgMgr.GetConfig()

	// Determine project ID from current directory
	cwd, err := os.Getwd()
	if err != nil {
		return cmd.Help()
	}
	projectID := filepath.Base(cwd)

	// Initialize state store (use config directory)
	configDir := filepath.Dir(cfg.ConfigPath)
	dbPath := filepath.Join(configDir, "geoffrussy.db")
	store, err := state.NewStore(dbPath)
	if err != nil {
		// If state store doesn't exist, show help
		return cmd.Help()
	}
	defer store.Close()

	// Try to get project - if it doesn't exist, just show help
	_, err = store.GetProject(projectID)
	if err != nil {
		return cmd.Help()
	}

	// Initialize managers
	gitMgr := git.NewManager(".")
	checkpointMgr := checkpoint.NewManager(store, gitMgr, configDir)
	resumeMgr := resume.NewManager(store, checkpointMgr)

	// Check for incomplete work
	info, err := resumeMgr.DetectIncompleteWork(projectID)
	if err != nil {
		return cmd.Help()
	}

	// If no incomplete work, just show help
	if !info.HasIncompleteWork {
		fmt.Println("✅ Project is complete!")
		fmt.Println()
		return cmd.Help()
	}

	// Show incomplete work detected
	fmt.Println("🔔 Incomplete Work Detected")
	fmt.Println("═══════════════════════════")
	fmt.Println()
	fmt.Println(info.Summary)
	fmt.Println()
	fmt.Println("💡 Tip: Run 'geoffrussy resume' to continue where you left off")
	fmt.Println("     Or run 'geoffrussy status' to see detailed progress")
	fmt.Println()

	return nil
}
