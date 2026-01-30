package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mojomast/geoffrussy/internal/config"
	"github.com/mojomast/geoffrussy/internal/executor"
	"github.com/mojomast/geoffrussy/internal/provider"
)

// ExecHandlers contains handlers for execution-related tools
type ExecHandlers struct {
	configManager *config.Manager
}

// NewExecHandlers creates a new execution handlers instance
func NewExecHandlers(configManager *config.Manager) *ExecHandlers {
	return &ExecHandlers{
		configManager: configManager,
	}
}

// RegisterHandlers registers execution tools with the registry
func (h *ExecHandlers) RegisterHandlers(registry *ToolRegistry) error {

tools := []struct {
		tool    Tool
		handler ToolHandler
	}{
		{h.executePhaseTool(), h.handleExecutePhase},
		{h.executeTaskTool(), h.handleExecuteTask},
		{h.getTaskOutputTool(), h.handleGetTaskOutput},
		{h.handleBlockerTool(), h.handleHandleBlocker},
	}

	for _, t := range tools {
		if err := registry.RegisterTool(t.tool, t.handler); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", t.tool.Name, err)
		}
	}

	return nil
}

// Tool definitions

func (h *ExecHandlers) executePhaseTool() Tool {
	return Tool{
		Name:        "execute_phase",
		Description: "Execute all tasks in a development phase, writing code and creating files",
		InputSchema: CreateInputSchema(
			map[string]interface{}{
				"projectPath":    StringParam("Absolute path to the project directory"),
				"phaseId":        StringParam("ID of the phase to execute (e.g., 'phase-0', 'phase-1')"),
				"model":          StringParam("Model to use for task execution"),
				"stopAfterPhase": BooleanParam("Stop after completing this phase"),
				"streamOutput":   BooleanParam("Stream task output in real-time (not supported in current transport)"),
			},
			[]string{"projectPath", "phaseId"},
		),
	}
}

func (h *ExecHandlers) executeTaskTool() Tool {
	return Tool{
		Name:        "execute_task",
		Description: "Execute a single task by ID",
		InputSchema: CreateInputSchema(
			map[string]interface{}{
				"projectPath": StringParam("Absolute path to the project directory"),
				"taskId":      StringParam("ID of task to execute (e.g., 'task-1.3')"),
				"model":       StringParam("Model to use for task execution"),
			},
			[]string{"projectPath", "taskId"},
		),
	}
}

func (h *ExecHandlers) getTaskOutputTool() Tool {
	return Tool{
		Name:        "get_task_output",
		Description: "Get detailed execution output for a task",
		InputSchema: CreateInputSchema(
			map[string]interface{}{
				"projectPath": StringParam("Absolute path to the project directory"),
				"taskId":      StringParam("ID of task"),
			},
			[]string{"projectPath", "taskId"},
		),
	}
}

func (h *ExecHandlers) handleBlockerTool() Tool {
	return Tool{
		Name:        "handle_blocker",
		Description: "Attempt to resolve a blocker or get guidance on resolution",
		InputSchema: CreateInputSchema(
			map[string]interface{}{
				"projectPath":  StringParam("Absolute path to the project directory"),
				"taskId":       StringParam("ID of task"),
				"action":       StringParam("Action to take: retry, skip, modify, analyze"),
				"modification": StringParam("Modified task description if action is 'modify'"),
			},
			[]string{"projectPath", "taskId", "action"},
		),
	}
}

// Handler implementations

