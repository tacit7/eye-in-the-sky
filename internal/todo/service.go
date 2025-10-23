package todo

import (
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/todo/db"
	"github.com/tacit7/eye-in-the-sky/internal/todo/models"
	"github.com/tacit7/eye-in-the-sky/internal/todo/repository"
	"github.com/tacit7/eye-in-the-sky/internal/todo/util"
)

// Service provides high-level operations for the todo backend.
type Service struct {
	db        *db.DB
	projects  *repository.ProjectRepo
	tasks     *repository.TaskRepo
	notes     *repository.NoteRepo
}

// NewService creates a new TodoService.
func NewService(database *db.DB) *Service {
	return &Service{
		db:       database,
		projects: repository.NewProjectRepo(database),
		tasks:    repository.NewTaskRepo(database),
		notes:    repository.NewNoteRepo(database),
	}
}

// Close closes the service and database connection.
func (s *Service) Close() error {
	return s.db.Close()
}

// ============================================================================
// Project Operations
// ============================================================================

// CreateProject creates a new project.
func (s *Service) CreateProject(uuid, name string) (*models.Project, error) {
	if err := util.ValidateProjectName(name); err != nil {
		return nil, err
	}
	return s.projects.CreateProject(uuid, name)
}

// GetProject retrieves a project by ID.
func (s *Service) GetProject(projectID int) (*models.Project, error) {
	return s.projects.GetProjectByID(projectID)
}

// ListProjects returns all active projects.
func (s *Service) ListProjects() ([]models.Project, error) {
	return s.projects.ListProjects()
}

// UpdateProject updates a project's name.
func (s *Service) UpdateProject(projectID int, name string) (*models.Project, error) {
	if err := util.ValidateProjectName(name); err != nil {
		return nil, err
	}
	return s.projects.UpdateProject(projectID, name)
}

// DeleteProject marks a project as archived.
func (s *Service) DeleteProject(projectID int) error {
	return s.projects.SoftDeleteProject(projectID)
}

// GetOrCreateProjectFromRepo finds or creates a project by git repo slug.
func (s *Service) GetOrCreateProjectFromRepo(repoSlug, uuid string) (*models.Project, error) {
	return s.projects.CreateOrGetProjectByGitRepo(repoSlug, uuid)
}

// ============================================================================
// Workflow Operations
// ============================================================================

// GetWorkflow returns the workflow states for a project.
func (s *Service) GetWorkflow(projectID int) ([]models.WorkflowState, error) {
	return s.projects.GetWorkflowStates(projectID)
}

// SyncWorkflowFromYAML syncs workflow states from parsed YAML.
func (s *Service) SyncWorkflowFromYAML(projectID int, yamlData []byte) error {
	wf, err := util.ParseWorkflowYAML(yamlData)
	if err != nil {
		return err
	}

	states := make([]struct {
		Code        string
		DisplayName string
	}, len(wf.Workflow))

	for i, state := range wf.Workflow {
		states[i].Code = state.Code
		states[i].DisplayName = state.Label
	}

	return s.projects.SyncWorkflowFromYAML(projectID, states)
}

// ============================================================================
// Task Operations
// ============================================================================

// CreateTask creates a new task in a project.
func (s *Service) CreateTask(projectID int, description string, parentID *int) (*models.Task, error) {
	if err := util.ValidateDescription(description); err != nil {
		return nil, err
	}

	input := models.CreateTaskInput{
		Description: description,
		ParentID:    parentID,
	}

	return s.tasks.CreateTask(projectID, input)
}

// GetTask retrieves a task by ID.
func (s *Service) GetTask(taskID int) (*models.Task, error) {
	return s.tasks.FindByID(taskID)
}

// ListTasks retrieves tasks for a project with optional filters.
func (s *Service) ListTasks(projectID int, filters *models.Filters) ([]models.Task, error) {
	if filters == nil {
		filters = &models.Filters{IsActive: true}
	}
	return s.tasks.List(projectID, *filters, models.SortByPosition)
}

