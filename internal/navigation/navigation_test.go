package navigation

import (
	"os"
	"testing"
	"time"

	"github.com/mojomast/geoffrussy/internal/git"
	"github.com/mojomast/geoffrussy/internal/state"
)

func TestNavigateToStage(t *testing.T) {
	// Create temporary database
	tmpDB := t.TempDir() + "/test.db"
	defer os.Remove(tmpDB)

	store, err := state.NewStore(tmpDB)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Create git manager (won't actually commit in test)
	gitMgr := git.NewManager(".")

	// Create navigator
	nav := NewNavigator(store, gitMgr)

	// Create test project
	project := &state.Project{
		ID:           "test-project-1",
		Name:         "Test Project",
		CreatedAt:    time.Now(),
		CurrentStage: state.StageDesign,
		CurrentPhase: "",
	}

	if err := store.CreateProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Navigate back to interview
	result, err := nav.NavigateToStage(project.ID, state.StageInterview)
	if err != nil {
		t.Errorf("failed to navigate to interview: %v", err)
	}

	if result == nil {
		t.Fatal("expected navigation result, got nil")
	}

	if result.FromStage != state.StageDesign {
		t.Errorf("expected from stage design, got %s", result.FromStage)
	}

	if result.ToStage != state.StageInterview {
		t.Errorf("expected to stage interview, got %s", result.ToStage)
	}

	// Verify project stage was updated
	updatedProject, err := store.GetProject(project.ID)
	if err != nil {
		t.Errorf("failed to get updated project: %v", err)
	}

	if updatedProject.CurrentStage != state.StageInterview {
		t.Errorf("expected current stage interview, got %s", updatedProject.CurrentStage)
	}
}

func TestValidateNavigation(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	defer os.Remove(tmpDB)

	store, err := state.NewStore(tmpDB)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	gitMgr := git.NewManager(".")
	nav := NewNavigator(store, gitMgr)

	testCases := []struct {
		name        string
		from        state.Stage
		to          state.Stage
		expectError bool
	}{
		{
			name:        "Can go back from design to interview",
			from:        state.StageDesign,
			to:          state.StageInterview,
			expectError: false,
		},
		{
			name:        "Can go forward one stage",
			from:        state.StageInterview,
			to:          state.StageDesign,
			expectError: false,
		},
		{
			name:        "Cannot skip stages",
			from:        state.StageInterview,
			to:          state.StagePlan,
			expectError: true,
		},
		{
			name:        "Cannot navigate to same stage",
			from:        state.StageDesign,
			to:          state.StageDesign,
			expectError: true,
		},
		{
			name:        "Can go back multiple stages",
			from:        state.StageDevelop,
			to:          state.StageInterview,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := nav.ValidateNavigation(tc.from, tc.to)
			if tc.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestGetNavigationOptions(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	defer os.Remove(tmpDB)

	store, err := state.NewStore(tmpDB)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	gitMgr := git.NewManager(".")
	nav := NewNavigator(store, gitMgr)

	// Create project at design stage
	project := &state.Project{
		ID:           "test-project-2",
		Name:         "Test Project",
		CreatedAt:    time.Now(),
		CurrentStage: state.StageDesign,
		CurrentPhase: "",
	}

	if err := store.CreateProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Get navigation options
	options, err := nav.GetNavigationOptions(project.ID)
	if err != nil {
		t.Errorf("failed to get navigation options: %v", err)
	}

	if options == nil {
		t.Fatal("expected navigation options, got nil")
	}

	if options.CurrentStage != state.StageDesign {
		t.Errorf("expected current stage design, got %s", options.CurrentStage)
	}

	// Should be able to go back to init and interview
	if len(options.CanGoBack) != 2 {
		t.Errorf("expected 2 backward options, got %d", len(options.CanGoBack))
	}

	// Should be able to go forward to plan
	if options.NextStage != state.StagePlan {
		t.Errorf("expected next stage plan, got %s", options.NextStage)
	}

	if len(options.CanGoForward) != 1 {
		t.Errorf("expected 1 forward option, got %d", len(options.CanGoForward))
	}
}

func TestDetermineArtifacts(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	defer os.Remove(tmpDB)

	store, err := state.NewStore(tmpDB)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	gitMgr := git.NewManager(".")
	nav := NewNavigator(store, gitMgr)

	// Create project
	project := &state.Project{
		ID:           "test-project-3",
		Name:         "Test Project",
		CreatedAt:    time.Now(),
		CurrentStage: state.StagePlan,
		CurrentPhase: "",
	}

	if err := store.CreateProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Add some artifacts
	interviewData := &state.InterviewData{
		ProjectID:        project.ID,
		ProjectName:      project.Name,
		CreatedAt:        time.Now(),
		ProblemStatement: "Test problem",
	}

	if err := store.SaveInterviewData(project.ID, interviewData); err != nil {
		t.Fatalf("failed to save interview data: %v", err)
	}

	// Test determining artifacts when going back
	result := &NavigationResult{
		PreservedWork:        []string{},
		RegeneratedArtifacts: []string{},
	}

	err = nav.determineArtifacts(project.ID, state.StagePlan, state.StageInterview, result)
	if err != nil {
		t.Errorf("failed to determine artifacts: %v", err)
	}

	// Should preserve interview data
	if len(result.PreservedWork) == 0 {
		t.Error("expected preserved work, got none")
	}

	// Should indicate devplan may need regeneration
	if len(result.RegeneratedArtifacts) == 0 {
		t.Error("expected regenerated artifacts, got none")
	}
}

func TestHistoryTracker(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	defer os.Remove(tmpDB)

	store, err := state.NewStore(tmpDB)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	gitMgr := git.NewManager(".")
	tracker := NewHistoryTracker(store, gitMgr)

	// Record a navigation event
	err = tracker.RecordNavigation("test-project", state.StageDesign, state.StageInterview)
	if err != nil {
		t.Errorf("failed to record navigation: %v", err)
	}

	// Get history (will be empty in this implementation, but should not error)
	history, err := tracker.GetNavigationHistory("test-project")
	if err != nil {
		t.Errorf("failed to get navigation history: %v", err)
	}

	// Should return empty slice without error
	if history == nil {
		t.Error("expected empty history slice, got nil")
	}
}

func TestCheckPrerequisites(t *testing.T) {
	tmpDB := t.TempDir() + "/test.db"
	defer os.Remove(tmpDB)

	store, err := state.NewStore(tmpDB)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	gitMgr := git.NewManager(".")
	nav := NewNavigator(store, gitMgr)

	testCases := []struct {
		name        string
		from        state.Stage
		to          state.Stage
		expectError bool
	}{
		{
			name:        "Interview to Design",
			from:        state.StageInterview,
			to:          state.StageDesign,
			expectError: false,
		},
		{
			name:        "Design to Plan",
			from:        state.StageDesign,
			to:          state.StagePlan,
			expectError: false,
		},
		{
			name:        "Plan to Review",
			from:        state.StagePlan,
			to:          state.StageReview,
			expectError: false,
		},
		{
			name:        "Review to Develop",
			from:        state.StageReview,
			to:          state.StageDevelop,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := nav.checkPrerequisites(tc.from, tc.to)
			if tc.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}
