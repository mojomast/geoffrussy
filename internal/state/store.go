package state

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Store represents the state store
type Store struct {
	db               *sql.DB
	migrationManager *MigrationManager
}

// NewStore creates a new state store
func NewStore(dbPath string) (*Store, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}
	
	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	
	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}
	
	store := &Store{
		db:               db,
		migrationManager: NewMigrationManager(db),
	}
	
	// Run migrations
	if err := store.migrationManager.Migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	
	return store, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// DB returns the underlying database connection
func (s *Store) DB() *sql.DB {
	return s.db
}

// MigrationManager returns the migration manager
func (s *Store) MigrationManager() *MigrationManager {
	return s.migrationManager
}

// HealthCheck verifies the database is accessible and not corrupted
func (s *Store) HealthCheck() error {
	// Try a simple query
	var result int
	err := s.db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	
	if result != 1 {
		return fmt.Errorf("database health check returned unexpected result: %d", result)
	}
	
	// Check schema version
	version, err := s.migrationManager.CurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get schema version: %w", err)
	}
	
	if version == 0 {
		return fmt.Errorf("database schema not initialized")
	}
	
	return nil
}

// BeginTx starts a new transaction
func (s *Store) BeginTx() (*sql.Tx, error) {
	return s.db.Begin()
}

// Project operations

