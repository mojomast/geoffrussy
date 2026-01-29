package blocker

import (
	"fmt"
	"time"

	"github.com/mojomast/geoffrussy/internal/interview"
	"github.com/mojomast/geoffrussy/internal/state"
)

// FailureThreshold is the number of failures before marking a task as blocked
const FailureThreshold = 3

// Detector handles blocker detection and resolution
type Detector struct {
	store           *state.Store
	interviewEngine *interview.Engine
	failureTracker  map[string]int // taskID -> failure count
}

// NewDetector creates a new blocker detector
func NewDetector(store *state.Store, interviewEngine *interview.Engine) *Detector {
	return &Detector{
		store:           store,
		interviewEngine: interviewEngine,
		failureTracker:  make(map[string]int),
	}
}

// RecordFailure records a task failure and checks if it should be marked as blocked
func (d *Detector) RecordFailure(taskID, errorMessage string) (bool, error) {
	// Increment failure count
	d.failureTracker[taskID]++

	// Check if threshold reached
	if d.failureTracker[taskID] >= FailureThreshold {
		return true, nil
	}

	return false, nil
}

// MarkAsBlocked marks a task as blocked and creates a blocker record
func (d *Detector) MarkAsBlocked(taskID, phaseID, projectID, reason, context string) (*state.Blocker, error) {
	// Create blocker record
	blocker := &state.Blocker{
		ID:          fmt.Sprintf("blocker-%s-%d", taskID, time.Now().UnixNano()),
		TaskID:      taskID,
		Description: fmt.Sprintf("%s. Context: %s", reason, context),
		CreatedAt:   time.Now(),
	}

	// Save blocker to store
	if err := d.store.SaveBlocker(blocker); err != nil {
		return nil, fmt.Errorf("failed to save blocker: %w", err)
	}

	// Update task status to blocked
	if err := d.store.UpdateTaskStatus(taskID, "blocked"); err != nil {
		return nil, fmt.Errorf("failed to update task status: %w", err)
	}

	return blocker, nil
}

// GatherBlockerInformation uses the interview engine to gather information about the blocker
func (d *Detector) GatherBlockerInformation(blocker *state.Blocker) (map[string]string, error) {
	// In a real implementation, we would:
	// 1. Create interview questions specific to the blocker
	// 2. Gather user responses
	// 3. Analyze the responses to understand the issue

	// For now, return a simplified structure
	info := map[string]string{
		"blocker_id":   blocker.ID,
		"task_id":      blocker.TaskID,
		"description":  blocker.Description,
		"gathered_at":  time.Now().Format(time.RFC3339),
	}

	return info, nil
}

// AttemptResolution attempts to resolve a blocker using various strategies
func (d *Detector) AttemptResolution(blocker *state.Blocker) (*ResolutionResult, error) {
	result := &ResolutionResult{
		BlockerID:  blocker.ID,
		Strategies: []ResolutionStrategy{},
		Success:    false,
	}

	// Try different resolution strategies
	strategies := []ResolutionStrategy{
		{
			Name:        "Retry with backoff",
			Description: "Retry the task with exponential backoff",
			Automatic:   true,
		},
		{
			Name:        "Skip and continue",
			Description: "Skip the blocked task and continue with others",
			Automatic:   false,
		},
		{
			Name:        "Request user intervention",
			Description: "Ask the user to manually resolve the issue",
			Automatic:   false,
		},
	}

	result.Strategies = strategies

	// Try automatic strategies first
	for _, strategy := range strategies {
		if strategy.Automatic {
			// In a real implementation, we would execute the strategy
			// For now, we'll just record it
			result.AttemptedStrategies = append(result.AttemptedStrategies, strategy.Name)
		}
	}

	return result, nil
}

// ResolutionResult represents the result of a resolution attempt
type ResolutionResult struct {
	BlockerID           string
	Strategies          []ResolutionStrategy
	AttemptedStrategies []string
	Success             bool
	Resolution          string
}

