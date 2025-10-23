package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/todo/db"
	"github.com/tacit7/eye-in-the-sky/internal/todo/models"
)

// ProjectRepo handles project-related database operations.
type ProjectRepo struct {
	db *db.DB
}

// NewProjectRepo creates a new ProjectRepo.
func NewProjectRepo(database *db.DB) *ProjectRepo {
	return &ProjectRepo{db: database}
}

// CreateProject creates a new project.
func (pr *ProjectRepo) CreateProject(uuid, name string) (*models.Project, error) {
	result, err := pr.db.Exec(
		"INSERT INTO projects (uuid, name) VALUES (?, ?)",
		uuid, name,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert project: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	project := &models.Project{
		ID:        int(id),
		UUID:      uuid,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return project, nil
}

// CreateOrGetProjectByGitRepo finds or creates a project by git repo slug.
func (pr *ProjectRepo) CreateOrGetProjectByGitRepo(repoSlug, uuid string) (*models.Project, error) {
	// Try to find existing project
	project := &models.Project{}
	err := pr.db.QueryRow(
		"SELECT id, uuid, name, repo_slug, created_at, updated_at, archived_at FROM projects WHERE repo_slug = ?",
		repoSlug,
	).Scan(&project.ID, &project.UUID, &project.Name, &project.RepoSlug, &project.CreatedAt, &project.UpdatedAt, &project.ArchivedAt)

	if err == sql.ErrNoRows {
		// Create new project
		return pr.CreateProject(uuid, repoSlug)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query project: %w", err)
	}

	return project, nil
}

// GetProjectByID retrieves a project by ID.
func (pr *ProjectRepo) GetProjectByID(id int) (*models.Project, error) {
	project := &models.Project{}
	err := pr.db.QueryRow(
		"SELECT id, uuid, name, repo_slug, created_at, updated_at, archived_at FROM projects WHERE id = ?",
		id,
	).Scan(&project.ID, &project.UUID, &project.Name, &project.RepoSlug, &project.CreatedAt, &project.UpdatedAt, &project.ArchivedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query project: %w", err)
	}

	return project, nil
}

// ListProjects returns all non-archived projects.
func (pr *ProjectRepo) ListProjects() ([]models.Project, error) {
	rows, err := pr.db.Query(
		"SELECT id, uuid, name, repo_slug, created_at, updated_at, archived_at FROM projects WHERE archived_at IS NULL ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		project := models.Project{}
		if err := rows.Scan(&project.ID, &project.UUID, &project.Name, &project.RepoSlug, &project.CreatedAt, &project.UpdatedAt, &project.ArchivedAt); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, rows.Err()
}

// UpdateProject updates a project's name.
func (pr *ProjectRepo) UpdateProject(id int, name string) (*models.Project, error) {
	_, err := pr.db.Exec(
		"UPDATE projects SET name = ? WHERE id = ?",
		name, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return pr.GetProjectByID(id)
}

// SoftDeleteProject marks a project as archived.
func (pr *ProjectRepo) SoftDeleteProject(id int) error {
	_, err := pr.db.Exec(
		"UPDATE projects SET archived_at = CURRENT_TIMESTAMP WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to soft delete project: %w", err)
	}
	return nil
}

// SyncWorkflowFromYAML syncs workflow states from a YAML definition.
// WorkflowYAML should contain a slice of states with Code and DisplayName.
type WorkflowYAML struct {
	States []struct {
		Code        string `yaml:"code"`
		DisplayName string `yaml:"label"`
	} `yaml:"workflow"`
}

// SyncWorkflowFromYAML upserts workflow states for a project.
func (pr *ProjectRepo) SyncWorkflowFromYAML(projectID int, states []struct {
	Code        string
	DisplayName string
}) error {
	for i, state := range states {
		// Check if state exists
		var existingID int
		err := pr.db.QueryRow(
			"SELECT id FROM workflow_states WHERE project_id = ? AND code = ?",
			projectID, state.Code,
		).Scan(&existingID)

		if err == sql.ErrNoRows {
			// Insert new state
			_, err := pr.db.Exec(
				"INSERT INTO workflow_states (project_id, code, display_name, position) VALUES (?, ?, ?, ?)",
				projectID, state.Code, state.DisplayName, i,
			)
			if err != nil {
				return fmt.Errorf("failed to insert workflow state: %w", err)
			}
		} else if err == nil {
			// Update existing state
			_, err := pr.db.Exec(
				"UPDATE workflow_states SET display_name = ?, position = ? WHERE id = ?",
				state.DisplayName, i, existingID,
			)
			if err != nil {
				return fmt.Errorf("failed to update workflow state: %w", err)
			}
		} else {
			return fmt.Errorf("failed to query workflow state: %w", err)
		}
	}

	return nil
}

// GetWorkflowStates returns all states for a project.
func (pr *ProjectRepo) GetWorkflowStates(projectID int) ([]models.WorkflowState, error) {
	rows, err := pr.db.Query(
		"SELECT id, project_id, code, display_name, position, created_at, updated_at FROM workflow_states WHERE project_id = ? ORDER BY position ASC",
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow states: %w", err)
	}
	defer rows.Close()

	var states []models.WorkflowState
	for rows.Next() {
		state := models.WorkflowState{}
		if err := rows.Scan(&state.ID, &state.ProjectID, &state.Code, &state.DisplayName, &state.Position, &state.CreatedAt, &state.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan workflow state: %w", err)
		}
		states = append(states, state)
	}

	return states, rows.Err()
}
