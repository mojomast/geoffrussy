package cli

import (
	"fmt"

	"github.com/mojomast/geoffrussy/internal/reviewer"
	"github.com/mojomast/geoffrussy/internal/state"
	"github.com/spf13/cobra"
)

var (
	reviewModel string
	reviewApply bool
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Review development plan phases",
	Long: `Review development plan phases for clarity, completeness,
dependencies, scope, and other quality metrics. Optionally apply
suggested improvements.`,
	RunE: runReview,
}

func init() {
	reviewCmd.Flags().StringVar(&reviewModel, "model", "", "Model to use for review")
	reviewCmd.Flags().BoolVar(&reviewApply, "apply", false, "Apply improvements automatically")
}

func runReview(cmd *cobra.Command, args []string) error {
	fmt.Println("🔍 Reviewing Development Plan...")
	fmt.Println("⚠️  This command requires full provider integration")
	fmt.Println("   Implementation in progress...")
	
	// TODO: Full implementation requires:
	// - Provider selection based on model
	// - Phase review with improvement suggestions
	// - Interactive improvement application
	
	return fmt.Errorf("review command not yet fully implemented")
}

func applyImprovements(rev *reviewer.Reviewer, store *state.Store, phases []state.Phase, report *reviewer.ReviewReport) error {
	// Stub implementation
	return fmt.Errorf("not yet implemented")
}
