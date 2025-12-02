package todo

import (
	"fmt"

	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/gitinfo"
	"github.com/tacit7/eye-in-the-sky/internal/todo/models"
	"github.com/tacit7/eye-in-the-sky/internal/todo/repository"
	"github.com/tacit7/eye-in-the-sky/internal/todo/util"
)

// Service provides high-level operations for the todo backend.
type Service struct {
	db        *database.DB
	projects  *repository.ProjectRepo
	tasks     *repository.TaskRepo
	notes     *repository.NoteRepo
}

// NewService creates a new TodoService.
func NewService(database *database.DB) *Service {
	return &Service{
		db:       database,
		projects: repository.NewProjectRepo(database),
		tasks:    repository.NewTaskRepo(database),
		notes:    repository.NewNoteRepo(database),
	}
}

// Close closes the service (note: does not close the database as it's managed externally).
func (s *Service) Close() error {
	// Database is managed externally; do nothing
	return nil
}

// Conn returns the underlying database connection for transaction management.
func (s *Service) Conn() *database.DB {
	return s.db
}

// GetDB returns the underlying database for direct access (for MCP handlers).
func (s *Service) GetDB() *database.DB {
	return s.db
}

// GetTasksRepo returns the task repository for direct access.
func (s *Service) GetTasksRepo() *repository.TaskRepo {
	return s.tasks
}

// GetProjectsRepo returns the project repository for direct access.
func (s *Service) GetProjectsRepo() *repository.ProjectRepo {
	return s.projects
}

// GetNotesRepo returns the notes repository for direct access.
func (s *Service) GetNotesRepo() *repository.NoteRepo {
	return s.notes
}

// ============================================================================
// Project Operations
// ============================================================================

// CreateProject creates a new project.
func (s *Service) CreateProject(name string, path *string, remoteURL *string) (*models.Project, error) {
	if err := util.ValidateProjectName(name); err != nil {
		return nil, err
	}
	return s.projects.CreateProject(name, path, remoteURL)
}

// GetProject retrieves a project by ID.
func (s *Service) GetProject(projectID int) (*models.Project, error) {
	return s.projects.GetProjectByID(projectID)
}

// ListProjects returns all active projects.
func (s *Service) ListProjects() ([]models.Project, error) {
	return s.projects.ListProjects()
}

// UpdateProject updates a project's metadata.
func (s *Service) UpdateProject(projectID int, updates map[string]interface{}) (*models.Project, error) {
	return s.projects.UpdateProject(projectID, updates)
}

// DetectCurrentProject auto-detects the current project based on git repository info.
// It tries to match by remote_url first (more reliable), then falls back to path matching.
func (s *Service) DetectCurrentProject() (*models.Project, error) {
	// Try to get git remote URL (more reliable as it's unique across clones)
	remoteURL, err := gitinfo.RemoteURL()
	if err == nil && remoteURL != "" {
		project, err := s.projects.GetProjectByRemoteURL(remoteURL)
		if err == nil {
			return project, nil
		}
		// If no match by remote URL, continue to try path
	}

	// Fallback to path matching
	repoPath, err := gitinfo.RepoPath()
	if err != nil {
		return nil, fmt.Errorf("not in a git repository or failed to detect repo info: %w", err)
	}

	if repoPath == "" {
		return nil, fmt.Errorf("could not determine git repository path")
	}

	project, err := s.projects.GetProjectByPath(repoPath)
	if err != nil {
		return nil, fmt.Errorf("no project found for current directory (path: %s): %w", repoPath, err)
	}

	return project, nil
}

// ============================================================================
// Workflow Operations
// ============================================================================

// GetWorkflow returns all global workflow states.
func (s *Service) GetWorkflow() ([]models.WorkflowState, error) {
	return s.projects.GetWorkflowStates()
}

// ============================================================================
// Task Operations
// ============================================================================

// CreateTask creates a new task in a project.
func (s *Service) CreateTask(projectID int, title string) (*models.Task, error) {
	if err := util.ValidateDescription(title); err != nil {
		return nil, err
	}

	input := models.CreateTaskInput{
		Title: title,
	}

	return s.tasks.CreateTask(projectID, input)
}