// CreateProject creates a new project
func (s *Store) CreateProject(project *Project) error {
	query := `
		INSERT INTO projects (id, name, created_at, current_stage, current_phase_id)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query,
		project.ID,
		project.Name,
		project.CreatedAt,
		project.CurrentStage,
		project.CurrentPhase,
	)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	return nil
}

// GetProject retrieves a project by ID
func (s *Store) GetProject(id string) (*Project, error) {
	query := `
		SELECT id, name, created_at, current_stage, current_phase_id
		FROM projects
		WHERE id = ?
	`
	var project Project
	err := s.db.QueryRow(query, id).Scan(
		&project.ID,
		&project.Name,
		&project.CreatedAt,
		&project.CurrentStage,
		&project.CurrentPhase,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return &project, nil
}

// UpdateProject updates an existing project
func (s *Store) UpdateProject(project *Project) error {
	query := `
		UPDATE projects
		SET name = ?, current_stage = ?, current_phase_id = ?
		WHERE id = ?
	`
	result, err := s.db.Exec(query,
		project.Name,
		project.CurrentStage,
		project.CurrentPhase,
		project.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("project not found: %s", project.ID)
	}
	
	return nil
}

// UpdateProjectStage updates the current stage of a project
func (s *Store) UpdateProjectStage(id string, stage Stage) error {
	query := `
		UPDATE projects
		SET current_stage = ?
		WHERE id = ?
	`
	result, err := s.db.Exec(query, stage, id)
	if err != nil {
		return fmt.Errorf("failed to update project stage: %w", err)
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("project not found: %s", id)
	}
	
	return nil
}

// Interview data operations

// SaveInterviewData saves interview data for a project
func (s *Store) SaveInterviewData(projectID string, data *InterviewData) error {
	// Convert data to JSON
	jsonData, err := marshalJSON(data)
	if err != nil {
		return fmt.Errorf("failed to marshal interview data: %w", err)
	}
	
	query := `
		INSERT INTO interview_data (project_id, data, completed_at)
		VALUES (?, ?, ?)
		ON CONFLICT(project_id) DO UPDATE SET
			data = excluded.data,
			completed_at = excluded.completed_at
	`
	_, err = s.db.Exec(query, projectID, jsonData, data.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save interview data: %w", err)
	}
	return nil
}

// GetInterviewData retrieves interview data for a project
func (s *Store) GetInterviewData(projectID string) (*InterviewData, error) {
	query := `
		SELECT data
		FROM interview_data
		WHERE project_id = ?
	`
	var jsonData string
	err := s.db.QueryRow(query, projectID).Scan(&jsonData)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("interview data not found for project: %s", projectID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get interview data: %w", err)
	}
	
	var data InterviewData
	if err := unmarshalJSON(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal interview data: %w", err)
	}
	
	return &data, nil
}

// Architecture operations

// SaveArchitecture saves architecture for a project
func (s *Store) SaveArchitecture(projectID string, arch *Architecture) error {
	query := `
		INSERT INTO architectures (project_id, content, created_at)
		VALUES (?, ?, ?)
		ON CONFLICT(project_id) DO UPDATE SET
			content = excluded.content,
			created_at = excluded.created_at
	`
	_, err := s.db.Exec(query, projectID, arch.Content, arch.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save architecture: %w", err)
	}
	return nil
}

// GetArchitecture retrieves architecture for a project
func (s *Store) GetArchitecture(projectID string) (*Architecture, error) {
	query := `
		SELECT project_id, content, created_at
		FROM architectures
		WHERE project_id = ?
	`
	var arch Architecture
	err := s.db.QueryRow(query, projectID).Scan(
		&arch.ProjectID,
		&arch.Content,
		&arch.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("architecture not found for project: %s", projectID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get architecture: %w", err)
	}
	return &arch, nil
}

// Phase operations

// SavePhase saves a phase
func (s *Store) SavePhase(phase *Phase) error {
	query := `
		INSERT INTO phases (id, project_id, number, title, content, status, created_at, started_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			number = excluded.number,
			title = excluded.title,
			content = excluded.content,
			status = excluded.status,
			started_at = excluded.started_at,
			completed_at = excluded.completed_at
	`
	_, err := s.db.Exec(query,
		phase.ID,
		phase.ProjectID,
		phase.Number,
		phase.Title,
		phase.Content,
		phase.Status,
		phase.CreatedAt,
		phase.StartedAt,
		phase.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save phase: %w", err)
	}
	return nil
}

// GetPhase retrieves a phase by ID
func (s *Store) GetPhase(id string) (*Phase, error) {
	query := `
		SELECT id, project_id, number, title, content, status, created_at, started_at, completed_at
		FROM phases
		WHERE id = ?
	`
	var phase Phase
	err := s.db.QueryRow(query, id).Scan(
		&phase.ID,
		&phase.ProjectID,
		&phase.Number,
		&phase.Title,
		&phase.Content,
		&phase.Status,
		&phase.CreatedAt,
		&phase.StartedAt,
		&phase.CompletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("phase not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get phase: %w", err)
	}
	return &phase, nil
}

// ListPhases retrieves all phases for a project
func (s *Store) ListPhases(projectID string) ([]*Phase, error) {
	query := `
		SELECT id, project_id, number, title, content, status, created_at, started_at, completed_at
		FROM phases
		WHERE project_id = ?
		ORDER BY number ASC
	`
	rows, err := s.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list phases: %w", err)
	}
	defer rows.Close()
	
	var phases []*Phase
	for rows.Next() {
		var phase Phase
		err := rows.Scan(
			&phase.ID,
			&phase.ProjectID,
			&phase.Number,
			&phase.Title,
			&phase.Content,
			&phase.Status,
			&phase.CreatedAt,
			&phase.StartedAt,
			&phase.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan phase: %w", err)
		}
		phases = append(phases, &phase)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating phases: %w", err)
	}
	
	return phases, nil
}

// UpdatePhaseStatus updates the status of a phase
func (s *Store) UpdatePhaseStatus(id string, status PhaseStatus) error {
	now := time.Now()
	var query string
	var args []interface{}
	
	switch status {
	case PhaseInProgress:
		query = `
			UPDATE phases
			SET status = ?, started_at = COALESCE(started_at, ?)
			WHERE id = ?
		`
		args = []interface{}{status, now, id}
	case PhaseCompleted:
		query = `
			UPDATE phases
			SET status = ?, completed_at = ?
			WHERE id = ?
		`
		args = []interface{}{status, now, id}
	default:
		query = `
			UPDATE phases
			SET status = ?
			WHERE id = ?
		`
		args = []interface{}{status, id}
	}
	
	result, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update phase status: %w", err)
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("phase not found: %s", id)
	}
	
	return nil
}

// Task operations

// SaveTask saves a task
func (s *Store) SaveTask(task *Task) error {
	query := `
		INSERT INTO tasks (id, phase_id, number, description, status, started_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			number = excluded.number,
			description = excluded.description,
			status = excluded.status,
			started_at = excluded.started_at,
			completed_at = excluded.completed_at
	`
	_, err := s.db.Exec(query,
		task.ID,
		task.PhaseID,
		task.Number,
		task.Description,
		task.Status,
		task.StartedAt,
		task.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}
	return nil
}

// GetTask retrieves a task by ID
func (s *Store) GetTask(id string) (*Task, error) {
	query := `
		SELECT id, phase_id, number, description, status, started_at, completed_at
		FROM tasks
		WHERE id = ?
	`
	var task Task
	err := s.db.QueryRow(query, id).Scan(
		&task.ID,
		&task.PhaseID,
		&task.Number,
		&task.Description,
		&task.Status,
		&task.StartedAt,
		&task.CompletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

// UpdateTaskStatus updates the status of a task
func (s *Store) UpdateTaskStatus(id string, status TaskStatus) error {
	now := time.Now()
	var query string
	var args []interface{}
	
	switch status {
	case TaskInProgress:
		query = `
			UPDATE tasks
			SET status = ?, started_at = COALESCE(started_at, ?)
			WHERE id = ?
		`
		args = []interface{}{status, now, id}
	case TaskCompleted:
		query = `
			UPDATE tasks
			SET status = ?, completed_at = ?
			WHERE id = ?
		`
		args = []interface{}{status, now, id}
	default:
		query = `
			UPDATE tasks
			SET status = ?
			WHERE id = ?
		`
		args = []interface{}{status, id}
	}
	
	result, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("task not found: %s", id)
	}
	
	return nil
}

// ListTasks retrieves all tasks for a phase
func (s *Store) ListTasks(phaseID string) ([]Task, error) {
	query := `
		SELECT id, phase_id, number, description, status, started_at, completed_at
		FROM tasks
		WHERE phase_id = ?
		ORDER BY number
	`
	rows, err := s.db.Query(query, phaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(
			&task.ID,
			&task.PhaseID,
			&task.Number,
			&task.Description,
			&task.Status,
			&task.StartedAt,
			&task.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tasks: %w", err)
	}

	return tasks, nil
}

// Helper functions for JSON marshaling

func marshalJSON(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func unmarshalJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

// Checkpoint operations

// SaveCheckpoint saves a checkpoint
func (s *Store) SaveCheckpoint(checkpoint *Checkpoint) error {
	// Convert metadata to JSON
	var metadataJSON string
	if checkpoint.Metadata != nil {
		jsonData, err := marshalJSON(checkpoint.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal checkpoint metadata: %w", err)
		}
		metadataJSON = jsonData
	}
	
	query := `
		INSERT INTO checkpoints (id, project_id, name, git_tag, created_at, metadata)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			git_tag = excluded.git_tag,
			metadata = excluded.metadata
	`
	_, err := s.db.Exec(query,
		checkpoint.ID,
		checkpoint.ProjectID,
		checkpoint.Name,
		checkpoint.GitTag,
		checkpoint.CreatedAt,
		metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to save checkpoint: %w", err)
	}
	return nil
}

// GetCheckpoint retrieves a checkpoint by ID
func (s *Store) GetCheckpoint(id string) (*Checkpoint, error) {
	query := `
		SELECT id, project_id, name, git_tag, created_at, metadata
		FROM checkpoints
		WHERE id = ?
	`
	var checkpoint Checkpoint
	var metadataJSON sql.NullString
	
	err := s.db.QueryRow(query, id).Scan(
		&checkpoint.ID,
		&checkpoint.ProjectID,
		&checkpoint.Name,
		&checkpoint.GitTag,
		&checkpoint.CreatedAt,
		&metadataJSON,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("checkpoint not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint: %w", err)
	}
	
	// Unmarshal metadata if present
	if metadataJSON.Valid && metadataJSON.String != "" {
		var metadata map[string]string
		if err := unmarshalJSON(metadataJSON.String, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal checkpoint metadata: %w", err)
		}
		checkpoint.Metadata = metadata
	}
	
	return &checkpoint, nil
}

// ListCheckpoints retrieves all checkpoints for a project
func (s *Store) ListCheckpoints(projectID string) ([]*Checkpoint, error) {
	query := `
		SELECT id, project_id, name, git_tag, created_at, metadata
		FROM checkpoints
		WHERE project_id = ?
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list checkpoints: %w", err)
	}
	defer rows.Close()
	
	var checkpoints []*Checkpoint
	for rows.Next() {
		var checkpoint Checkpoint
		var metadataJSON sql.NullString
		
		err := rows.Scan(
			&checkpoint.ID,
			&checkpoint.ProjectID,
			&checkpoint.Name,
			&checkpoint.GitTag,
			&checkpoint.CreatedAt,
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan checkpoint: %w", err)
		}
		
		// Unmarshal metadata if present
		if metadataJSON.Valid && metadataJSON.String != "" {
			var metadata map[string]string
			if err := unmarshalJSON(metadataJSON.String, &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal checkpoint metadata: %w", err)
			}
			checkpoint.Metadata = metadata
		}
		
		checkpoints = append(checkpoints, &checkpoint)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating checkpoints: %w", err)
	}
	
	return checkpoints, nil
}

// Token usage operations

// RecordTokenUsage records token usage
func (s *Store) RecordTokenUsage(usage *TokenUsage) error {
	query := `
		INSERT INTO token_usage (project_id, phase_id, task_id, provider, model, tokens_input, tokens_output, cost, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	// Handle nullable phase_id and task_id
	var phaseID, taskID interface{}
	if usage.PhaseID != "" {
		phaseID = usage.PhaseID
	} else {
		phaseID = nil
	}
	if usage.TaskID != "" {
		taskID = usage.TaskID
	} else {
		taskID = nil
	}
	
	result, err := s.db.Exec(query,
		usage.ProjectID,
		phaseID,
		taskID,
		usage.Provider,
		usage.Model,
		usage.TokensInput,
		usage.TokensOutput,
		usage.Cost,
		usage.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to record token usage: %w", err)
	}
	
	// Get the auto-generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get token usage ID: %w", err)
	}
	usage.ID = int(id)
	
	return nil
}

// GetTotalCost retrieves the total cost for a project
func (s *Store) GetTotalCost(projectID string) (float64, error) {
	query := `
		SELECT COALESCE(SUM(cost), 0)
		FROM token_usage
		WHERE project_id = ?
	`
	var totalCost float64
	err := s.db.QueryRow(query, projectID).Scan(&totalCost)
	if err != nil {
		return 0, fmt.Errorf("failed to get total cost: %w", err)
	}
	return totalCost, nil
}

// GetTokenStats retrieves token statistics for a project
func (s *Store) GetTokenStats(projectID string) (*TokenStats, error) {
	// Get total tokens
	query := `
		SELECT 
			COALESCE(SUM(tokens_input), 0) as total_input,
			COALESCE(SUM(tokens_output), 0) as total_output
		FROM token_usage
		WHERE project_id = ?
	`
	var stats TokenStats
	err := s.db.QueryRow(query, projectID).Scan(&stats.TotalInput, &stats.TotalOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to get token stats: %w", err)
	}
	
	// Get by provider
	stats.ByProvider = make(map[string]int)
	providerQuery := `
		SELECT provider, SUM(tokens_input + tokens_output) as total
		FROM token_usage
		WHERE project_id = ?
		GROUP BY provider
	`
	rows, err := s.db.Query(providerQuery, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider stats: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var provider string
		var total int
		if err := rows.Scan(&provider, &total); err != nil {
			return nil, fmt.Errorf("failed to scan provider stats: %w", err)
		}
		stats.ByProvider[provider] = total
	}
	
	// Get by phase
	stats.ByPhase = make(map[string]int)
	phaseQuery := `
		SELECT phase_id, SUM(tokens_input + tokens_output) as total
		FROM token_usage
		WHERE project_id = ? AND phase_id IS NOT NULL
		GROUP BY phase_id
	`
	rows, err = s.db.Query(phaseQuery, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get phase stats: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var phaseID string
		var total int
		if err := rows.Scan(&phaseID, &total); err != nil {
			return nil, fmt.Errorf("failed to scan phase stats: %w", err)
		}
		stats.ByPhase[phaseID] = total
	}
	
	stats.LastUpdated = time.Now()
	return &stats, nil
}

// CacheTokenStats caches token statistics for faster retrieval
func (s *Store) CacheTokenStats(projectID string, stats *TokenStats) error {
	byProviderJSON, err := marshalJSON(stats.ByProvider)
	if err != nil {
		return fmt.Errorf("failed to marshal provider stats: %w", err)
	}
	
	byPhaseJSON, err := marshalJSON(stats.ByPhase)
	if err != nil {
		return fmt.Errorf("failed to marshal phase stats: %w", err)
	}
	
	query := `
		INSERT OR REPLACE INTO token_stats_cache (project_id, total_input, total_output, by_provider, by_phase, last_updated)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = s.db.Exec(query,
		projectID,
		stats.TotalInput,
		stats.TotalOutput,
		byProviderJSON,
		byPhaseJSON,
		stats.LastUpdated,
	)
	if err != nil {
		return fmt.Errorf("failed to cache token stats: %w", err)
	}
	
	return nil
}

// GetCachedTokenStats retrieves cached token statistics
func (s *Store) GetCachedTokenStats(projectID string) (*TokenStats, error) {
	query := `
		SELECT total_input, total_output, by_provider, by_phase, last_updated
		FROM token_stats_cache
		WHERE project_id = ?
	`
	var stats TokenStats
	var byProviderJSON, byPhaseJSON string
	
	err := s.db.QueryRow(query, projectID).Scan(
		&stats.TotalInput,
		&stats.TotalOutput,
		&byProviderJSON,
		&byPhaseJSON,
		&stats.LastUpdated,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cached stats not found for project: %s", projectID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cached token stats: %w", err)
	}
	
	if err := unmarshalJSON(byProviderJSON, &stats.ByProvider); err != nil {
		return nil, fmt.Errorf("failed to unmarshal provider stats: %w", err)
	}
	
	if err := unmarshalJSON(byPhaseJSON, &stats.ByPhase); err != nil {
		return nil, fmt.Errorf("failed to unmarshal phase stats: %w", err)
	}
	
	return &stats, nil
}

