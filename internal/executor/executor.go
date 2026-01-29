package executor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mojomast/geoffrussy/internal/provider"
	"github.com/mojomast/geoffrussy/internal/state"
)

// UpdateType represents the type of task update
type UpdateType string

const (
	TaskStarted   UpdateType = "started"
	TaskProgress  UpdateType = "progress"
	TaskCompleted UpdateType = "completed"
	TaskError     UpdateType = "error"
	TaskBlocked   UpdateType = "blocked"
	TaskPaused    UpdateType = "paused"
	TaskResumed   UpdateType = "resumed"
	TaskSkipped   UpdateType = "skipped"
)

// TaskUpdate represents a real-time update from task execution
type TaskUpdate struct {
	TaskID    string
	PhaseID   string
	Type      UpdateType
	Content   string
	Timestamp time.Time
	Error     error
}

// Executor handles task and phase execution
type Executor struct {
	store      *state.Store
	provider   provider.Provider
	updateChan chan TaskUpdate
	ctx        context.Context
	cancel     context.CancelFunc
	paused     bool
	pauseMu    sync.RWMutex
	pauseCond  *sync.Cond
}

// NewExecutor creates a new task executor
func NewExecutor(store *state.Store, provider provider.Provider) *Executor {
	ctx, cancel := context.WithCancel(context.Background())
	mu := &sync.RWMutex{}
	
	return &Executor{
		store:      store,
		provider:   provider,
		updateChan: make(chan TaskUpdate, 100),
		ctx:        ctx,
		cancel:     cancel,
		paused:     false,
		pauseCond:  sync.NewCond(mu),
	}
}

// ExecutePhase executes all tasks in a phase
func (e *Executor) ExecutePhase(phaseID string) error {
	// Get phase from store
	phase, err := e.store.GetPhase(phaseID)
	if err != nil {
		return fmt.Errorf("failed to get phase: %w", err)
	}

	// Update phase status to in_progress
	if err := e.store.UpdatePhaseStatus(phaseID, state.PhaseInProgress); err != nil {
		return fmt.Errorf("failed to update phase status: %w", err)
	}

	// Send phase started update
	e.sendUpdate(TaskUpdate{
		PhaseID:   phaseID,
		Type:      TaskStarted,
		Content:   fmt.Sprintf("Starting phase: %s", phase.Title),
		Timestamp: time.Now(),
	})

	// Get all tasks for this phase
	tasks, err := e.store.ListTasks(phaseID)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	for _, task := range tasks {
		if task.Status == state.TaskCompleted {
			continue
		}
		if err := e.ExecuteTask(task.ID); err != nil {
			// If task failed, stop phase execution
			e.sendUpdate(TaskUpdate{
				PhaseID:   phaseID,
				Type:      TaskError,
				Content:   fmt.Sprintf("Phase stopped due to task error: %v", err),
				Timestamp: time.Now(),
				Error:     err,
			})
			return err
		}
	}

	// Update phase status to completed
	if err := e.store.UpdatePhaseStatus(phaseID, state.PhaseCompleted); err != nil {
		return fmt.Errorf("failed to update phase status: %w", err)
	}

	// Send phase completed update
	e.sendUpdate(TaskUpdate{
		PhaseID:   phaseID,
		Type:      TaskCompleted,
		Content:   fmt.Sprintf("Completed phase: %s", phase.Title),
		Timestamp: time.Now(),
	})

	return nil
}

// ExecuteTask executes a single task
func (e *Executor) ExecuteTask(taskID string) error {
	// Check if paused
	e.checkPause()

	// Check if context is cancelled
	select {
	case <-e.ctx.Done():
		return fmt.Errorf("execution cancelled")
	default:
	}

	// Get task from store
	task, err := e.store.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Update task status to in_progress
	if err := e.store.UpdateTaskStatus(taskID, state.TaskInProgress); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Send task started update
	e.sendUpdate(TaskUpdate{
		TaskID:    taskID,
		PhaseID:   task.PhaseID,
		Type:      TaskStarted,
		Content:   fmt.Sprintf("Starting task: %s", task.Description),
		Timestamp: time.Now(),
	})

	// Execute the task using the provider
	// This is a simplified implementation
	// In a real implementation, we would:
	// 1. Parse the task description
	// 2. Generate code using the LLM
	// 3. Write files
	// 4. Run tests
	// 5. Handle errors and retries

	// Simulate task execution
	time.Sleep(100 * time.Millisecond)

	// Send progress updates
	e.sendUpdate(TaskUpdate{
		TaskID:    taskID,
		PhaseID:   task.PhaseID,
		Type:      TaskProgress,
		Content:   "Executing task...",
		Timestamp: time.Now(),
	})

	// Update task status to completed
	if err := e.store.UpdateTaskStatus(taskID, state.TaskCompleted); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Send task completed update
	e.sendUpdate(TaskUpdate{
		TaskID:    taskID,
		PhaseID:   task.PhaseID,
		Type:      TaskCompleted,
		Content:   fmt.Sprintf("Completed task: %s", task.Description),
		Timestamp: time.Now(),
	})

	return nil
}

