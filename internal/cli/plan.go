package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	planModel  string
	planMerge  string
	planSplit  string
	planReorder bool
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Generate or manipulate development plan",
	Long: `Generate development plan from architecture or manipulate
existing plan by merging, splitting, or reordering phases.`,
	RunE: runPlan,
}

func init() {
	planCmd.Flags().StringVar(&planModel, "model", "", "Model to use for plan generation")
	planCmd.Flags().StringVar(&planMerge, "merge", "", "Merge phases (format: 1,2)")
	planCmd.Flags().StringVar(&planSplit, "split", "", "Split phase (format: 1:3 - split phase 1 at task 3)")
	planCmd.Flags().BoolVar(&planReorder, "reorder", false, "Reorder phases interactively")
}

func runPlan(cmd *cobra.Command, args []string) error {
	fmt.Println("📋 Development Plan Management...")
	fmt.Println("⚠️  This command requires full provider integration")
	fmt.Println("   Implementation in progress...")
	
	// TODO: Full implementation requires:
	// - Provider selection based on model
	// - DevPlan generation from architecture
	// - Phase manipulation support
	// - Type conversion between state.Phase and devplan.Phase
	
	return fmt.Errorf("plan command not yet fully implemented")
}