// InvalidateTokenStatsCache removes cached token statistics
func (s *Store) InvalidateTokenStatsCache(projectID string) error {
	query := `DELETE FROM token_stats_cache WHERE project_id = ?`
	_, err := s.db.Exec(query, projectID)
	if err != nil {
		return fmt.Errorf("failed to invalidate token stats cache: %w", err)
	}
	return nil
}

// Cost statistics operations

// GetCostStats retrieves cost statistics for a project
func (s *Store) GetCostStats(projectID string) (*CostStats, error) {
	// Get total cost
	query := `
		SELECT COALESCE(SUM(cost), 0)
		FROM token_usage
		WHERE project_id = ?
	`
	var stats CostStats
	err := s.db.QueryRow(query, projectID).Scan(&stats.TotalCost)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost stats: %w", err)
	}
	
	// Get by provider
	stats.ByProvider = make(map[string]float64)
	providerQuery := `
		SELECT provider, SUM(cost) as total
		FROM token_usage
		WHERE project_id = ?
		GROUP BY provider
	`
	rows, err := s.db.Query(providerQuery, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider cost stats: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var provider string
		var total float64
		if err := rows.Scan(&provider, &total); err != nil {
			return nil, fmt.Errorf("failed to scan provider cost stats: %w", err)
		}
		stats.ByProvider[provider] = total
	}
	
	// Get by phase
	stats.ByPhase = make(map[string]float64)
	phaseQuery := `
		SELECT phase_id, SUM(cost) as total
		FROM token_usage
		WHERE project_id = ? AND phase_id IS NOT NULL
		GROUP BY phase_id
	`
	rows, err = s.db.Query(phaseQuery, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get phase cost stats: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var phaseID string
		var total float64
		if err := rows.Scan(&phaseID, &total); err != nil {
			return nil, fmt.Errorf("failed to scan phase cost stats: %w", err)
		}
		stats.ByPhase[phaseID] = total
	}
	
	stats.LastUpdated = time.Now()
	return &stats, nil
}

