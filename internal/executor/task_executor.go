package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mojomast/geoffrussy/internal/provider"
	"github.com/mojomast/geoffrussy/internal/state"
)

// SendUpdateFunc is the type of function used to send updates
type SendUpdateFunc func(update TaskUpdate)

// TaskExecutor implements actual task execution using LLM
type TaskExecutor struct {
	store      *state.Store
	provider   provider.Provider
	modelName  string
	ctx        context.Context
	sendUpdate SendUpdateFunc // Function to send updates through TUI
	phaseID    string         // For update messages
	taskID     string         // For update messages
}

// NewTaskExecutor creates a new task executor that actually implements tasks
func NewTaskExecutor(store *state.Store, prov provider.Provider, sendUpdateFn SendUpdateFunc, modelName string) *TaskExecutor {
	return &TaskExecutor{
		store:      store,
		provider:   prov,
		modelName:  modelName,
		ctx:        context.Background(),
		sendUpdate: sendUpdateFn,
	}
}

// CodeGenerationResponse represents a LLM response for code generation
type CodeGenerationResponse struct {
	Explanation string    `json:"explanation"`
	Files       []File    `json:"files"`
	Commands    []Command `json:"commands,omitempty"`
	Tests       []Test    `json:"tests,omitempty"`
}

type File struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Language string `json:"language,omitempty"`
}

type Command struct {
	Command   string `json:"command"`
	Directory string `json:"directory,omitempty"`
}

type Test struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

