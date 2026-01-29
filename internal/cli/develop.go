package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mojomast/geoffrussy/internal/blocker"
	"github.com/mojomast/geoffrussy/internal/config"
	"github.com/mojomast/geoffrussy/internal/detour"
	"github.com/mojomast/geoffrussy/internal/devplan"
	"github.com/mojomast/geoffrussy/internal/executor"
	"github.com/mojomast/geoffrussy/internal/interview"
	"github.com/mojomast/geoffrussy/internal/provider"
	"github.com/mojomast/geoffrussy/internal/state"
	"github.com/spf13/cobra"
)

var (
	developModel string
	developPhase string
)

var developCmd = &cobra.Command{
	Use:   "develop",
	Short: "Execute development phases",
	Long: `Execute development phases and tasks with real-time monitoring.
Handles detours and blockers automatically.`,
	RunE: runDevelop,
}

func init() {
	developCmd.Flags().StringVar(&developModel, "model", "", "Model to use for development")
	developCmd.Flags().StringVar(&developPhase, "phase", "", "Specific phase ID to execute")
}

func runDevelop(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Starting Development Execution...")

	// 1. Load Configuration
	cfgMgr := config.NewManager()
	if err := cfgMgr.Load(nil); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	projectID := filepath.Base(cwd)

	// 2. Initialize Store
	dbPath := filepath.Join(cwd, ".geoffrussy", "state.db")
	store, err := state.NewStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open state store: %w", err)
	}
	defer store.Close()

	project, err := store.GetProject(projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w. Please run 'geoffrussy init' first", err)
	}

	// 3. Initialize Provider
	providerName, modelName, err := getProviderAndModel(cfgMgr, "develop", developModel)
	if err != nil {
		return fmt.Errorf("failed to get provider and model: %w", err)
	}

	bridge := provider.NewBridge()
	if err := setupProvider(bridge, cfgMgr, providerName); err != nil {
		return fmt.Errorf("failed to setup provider: %w", err)
	}

	prov, err := bridge.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	fmt.Printf("📦 Using Provider: %s\n", providerName)
	fmt.Printf("🤖 Using Model: %s\n", modelName)

	// 4. Initialize Components
	interviewEngine := interview.NewEngine(store, prov, modelName)
	devplanGenerator := devplan.NewGenerator(prov, modelName)

	// We initialize these even if not strictly used by current Executor implementation
	// to ensure all components are ready for the "full implementation" context.
	_ = detour.NewManager(store, interviewEngine, devplanGenerator)
	_ = blocker.NewDetector(store, interviewEngine)

	// 5. Determine Phase
	phaseID := developPhase
	if phaseID == "" {
		if project.CurrentPhase != "" {
			phaseID = project.CurrentPhase
		} else {
			// Find first non-completed phase
			phases, err := store.ListPhases(projectID)
			if err != nil {
				return fmt.Errorf("failed to list phases: %w", err)
			}
			for _, p := range phases {
				if p.Status != state.PhaseCompleted {
					phaseID = p.ID
					break
				}
			}
		}
	}

	if phaseID == "" {
		return fmt.Errorf("no active phase found to execute")
	}

	phase, err := store.GetPhase(phaseID)
	if err != nil {
		return fmt.Errorf("failed to get phase %s: %w", phaseID, err)
	}
	fmt.Printf("📋 Executing Phase: %s (%s)\n", phase.Title, phase.ID)

	// 6. Initialize Executor and Monitor
	exec := executor.NewExecutor(store, prov)
	mon := executor.NewMonitor(exec)

	// 7. Start Execution
	// Run execution in a separate goroutine so Monitor can run in main thread
	go func() {
		// Give the monitor a moment to start
		time.Sleep(500 * time.Millisecond)

		if err := exec.ExecutePhase(phaseID); err != nil {
			// Errors are reported via the update channel usually,
			// but we can also log here if needed or if ExecutePhase returns early
			// We can't easily log to stdout here because the TUI has taken over
		}
		// We might want to close the executor or signal completion here
		// But Monitor handles Ctrl+C/Quit
	}()

	// 8. Run Monitor (Blocking)
	if err := mon.Run(); err != nil {
		return fmt.Errorf("monitor error: %w", err)
	}

	return nil
}