// GetMostExpensiveCalls retrieves the most expensive API calls
func (s *Store) GetMostExpensiveCalls(projectID string, limit int) ([]*TokenUsage, error) {
	query := `
		SELECT id, project_id, phase_id, task_id, provider, model, tokens_input, tokens_output, cost, timestamp
		FROM token_usage
		WHERE project_id = ?
		ORDER BY cost DESC
		LIMIT ?
	`
	rows, err := s.db.Query(query, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get most expensive calls: %w", err)
	}
	defer rows.Close()
	
	var calls []*TokenUsage
	for rows.Next() {
		var usage TokenUsage
		var phaseID, taskID sql.NullString
		
		err := rows.Scan(
			&usage.ID,
			&usage.ProjectID,
			&phaseID,
			&taskID,
			&usage.Provider,
			&usage.Model,
			&usage.TokensInput,
			&usage.TokensOutput,
			&usage.Cost,
			&usage.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan token usage: %w", err)
		}
		
		if phaseID.Valid {
			usage.PhaseID = phaseID.String
		}
		if taskID.Valid {
			usage.TaskID = taskID.String
		}
		
		calls = append(calls, &usage)
	}
	
	return calls, nil
}

// GetTokenUsageByTimeRange retrieves token usage within a time range
func (s *Store) GetTokenUsageByTimeRange(projectID string, startTime, endTime time.Time) ([]*TokenUsage, error) {
	query := `
		SELECT id, project_id, phase_id, task_id, provider, model, tokens_input, tokens_output, cost, timestamp
		FROM token_usage
		WHERE project_id = ? AND timestamp BETWEEN ? AND ?
		ORDER BY timestamp ASC
	`
	rows, err := s.db.Query(query, projectID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get token usage by time range: %w", err)
	}
	defer rows.Close()
	
	var usages []*TokenUsage
	for rows.Next() {
		var usage TokenUsage
		var phaseID, taskID sql.NullString
		
		err := rows.Scan(
			&usage.ID,
			&usage.ProjectID,
			&phaseID,
			&taskID,
			&usage.Provider,
			&usage.Model,
			&usage.TokensInput,
			&usage.TokensOutput,
			&usage.Cost,
			&usage.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan token usage: %w", err)
		}
		
		if phaseID.Valid {
			usage.PhaseID = phaseID.String
		}
		if taskID.Valid {
			usage.TaskID = taskID.String
		}
		
		usages = append(usages, &usage)
	}
	
	return usages, nil
}

