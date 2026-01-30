package mcp

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/mojomast/geoffrussy/internal/checkpoint"
	"github.com/mojomast/geoffrussy/internal/config"
	"github.com/mojomast/geoffrussy/internal/git"
	"github.com/mojomast/geoffrussy/internal/state"
	"github.com/mojomast/geoffrussy/internal/token"
)

// SimpleToolHandlers contains basic MCP tool handlers for Geoffrey
// This is a simplified version that doesn't require full CLI integration
type SimpleToolHandlers struct {
	configManager *config.Manager
	projectRoot   string
}

// NewSimpleToolHandlers creates a new simple tool handlers instance
func NewSimpleToolHandlers(configManager *config.Manager, projectRoot string) *SimpleToolHandlers {
	return &SimpleToolHandlers{
		configManager: configManager,
		projectRoot:   projectRoot,
	}
}

// RegisterBasicTools registers basic Geoffrey tools with the registry
func (h *SimpleToolHandlers) RegisterBasicTools(registry *ToolRegistry) error {
	tools := []struct {
		tool    Tool
		handler ToolHandler
	}{
		{h.getStatusTool(), h.handleGetStatus},
		{h.getStatsTool(), h.handleGetStats},
		{h.listPhasesTool(), h.handleListPhases},
		{h.listCheckpointsTool(), h.handleListCheckpoints},
		{h.createCheckpointTool(), h.handleCreateCheckpoint},
	}

	for _, t := range tools {
		if err := registry.RegisterTool(t.tool, t.handler); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", t.tool.Name, err)
		}
	}

	return nil
}

// Tool definitions

func (h *SimpleToolHandlers) getStatusTool() Tool {
	return Tool{
		Name:        "get_status",
		Description: "Get current project status including stage, progress, and active tasks",
		InputSchema: CreateInputSchema(
			map[string]interface{}{
				"projectPath": StringParam("Absolute path to the project directory"),
			},
			[]string{"projectPath"},
		),
	}
}

func (h *SimpleToolHandlers) getStatsTool() Tool {
	return Tool{
		Name:        "get_stats",
		Description: "Get token usage and cost statistics for the project",
		InputSchema: CreateInputSchema(
			map[string]interface{}{
				"projectPath": StringParam("Absolute path to the project directory"),
			},
			[]string{"projectPath"},
		),
	}
}

func (h *SimpleToolHandlers) listPhasesTool() Tool {
	return Tool{
		Name:        "list_phases",
		Description: "List all development phases with their status and tasks",
		InputSchema: CreateInputSchema(
			map[string]interface{}{
				"projectPath": StringParam("Absolute path to the project directory"),
			},
			[]string{"projectPath"},
		),
	}
}

func (h *SimpleToolHandlers) createCheckpointTool() Tool {
	return Tool{
		Name:        "create_checkpoint",
		Description: "Create a checkpoint to save current project state",
		InputSchema: CreateInputSchema(
			map[string]interface{}{
				"projectPath": StringParam("Absolute path to the project directory"),
				"name":        StringParam("Name for the checkpoint"),
			},
			[]string{"projectPath", "name"},
		),
	}
}

func (h *SimpleToolHandlers) listCheckpointsTool() Tool {
	return Tool{
		Name:        "list_checkpoints",
		Description: "List all checkpoints for the project",
		InputSchema: CreateInputSchema(
			map[string]interface{}{
				"projectPath": StringParam("Absolute path to the project directory"),
			},
			[]string{"projectPath"},
		),
	}
}

// Tool handlers

func (h *SimpleToolHandlers) handleGetStatus(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	projectPath, ok := args["projectPath"].(string)
	if !ok {
		return ErrorResult("projectPath must be a string"), nil
	}

	// Open state store
	dbPath := filepath.Join(projectPath, ".geoffrussy", "state.db")
	store, err := state.NewStore(dbPath)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open state store: %v", err)), nil
	}
	defer store.Close()

	projectID := filepath.Base(projectPath)
	project, err := store.GetProject(projectID)
	if err != nil {
		return ErrorResult("Project not found"), nil
	}

	// Calculate progress
	progress, err := store.CalculateProgress(projectID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to calculate progress: %v", err)), nil
	}

	// Build status report
	status := fmt.Sprintf("📊 Project Status: %s\n", project.Name)
	status += fmt.Sprintf("═══════════════════════════════\n")
	status += fmt.Sprintf("Stage: %s\n", project.CurrentStage)
	status += fmt.Sprintf("Progress: %.1f%%\n", progress.CompletionPercentage)
	status += fmt.Sprintf("Tasks: %d/%d completed\n", progress.CompletedTasks, progress.TotalTasks)
	status += fmt.Sprintf("Phases: %d/%d completed\n", progress.CompletedPhases, progress.TotalPhases)

	if progress.InProgressTasks > 0 {
		status += fmt.Sprintf("In Progress: %d tasks\n", progress.InProgressTasks)
	}
	if progress.BlockedTasks > 0 {
		status += fmt.Sprintf("Blocked: %d tasks\n", progress.BlockedTasks)
	}

	return SuccessResult(status), nil
}