// StreamOutput returns a channel for receiving task updates
func (e *Executor) StreamOutput() <-chan TaskUpdate {
	return e.updateChan
}

// PauseExecution pauses the current execution
func (e *Executor) PauseExecution() error {
	e.pauseMu.Lock()
	defer e.pauseMu.Unlock()

	if e.paused {
		return fmt.Errorf("execution already paused")
	}

	e.paused = true

	// Send pause update
	e.sendUpdate(TaskUpdate{
		Type:      TaskPaused,
		Content:   "Execution paused",
		Timestamp: time.Now(),
	})

	return nil
}

// ResumeExecution resumes paused execution
func (e *Executor) ResumeExecution() error {
	e.pauseMu.Lock()
	defer e.pauseMu.Unlock()

	if !e.paused {
		return fmt.Errorf("execution not paused")
	}

	e.paused = false
	e.pauseCond.Broadcast()

	// Send resume update
	e.sendUpdate(TaskUpdate{
		Type:      TaskResumed,
		Content:   "Execution resumed",
		Timestamp: time.Now(),
	})

	return nil
}

// SkipTask skips the current task
func (e *Executor) SkipTask(taskID string) error {
	// Update task status to skipped
	if err := e.store.UpdateTaskStatus(taskID, state.TaskSkipped); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Send skip update
	e.sendUpdate(TaskUpdate{
		TaskID:    taskID,
		Type:      TaskSkipped,
		Content:   "Task skipped",
		Timestamp: time.Now(),
	})

	return nil
}

// MarkBlocked marks a task as blocked
func (e *Executor) MarkBlocked(taskID, reason string) error {
	// Update task status to blocked
	if err := e.store.UpdateTaskStatus(taskID, state.TaskBlocked); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Get task to get phase ID
	task, err := e.store.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Save blocker to store
	blocker := &state.Blocker{
		ID:          fmt.Sprintf("blocker-%s-%d", taskID, time.Now().Unix()),
		TaskID:      taskID,
		Description: reason,
		CreatedAt:   time.Now(),
	}

	if err := e.store.SaveBlocker(blocker); err != nil {
		return fmt.Errorf("failed to save blocker: %w", err)
	}

	// Send blocked update
	e.sendUpdate(TaskUpdate{
		TaskID:    taskID,
		PhaseID:   task.PhaseID,
		Type:      TaskBlocked,
		Content:   fmt.Sprintf("Task blocked: %s", reason),
		Timestamp: time.Now(),
	})

	return nil
}

// ResolveBlocker resolves a blocker and resumes execution
func (e *Executor) ResolveBlocker(taskID, resolution string) error {
	// Get task to get phase ID
	task, err := e.store.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Get phase to get project ID
	phase, err := e.store.GetPhase(task.PhaseID)
	if err != nil {
		return fmt.Errorf("failed to get phase: %w", err)
	}

	// Get active blockers for this project
	blockers, err := e.store.ListActiveBlockers(phase.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to list blockers: %w", err)
	}

	// Find blocker for this task
	var blockerID string
	for _, blocker := range blockers {
		if blocker.TaskID == taskID {
			blockerID = blocker.ID
			break
		}
	}

	if blockerID == "" {
		return fmt.Errorf("no active blocker found for task %s", taskID)
	}

	// Resolve the blocker
	if err := e.store.ResolveBlocker(blockerID, resolution); err != nil {
		return fmt.Errorf("failed to resolve blocker: %w", err)
	}

	// Update task status back to pending
	// Note: We use TaskNotStarted as pending state
	if err := e.store.UpdateTaskStatus(taskID, state.TaskNotStarted); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

// Close closes the executor and cleans up resources
func (e *Executor) Close() {
	e.cancel()
	close(e.updateChan)
}

// checkPause checks if execution is paused and waits if necessary
func (e *Executor) checkPause() {
	e.pauseMu.RLock()
	paused := e.paused
	e.pauseMu.RUnlock()

	if paused {
		e.pauseCond.L.Lock()
		for e.paused {
			e.pauseCond.Wait()
		}
		e.pauseCond.L.Unlock()
	}
}

// sendUpdate sends an update to the update channel
func (e *Executor) sendUpdate(update TaskUpdate) {
	select {
	case e.updateChan <- update:
	case <-e.ctx.Done():
		// Context cancelled, don't send update
	default:
		// Channel full, drop update
		// In a production system, we might want to log this
	}
}
