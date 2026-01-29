package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mojomast/geoffrussy/internal/provider"
	"github.com/mojomast/geoffrussy/internal/state"
)

// CodeGenerationResponse represents the LLM response for code generation
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

// NewTaskExecutor creates a new task executor that actually implements tasks
func NewTaskExecutor(store *state.Store, prov provider.Provider) *TaskExecutor {
	return &TaskExecutor{
		store:    store,
		provider: prov,
		ctx:      context.Background(),
	}
}

// TaskExecutor implements actual task execution using LLM
type TaskExecutor struct {
	store    *state.Store
	provider provider.Provider
	ctx      context.Context
}

// ExecuteTask executes a single task using LLM to generate code
func (te *TaskExecutor) ExecuteTask(taskID string) error {
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
	fmt.Printf("🤖 Using model: %s\n", modelName)

	// Show task being worked on
	fmt.Printf("\n🎯 Task: %s\n", task.Description)
	fmt.Printf("📋 Phase: %s\n", phase.Title)
	fmt.Println()

	// Call LLM to generate code
	fmt.Printf("📝 Calling LLM to generate code...\n")
	fmt.Printf("   (This may take 30-60 seconds for complex tasks)\n")
	fmt.Println()

	response, err := te.provider.Call(modelName, prompt)
	if err != nil {
		return fmt.Errorf("failed to call LLM: %w", err)
	}

	fmt.Printf("\n✓ LLM responded with %d tokens (input: %d, output: %d)\n",
		response.TokensInput+response.TokensOutput,
		response.TokensInput,
		response.TokensOutput)

	// Parse response
	var codeResp CodeGenerationResponse
	if err := json.Unmarshal([]byte(response.Content), &codeResp); err != nil {
		// If JSON parsing fails, treat the entire response as code
		fmt.Printf("⚠️  JSON parsing failed, treating as markdown\n")
		fmt.Printf("\n💭 LLM Response Preview:\n")
		fmt.Printf("%s\n", response.Content[:min(500, len(response.Content))])
		fmt.Printf("... (truncated, %d total chars)\n\n", len(response.Content))
		codeResp = CodeGenerationResponse{
			Explanation: response.Content,
			Files: []File{
				{
					Path:    "output.md",
					Content: response.Content,
				},
			},
		}
	} else {
		// Show LLM's explanation
		if codeResp.Explanation != "" {
			fmt.Printf("\n💭 LLM Explanation:\n")
			fmt.Printf("%s\n", codeResp.Explanation)
			fmt.Println()
		}
	}

	fmt.Printf("📦 Generated %d file(s)\n", len(codeResp.Files))

	// Create files
	for i, file := range codeResp.Files {
		fmt.Printf("   Writing file %d/%d: %s\n", i+1, len(codeResp.Files), file.Path)

		// Show preview of file content (first 200 chars)
		preview := file.Content
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		fmt.Printf("   📝 Content preview: %s\n", preview)

		if err := te.writeFile(file); err != nil {
			return fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}
		fmt.Printf("   ✓ Created: %s (%d bytes)\n", file.Path, len(file.Content))
	}

	// Execute commands (optional - might be dangerous in auto-execution)
	if len(codeResp.Commands) > 0 {
		// For safety, just log commands for now
		// TODO: Implement command execution with confirmation
		fmt.Printf("📝 Commands to run:\n")
		for _, cmd := range codeResp.Commands {
			fmt.Printf("   %s\n", cmd.Command)
		}
	}

	return nil
}

func (te *TaskExecutor) getModelForTask(task *state.Task) string {
	// TODO: Get model from config or task specification
	// For now, use a sensible default
	return "openai/gpt-5-nano"
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
	promptBuilder.WriteString("3. Ensure code follows best practices for language/framework\n")
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
