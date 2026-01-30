package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mojomast/geoffrussy/internal/config"
	"github.com/mojomast/geoffrussy/internal/state"
	"github.com/mojomast/geoffrussy/internal/token"
)

// ResourceHandlers contains all MCP resource handlers for Geoffrey
type ResourceHandlers struct {
	configManager *config.Manager
	projectRoot   string
}

// NewResourceHandlers creates a new resource handlers instance
func NewResourceHandlers(configManager *config.Manager, projectRoot string) *ResourceHandlers {
	return &ResourceHandlers{
		configManager: configManager,
		projectRoot:   projectRoot,
	}
}

// RegisterAllResources registers all Geoffrey resources with the registry
func (h *ResourceHandlers) RegisterAllResources(registry *ResourceRegistry) error {
	resources := []struct {
		resource Resource
		handler  ResourceHandler
	}{
		{h.statusResource(), h.handleStatusResource},
		{h.architectureResource(), h.handleArchitectureResource},
		{h.devplanResource(), h.handleDevPlanResource},
		{h.phasesResource(), h.handlePhasesResource},
		{h.interviewResource(), h.handleInterviewResource},
		{h.checkpointsResource(), h.handleCheckpointsResource},
		{h.statsResource(), h.handleStatsResource},
	}

	for _, r := range resources {
		if err := registry.RegisterResource(r.resource, r.handler); err != nil {
			return fmt.Errorf("failed to register resource %s: %w", r.resource.URI, err)
		}
	}

	return nil
}

// Resource definitions

func (h *ResourceHandlers) statusResource() Resource {
	return Resource{
		URI:         "project://status",
		Name:        "Project Status",
		Description: "Current project status, stage, and progress information",
		MimeType:    "application/json",
	}
}

func (h *ResourceHandlers) architectureResource() Resource {
	return Resource{
		URI:         "project://architecture",
		Name:        "Architecture Document",
		Description: "Generated system architecture documentation",
		MimeType:    "text/markdown",
	}
}

func (h *ResourceHandlers) devplanResource() Resource {
	return Resource{
		URI:         "project://devplan",
		Name:        "Development Plan",
		Description: "Complete development plan with all phases and tasks",
		MimeType:    "application/json",
	}
}

func (h *ResourceHandlers) phasesResource() Resource {
	return Resource{
		URI:         "project://phases",
		Name:        "All Phases",
		Description: "List of all development phases with status",
		MimeType:    "application/json",
	}
}

func (h *ResourceHandlers) interviewResource() Resource {
	return Resource{
		URI:         "project://interview",
		Name:        "Interview Data",
		Description: "Collected requirements from interview process",
		MimeType:    "application/json",
	}
}

func (h *ResourceHandlers) checkpointsResource() Resource {
	return Resource{
		URI:         "project://checkpoints",
		Name:        "Checkpoints",
		Description: "List of all saved checkpoints",
		MimeType:    "application/json",
	}
}

func (h *ResourceHandlers) statsResource() Resource {
	return Resource{
		URI:         "project://stats",
		Name:        "Statistics",
		Description: "Token usage and cost statistics",
		MimeType:    "application/json",
	}
}

// Resource handlers

func (h *ResourceHandlers) handleStatusResource(ctx context.Context, uri string) (*ReadResourceResult, error) {
	store, projectID, err := h.getStore()
	if err != nil {
		return nil, err
	}
	defer store.Close()

	project, err := store.GetProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	progress, err := store.CalculateProgress(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate progress: %w", err)
	}

	status := map[string]interface{}{
		"projectId":            projectID,
		"projectName":          project.Name,
		"currentStage":         project.CurrentStage,
		"currentPhase":         project.CurrentPhase,
		"completionPercentage": progress.CompletionPercentage,
		"totalTasks":           progress.TotalTasks,
		"completedTasks":       progress.CompletedTasks,
		"inProgressTasks":      progress.InProgressTasks,
		"blockedTasks":         progress.BlockedTasks,
		"totalPhases":          progress.TotalPhases,
		"completedPhases":      progress.CompletedPhases,
		"inProgressPhases":     progress.InProgressPhases,
		"blockedPhases":        progress.BlockedPhases,
	}

	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal status: %w", err)
	}

	return &ReadResourceResult{
		Contents: []Content{
			{
				Type:     "text",
				Text:     string(data),
				MimeType: "application/json",
			},
		},
	}, nil
}

