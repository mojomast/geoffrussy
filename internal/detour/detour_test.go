package detour

import (
	"testing"
	"time"

	"github.com/mojomast/geoffrussy/internal/devplan"
	"github.com/mojomast/geoffrussy/internal/state"
)

func TestRequestDetour(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	manager := NewManager(store, nil, nil)

	detour, err := manager.RequestDetour("project-1", "phase-1", "task-1", "Add new feature", "User requested")
	if err != nil {
		t.Fatalf("failed to request detour: %v", err)
	}

	if detour.ProjectID != "project-1" {
		t.Errorf("expected project ID 'project-1', got '%s'", detour.ProjectID)
	}

	if detour.PhaseID != "phase-1" {
		t.Errorf("expected phase ID 'phase-1', got '%s'", detour.PhaseID)
	}

	if detour.TaskID != "task-1" {
		t.Errorf("expected task ID 'task-1', got '%s'", detour.TaskID)
	}

	if detour.Description != "Add new feature" {
		t.Errorf("expected description 'Add new feature', got '%s'", detour.Description)
	}

	if detour.Reason != "User requested" {
		t.Errorf("expected reason 'User requested', got '%s'", detour.Reason)
	}

	if detour.Status != DetourPending {
		t.Errorf("expected status 'pending', got '%s'", detour.Status)
	}

	if detour.CreatedAt.IsZero() {
		t.Error("expected created at to be set")
	}
}

func TestGatherDetourInformation(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	manager := NewManager(store, nil, nil)

	detour := &Detour{
		ID:          "detour-1",
		ProjectID:   "project-1",
		PhaseID:     "phase-1",
		TaskID:      "task-1",
		Description: "Add new feature",
		Reason:      "User requested",
		Status:      DetourPending,
		CreatedAt:   time.Now(),
	}

	err = manager.GatherDetourInformation(detour)
	if err != nil {
		t.Fatalf("failed to gather detour information: %v", err)
	}

	if detour.Status != DetourPlanned {
		t.Errorf("expected status 'planned', got '%s'", detour.Status)
	}
}

func TestGatherDetourInformation_InvalidStatus(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	manager := NewManager(store, nil, nil)

	detour := &Detour{
		ID:          "detour-1",
		ProjectID:   "project-1",
		PhaseID:     "phase-1",
		TaskID:      "task-1",
		Description: "Add new feature",
		Reason:      "User requested",
		Status:      DetourCompleted, // Invalid status
		CreatedAt:   time.Now(),
	}

	err = manager.GatherDetourInformation(detour)
	if err == nil {
		t.Error("expected error when gathering information for non-pending detour")
	}
}

func TestGenerateDetourTasks(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	manager := NewManager(store, nil, nil)

	detour := &Detour{
		ID:          "detour-1",
		ProjectID:   "project-1",
		PhaseID:     "phase-1",
		TaskID:      "task-1",
		Description: "Add new feature",
		Reason:      "User requested",
		Status:      DetourPlanned,
		CreatedAt:   time.Now(),
	}

	tasks := manager.generateDetourTasks(detour)

	if len(tasks) == 0 {
		t.Error("expected at least one task to be generated")
	}

	if tasks[0].Description == "" {
		t.Error("expected task description to be set")
	}

	if tasks[0].Status != devplan.TaskNotStarted {
		t.Errorf("expected task status 'not_started', got '%s'", tasks[0].Status)
	}
}

func TestUpdateDevPlan_InvalidStatus(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	manager := NewManager(store, nil, nil)

	detour := &Detour{
		ID:          "detour-1",
		ProjectID:   "project-1",
		PhaseID:     "phase-1",
		TaskID:      "task-1",
		Description: "Add new feature",
		Reason:      "User requested",
		Status:      DetourPending, // Invalid status for update
		CreatedAt:   time.Now(),
	}

	err = manager.UpdateDevPlan(detour, "task-1")
	if err == nil {
		t.Error("expected error when updating devplan for non-planned detour")
	}
}

func TestValidateDetourDependencies(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	manager := NewManager(store, nil, nil)

	detour := &Detour{
		ID:          "detour-1",
		ProjectID:   "project-1",
		PhaseID:     "phase-1",
		TaskID:      "task-1",
		Description: "Add new feature",
		Reason:      "User requested",
		Status:      DetourPlanned,
		CreatedAt:   time.Now(),
	}

	phase := &devplan.Phase{
		ID:     "phase-1",
		Number: 1,
		Title:  "Test Phase",
		Tasks: []devplan.Task{
			{
				ID:          "task-1",
				Number:      "1.1",
				Description: "Existing task",
				Status:      devplan.TaskNotStarted,
			},
		},
	}

	valid, conflicts := manager.ValidateDetourDependencies(detour, phase)

	if !valid {
		t.Errorf("expected detour to be valid, got conflicts: %v", conflicts)
	}

	if len(conflicts) > 0 {
		t.Errorf("expected no conflicts, got %d", len(conflicts))
	}
}

