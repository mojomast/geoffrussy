package blocker

import (
	"fmt"
	"testing"
	"time"

	"github.com/mojomast/geoffrussy/internal/state"
)

func TestRecordFailure(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	detector := NewDetector(store, nil)

	// Record first failure
	blocked, err := detector.RecordFailure("task-1", "error 1")
	if err != nil {
		t.Fatalf("failed to record failure: %v", err)
	}

	if blocked {
		t.Error("expected task not to be blocked after first failure")
	}

	if detector.GetFailureCount("task-1") != 1 {
		t.Errorf("expected failure count 1, got %d", detector.GetFailureCount("task-1"))
	}

	// Record second failure
	blocked, err = detector.RecordFailure("task-1", "error 2")
	if err != nil {
		t.Fatalf("failed to record failure: %v", err)
	}

	if blocked {
		t.Error("expected task not to be blocked after second failure")
	}

	if detector.GetFailureCount("task-1") != 2 {
		t.Errorf("expected failure count 2, got %d", detector.GetFailureCount("task-1"))
	}

	// Record third failure (should trigger blocking)
	blocked, err = detector.RecordFailure("task-1", "error 3")
	if err != nil {
		t.Fatalf("failed to record failure: %v", err)
	}

	if !blocked {
		t.Error("expected task to be blocked after third failure")
	}

	if detector.GetFailureCount("task-1") != 3 {
		t.Errorf("expected failure count 3, got %d", detector.GetFailureCount("task-1"))
	}
}