func (h *ResourceHandlers) handleArchitectureResource(ctx context.Context, uri string) (*ReadResourceResult, error) {
	store, projectID, err := h.getStore()
	if err != nil {
		return nil, err
	}
	defer store.Close()

	arch, err := store.GetArchitecture(projectID)
	if err != nil {
		return nil, fmt.Errorf("architecture not found: %w", err)
	}

	return &ReadResourceResult{
		Contents: []Content{
			{
				Type:     "text",
				Text:     arch.Content,
				MimeType: "text/markdown",
			},
		},
	}, nil
}

func (h *ResourceHandlers) handleDevPlanResource(ctx context.Context, uri string) (*ReadResourceResult, error) {
	store, projectID, err := h.getStore()
	if err != nil {
		return nil, err
	}
	defer store.Close()

	phases, err := store.ListPhases(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list phases: %w", err)
	}

	// Build dev plan structure
	devPlan := map[string]interface{}{
		"projectId":   projectID,
		"totalPhases": len(phases),
		"phases":      phases,
	}

	data, err := json.MarshalIndent(devPlan, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dev plan: %w", err)
	}

	return &ReadResourceResult{
		Contents: []Content{
			{
				Type:     "text",
				Text:     string(data),
				MimeType: "application/json",
			},
		},
	}, nil
}

func (h *ResourceHandlers) handlePhasesResource(ctx context.Context, uri string) (*ReadResourceResult, error) {
	store, projectID, err := h.getStore()
	if err != nil {
		return nil, err
	}
	defer store.Close()

	phases, err := store.ListPhases(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list phases: %w", err)
	}

	data, err := json.MarshalIndent(phases, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal phases: %w", err)
	}

	return &ReadResourceResult{
		Contents: []Content{
			{
				Type:     "text",
				Text:     string(data),
				MimeType: "application/json",
			},
		},
	}, nil
}

func (h *ResourceHandlers) handleInterviewResource(ctx context.Context, uri string) (*ReadResourceResult, error) {
	store, projectID, err := h.getStore()
	if err != nil {
		return nil, err
	}
	defer store.Close()

	interview, err := store.GetInterviewData(projectID)
	if err != nil {
		return nil, fmt.Errorf("interview data not found: %w", err)
	}

	data, err := json.MarshalIndent(interview, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal interview data: %w", err)
	}

	return &ReadResourceResult{
		Contents: []Content{
			{
				Type:     "text",
				Text:     string(data),
				MimeType: "application/json",
			},
		},
	}, nil
}

func (h *ResourceHandlers) handleCheckpointsResource(ctx context.Context, uri string) (*ReadResourceResult, error) {
	store, projectID, err := h.getStore()
	if err != nil {
		return nil, err
	}
	defer store.Close()

	checkpoints, err := store.ListCheckpoints(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list checkpoints: %w", err)
	}

	data, err := json.MarshalIndent(checkpoints, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal checkpoints: %w", err)
	}

	return &ReadResourceResult{
		Contents: []Content{
			{
				Type:     "text",
				Text:     string(data),
				MimeType: "application/json",
			},
		},
	}, nil
}

func (h *ResourceHandlers) handleStatsResource(ctx context.Context, uri string) (*ReadResourceResult, error) {
	store, projectID, err := h.getStore()
	if err != nil {
		return nil, err
	}
	defer store.Close()

	counter := token.NewCounter(store)
	tokenStats, err := counter.GetTotalTokens(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token stats: %w", err)
	}

	costEstimator := token.NewCostEstimator(store)
	costStats, err := costEstimator.GetCostStats(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost stats: %w", err)
	}

	stats := map[string]interface{}{
		"totalCost":      costStats.TotalCost,
		"totalInput":     tokenStats.TotalInput,
		"totalOutput":    tokenStats.TotalOutput,
		"byProvider":     tokenStats.ByProvider,
		"byPhase":        tokenStats.ByPhase,
		"costByProvider": costStats.ByProvider,
		"costByPhase":    costStats.ByPhase,
		"lastUpdated":    tokenStats.LastUpdated,
	}

	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stats: %w", err)
	}

	return &ReadResourceResult{
		Contents: []Content{
			{
				Type:     "text",
				Text:     string(data),
				MimeType: "application/json",
			},
		},
	}, nil
}

// Helper method to get store and project ID
func (h *ResourceHandlers) getStore() (*state.Store, string, error) {
	// For resources, we use the configured project root
	projectPath := h.projectRoot
	if projectPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, "", fmt.Errorf("failed to get current directory: %w", err)
		}
		projectPath = cwd
	}

	dbPath := filepath.Join(projectPath, ".geoffrussy", "state.db")
	store, err := state.NewStore(dbPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open state store: %w", err)
	}

	projectID := filepath.Base(projectPath)
	return store, projectID, nil
}