func (h *SimpleToolHandlers) handleGetStats(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	projectPath, ok := args["projectPath"].(string)
	if !ok {
		return ErrorResult("projectPath must be a string"), nil
	}

	// Open state store
	dbPath := filepath.Join(projectPath, ".geoffrussy", "state.db")
	store, err := state.NewStore(dbPath)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open state store: %v", err)), nil
	}
	defer store.Close()

	projectID := filepath.Base(projectPath)
	_, err = store.GetProject(projectID)
	if err != nil {
		return ErrorResult("Project not found"), nil
	}

	// Get token stats
	counter := token.NewCounter(store)
	tokenStats, err := counter.GetTotalTokens(projectID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to get token stats: %v", err)), nil
	}

	// Get cost stats
	costEstimator := token.NewCostEstimator(store)
	costStats, err := costEstimator.GetCostStats(projectID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to get cost stats: %v", err)), nil
	}

	// Build stats report
	stats := fmt.Sprintf("📊 Token Usage & Cost Statistics\n")
	stats += fmt.Sprintf("═══════════════════════════════\n")
	stats += fmt.Sprintf("Total Cost: $%.4f\n", costStats.TotalCost)
	stats += fmt.Sprintf("Total Input: %d tokens\n", tokenStats.TotalInput)
	stats += fmt.Sprintf("Total Output: %d tokens\n", tokenStats.TotalOutput)
	stats += fmt.Sprintf("Grand Total: %d tokens\n", tokenStats.TotalInput+tokenStats.TotalOutput)

	if len(tokenStats.ByProvider) > 0 {
		stats += fmt.Sprintf("\nBy Provider:\n")
		for provider, tokens := range tokenStats.ByProvider {
			cost := costStats.ByProvider[provider]
			stats += fmt.Sprintf("  %s: %d tokens ($%.4f)\n", provider, tokens, cost)
		}
	}

	return SuccessResult(stats), nil
}

func (h *SimpleToolHandlers) handleListPhases(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	projectPath, ok := args["projectPath"].(string)
	if !ok {
		return ErrorResult("projectPath must be a string"), nil
	}

	// Open state store
	dbPath := filepath.Join(projectPath, ".geoffrussy", "state.db")
	store, err := state.NewStore(dbPath)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open state store: %v", err)), nil
	}
	defer store.Close()

	projectID := filepath.Base(projectPath)
	phases, err := store.ListPhases(projectID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to list phases: %v", err)), nil
	}

	if len(phases) == 0 {
		return SuccessResult("No phases found. Run 'geoffrussy plan' to generate development plan."), nil
	}

	result := "📋 Development Phases:\n"
	result += "═════════════════════\n\n"
	for _, phase := range phases {
		statusIcon := getStatusIcon(phase.Status)
		result += fmt.Sprintf("%s Phase %d: %s\n", statusIcon, phase.Number, phase.Title)
		result += fmt.Sprintf("   Status: %s\n", phase.Status)

		// Get tasks for this phase
		tasks, err := store.ListTasks(phase.ID)
		if err == nil && len(tasks) > 0 {
			completed := 0
			for _, task := range tasks {
				if task.Status == state.TaskCompleted {
					completed++
				}
			}
			result += fmt.Sprintf("   Tasks: %d/%d completed\n", completed, len(tasks))
		}
		result += "\n"
	}

	return SuccessResult(result), nil
}

func (h *SimpleToolHandlers) handleCreateCheckpoint(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	projectPath, ok := args["projectPath"].(string)
	if !ok {
		return ErrorResult("projectPath must be a string"), nil
	}

	name, ok := args["name"].(string)
	if !ok {
		return ErrorResult("name must be a string"), nil
	}

	// Open state store
	dbPath := filepath.Join(projectPath, ".geoffrussy", "state.db")
	store, err := state.NewStore(dbPath)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open state store: %v", err)), nil
	}
	defer store.Close()

	projectID := filepath.Base(projectPath)
	_, err = store.GetProject(projectID)
	if err != nil {
		return ErrorResult("Project not found"), nil
	}

	// Create checkpoint
	gitManager := git.NewManager(projectPath)
	dataDir := filepath.Join(projectPath, ".geoffrussy")
	checkpointMgr := checkpoint.NewManager(store, gitManager, dataDir)

	cp, err := checkpointMgr.CreateCheckpoint(projectID, name, nil)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to create checkpoint: %v", err)), nil
	}

	return SuccessResult(fmt.Sprintf("✅ Checkpoint created: %s (ID: %s)", name, cp.ID)), nil
}

func (h *SimpleToolHandlers) handleListCheckpoints(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	projectPath, ok := args["projectPath"].(string)
	if !ok {
		return ErrorResult("projectPath must be a string"), nil
	}

	// Open state store
	dbPath := filepath.Join(projectPath, ".geoffrussy", "state.db")
	store, err := state.NewStore(dbPath)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to open state store: %v", err)), nil
	}
	defer store.Close()

	projectID := filepath.Base(projectPath)
	checkpoints, err := store.ListCheckpoints(projectID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to list checkpoints: %v", err)), nil
	}

	if len(checkpoints) == 0 {
		return SuccessResult("No checkpoints found"), nil
	}

	result := "📋 Checkpoints:\n"
	for _, cp := range checkpoints {
		result += fmt.Sprintf("  • %s - %s (ID: %s)\n", cp.CreatedAt.Format("2006-01-02 15:04"), cp.Name, cp.ID)
	}

	return SuccessResult(result), nil
}

// Helper functions

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