// ExecuteTask executes a single task using LLM to generate code
func (te *TaskExecutor) ExecuteTask(taskID string) error {
	// Store IDs for update messages
	te.taskID = taskID

	// Get task from store
	task, err := te.store.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Get phase to understand context
	phase, err := te.store.GetPhase(task.PhaseID)
	if err != nil {
		return fmt.Errorf("failed to get phase: %w", err)
	}

	// Store phase ID for updates
	te.phaseID = phase.ID

	// Get project
	project, err := te.store.GetProject(phase.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Get interview data for context
	interviewData, err := te.store.GetInterviewData(project.ID)
	if err != nil {
		return fmt.Errorf("failed to get interview data: %w", err)
	}

	// Get architecture for context
	architecture, err := te.store.GetArchitecture(project.ID)
	if err != nil {
		return fmt.Errorf("failed to get architecture: %w", err)
	}

	// Build prompt for LLM
	prompt := te.buildExecutionPrompt(task, phase, interviewData, architecture)

	// Determine model to use
	modelName := te.getModelForTask(task)

	// Show task being worked on (through TUI)
	te.sendUpdate(TaskUpdate{
		TaskID:    taskID,
		PhaseID:   phase.ID,
		Type:      TaskProgress,
		Content:   fmt.Sprintf("Starting task: %s\nUsing model: %s", task.Description, modelName),
		Timestamp: time.Now(),
	})

	// Call LLM to generate code
	response, err := te.provider.Call(modelName, prompt)
	if err != nil {
		te.sendUpdate(TaskUpdate{
			TaskID:    taskID,
			PhaseID:   phase.ID,
			Type:      TaskError,
			Content:   fmt.Sprintf("LLM call failed: %v", err),
			Error:     err,
			Timestamp: time.Now(),
		})
		return fmt.Errorf("failed to call LLM: %w", err)
	}

	te.sendUpdate(TaskUpdate{
		TaskID:    taskID,
		PhaseID:   phase.ID,
		Type:      TaskProgress,
		Content:   fmt.Sprintf("LLM responded with %d tokens", response.TokensInput+response.TokensOutput),
		Timestamp: time.Now(),
	})

	// Parse response
	var codeResp CodeGenerationResponse
	if err := json.Unmarshal([]byte(response.Content), &codeResp); err != nil {
		// If JSON parsing fails, treat as entire response as code
		te.sendUpdate(TaskUpdate{
			TaskID:    taskID,
			PhaseID:   phase.ID,
			Type:      TaskProgress,
			Content:   fmt.Sprintf("JSON parsing failed, treating as markdown"),
			Timestamp: time.Now(),
		})
		codeResp = CodeGenerationResponse{
			Explanation: response.Content,
			Files: []File{
				{
					Path:    "output.md",
					Content: response.Content,
				},
			},
		}
	}

	// Show LLM's explanation
	if codeResp.Explanation != "" {
		te.sendUpdate(TaskUpdate{
			TaskID:    taskID,
			PhaseID:   phase.ID,
			Type:      TaskProgress,
			Content:   fmt.Sprintf("Explanation: %s", truncateString(codeResp.Explanation, 200)),
			Timestamp: time.Now(),
		})
	}

	te.sendUpdate(TaskUpdate{
		TaskID:    taskID,
		PhaseID:   phase.ID,
		Type:      TaskProgress,
		Content:   fmt.Sprintf("Generated %d file(s)", len(codeResp.Files)),
		Timestamp: time.Now(),
	})

	// Create files
	for i, file := range codeResp.Files {
		preview := truncateString(file.Content, 200)

		te.sendUpdate(TaskUpdate{
			TaskID:    taskID,
			PhaseID:   phase.ID,
			Type:      TaskProgress,
			Content:   fmt.Sprintf("Writing file %d/%d: %s\nPreview: %s", i+1, len(codeResp.Files), file.Path, preview),
			Timestamp: time.Now(),
		})

		if err := te.writeFile(file); err != nil {
			return fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}

		te.sendUpdate(TaskUpdate{
			TaskID:    taskID,
			PhaseID:   phase.ID,
			Type:      TaskProgress,
			Content:   fmt.Sprintf("Created: %s (%d bytes)", file.Path, len(file.Content)),
			Timestamp: time.Now(),
		})
	}

	// Execute commands (optional - might be dangerous in auto-execution)
	if len(codeResp.Commands) > 0 {
		cmdList := fmt.Sprintf("%d commands", len(codeResp.Commands))
		te.sendUpdate(TaskUpdate{
			TaskID:    taskID,
			PhaseID:   phase.ID,
			Type:      TaskProgress,
			Content:   cmdList,
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (te *TaskExecutor) getModelForTask(task *state.Task) string {
	return te.modelName
}

func (te *TaskExecutor) buildExecutionPrompt(
	task *state.Task,
	phase *state.Phase,
	interviewData *state.InterviewData,
	architecture *state.Architecture,
) string {
	promptBuilder := strings.Builder{}

	promptBuilder.WriteString("You are an expert software developer tasked with implementing a specific task.\n\n")

	promptBuilder.WriteString("PROJECT CONTEXT:\n")
	promptBuilder.WriteString(fmt.Sprintf("Project: %s\n", interviewData.ProjectName))
	promptBuilder.WriteString(fmt.Sprintf("Problem: %s\n\n", interviewData.ProblemStatement))

	promptBuilder.WriteString("PHASE: ")
	promptBuilder.WriteString(phase.Title)
	promptBuilder.WriteString("\n\n")

	promptBuilder.WriteString("TASK: ")
	promptBuilder.WriteString(task.Description)
	promptBuilder.WriteString("\n\n")

	// Add architecture context
	if architecture != nil && len(architecture.Content) > 0 {
		promptBuilder.WriteString("ARCHITECTURE CONTEXT:\n")
		promptBuilder.WriteString(architecture.Content[:min(2000, len(architecture.Content))])
		promptBuilder.WriteString("\n\n")
	}

	promptBuilder.WriteString("INSTRUCTIONS:\n")
	promptBuilder.WriteString("1. Analyze the task and architecture context\n")
	promptBuilder.WriteString("2. Generate working code that implements the task\n")
	promptBuilder.WriteString("3. Ensure code follows best practices for the language/framework\n")
	promptBuilder.WriteString("4. Return your response as JSON with the following structure:\n\n")

	promptBuilder.WriteString(`{
  "explanation": "Brief explanation of your approach",
  "files": [
    {
      "path": "relative/path/to/file.ext",
      "content": "file content here",
      "language": "programming language (optional)"
    }
  ],
  "commands": [
    {
      "command": "shell command to run",
      "directory": "optional directory (default to current)"
    }
  ],
  "tests": [
    {
      "name": "test description",
      "command": "command to run test"
    }
  ]
}`)

	promptBuilder.WriteString("\n\nExecute the task now and return valid JSON.")

	return promptBuilder.String()
}

func (te *TaskExecutor) writeFile(file File) error {
	// Create directory if needed
	dir := filepath.Dir(file.Path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Write file
	if err := os.WriteFile(file.Path, []byte(file.Content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// truncateString truncates a string to max length with "..." suffix
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