// Rate limit operations

// SaveRateLimit saves rate limit information
func (s *Store) SaveRateLimit(provider string, info *RateLimitInfo) error {
	query := `
		INSERT INTO rate_limits (provider, requests_remaining, requests_limit, reset_at, checked_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query,
		provider,
		info.RequestsRemaining,
		info.RequestsLimit,
		info.ResetAt,
		info.CheckedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save rate limit: %w", err)
	}
	return nil
}

// GetRateLimit retrieves the most recent rate limit information for a provider
func (s *Store) GetRateLimit(provider string) (*RateLimitInfo, error) {
	query := `
		SELECT provider, requests_remaining, requests_limit, reset_at, checked_at
		FROM rate_limits
		WHERE provider = ?
		ORDER BY checked_at DESC
		LIMIT 1
	`
	var info RateLimitInfo
	err := s.db.QueryRow(query, provider).Scan(
		&info.Provider,
		&info.RequestsRemaining,
		&info.RequestsLimit,
		&info.ResetAt,
		&info.CheckedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("rate limit not found for provider: %s", provider)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit: %w", err)
	}
	return &info, nil
}

// Quota operations

// SaveQuota saves quota information
func (s *Store) SaveQuota(provider string, info *QuotaInfo) error {
	query := `
		INSERT INTO quotas (provider, tokens_remaining, tokens_limit, cost_remaining, cost_limit, reset_at, checked_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query,
		provider,
		info.TokensRemaining,
		info.TokensLimit,
		info.CostRemaining,
		info.CostLimit,
		info.ResetAt,
		info.CheckedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save quota: %w", err)
	}
	return nil
}

