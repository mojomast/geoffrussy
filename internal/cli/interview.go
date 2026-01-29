package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	interviewResume bool
	interviewModel  string
)

var interviewCmd = &cobra.Command{
	Use:   "interview",
	Short: "Start or resume project interview",
	Long: `Start a new project interview or resume an existing one.
The interview gathers essential information about your project through
a structured five-phase process.`,
	RunE: runInterview,
}

func init() {
	interviewCmd.Flags().BoolVar(&interviewResume, "resume", false, "Resume existing interview")
	interviewCmd.Flags().StringVar(&interviewModel, "model", "", "Model to use for interview")
}

func runInterview(cmd *cobra.Command, args []string) error {
	fmt.Println("🎤 Starting Project Interview...")
	fmt.Println("⚠️  This command requires full provider integration")
	fmt.Println("   Implementation in progress...")
	
	// TODO: Full implementation requires:
	// - Provider selection based on model
	// - Proper session management
	// - Interactive TUI with bubbletea
	
	return fmt.Errorf("interview command not yet fully implemented")
}
