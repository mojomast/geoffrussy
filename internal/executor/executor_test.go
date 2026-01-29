package executor

import (
	"testing"
	"time"

	"github.com/mojomast/geoffrussy/internal/provider"
	"github.com/mojomast/geoffrussy/internal/state"
)

func setupTestExecutor(t *testing.T) (*Executor, *state.Store) {
	// Create in-memory store
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Create mock provider (nil is acceptable for testing)
	var mockProvider provider.Provider = nil

	// Create executor
	executor := NewExecutor(store, mockProvider)

	return executor, store
}

func TestNewExecutor(t *testing.T) {
	executor, store := setupTestExecutor(t)
	defer store.Close()
	defer executor.Close()

	if executor == nil {
		t.Fatal("expected executor to be created")
	}

	if executor.store == nil {
		t.Error("expected store to be set")
	}

	// Provider can be nil in tests
	// if executor.provider == nil {
	// 	t.Error("expected provider to be set")
	// }

	if executor.updateChan == nil {
		t.Error("expected update channel to be created")
	}
}

func TestExecutor_PauseResume(t *testing.T) {
	executor, store := setupTestExecutor(t)
	defer store.Close()
	defer executor.Close()

	// Test pause
	if err := executor.PauseExecution(); err != nil {
		t.Errorf("failed to pause execution: %v", err)
	}

	if !executor.paused {
		t.Error("expected execution to be paused")
	}

	// Test pause when already paused
	if err := executor.PauseExecution(); err == nil {
		t.Error("expected error when pausing already paused execution")
	}

	// Test resume
	if err := executor.ResumeExecution(); err != nil {
		t.Errorf("failed to resume execution: %v", err)
	}

	if executor.paused {
		t.Error("expected execution to be resumed")
	}

	// Test resume when not paused
	if err := executor.ResumeExecution(); err == nil {
		t.Error("expected error when resuming non-paused execution")
	}
}