// GetQuota retrieves the most recent quota information for a provider
func (s *Store) GetQuota(provider string) (*QuotaInfo, error) {
	query := `
		SELECT provider, tokens_remaining, tokens_limit, cost_remaining, cost_limit, reset_at, checked_at
		FROM quotas
		WHERE provider = ?
		ORDER BY checked_at DESC
		LIMIT 1
	`
	var info QuotaInfo
	var tokensRemaining, tokensLimit sql.NullInt64
	var costRemaining, costLimit sql.NullFloat64
	
	err := s.db.QueryRow(query, provider).Scan(
		&info.Provider,
		&tokensRemaining,
		&tokensLimit,
		&costRemaining,
		&costLimit,
		&info.ResetAt,
		&info.CheckedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("quota not found for provider: %s", provider)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}
	
	// Convert nullable fields
	if tokensRemaining.Valid {
		val := int(tokensRemaining.Int64)
		info.TokensRemaining = &val
	}
	if tokensLimit.Valid {
		val := int(tokensLimit.Int64)
		info.TokensLimit = &val
	}
	if costRemaining.Valid {
		info.CostRemaining = &costRemaining.Float64
	}
	if costLimit.Valid {
		info.CostLimit = &costLimit.Float64
	}
	
	return &info, nil
}

// Blocker operations