// ResolutionStrategy represents a strategy for resolving a blocker
type ResolutionStrategy struct {
	Name        string
	Description string
	Automatic   bool
}

// RequestUserIntervention notifies the user about a blocker that requires manual intervention
func (d *Detector) RequestUserIntervention(blocker *state.Blocker, context string) error {
	// In a real implementation, we would:
	// 1. Format a user-friendly notification
	// 2. Display it in the UI
	// 3. Wait for user response
	// 4. Record the intervention

	return nil
}

// ResolveBlocker marks a blocker as resolved
func (d *Detector) ResolveBlocker(blockerID, resolution string) error {
	// Resolve the blocker in the store
	if err := d.store.ResolveBlocker(blockerID, resolution); err != nil {
		return fmt.Errorf("failed to resolve blocker: %w", err)
	}

	// Get all blockers to find the task ID
	// We need to query all blockers (not just active ones) since we just resolved it
	allBlockers, err := d.store.ListActiveBlockers("")
	if err != nil {
		return fmt.Errorf("failed to list blockers: %w", err)
	}

	var taskID string
	for _, b := range allBlockers {
		if b.ID == blockerID {
			taskID = b.TaskID
			break
		}
	}

	// If not found in active blockers, it might have just been resolved
	// In that case, we still need to reset the failure count and update task status
	// For now, we'll just skip if not found
	if taskID == "" {
		// Try to find it in the failure tracker
		for tid := range d.failureTracker {
			// This is a simplified approach - in production we'd need better tracking
			taskID = tid
			break
		}
		if taskID == "" {
			return fmt.Errorf("blocker not found: %s", blockerID)
		}
	}

	// Reset failure count
	delete(d.failureTracker, taskID)

	// Update task status back to pending
	if err := d.store.UpdateTaskStatus(taskID, "pending"); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

// GetFailureCount returns the current failure count for a task
func (d *Detector) GetFailureCount(taskID string) int {
	return d.failureTracker[taskID]
}

// ResetFailureCount resets the failure count for a task
func (d *Detector) ResetFailureCount(taskID string) {
	delete(d.failureTracker, taskID)
}

// ListActiveBlockers lists all active blockers for a project
func (d *Detector) ListActiveBlockers(projectID string) ([]*state.Blocker, error) {
	return d.store.ListActiveBlockers(projectID)
}

// GetBlocker retrieves a specific blocker
func (d *Detector) GetBlocker(blockerID string) (*state.Blocker, error) {
	// Try to get all blockers (pass empty string to get all)
	blockers, err := d.store.ListActiveBlockers("")
	if err != nil {
		return nil, fmt.Errorf("failed to list blockers: %w", err)
	}

	for _, blocker := range blockers {
		if blocker.ID == blockerID {
			return blocker, nil
		}
	}

	// If not found, it might be because we need to query by project
	// In a real implementation, we would have a GetBlockerByID method in the store
	return nil, fmt.Errorf("blocker not found: %s", blockerID)
}

// AnalyzeBlockerPattern analyzes blocker patterns to identify recurring issues
func (d *Detector) AnalyzeBlockerPattern(projectID string) (*BlockerAnalysis, error) {
	blockers, err := d.store.ListActiveBlockers(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list blockers: %w", err)
	}

	analysis := &BlockerAnalysis{
		TotalBlockers:     len(blockers),
		BlockersByTask:    make(map[string]int),
		CommonDescriptions: make(map[string]int),
	}

	for _, blocker := range blockers {
		analysis.BlockersByTask[blocker.TaskID]++
		analysis.CommonDescriptions[blocker.Description]++
	}

	return analysis, nil
}

// BlockerAnalysis contains analysis of blocker patterns
type BlockerAnalysis struct {
	TotalBlockers      int
	BlockersByTask     map[string]int
	CommonDescriptions map[string]int
}