func TestDetourStatusTransitions(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  DetourStatus
		expectedStatus DetourStatus
		operation      string
	}{
		{
			name:           "Pending to Gathering",
			initialStatus:  DetourPending,
			expectedStatus: DetourPlanned,
			operation:      "gather",
		},
		{
			name:           "Planned to Active",
			initialStatus:  DetourPlanned,
			expectedStatus: DetourActive,
			operation:      "update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := state.NewStore(":memory:")
			if err != nil {
				t.Fatalf("failed to create store: %v", err)
			}
			defer store.Close()

			manager := NewManager(store, nil, nil)

			detour := &Detour{
				ID:          "detour-1",
				ProjectID:   "project-1",
				PhaseID:     "phase-1",
				TaskID:      "task-1",
				Description: "Add new feature",
				Reason:      "User requested",
				Status:      tt.initialStatus,
				CreatedAt:   time.Now(),
			}

			switch tt.operation {
			case "gather":
				err = manager.GatherDetourInformation(detour)
			case "update":
				// Create a test phase first
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
					Status:    "not_started",
					CreatedAt: time.Now(),
				}
				if err := store.SavePhase(phase); err != nil {
					t.Fatalf("failed to save phase: %v", err)
				}

				err = manager.UpdateDevPlan(detour, "task-1")
			}

			if err != nil {
				t.Fatalf("operation failed: %v", err)
			}

			if detour.Status != tt.expectedStatus {
				t.Errorf("expected status '%s', got '%s'", tt.expectedStatus, detour.Status)
			}
		})
	}
}

func TestSaveDetour(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	manager := NewManager(store, nil, nil)

	detour := &Detour{
		ID:          "detour-1",
		ProjectID:   "project-1",
		PhaseID:     "phase-1",
		TaskID:      "task-1",
		Description: "Add new feature",
		Reason:      "User requested",
		Status:      DetourPlanned,
		CreatedAt:   time.Now(),
	}

	err = manager.SaveDetour(detour)
	if err != nil {
		t.Fatalf("failed to save detour: %v", err)
	}
}

func TestTrackDetourInDirectory(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	manager := NewManager(store, nil, nil)

	detour := &Detour{
		ID:          "detour-1",
		ProjectID:   "project-1",
		PhaseID:     "phase-1",
		TaskID:      "task-1",
		Description: "Add new feature",
		Reason:      "User requested",
		Status:      DetourPlanned,
		CreatedAt:   time.Now(),
	}

	err = manager.TrackDetourInDirectory(detour, "./detours")
	if err != nil {
		t.Fatalf("failed to track detour in directory: %v", err)
	}
}

func TestGetDetourDependencies(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	manager := NewManager(store, nil, nil)

	detour := &Detour{
		ID:          "detour-1",
		ProjectID:   "project-1",
		PhaseID:     "phase-1",
		TaskID:      "task-1",
		Description: "Add new feature",
		Reason:      "User requested",
		Status:      DetourPlanned,
		CreatedAt:   time.Now(),
		NewTasks: []devplan.Task{
			{
				ID:          "detour-task-1",
				Number:      "detour-1",
				Description: "Implement detour feature",
				Status:      devplan.TaskNotStarted,
			},
		},
	}

	deps, err := manager.GetDetourDependencies(detour)
	if err != nil {
		t.Fatalf("failed to get detour dependencies: %v", err)
	}

	if deps == nil {
		t.Error("expected dependencies to be non-nil")
	}
}

func TestUpdateTaskDependencies(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	manager := NewManager(store, nil, nil)

	detour := &Detour{
		ID:          "detour-1",
		ProjectID:   "project-1",
		PhaseID:     "phase-1",
		TaskID:      "task-1",
		Description: "Add new feature",
		Reason:      "User requested",
		Status:      DetourPlanned,
		CreatedAt:   time.Now(),
	}

	affectedTaskIDs := []string{"task-2", "task-3"}

	err = manager.UpdateTaskDependencies(detour, affectedTaskIDs)
	if err != nil {
		t.Fatalf("failed to update task dependencies: %v", err)
	}
}

func TestExportDetourMarkdown(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	manager := NewManager(store, nil, nil)

	completedAt := time.Now()
	detour := &Detour{
		ID:          "detour-1",
		ProjectID:   "project-1",
		PhaseID:     "phase-1",
		TaskID:      "task-1",
		Description: "Add new feature",
		Reason:      "User requested",
		Status:      DetourCompleted,
		CreatedAt:   time.Now(),
		CompletedAt: &completedAt,
		NewTasks: []devplan.Task{
			{
				ID:                  "detour-task-1",
				Number:              "detour-1",
				Description:         "Implement detour feature",
				AcceptanceCriteria:  []string{"Feature works", "Tests pass"},
				ImplementationNotes: []string{"Use existing API"},
				Status:              devplan.TaskCompleted,
			},
		},
	}

	md, err := manager.ExportDetourMarkdown(detour)
	if err != nil {
		t.Fatalf("failed to export detour markdown: %v", err)
	}

	if md == "" {
		t.Error("expected markdown to be non-empty")
	}

	// Check that markdown contains key information
	if !contains(md, "detour-1") {
		t.Error("expected markdown to contain detour ID")
	}

	if !contains(md, "Add new feature") {
		t.Error("expected markdown to contain description")
	}

	if !contains(md, "User requested") {
		t.Error("expected markdown to contain reason")
	}

	if !contains(md, "Implement detour feature") {
		t.Error("expected markdown to contain task description")
	}
}

func TestExportDetourMarkdown_NoTasks(t *testing.T) {
	store, err := state.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	manager := NewManager(store, nil, nil)

	detour := &Detour{
		ID:          "detour-1",
		ProjectID:   "project-1",
		PhaseID:     "phase-1",
		TaskID:      "task-1",
		Description: "Add new feature",
		Reason:      "User requested",
		Status:      DetourPending,
		CreatedAt:   time.Now(),
		NewTasks:    []devplan.Task{},
	}

	md, err := manager.ExportDetourMarkdown(detour)
	if err != nil {
		t.Fatalf("failed to export detour markdown: %v", err)
	}

	if md == "" {
		t.Error("expected markdown to be non-empty")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || contains(s[1:], substr)))
}