// SaveBlocker saves a blocker
func (s *Store) SaveBlocker(blocker *Blocker) error {
	query := `
		INSERT INTO blockers (id, task_id, description, resolution, created_at, resolved_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			description = excluded.description,
			resolution = excluded.resolution,
			resolved_at = excluded.resolved_at
	`
	_, err := s.db.Exec(query,
		blocker.ID,
		blocker.TaskID,
		blocker.Description,
		blocker.Resolution,
		blocker.CreatedAt,
		blocker.ResolvedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save blocker: %w", err)
	}
	return nil
}

// ResolveBlocker marks a blocker as resolved
func (s *Store) ResolveBlocker(id string, resolution string) error {
	now := time.Now()
	query := `
		UPDATE blockers
		SET resolution = ?, resolved_at = ?
		WHERE id = ?
	`
	result, err := s.db.Exec(query, resolution, now, id)
	if err != nil {
		return fmt.Errorf("failed to resolve blocker: %w", err)
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("blocker not found: %s", id)
	}
	
	return nil
}

// ListActiveBlockers retrieves all active (unresolved) blockers for a project
func (s *Store) ListActiveBlockers(projectID string) ([]*Blocker, error) {
	query := `
		SELECT b.id, b.task_id, b.description, b.resolution, b.created_at, b.resolved_at
		FROM blockers b
		JOIN tasks t ON b.task_id = t.id
		JOIN phases p ON t.phase_id = p.id
		WHERE p.project_id = ? AND b.resolved_at IS NULL
		ORDER BY b.created_at DESC
	`
	rows, err := s.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list active blockers: %w", err)
	}
	defer rows.Close()
	
	var blockers []*Blocker
	for rows.Next() {
		var blocker Blocker
		var resolution sql.NullString
		
		err := rows.Scan(
			&blocker.ID,
			&blocker.TaskID,
			&blocker.Description,
			&resolution,
			&blocker.CreatedAt,
			&blocker.ResolvedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan blocker: %w", err)
		}
		
		if resolution.Valid {
			blocker.Resolution = resolution.String
		}
		
		blockers = append(blockers, &blocker)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating blockers: %w", err)
	}
	
	return blockers, nil
}

// Configuration operations

// SetConfig sets a configuration value
func (s *Store) SetConfig(key string, value string) error {
	query := `
		INSERT INTO config (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = excluded.updated_at
	`
	_, err := s.db.Exec(query, key, value, time.Now())
	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}
	return nil
}

// GetConfig retrieves a configuration value
func (s *Store) GetConfig(key string) (string, error) {
	query := `
		SELECT value
		FROM config
		WHERE key = ?
	`
	var value string
	err := s.db.QueryRow(query, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("config key not found: %s", key)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get config: %w", err)
	}
	return value, nil
}