// ListTasksWithSort retrieves tasks for a project sorted by a specific field.
func (s *Service) ListTasksWithSort(projectID int, filters models.Filters, sortBy models.SortOrder) ([]models.Task, error) {
	return s.tasks.List(projectID, filters, sortBy)
}

// UpdateTaskDescription updates a task's description.
func (s *Service) UpdateTaskDescription(taskID int, description string) (*models.Task, error) {
	if err := util.ValidateDescription(description); err != nil {
		return nil, err
	}
	return s.tasks.UpdateDescription(taskID, description)
}

// SetTaskState changes a task's workflow state.
func (s *Service) SetTaskState(taskID int, stateCode string) (*models.Task, error) {
	return s.tasks.MoveToState(taskID, stateCode)
}

// SetTaskPriority sets a task's priority (1-5).
func (s *Service) SetTaskPriority(taskID int, priority *int) (*models.Task, error) {
	if err := util.ValidatePriority(priority); err != nil {
		return nil, err
	}
	return s.tasks.SetPriority(taskID, priority)
}

// SetTaskWeight sets a task's weight.
func (s *Service) SetTaskWeight(taskID int, weight *int) (*models.Task, error) {
	if err := util.ValidateWeight(weight); err != nil {
		return nil, err
	}
	return s.tasks.SetWeight(taskID, weight)
}

// SetTaskDueDate sets a task's due date.
func (s *Service) SetTaskDueDate(taskID int, dueDate *time.Time) (*models.Task, error) {
	if err := util.ValidateDueDate(dueDate); err != nil {
		return nil, err
	}
	return s.tasks.SetDue(taskID, dueDate)
}

// SetTaskParent sets a task's parent task.
func (s *Service) SetTaskParent(taskID int, parentID *int) (*models.Task, error) {
	if err := util.ValidateParentID(taskID, parentID); err != nil {
		return nil, err
	}
	return s.tasks.SetParent(taskID, parentID)
}

// ReorderTask changes a task's position.
func (s *Service) ReorderTask(taskID int, newPosition int) (*models.Task, error) {
	return s.tasks.Reorder(taskID, newPosition)
}

// DeleteTask marks a task and its children as archived.
func (s *Service) DeleteTask(taskID int) error {
	return s.tasks.Delete(taskID)
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
func (s *Service) AddTaskTag(taskID int, tagName string) (*models.Tag, error) {
	if err := util.ValidateTagName(tagName); err != nil {
		return nil, err
	}
	return s.tasks.AddTag(taskID, tagName)
}

// RemoveTaskTag removes a tag from a task.
func (s *Service) RemoveTaskTag(taskID int, tagName string) error {
	return s.tasks.RemoveTag(taskID, tagName)
}

// ============================================================================
// Note Operations
// ============================================================================

// AddTaskNote adds a markdown note to a task.
func (s *Service) AddTaskNote(taskID int, bodyMarkdown string) (*models.Note, error) {
	if err := util.ValidateNoteContent(bodyMarkdown); err != nil {
		return nil, err
	}
	return s.tasks.AddNote(taskID, bodyMarkdown)
}

// GetNotesByTask retrieves all notes for a task.
func (s *Service) GetNotesByTask(taskID int) ([]models.Note, error) {
	return s.notes.GetNotesByTaskID(taskID)
}

// GetLatestNoteForTask retrieves the most recent note for a task.
func (s *Service) GetLatestNoteForTask(taskID int) (*models.Note, error) {
	return s.notes.GetLatestNoteByTaskID(taskID)
}

// UpdateNote updates a note's content.
func (s *Service) UpdateNote(noteID int, bodyMarkdown string) (*models.Note, error) {
	if err := util.ValidateNoteContent(bodyMarkdown); err != nil {
		return nil, err
	}
	return s.notes.UpdateNoteContent(noteID, bodyMarkdown)
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