func TestMarkAsBlocked(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Create test project, phase, and task
	project := &state.Project{
		ID:        "project-1",
		Name:      "Test Project",
		CreatedAt: time.Now(),
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	phase := &state.Phase{
		ID:        "phase-1",
		ProjectID: "project-1",
		Number:    1,
		Title:     "Test Phase",
		Status:    "in_progress",
		CreatedAt: time.Now(),
	}
	if err := store.SavePhase(phase); err != nil {
		t.Fatalf("failed to save phase: %v", err)
	}

	task := &state.Task{
		ID:          "task-1",
		PhaseID:     "phase-1",
		Number:      "1",
		Description: "Test Task",
		Status:      "in_progress",
	}
	if err := store.SaveTask(task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	detector := NewDetector(store, nil)

	blocker, err := detector.MarkAsBlocked("task-1", "phase-1", "project-1", "Task failed repeatedly", "Error context")
	if err != nil {
		t.Fatalf("failed to mark as blocked: %v", err)
	}

	if blocker.TaskID != "task-1" {
		t.Errorf("expected task ID 'task-1', got '%s'", blocker.TaskID)
	}

	if blocker.Description == "" {
		t.Error("expected description to be set")
	}

	// Verify task status was updated
	updatedTask, err := store.GetTask("task-1")
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	if updatedTask.Status != "blocked" {
		t.Errorf("expected task status 'blocked', got '%s'", updatedTask.Status)
	}
}

func TestGatherBlockerInformation(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	detector := NewDetector(store, nil)

	blocker := &state.Blocker{
		ID:          "blocker-1",
		TaskID:      "task-1",
		Description: "Task failed",
		CreatedAt:   time.Now(),
	}

	info, err := detector.GatherBlockerInformation(blocker)
	if err != nil {
		t.Fatalf("failed to gather blocker information: %v", err)
	}

	if info["blocker_id"] != "blocker-1" {
		t.Errorf("expected blocker_id 'blocker-1', got '%s'", info["blocker_id"])
	}

	if info["task_id"] != "task-1" {
		t.Errorf("expected task_id 'task-1', got '%s'", info["task_id"])
	}

	if info["description"] != "Task failed" {
		t.Errorf("expected description 'Task failed', got '%s'", info["description"])
	}
}

func TestAttemptResolution(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	detector := NewDetector(store, nil)

	blocker := &state.Blocker{
		ID:          "blocker-1",
		TaskID:      "task-1",
		Description: "Task failed",
		CreatedAt:   time.Now(),
	}

	result, err := detector.AttemptResolution(blocker)
	if err != nil {
		t.Fatalf("failed to attempt resolution: %v", err)
	}

	if result.BlockerID != "blocker-1" {
		t.Errorf("expected blocker ID 'blocker-1', got '%s'", result.BlockerID)
	}

	if len(result.Strategies) == 0 {
		t.Error("expected at least one resolution strategy")
	}

	// Check that automatic strategies were attempted
	if len(result.AttemptedStrategies) == 0 {
		t.Error("expected at least one automatic strategy to be attempted")
	}
}

func TestResolveBlocker(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Create test project, phase, and task
	project := &state.Project{
		ID:        "project-1",
		Name:      "Test Project",
		CreatedAt: time.Now(),
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	phase := &state.Phase{
		ID:        "phase-1",
		ProjectID: "project-1",
		Number:    1,
		Title:     "Test Phase",
		Status:    "in_progress",
		CreatedAt: time.Now(),
	}
	if err := store.SavePhase(phase); err != nil {
		t.Fatalf("failed to save phase: %v", err)
	}

	task := &state.Task{
		ID:          "task-1",
		PhaseID:     "phase-1",
		Number:      "1",
		Description: "Test Task",
		Status:      "blocked",
	}
	if err := store.SaveTask(task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	detector := NewDetector(store, nil)

	// Create a blocker
	blocker := &state.Blocker{
		ID:          "blocker-1",
		TaskID:      "task-1",
		Description: "Task failed",
		CreatedAt:   time.Now(),
	}
	if err := store.SaveBlocker(blocker); err != nil {
		t.Fatalf("failed to save blocker: %v", err)
	}

	// Set failure count
	detector.failureTracker["task-1"] = 3

	// Resolve the blocker
	err = detector.ResolveBlocker("blocker-1", "Fixed the issue")
	if err != nil {
		t.Fatalf("failed to resolve blocker: %v", err)
	}

	// Verify failure count was reset
	if detector.GetFailureCount("task-1") != 0 {
		t.Errorf("expected failure count 0, got %d", detector.GetFailureCount("task-1"))
	}

	// Verify task status was updated
	updatedTask, err := store.GetTask("task-1")
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	if updatedTask.Status != "pending" {
		t.Errorf("expected task status 'pending', got '%s'", updatedTask.Status)
	}
}

func TestGetFailureCount(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	detector := NewDetector(store, nil)

	// Initially should be 0
	if detector.GetFailureCount("task-1") != 0 {
		t.Errorf("expected failure count 0, got %d", detector.GetFailureCount("task-1"))
	}

	// Record some failures
	detector.RecordFailure("task-1", "error 1")
	detector.RecordFailure("task-1", "error 2")

	if detector.GetFailureCount("task-1") != 2 {
		t.Errorf("expected failure count 2, got %d", detector.GetFailureCount("task-1"))
	}
}

func TestResetFailureCount(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	detector := NewDetector(store, nil)

	// Record some failures
	detector.RecordFailure("task-1", "error 1")
	detector.RecordFailure("task-1", "error 2")

	if detector.GetFailureCount("task-1") != 2 {
		t.Errorf("expected failure count 2, got %d", detector.GetFailureCount("task-1"))
	}

	// Reset
	detector.ResetFailureCount("task-1")

	if detector.GetFailureCount("task-1") != 0 {
		t.Errorf("expected failure count 0 after reset, got %d", detector.GetFailureCount("task-1"))
	}
}

func TestListActiveBlockers(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Create test project, phase, and tasks
	project := &state.Project{
		ID:        "project-1",
		Name:      "Test Project",
		CreatedAt: time.Now(),
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	phase := &state.Phase{
		ID:        "phase-1",
		ProjectID: "project-1",
		Number:    1,
		Title:     "Test Phase",
		Status:    "in_progress",
		CreatedAt: time.Now(),
	}
	if err := store.SavePhase(phase); err != nil {
		t.Fatalf("failed to save phase: %v", err)
	}

	task1 := &state.Task{
		ID:          "task-1",
		PhaseID:     "phase-1",
		Number:      "1",
		Description: "Test Task 1",
		Status:      "blocked",
	}
	if err := store.SaveTask(task1); err != nil {
		t.Fatalf("failed to save task 1: %v", err)
	}

	task2 := &state.Task{
		ID:          "task-2",
		PhaseID:     "phase-1",
		Number:      "2",
		Description: "Test Task 2",
		Status:      "blocked",
	}
	if err := store.SaveTask(task2); err != nil {
		t.Fatalf("failed to save task 2: %v", err)
	}

	detector := NewDetector(store, nil)

	// Create some blockers
	blocker1 := &state.Blocker{
		ID:          "blocker-1",
		TaskID:      "task-1",
		Description: "Task 1 failed",
		CreatedAt:   time.Now(),
	}
	if err := store.SaveBlocker(blocker1); err != nil {
		t.Fatalf("failed to save blocker 1: %v", err)
	}

	blocker2 := &state.Blocker{
		ID:          "blocker-2",
		TaskID:      "task-2",
		Description: "Task 2 failed",
		CreatedAt:   time.Now(),
	}
	if err := store.SaveBlocker(blocker2); err != nil {
		t.Fatalf("failed to save blocker 2: %v", err)
	}

	blockers, err := detector.ListActiveBlockers("project-1")
	if err != nil {
		t.Fatalf("failed to list active blockers: %v", err)
	}

	if len(blockers) != 2 {
		t.Errorf("expected 2 blockers, got %d", len(blockers))
	}
}

func TestAnalyzeBlockerPattern(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Create test project, phase, and tasks
	project := &state.Project{
		ID:        "project-1",
		Name:      "Test Project",
		CreatedAt: time.Now(),
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	phase := &state.Phase{
		ID:        "phase-1",
		ProjectID: "project-1",
		Number:    1,
		Title:     "Test Phase",
		Status:    "in_progress",
		CreatedAt: time.Now(),
	}
	if err := store.SavePhase(phase); err != nil {
		t.Fatalf("failed to save phase: %v", err)
	}

	task1 := &state.Task{
		ID:          "task-1",
		PhaseID:     "phase-1",
		Number:      "1",
		Description: "Test Task 1",
		Status:      "blocked",
	}
	if err := store.SaveTask(task1); err != nil {
		t.Fatalf("failed to save task 1: %v", err)
	}

	task2 := &state.Task{
		ID:          "task-2",
		PhaseID:     "phase-1",
		Number:      "2",
		Description: "Test Task 2",
		Status:      "blocked",
	}
	if err := store.SaveTask(task2); err != nil {
		t.Fatalf("failed to save task 2: %v", err)
	}

	detector := NewDetector(store, nil)

	// Create some blockers with patterns
	blocker1 := &state.Blocker{
		ID:          "blocker-1",
		TaskID:      "task-1",
		Description: "Network timeout",
		CreatedAt:   time.Now(),
	}
	if err := store.SaveBlocker(blocker1); err != nil {
		t.Fatalf("failed to save blocker 1: %v", err)
	}

	blocker2 := &state.Blocker{
		ID:          "blocker-2",
		TaskID:      "task-1",
		Description: "Network timeout",
		CreatedAt:   time.Now(),
	}
	if err := store.SaveBlocker(blocker2); err != nil {
		t.Fatalf("failed to save blocker 2: %v", err)
	}

	blocker3 := &state.Blocker{
		ID:          "blocker-3",
		TaskID:      "task-2",
		Description: "Database error",
		CreatedAt:   time.Now(),
	}
	if err := store.SaveBlocker(blocker3); err != nil {
		t.Fatalf("failed to save blocker 3: %v", err)
	}

	analysis, err := detector.AnalyzeBlockerPattern("project-1")
	if err != nil {
		t.Fatalf("failed to analyze blocker pattern: %v", err)
	}

	if analysis.TotalBlockers != 3 {
		t.Errorf("expected 3 total blockers, got %d", analysis.TotalBlockers)
	}

	if analysis.BlockersByTask["task-1"] != 2 {
		t.Errorf("expected 2 blockers for task-1, got %d", analysis.BlockersByTask["task-1"])
	}

	if analysis.BlockersByTask["task-2"] != 1 {
		t.Errorf("expected 1 blocker for task-2, got %d", analysis.BlockersByTask["task-2"])
	}

	if analysis.CommonDescriptions["Network timeout"] != 2 {
		t.Errorf("expected 2 'Network timeout' descriptions, got %d", analysis.CommonDescriptions["Network timeout"])
	}
}

func TestRequestUserIntervention(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	detector := NewDetector(store, nil)

	blocker := &state.Blocker{
		ID:          "blocker-1",
		TaskID:      "task-1",
		Description: "Task failed",
		CreatedAt:   time.Now(),
	}

	err = detector.RequestUserIntervention(blocker, "Additional context")
	if err != nil {
		t.Fatalf("failed to request user intervention: %v", err)
	}
}

func TestResolutionStrategies(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	detector := NewDetector(store, nil)

	blocker := &state.Blocker{
		ID:          "blocker-1",
		TaskID:      "task-1",
		Description: "Task failed",
		CreatedAt:   time.Now(),
	}

	result, err := detector.AttemptResolution(blocker)
	if err != nil {
		t.Fatalf("failed to attempt resolution: %v", err)
	}

	// Check that we have both automatic and manual strategies
	hasAutomatic := false
	hasManual := false

	for _, strategy := range result.Strategies {
		if strategy.Automatic {
			hasAutomatic = true
		} else {
			hasManual = true
		}
	}

	if !hasAutomatic {
		t.Error("expected at least one automatic strategy")
	}

	if !hasManual {
		t.Error("expected at least one manual strategy")
	}
}

func TestGetBlocker(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Create test project, phase, and task
	project := &state.Project{
		ID:        "project-1",
		Name:      "Test Project",
		CreatedAt: time.Now(),
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	phase := &state.Phase{
		ID:        "phase-1",
		ProjectID: "project-1",
		Number:    1,
		Title:     "Test Phase",
		Status:    "in_progress",
		CreatedAt: time.Now(),
	}
	if err := store.SavePhase(phase); err != nil {
		t.Fatalf("failed to save phase: %v", err)
	}

	task := &state.Task{
		ID:          "task-1",
		PhaseID:     "phase-1",
		Number:      "1",
		Description: "Test Task",
		Status:      "blocked",
	}
	if err := store.SaveTask(task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	detector := NewDetector(store, nil)

	// Create a blocker
	blocker := &state.Blocker{
		ID:          "blocker-1",
		TaskID:      "task-1",
		Description: "Task failed",
		CreatedAt:   time.Now(),
	}
	if err := store.SaveBlocker(blocker); err != nil {
		t.Fatalf("failed to save blocker: %v", err)
	}

	// Get the blocker - need to use ListActiveBlockers with project ID
	blockers, err := detector.ListActiveBlockers("project-1")
	if err != nil {
		t.Fatalf("failed to list blockers: %v", err)
	}

	if len(blockers) == 0 {
		t.Fatal("expected at least one blocker")
	}

	retrieved := blockers[0]
	if retrieved.ID != "blocker-1" {
		t.Errorf("expected blocker ID 'blocker-1', got '%s'", retrieved.ID)
	}

	if retrieved.TaskID != "task-1" {
		t.Errorf("expected task ID 'task-1', got '%s'", retrieved.TaskID)
	}
}

func TestGetBlocker_NotFound(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	detector := NewDetector(store, nil)

	// Try to get a non-existent blocker
	_, err = detector.GetBlocker("non-existent")
	if err == nil {
		t.Error("expected error when getting non-existent blocker")
	}
}

func TestBlockerWorkflow(t *testing.T) {
	// This test demonstrates the complete blocker workflow
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Create test project, phase, and task
	project := &state.Project{
		ID:        "project-1",
		Name:      "Test Project",
		CreatedAt: time.Now(),
	}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	phase := &state.Phase{
		ID:        "phase-1",
		ProjectID: "project-1",
		Number:    1,
		Title:     "Test Phase",
		Status:    "in_progress",
		CreatedAt: time.Now(),
	}
	if err := store.SavePhase(phase); err != nil {
		t.Fatalf("failed to save phase: %v", err)
	}

	task := &state.Task{
		ID:          "task-1",
		PhaseID:     "phase-1",
		Number:      "1",
		Description: "Test Task",
		Status:      "in_progress",
	}
	if err := store.SaveTask(task); err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	detector := NewDetector(store, nil)

	// Step 1: Record failures
	for i := 0; i < 3; i++ {
		blocked, err := detector.RecordFailure("task-1", fmt.Sprintf("error %d", i+1))
		if err != nil {
			t.Fatalf("failed to record failure %d: %v", i+1, err)
		}

		if i < 2 && blocked {
			t.Errorf("expected task not to be blocked after failure %d", i+1)
		}

		if i == 2 && !blocked {
			t.Error("expected task to be blocked after 3 failures")
		}
	}

	// Step 2: Mark as blocked
	blocker, err := detector.MarkAsBlocked("task-1", "phase-1", "project-1", "Task failed repeatedly", "Error context")
	if err != nil {
		t.Fatalf("failed to mark as blocked: %v", err)
	}

	// Step 3: Gather information
	info, err := detector.GatherBlockerInformation(blocker)
	if err != nil {
		t.Fatalf("failed to gather blocker information: %v", err)
	}

	if info["blocker_id"] != blocker.ID {
		t.Errorf("expected blocker_id '%s', got '%s'", blocker.ID, info["blocker_id"])
	}

	// Step 4: Attempt resolution
	result, err := detector.AttemptResolution(blocker)
	if err != nil {
		t.Fatalf("failed to attempt resolution: %v", err)
	}

	if len(result.Strategies) == 0 {
		t.Error("expected at least one resolution strategy")
	}

	// Step 5: Resolve blocker
	err = detector.ResolveBlocker(blocker.ID, "Fixed the issue")
	if err != nil {
		t.Fatalf("failed to resolve blocker: %v", err)
	}

	// Verify failure count was reset
	if detector.GetFailureCount("task-1") != 0 {
		t.Errorf("expected failure count 0 after resolution, got %d", detector.GetFailureCount("task-1"))
	}

	// Verify task status was updated
	updatedTask, err := store.GetTask("task-1")
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	if updatedTask.Status != "pending" {
		t.Errorf("expected task status 'pending' after resolution, got '%s'", updatedTask.Status)
	}
}
