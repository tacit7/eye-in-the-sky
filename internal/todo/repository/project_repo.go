package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/todo/models"
)

// ProjectRepo handles project-related database operations.
type ProjectRepo struct {
	db *database.DB
}

// NewProjectRepo creates a new ProjectRepo.
func NewProjectRepo(database *database.DB) *ProjectRepo {
	return &ProjectRepo{db: database}
}

// CreateProject creates a new project.
func (pr *ProjectRepo) CreateProject(id, name string) (*models.Project, error) {
	_, err := pr.db.Exec(
		"INSERT INTO projects (id, name, id_algorithm, created_at, updated_at, active) VALUES (?, ?, ?, ?, ?, ?)",
		id, name, "uuidv5", time.Now(), time.Now(), 1,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert project: %w", err)
	}

	return pr.GetProjectByID(id)
}

// GetProjectByID retrieves a project by ID.
func (pr *ProjectRepo) GetProjectByID(id string) (*models.Project, error) {
	project := &models.Project{}
	err := pr.db.QueryRow(
		`SELECT id, name, path, remote_url, subpath, module, salt, id_algorithm, created_at, updated_at, last_commit, active
		 FROM projects WHERE id = ?`,
		id,
	).Scan(&project.ID, &project.Name, &project.Path, &project.RemoteURL, &project.Subpath, &project.Module, &project.Salt, &project.IDAlgorithm, &project.CreatedAt, &project.UpdatedAt, &project.LastCommit, &project.Active)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query project: %w", err)
	}

	return project, nil
}

// ListProjects returns all active projects.
func (pr *ProjectRepo) ListProjects() ([]models.Project, error) {
	rows, err := pr.db.Query(
		`SELECT id, name, path, remote_url, subpath, module, salt, id_algorithm, created_at, updated_at, last_commit, active
		 FROM projects WHERE active = 1 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		project := models.Project{}
		if err := rows.Scan(&project.ID, &project.Name, &project.Path, &project.RemoteURL, &project.Subpath, &project.Module, &project.Salt, &project.IDAlgorithm, &project.CreatedAt, &project.UpdatedAt, &project.LastCommit, &project.Active); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, rows.Err()
}

// UpdateProject updates a project's metadata.
func (pr *ProjectRepo) UpdateProject(id string, updates map[string]interface{}) (*models.Project, error) {
	// Build dynamic UPDATE query
	query := "UPDATE projects SET updated_at = ?"
	args := []interface{}{time.Now()}

	if name, ok := updates["name"].(string); ok {
		query += ", name = ?"
		args = append(args, name)
	}
	if path, ok := updates["path"].(string); ok {
		query += ", path = ?"
		args = append(args, path)
	}
	if remoteURL, ok := updates["remote_url"].(string); ok {
		query += ", remote_url = ?"
		args = append(args, remoteURL)
	}
	if lastCommit, ok := updates["last_commit"].(string); ok {
		query += ", last_commit = ?"
		args = append(args, lastCommit)
	}

	query += " WHERE id = ?"
	args = append(args, id)

	_, err := pr.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return pr.GetProjectByID(id)
}

// GetWorkflowStates returns all workflow states (global, not per-project).
func (pr *ProjectRepo) GetWorkflowStates() ([]models.WorkflowState, error) {
	rows, err := pr.db.Query(
		"SELECT id, name, position, color FROM workflow_states ORDER BY position ASC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow states: %w", err)
	}
	defer rows.Close()

	var states []models.WorkflowState
	for rows.Next() {
		state := models.WorkflowState{}
		if err := rows.Scan(&state.ID, &state.Name, &state.Position, &state.Color); err != nil {
			return nil, fmt.Errorf("failed to scan workflow state: %w", err)
		}
		states = append(states, state)
	}

	return states, rows.Err()
}
