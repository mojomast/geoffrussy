package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display project status",
	Long: `Display current project status including stage, phase progress,
blockers, and token usage statistics.`,
	RunE: runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	fmt.Println("📊 Project Status")
	fmt.Println("============================================================")
	fmt.Println("⚠️  This command requires full project tracking integration")
	fmt.Println("   Implementation in progress...")
	
	// TODO: Full implementation requires:
	// - Project state tracking
	// - Phase progress calculation
	// - Blocker management
	// - Token usage display
	
	return fmt.Errorf("status command not yet fully implemented")
}

func displayProgressBar(percent int) {
	barLength := 40
	filled := percent * barLength / 100
	bar := "["
	for i := 0; i < barLength; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += "]"
	fmt.Printf("  %s\n", bar)
}
