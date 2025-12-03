package repository

import (
	"database/sql"
	"fmt"

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
// Note: eits.db uses INTEGER autoincrement for id, so we don't pass id parameter
func (pr *ProjectRepo) CreateProject(name string, path *string, remoteURL *string) (*models.Project, error) {
	result, err := pr.db.Exec(
		"INSERT INTO projects (name, path, remote_url, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		name, path, remoteURL, now(), now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert project: %w", err)
	}

	// Get the auto-generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return pr.GetProjectByID(int(id))
}

// GetProjectByID retrieves a project by ID.
func (pr *ProjectRepo) GetProjectByID(id int) (*models.Project, error) {
	project := &models.Project{}
	err := pr.db.QueryRow(
		`SELECT id, name, path, remote_url, created_at, updated_at
		 FROM projects WHERE id = ?`,
		id,
	).Scan(&project.ID, &project.Name, &project.Path, &project.RemoteURL, &project.CreatedAt, &project.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query project: %w", err)
	}

	return project, nil
}

// ListProjects returns all projects.
func (pr *ProjectRepo) ListProjects() ([]models.Project, error) {
	rows, err := pr.db.Query(
		`SELECT id, name, path, remote_url, created_at, updated_at
		 FROM projects ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		project := models.Project{}
		if err := rows.Scan(&project.ID, &project.Name, &project.Path, &project.RemoteURL, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, rows.Err()
}

// UpdateProject updates a project's metadata.
func (pr *ProjectRepo) UpdateProject(id int, updates map[string]interface{}) (*models.Project, error) {
	// Build dynamic UPDATE query
	query := "UPDATE projects SET updated_at = ?"
	args := []interface{}{now()}

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

	query += " WHERE id = ?"
	args = append(args, id)

	_, err := pr.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return pr.GetProjectByID(id)
}

// GetProjectByPath retrieves a project by its path.
func (pr *ProjectRepo) GetProjectByPath(path string) (*models.Project, error) {
	project := &models.Project{}
	err := pr.db.QueryRow(
		`SELECT id, name, path, remote_url, created_at, updated_at
		 FROM projects WHERE path = ?`,
		path,
	).Scan(&project.ID, &project.Name, &project.Path, &project.RemoteURL, &project.CreatedAt, &project.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found for path: %s", path)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query project by path: %w", err)
	}

	return project, nil
}

// GetProjectByRemoteURL retrieves a project by its remote URL.
func (pr *ProjectRepo) GetProjectByRemoteURL(remoteURL string) (*models.Project, error) {
	project := &models.Project{}
	err := pr.db.QueryRow(
		`SELECT id, name, path, remote_url, created_at, updated_at
		 FROM projects WHERE remote_url = ?`,
		remoteURL,
	).Scan(&project.ID, &project.Name, &project.Path, &project.RemoteURL, &project.CreatedAt, &project.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found for remote URL: %s", remoteURL)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query project by remote URL: %w", err)
	}

	return project, nil
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
