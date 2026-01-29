package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version   string
	cfgFile   string
	verbose   bool
	rootCmd   *cobra.Command
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
		Long: `Geoffrey is a next-generation AI-powered development orchestration platform
that reimagines human-AI collaboration on software projects.

The system prioritizes deep project understanding through a multi-stage iterative
pipeline: Interview → Architecture Design → DevPlan Generation → Phase Review.`,
		Version: version,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.geoffrussy/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
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
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Geoffrey version %s\n", version)
	},
}