func TestExecutor_StreamOutput(t *testing.T) {
	executor, store := setupTestExecutor(t)
	defer store.Close()
	defer executor.Close()

	// Get output channel
	outputChan := executor.StreamOutput()
	if outputChan == nil {
		t.Fatal("expected output channel to be returned")
	}

	// Send an update
	update := TaskUpdate{
		TaskID:    "task-1",
		Type:      TaskStarted,
		Content:   "Test update",
		Timestamp: time.Now(),
	}

	executor.sendUpdate(update)

	// Receive the update
	select {
	case received := <-outputChan:
		if received.TaskID != update.TaskID {
			t.Errorf("expected task ID %s, got %s", update.TaskID, received.TaskID)
		}
		if received.Type != update.Type {
			t.Errorf("expected type %s, got %s", update.Type, received.Type)
		}
		if received.Content != update.Content {
			t.Errorf("expected content %s, got %s", update.Content, received.Content)
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for update")
	}
}

func TestExecutor_SkipTask(t *testing.T) {
	executor, store := setupTestExecutor(t)
	defer store.Close()
	defer executor.Close()

	// Create a test project
	project := &state.Project{
		ID:        "test-project",
		Name:      "Test Project",
		CreatedAt: time.Now(),
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Create a test phase
	phase := &state.Phase{
		ID:        "phase-1",
		ProjectID: project.ID,
		Number:    1,
		Title:     "Test Phase",
		Status:    state.PhaseNotStarted,
		CreatedAt: time.Now(),
	}
	if err := store.SavePhase(phase); err != nil {
		t.Fatalf("failed to save phase: %v", err)
	}

	// Create a test task
	task := &state.Task{
		ID:          "task-1",
		PhaseID:     phase.ID,
		Number:      "1.1",
		Description: "Test Task",
		Status:      state.TaskNotStarted,
	}
	if err := store.SaveTask(task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	// Skip the task
	if err := executor.SkipTask(task.ID); err != nil {
		t.Errorf("failed to skip task: %v", err)
	}

	// Verify task status
	updatedTask, err := store.GetTask(task.ID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	if updatedTask.Status != state.TaskSkipped {
		t.Errorf("expected task status to be '%s', got %s", state.TaskSkipped, updatedTask.Status)
	}
}

func TestExecutor_MarkBlocked(t *testing.T) {
	executor, store := setupTestExecutor(t)
	defer store.Close()
	defer executor.Close()

	// Create a test project
	project := &state.Project{
		ID:        "test-project",
		Name:      "Test Project",
		CreatedAt: time.Now(),
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Create a test phase
	phase := &state.Phase{
		ID:        "phase-1",
		ProjectID: project.ID,
		Number:    1,
		Title:     "Test Phase",
		Status:    state.PhaseNotStarted,
		CreatedAt: time.Now(),
	}
	if err := store.SavePhase(phase); err != nil {
		t.Fatalf("failed to save phase: %v", err)
	}

	// Create a test task
	task := &state.Task{
		ID:          "task-1",
		PhaseID:     phase.ID,
		Number:      "1.1",
		Description: "Test Task",
		Status:      state.TaskInProgress,
	}
	if err := store.SaveTask(task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	// Mark task as blocked
	reason := "Test blocker reason"
	if err := executor.MarkBlocked(task.ID, reason); err != nil {
		t.Errorf("failed to mark task as blocked: %v", err)
	}

	// Verify task status
	updatedTask, err := store.GetTask(task.ID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	if updatedTask.Status != state.TaskBlocked {
		t.Errorf("expected task status to be '%s', got %s", state.TaskBlocked, updatedTask.Status)
	}

	// Verify blocker was created
	blockers, err := store.ListActiveBlockers(project.ID)
	if err != nil {
		t.Fatalf("failed to list blockers: %v", err)
	}

	if len(blockers) != 1 {
		t.Errorf("expected 1 blocker, got %d", len(blockers))
	}

	if blockers[0].TaskID != task.ID {
		t.Errorf("expected blocker task ID %s, got %s", task.ID, blockers[0].TaskID)
	}

	if blockers[0].Description != reason {
		t.Errorf("expected blocker description %s, got %s", reason, blockers[0].Description)
	}
}

func TestExecutor_ResolveBlocker(t *testing.T) {
	executor, store := setupTestExecutor(t)
	defer store.Close()
	defer executor.Close()

	// Create a test project
	project := &state.Project{
		ID:        "test-project",
		Name:      "Test Project",
		CreatedAt: time.Now(),
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Create a test phase
	phase := &state.Phase{
		ID:        "phase-1",
		ProjectID: project.ID,
		Number:    1,
		Title:     "Test Phase",
		Status:    state.PhaseNotStarted,
		CreatedAt: time.Now(),
	}
	if err := store.SavePhase(phase); err != nil {
		t.Fatalf("failed to save phase: %v", err)
	}

	// Create a test task
	task := &state.Task{
		ID:          "task-1",
		PhaseID:     phase.ID,
		Number:      "1.1",
		Description: "Test Task",
		Status:      state.TaskBlocked,
	}
	if err := store.SaveTask(task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	// Create a blocker
	blocker := &state.Blocker{
		ID:          "blocker-1",
		TaskID:      task.ID,
		Description: "Test blocker",
		CreatedAt:   time.Now(),
	}
	if err := store.SaveBlocker(blocker); err != nil {
		t.Fatalf("failed to save blocker: %v", err)
	}

	// Resolve the blocker
	resolution := "Test resolution"
	if err := executor.ResolveBlocker(task.ID, resolution); err != nil {
		t.Errorf("failed to resolve blocker: %v", err)
	}

	// Verify task status
	updatedTask, err := store.GetTask(task.ID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	if updatedTask.Status != state.TaskNotStarted {
		t.Errorf("expected task status to be '%s', got %s", state.TaskNotStarted, updatedTask.Status)
	}

	// Verify blocker was resolved
	blockers, err := store.ListActiveBlockers(project.ID)
	if err != nil {
		t.Fatalf("failed to list blockers: %v", err)
	}

	if len(blockers) != 0 {
		t.Errorf("expected 0 active blockers, got %d", len(blockers))
	}
}

func TestExecutor_ExecuteTask(t *testing.T) {
	executor, store := setupTestExecutor(t)
	defer store.Close()
	defer executor.Close()

	// Create a test project
	project := &state.Project{
		ID:        "test-project",
		Name:      "Test Project",
		CreatedAt: time.Now(),
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Create a test phase
	phase := &state.Phase{
		ID:        "phase-1",
		ProjectID: project.ID,
		Number:    1,
		Title:     "Test Phase",
		Status:    state.PhaseNotStarted,
		CreatedAt: time.Now(),
	}
	if err := store.SavePhase(phase); err != nil {
		t.Fatalf("failed to save phase: %v", err)
	}

	// Create a test task
	task := &state.Task{
		ID:          "task-1",
		PhaseID:     phase.ID,
		Number:      "1.1",
		Description: "Test Task",
		Status:      state.TaskNotStarted,
	}
	if err := store.SaveTask(task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	// Execute the task
	go func() {
		if err := executor.ExecuteTask(task.ID); err != nil {
			t.Errorf("failed to execute task: %v", err)
		}
	}()

	// Collect updates
	var updates []TaskUpdate
	timeout := time.After(2 * time.Second)

	for {
		select {
		case update := <-executor.StreamOutput():
			updates = append(updates, update)
			if update.Type == TaskCompleted {
				goto done
			}
		case <-timeout:
			t.Fatal("timeout waiting for task completion")
		}
	}

done:
	// Verify we received updates
	if len(updates) < 2 {
		t.Errorf("expected at least 2 updates, got %d", len(updates))
	}

	// Verify task status
	updatedTask, err := store.GetTask(task.ID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	if updatedTask.Status != state.TaskCompleted {
		t.Errorf("expected task status to be '%s', got %s", state.TaskCompleted, updatedTask.Status)
	}
}

func TestExecutor_ExecutePhase(t *testing.T) {
	executor, store := setupTestExecutor(t)
	defer store.Close()
	defer executor.Close()

	// Create a test project
	project := &state.Project{
		ID:        "test-project",
		Name:      "Test Project",
		CreatedAt: time.Now(),
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Create a test phase
	phase := &state.Phase{
		ID:        "phase-1",
		ProjectID: project.ID,
		Number:    1,
		Title:     "Test Phase",
		Status:    state.PhaseNotStarted,
		CreatedAt: time.Now(),
	}
	if err := store.SavePhase(phase); err != nil {
		t.Fatalf("failed to save phase: %v", err)
	}

	// Create a task for the phase so ExecutePhase has something to do
	task := &state.Task{
		ID:          "task-1",
		PhaseID:     phase.ID,
		Number:      "1.1",
		Description: "Test Task",
		Status:      state.TaskNotStarted,
	}
	if err := store.SaveTask(task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	// Execute the phase
	go func() {
		if err := executor.ExecutePhase(phase.ID); err != nil {
			t.Errorf("failed to execute phase: %v", err)
		}
	}()

	// Collect updates
	var updates []TaskUpdate
	timeout := time.After(2 * time.Second)

	for {
		select {
		case update := <-executor.StreamOutput():
			updates = append(updates, update)
			if update.Type == TaskCompleted && update.PhaseID == phase.ID {
				goto done
			}
		case <-timeout:
			t.Fatal("timeout waiting for phase completion")
		}
	}

done:
	// Verify we received updates
	if len(updates) < 2 {
		t.Errorf("expected at least 2 updates, got %d", len(updates))
	}

	// Verify phase status
	updatedPhase, err := store.GetPhase(phase.ID)
	if err != nil {
		t.Fatalf("failed to get phase: %v", err)
	}

	if updatedPhase.Status != state.PhaseCompleted {
		t.Errorf("expected phase status to be '%s', got %s", state.PhaseCompleted, updatedPhase.Status)
	}
}
