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
	resumeFromCheckpoint string
	resumeRestartStage   bool
	resumeStage          string
	resumeModel          string
	resumeProjectID      string
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume work on the current project",
	Long: `Resume work on the current project from where you left off.

This command detects incomplete work and helps you resume from:
  - The current state
  - A specific checkpoint
  - A specific pipeline stage

You can also choose to restart the current stage from the beginning.`,
	RunE: runResume,
}

func init() {
	resumeCmd.Flags().StringVar(&resumeFromCheckpoint, "checkpoint", "", "Resume from a specific checkpoint")
	resumeCmd.Flags().BoolVar(&resumeRestartStage, "restart-stage", false, "Restart the current stage from the beginning")
	resumeCmd.Flags().StringVar(&resumeStage, "stage", "", "Resume from a specific stage (interview, design, plan, review, develop)")
	resumeCmd.Flags().StringVar(&resumeModel, "model", "", "Model to use when resuming")
	resumeCmd.Flags().StringVar(&resumeProjectID, "project", "", "Project ID to resume (defaults to current directory)")
}

func runResume(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfgMgr := config.NewManager()
	if err := cfgMgr.Load(nil); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	cfg := cfgMgr.GetConfig()

	// Determine project ID
	projectID := resumeProjectID
	if projectID == "" {
		// Use current directory name as project ID
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		projectID = filepath.Base(cwd)
	}

	// Initialize state store (use config directory)
	configDir := filepath.Dir(cfg.ConfigPath)
	dbPath := filepath.Join(configDir, "geoffrussy.db")
	store, err := state.NewStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize state store: %w", err)
	}
	defer store.Close()

	// Initialize git manager
	gitMgr := git.NewManager(".")

	// Initialize checkpoint manager
	checkpointMgr := checkpoint.NewManager(store, gitMgr, configDir)

	// Initialize resume manager
	resumeMgr := resume.NewManager(store, checkpointMgr)

	// First, check if there's incomplete work
	fmt.Println("🔍 Checking for incomplete work...")
	info, err := resumeMgr.DetectIncompleteWork(projectID)
	if err != nil {
		return fmt.Errorf("failed to detect incomplete work: %w", err)
	}

	// Display summary
	fmt.Println("\n📊 Project Status")
	fmt.Println("─────────────────")
	fmt.Println(info.Summary)
	fmt.Println()

	// If no incomplete work, exit
	if !info.HasIncompleteWork {
		fmt.Println("✅ No incomplete work detected.")
		return nil
	}

	// Build resume options
	options := &resume.ResumeOptions{
		ProjectID:     projectID,
		RestartStage:  resumeRestartStage,
		SelectedModel: resumeModel,
	}

	// Set checkpoint if specified
	if resumeFromCheckpoint != "" {
		options.FromCheckpoint = resumeFromCheckpoint
	}

	// Set stage if specified
	if resumeStage != "" {
		stage, err := parseStage(resumeStage)
		if err != nil {
			return fmt.Errorf("invalid stage: %w", err)
		}
		options.Stage = &stage
	}

	// If no specific options provided, ask user what to do
	if resumeFromCheckpoint == "" && !resumeRestartStage && resumeStage == "" {
		fmt.Println("📋 Resume Options:")
		fmt.Println("  • Continue from current state (default)")

		// List available checkpoints
		checkpoints, err := resumeMgr.ListAvailableCheckpoints(projectID)
		if err == nil && len(checkpoints) > 0 {
			fmt.Printf("  • Resume from checkpoint (use --checkpoint <id>)\n")
			fmt.Println("\n  Available checkpoints:")
			for i, cp := range checkpoints {
				fmt.Printf("    %d. %s (ID: %s) - %s\n", i+1, cp.Name, cp.ID, cp.CreatedAt.Format("2006-01-02 15:04"))
			}
		}

		fmt.Println("\n  • Restart current stage (use --restart-stage)")
		fmt.Println("  • Jump to specific stage (use --stage <name>)")
		fmt.Println()
	}

	// Perform resume
	fmt.Println("🔄 Resuming work...")
	result, err := resumeMgr.Resume(options)
	if err != nil {
		return fmt.Errorf("failed to resume: %w", err)
	}

	// Display result
	fmt.Println("\n✅ Resume Complete")
	fmt.Println("───────────────────")
	fmt.Printf("Stage: %s\n", result.Stage)
	if result.PhaseID != "" {
		fmt.Printf("Phase: %s\n", result.PhaseID)
	}
	if result.RestoredFrom != "current" {
		fmt.Printf("Restored from: %s\n", result.RestoredFrom)
	}
	if result.ModelSelection != "" {
		fmt.Printf("Model: %s\n", result.ModelSelection)
	}
	fmt.Println()

	// Display next action
	fmt.Println("📌 Next Action:")
	fmt.Printf("   %s\n", result.NextAction)
	fmt.Println()

	return nil
}

// parseStage parses a stage string into a Stage enum
func parseStage(s string) (state.Stage, error) {
	switch s {
	case "init":
		return state.StageInit, nil
	case "interview":
		return state.StageInterview, nil
	case "design":
		return state.StageDesign, nil
	case "plan":
		return state.StagePlan, nil
	case "review":
		return state.StageReview, nil
	case "develop":
		return state.StageDevelop, nil
	case "complete":
		return state.StageComplete, nil
	default:
		return "", fmt.Errorf("unknown stage: %s (must be one of: init, interview, design, plan, review, develop, complete)", s)
	}
}