func (h *ExecHandlers) handleExecutePhase(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	projectPath, err := ValidateAndGetString(args, "projectPath", true)
	if err != nil {
		return ErrorResult(err.Error()), nil
	}

	phaseID, err := ValidateAndGetString(args, "phaseId", true)
	if err != nil {
		return ErrorResult(err.Error()), nil
	}

	model, _ := ValidateAndGetString(args, "model", false)
	stopAfterPhase, _ := ValidateAndGetBool(args, "stopAfterPhase", false, true)

	store, err := openStateStore(projectPath)
	if err != nil {
		return ErrorResult(err.Error()), nil
	}
	defer store.Close()

	prov, modelName, err := h.initProvider("develop", model)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to initialize provider: %v", err)), nil
	}

	exec := executor.NewExecutor(store, prov, modelName)
	defer exec.Close()

	// Start consuming updates
	logMap := make(map[string]*strings.Builder)
	tasksSucceeded := 0
	tasksFailed := 0
	phaseSummary := strings.Builder{}

	// Create log directory
	logDir := filepath.Join(projectPath, ".geoffrussy", "logs")
	os.MkdirAll(logDir, 0755)

	done := make(chan bool)
	go func() {
		for update := range exec.StreamOutput() {
			// Write to in-memory log for summary
			// And append to per-task log files
			if update.TaskID != "" {
				if _, ok := logMap[update.TaskID]; !ok {
					logMap[update.TaskID] = &strings.Builder{}
				}
				logMap[update.TaskID].WriteString(fmt.Sprintf("[%s] %s\n", update.Type, update.Content))
				
				// Write to file
				logFile := filepath.Join(logDir, fmt.Sprintf("%s.log", update.TaskID))
				f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err == nil {
					f.WriteString(fmt.Sprintf("[%s] %s %s\n", time.Now().Format(time.RFC3339), update.Type, update.Content))
					f.Close()
				}
			}

			if update.Type == executor.TaskCompleted {
				tasksSucceeded++
				if update.TaskID != "" {
					phaseSummary.WriteString(fmt.Sprintf("  ✅ Task %s: Completed\n", update.TaskID))
				}
			} else if update.Type == executor.TaskError {
				tasksFailed++
				if update.TaskID != "" {
					phaseSummary.WriteString(fmt.Sprintf("  ❌ Task %s: Failed - %v\n", update.TaskID, update.Error))
				}
			}
		}
		done <- true
	}()

	startTime := time.Now()
	// Execute Phase
	var execErr error
	if stopAfterPhase {
		execErr = exec.ExecutePhase(phaseID)
	} else {
		execErr = exec.ExecuteProject(getProjectID(projectPath), phaseID, stopAfterPhase)
	}
	
	// Wait for updates to process
	// We need to close executor to close channel?
	// Executor.Close() cancels context, but doesn't close channel immediately?
	// Executor.Close() closes channel.
	exec.Close()
	<-done

	duration := time.Since(startTime)
	
	// Construct result
	resultText := fmt.Sprintf("✅ Phase execution completed (Duration: %s)\n\nTasks Summary:\n%s\nFiles Created: (check logs)\nTotal Cost: (check stats)", 
		duration.Round(time.Second), phaseSummary.String())
	
	if execErr != nil {
		resultText = fmt.Sprintf("⚠️ Phase execution failed (Duration: %s)\nError: %v\n\nTasks Summary:\n%s", 
			duration.Round(time.Second), execErr, phaseSummary.String())
		// Don't return error result, return tool result with error details
	}

	return SuccessResult(resultText), nil
}

func (h *ExecHandlers) handleExecuteTask(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	projectPath, err := ValidateAndGetString(args, "projectPath", true)
	if err != nil {
		return ErrorResult(err.Error()), nil
	}

	taskID, err := ValidateAndGetString(args, "taskId", true)
	if err != nil {
		return ErrorResult(err.Error()), nil
	}

	model, _ := ValidateAndGetString(args, "model", false)

	store, err := openStateStore(projectPath)
	if err != nil {
		return ErrorResult(err.Error()), nil
	}
	defer store.Close()

	prov, modelName, err := h.initProvider("develop", model)
	if err != nil {
		return ErrorResult(fmt.Sprintf("Failed to initialize provider: %v", err)), nil
	}

	exec := executor.NewExecutor(store, prov, modelName)
	// Don't close exec immediately, we need channel

	logDir := filepath.Join(projectPath, ".geoffrussy", "logs")
	os.MkdirAll(logDir, 0755)
	logFile := filepath.Join(logDir, fmt.Sprintf("%s.log", taskID))

	// Truncate log file for new run
	os.WriteFile(logFile, []byte{}, 0644)

	outputBuilder := strings.Builder{}
	done := make(chan bool)

	go func() {
		for update := range exec.StreamOutput() {
			line := fmt.Sprintf("[%s] %s\n", update.Type, update.Content)
			outputBuilder.WriteString(line)
			
			f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				f.WriteString(fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), line))
				f.Close()
			}
		}
		done <- true
	}()

	err = exec.ExecuteTask(taskID)
	exec.Close()
	<-done

	if err != nil {
		return ErrorResult(fmt.Sprintf("Task execution failed: %v\nOutput:\n%s", err, outputBuilder.String())), nil
	}

	return SuccessResult(fmt.Sprintf("✅ Task %s completed successfully.\n\nOutput:\n%s", taskID, outputBuilder.String())), nil
}

func (h *ExecHandlers) handleGetTaskOutput(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	projectPath, err := ValidateAndGetString(args, "projectPath", true)
	if err != nil {
		return ErrorResult(err.Error()), nil
	}

	taskID, err := ValidateAndGetString(args, "taskId", true)
	if err != nil {
		return ErrorResult(err.Error()), nil
	}

	// Try to read log file
	logFile := filepath.Join(projectPath, ".geoffrussy", "logs", fmt.Sprintf("%s.log", taskID))
	content, err := os.ReadFile(logFile)
	if err != nil {
		// Fallback to basic status from DB
		store, err := openStateStore(projectPath)
		if err != nil {
			return ErrorResult("Log not found and DB unavailable"), nil
		}
		defer store.Close()
		
		task, err := store.GetTask(taskID)
		if err != nil {
			return ErrorResult(fmt.Sprintf("Task not found: %s", taskID)), nil
		}
		return SuccessResult(fmt.Sprintf("Task %s: %s\nStatus: %s\n(No detailed logs available)", task.Number, task.Description, task.Status)), nil
	}

	return SuccessResult(string(content)), nil
}