// GetTask retrieves a task by ID.
func (s *Service) GetTask(taskID string) (*models.Task, error) {
	return s.tasks.FindByID(taskID)
}

// ListTasks retrieves tasks for a project with optional filters.
func (s *Service) ListTasks(projectID int, filters *models.Filters) ([]models.Task, error) {
	if filters == nil {
		filters = &models.Filters{IsActive: true}
	}
	return s.tasks.List(projectID, *filters, models.SortByCreated)
}

// ListTasksWithSort retrieves tasks for a project sorted by a specific field.
func (s *Service) ListTasksWithSort(projectID int, filters models.Filters, sortBy models.SortOrder) ([]models.Task, error) {
	return s.tasks.List(projectID, filters, sortBy)
}

// SetTaskState changes a task's workflow state.
func (s *Service) SetTaskState(taskID string, stateID int) (*models.Task, error) {
	return s.tasks.MoveToState(taskID, stateID)
}

// HardDeleteTask permanently removes a task from the database.
func (s *Service) HardDeleteTask(taskID string) error {
	return s.tasks.HardDelete(taskID)
}

// SearchTasks performs full-text search on tasks.
func (s *Service) SearchTasks(projectID int, query string, limit, offset int) ([]models.SearchResult, error) {
	if err := util.ValidateSearchQuery(query); err != nil {
		return nil, err
	}
	if err := util.ValidatePaginationParams(limit, offset); err != nil {
		return nil, err
	}
	return s.tasks.Search(projectID, query, limit, offset)
}

// ============================================================================
// Tag Operations
// ============================================================================

// AddTaskTag adds a tag to a task.
func (s *Service) AddTaskTag(taskID string, tagName string) (*models.Tag, error) {
	if err := util.ValidateTagName(tagName); err != nil {
		return nil, err
	}
	return s.tasks.AddTag(taskID, tagName)
}

// RemoveTaskTag removes a tag from a task.
func (s *Service) RemoveTaskTag(taskID string, tagName string) error {
	return s.tasks.RemoveTag(taskID, tagName)
}

// ============================================================================
// Note Operations
// ============================================================================

// AddTaskNote adds a note to a task.
func (s *Service) AddTaskNote(taskID string, body string) (*models.Note, error) {
	if err := util.ValidateNoteContent(body); err != nil {
		return nil, err
	}
	return s.tasks.AddNote(taskID, body)
}

// GetNotesByTask retrieves all notes for a task.
func (s *Service) GetNotesByTask(taskID string) ([]models.Note, error) {
	return s.notes.GetNotesByTaskID(taskID)
}

// GetLatestNoteForTask retrieves the most recent note for a task.
func (s *Service) GetLatestNoteForTask(taskID string) (*models.Note, error) {
	return s.notes.GetLatestNoteByTaskID(taskID)
}

// UpdateNote updates a note's content.
func (s *Service) UpdateNote(noteID int, body string) (*models.Note, error) {
	if err := util.ValidateNoteContent(body); err != nil {
		return nil, err
	}
	return s.notes.UpdateNoteContent(noteID, body)
}

// DeleteNote removes a note.
func (s *Service) DeleteNote(noteID int) error {
	return s.notes.DeleteNote(noteID)
}

// GetProjectNoteStats returns note statistics for a project.
func (s *Service) GetProjectNoteStats(projectID int) (*repository.NoteStats, error) {
	return s.notes.GetProjectNoteStats(projectID)
}

// ============================================================================
// Database Operations
// ============================================================================

// Reindex rebuilds the FTS5 search index.
func (s *Service) Reindex() error {
	return s.db.Reindex()
}

// Vacuum performs database cleanup.
func (s *Service) Vacuum() error {
	return s.db.Vacuum()
}

// CheckAndMaybeReindex checks if a weekly reindex is needed.
func (s *Service) CheckAndMaybeReindex() error {
	return s.db.MaybeReindex()
}
