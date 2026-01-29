package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mojomast/geoffrussy/internal/blocker"
	"github.com/mojomast/geoffrussy/internal/state"
	"github.com/mojomast/geoffrussy/internal/token"
	"github.com/spf13/cobra"
)

var (
	statusPhaseFilter   []int
	statusStatusFilter  []string
	statusVerbose       bool
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display project status",
	Long: `Display current project status including stage, phase progress,
blockers, and token usage statistics.`,
	RunE: runStatus,
}

func init() {
	statusCmd.Flags().IntSliceVar(&statusPhaseFilter, "phase", []int{}, "Filter by phase numbers (comma-separated)")
	statusCmd.Flags().StringSliceVar(&statusStatusFilter, "status", []string{}, "Filter by status (not_started, in_progress, completed, blocked)")
	statusCmd.Flags().BoolVarP(&statusVerbose, "verbose", "v", false, "Show detailed information")
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Get current directory as project root
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Initialize state store
	dbPath := filepath.Join(projectRoot, ".geoffrussy", "state.db")
	store, err := state.NewStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize state store: %w", err)
	}
	defer store.Close()

	// Get project ID (use directory name for now)
	projectID := filepath.Base(projectRoot)

	// Check if project exists
	project, err := store.GetProject(projectID)
	if err != nil {
		fmt.Println("⚠️  No active project found in this directory")
		fmt.Println("   Run 'geoffrussy init' to initialize a new project")
		return nil
	}

	// Display header
	fmt.Println("📊 Project Status")
	fmt.Println("============================================================")
	fmt.Println()

	// Display project info
	fmt.Printf("📁 Project: %s\n", project.Name)
	fmt.Printf("🆔 ID: %s\n", projectID)
	fmt.Printf("📅 Started: %s\n", project.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("🏗️  Current Stage: %s\n", formatStage(project.CurrentStage))
	fmt.Println()

	// Calculate and display progress
	progress, err := store.CalculateProgress(projectID)
	if err != nil {
		return fmt.Errorf("failed to calculate progress: %w", err)
	}

	displayProgressSummary(progress)

	// Display phase-by-phase progress
	fmt.Println("\n📋 Phase Progress")
	fmt.Println("============================================================")

	filter := &state.ProgressFilter{
		PhaseNumbers: statusPhaseFilter,
	}

	// Convert status filter strings to PhaseStatus
	if len(statusStatusFilter) > 0 {
		for _, statusStr := range statusStatusFilter {
			filter.StatusFilter = append(filter.StatusFilter, state.PhaseStatus(statusStr))
		}
	}

	phaseProgress, err := store.GetFilteredProgress(projectID, filter)
	if err != nil {
		return fmt.Errorf("failed to get phase progress: %w", err)
	}

	for _, pp := range phaseProgress {
		displayPhaseProgress(pp, statusVerbose)
	}

	// Display active blockers
	blockerDetector := blocker.NewDetector(store, nil)
	blockers, err := blockerDetector.ListActiveBlockers(projectID)
	if err == nil && len(blockers) > 0 {
		fmt.Println("\n🚫 Active Blockers")
		fmt.Println("============================================================")
		for _, b := range blockers {
			fmt.Printf("  ⚠️  Task %s: %s\n", b.TaskID, b.Description)
		}
	}

	// Display token usage and costs
	if statusVerbose {
		fmt.Println("\n💰 Token Usage & Costs")
		fmt.Println("============================================================")

		tokenCounter := token.NewCounter(store)
		stats, err := tokenCounter.GetTotalTokens(projectID)
		if err == nil {
			fmt.Printf("  Total Input Tokens:  %d\n", stats.TotalInput)
			fmt.Printf("  Total Output Tokens: %d\n", stats.TotalOutput)
			totalTokens := stats.TotalInput + stats.TotalOutput
			fmt.Printf("  Total Tokens:        %d\n", totalTokens)

			if len(stats.ByPhase) > 0 {
				fmt.Println("\n  By Phase:")
				for phase, count := range stats.ByPhase {
					fmt.Printf("    Phase %s: %d tokens\n", phase, count)
				}
			}
		}

		costEstimator := token.NewCostEstimator(store)
		totalCost, err := costEstimator.GetTotalCost(projectID)
		if err == nil {
			fmt.Printf("\n  Total Cost: $%.2f\n", totalCost)
		}
	}

	fmt.Println()
	return nil
}

func displayProgressSummary(progress *state.ProgressStats) {
	fmt.Println("📈 Overall Progress")
	fmt.Println("------------------------------------------------------------")
	fmt.Printf("  Completion: %.1f%%\n", progress.CompletionPercentage)
	displayProgressBar(int(progress.CompletionPercentage))
	fmt.Printf("  Tasks: %d/%d completed (%d in progress, %d blocked)\n",
		progress.CompletedTasks,
		progress.TotalTasks,
		progress.InProgressTasks,
		progress.BlockedTasks,
	)
	fmt.Printf("  Phases: %d/%d completed (%d in progress, %d blocked)\n",
		progress.CompletedPhases,
		progress.TotalPhases,
		progress.InProgressPhases,
		progress.BlockedPhases,
	)

	// Display time tracking
	fmt.Printf("\n  ⏱️  Elapsed Time: %s\n", formatDuration(progress.ElapsedTime))
	if progress.EstimatedRemaining > 0 {
		fmt.Printf("  ⏳ Estimated Remaining: %s\n", formatDuration(progress.EstimatedRemaining))
	}
}

func displayPhaseProgress(progress *state.PhaseProgress, verbose bool) {
	statusIcon := getStatusIcon(progress.Status)
	fmt.Printf("\n%s Phase %d: %s\n", statusIcon, progress.PhaseNumber, progress.PhaseTitle)

	if progress.TotalTasks > 0 {
		fmt.Printf("  Progress: %.0f%% (%d/%d tasks completed)\n",
			progress.Percentage,
			progress.CompletedTasks,
			progress.TotalTasks,
		)

		if verbose {
			if progress.InProgressTasks > 0 {
				fmt.Printf("  🔄 In Progress: %d tasks\n", progress.InProgressTasks)
			}
			if progress.BlockedTasks > 0 {
				fmt.Printf("  🚫 Blocked: %d tasks\n", progress.BlockedTasks)
			}
		}
	}
}

func displayProgressBar(percent int) {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	barLength := 40
	filled := percent * barLength / 100
	bar := "  ["
	for i := 0; i < barLength; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += fmt.Sprintf("] %d%%", percent)
	fmt.Println(bar)
}

func getStatusIcon(status state.PhaseStatus) string {
	switch status {
	case state.PhaseCompleted:
		return "✅"
	case state.PhaseInProgress:
		return "🔄"
	case state.PhaseBlocked:
		return "🚫"
	default:
		return "⬜"
	}
}

func formatStage(stage state.Stage) string {
	switch stage {
	case state.StageInit:
		return "🔧 Initialization"
	case state.StageInterview:
		return "💬 Interview"
	case state.StageDesign:
		return "🎨 Design"
	case state.StagePlan:
		return "📋 Planning"
	case state.StageReview:
		return "🔍 Review"
	case state.StageDevelop:
		return "⚡ Development"
	case state.StageComplete:
		return "🎉 Complete"
	default:
		return string(stage)
	}
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 || days > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}

	return strings.Join(parts, " ")
}