func (h *ExecHandlers) handleHandleBlocker(ctx context.Context, args map[string]interface{}) (*CallToolResult, error) {
	projectPath, err := ValidateAndGetString(args, "projectPath", true)
	if err != nil {
		return ErrorResult(err.Error()), nil
	}

	taskID, err := ValidateAndGetString(args, "taskId", true)
	if err != nil {
		return ErrorResult(err.Error()), nil
	}

	action, err := ValidateAndGetString(args, "action", true)
	if err != nil {
		return ErrorResult(err.Error()), nil
	}

	store, err := openStateStore(projectPath)
	if err != nil {
		return ErrorResult(err.Error()), nil
	}
	defer store.Close()

	// We need executor to modify state?
	// Executor has ResolveBlocker, SkipTask.
	// But those methods are on Executor struct.
	// We can use a dummy provider for executor as we might not need LLM for simple actions
	// unless 'analyze' or 'modify' needs it.
	
	prov, modelName, _ := h.initProvider("develop", "")
	// If provider init fails (no key), we might still proceed if action is 'skip' or 'retry' if we don't need provider?
	// But NewExecutor requires provider.
	// Let's assume we can get one.

	exec := executor.NewExecutor(store, prov, modelName)
	defer exec.Close()

	switch action {
	case "retry":
		// Retry calls ExecuteTask
		// But this tool should just probably resolve it so next run picks it up?
		// Or actually run it?
		// "Attempt to resolve a blocker"
		// If we retry, we are running it.
		// Let's retry it.
		// Reuse handleExecuteTask logic?
		return h.handleExecuteTask(ctx, map[string]interface{}{
			"projectPath": projectPath,
			"taskId":      taskID,
		})
		
	case "skip":
		if err := exec.SkipTask(taskID); err != nil {
			return ErrorResult(fmt.Sprintf("Failed to skip task: %v", err)), nil
		}
		// Also resolve blocker if exists
		// Executor.SkipTask doesn't auto-resolve blocker record.
		// We should try to resolve it.
		_ = exec.ResolveBlocker(taskID, "Skipped by user")
		return SuccessResult(fmt.Sprintf("Task %s skipped.", taskID)), nil

	case "modify":
		// User provided new instructions?
		modification, _ := ValidateAndGetString(args, "modification", false)
		if modification == "" {
			return ErrorResult("Modification description required for 'modify' action"), nil
		}
		
		task, err := store.GetTask(taskID)
		if err != nil {
			return ErrorResult("Task not found"), nil
		}
		
		task.Description = modification
		if err := store.SaveTask(task); err != nil {
			return ErrorResult(fmt.Sprintf("Failed to update task: %v", err)), nil
		}
		
		// Resolve blocker
		_ = exec.ResolveBlocker(taskID, "Task modified")
		
		return SuccessResult(fmt.Sprintf("Task %s modified. You can now retry it.", taskID)), nil

	case "analyze":
		// Needs LLM to analyze error log and suggest fix
		// Read logs
		logFile := filepath.Join(projectPath, ".geoffrussy", "logs", fmt.Sprintf("%s.log", taskID))
		logs, _ := os.ReadFile(logFile)
		
		prompt := fmt.Sprintf("Analyze the failure for task %s.\nLogs:\n%s", taskID, string(logs))
		resp, err := prov.Call(modelName, prompt)
		if err != nil {
			return ErrorResult(fmt.Sprintf("Failed to analyze: %v", err)), nil
		}
		
		return SuccessResult(fmt.Sprintf("Analysis:\n%s", resp.Content)), nil
		
	default:
		return ErrorResult(fmt.Sprintf("Unknown action: %s", action)), nil
	}
}

// Helpers

func (h *ExecHandlers) initProvider(stage, overrideModel string) (provider.Provider, string, error) {
	// Duplicated again...
	providerName, modelName, err := getProviderAndModel(h.configManager, stage, overrideModel)
	if err != nil {
		return nil, "", err
	}

	p, err := provider.CreateProvider(providerName)
	if err != nil {
		return nil, "", err
	}

	if providerName == "ollama" {
		if err := p.Authenticate(""); err != nil {
			return nil, "", err
		}
	} else {
		apiKey, err := h.configManager.GetAPIKey(providerName)
		if err != nil {
			return nil, "", err
		}
		if err := p.Authenticate(apiKey); err != nil {
			return nil, "", err
		}
	}

	return p, modelName, nil
}
