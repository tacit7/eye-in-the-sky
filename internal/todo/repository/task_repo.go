package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/todo/models"
)

// TaskRepo handles task-related database operations.
type TaskRepo struct {
	db *database.DB
}

// NewTaskRepo creates a new TaskRepo.
func NewTaskRepo(database *database.DB) *TaskRepo {
	return &TaskRepo{db: database}
}

// CreateTask creates a new task in a project.
func (tr *TaskRepo) CreateTask(projectID string, input models.CreateTaskInput) (*models.Task, error) {
	// Validate input
	if strings.TrimSpace(input.Title) == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}

	// Generate UUID for task ID
	taskID := uuid.New().String()

	_, err := tr.db.Exec(
		`INSERT INTO tasks (id, project_id, title, description, state_id, priority, due_at, session_id, agent_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		taskID, projectID, input.Title, input.Description, input.StateID, input.Priority, input.DueAt, input.SessionID, input.AgentID, time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert task: %w", err)
	}

	return tr.FindByID(taskID)
}

// FindByID retrieves a task by ID with all related data.
func (tr *TaskRepo) FindByID(id string) (*models.Task, error) {
	task := &models.Task{}
	err := tr.db.QueryRow(
		`SELECT id, project_id, title, description, state_id, priority, due_at, completed_at, session_id, agent_id, created_at, updated_at, archived
		 FROM tasks WHERE id = ?`,
		id,
	).Scan(&task.ID, &task.ProjectID, &task.Title, &task.Description, &task.StateID, &task.Priority, &task.DueAt, &task.CompletedAt, &task.SessionID, &task.AgentID, &task.CreatedAt, &task.UpdatedAt, &task.Archived)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query task: %w", err)
	}

	// Load notes and tags
	if err := tr.loadTaskRelations(task); err != nil {
		return nil, err
	}

	return task, nil
}

// loadTaskRelations loads notes and tags for a task.
func (tr *TaskRepo) loadTaskRelations(task *models.Task) error {
	// Load notes
	rows, err := tr.db.Query(
		"SELECT id, task_id, author, body, created_at FROM task_notes WHERE task_id = ? ORDER BY created_at DESC",
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to query notes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		note := models.Note{}
		if err := rows.Scan(&note.ID, &note.TaskID, &note.Author, &note.Body, &note.CreatedAt); err != nil {
			return fmt.Errorf("failed to scan note: %w", err)
		}
		task.Notes = append(task.Notes, note)
	}

	// Load tags
	rows, err = tr.db.Query(
		`SELECT t.id, t.name, t.color FROM tags t
		 JOIN task_tags tt ON t.id = tt.tag_id
		 WHERE tt.task_id = ? ORDER BY t.name`,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		tag := models.Tag{}
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Color); err != nil {
			return fmt.Errorf("failed to scan tag: %w", err)
		}
		task.Tags = append(task.Tags, tag)
	}

	return nil
}

// Update updates a task's fields.
func (tr *TaskRepo) Update(taskID string, input models.UpdateTaskInput) (*models.Task, error) {
	updates := []string{}
	args := []interface{}{}

	if input.Title != nil {
		updates = append(updates, "title = ?")
		args = append(args, *input.Title)
	}

	if input.Description != nil {
		updates = append(updates, "description = ?")
		args = append(args, *input.Description)
	}

	if input.StateID != nil {
		updates = append(updates, "state_id = ?")
		args = append(args, *input.StateID)
	}

	if input.Priority != nil {
		updates = append(updates, "priority = ?")
		args = append(args, *input.Priority)
	}

	if input.DueAt != nil {
		updates = append(updates, "due_at = ?")
		args = append(args, *input.DueAt)
	}

	if input.CompletedAt != nil {
		updates = append(updates, "completed_at = ?")
		args = append(args, *input.CompletedAt)
	}

	if input.Archived != nil {
		updates = append(updates, "archived = ?")
		args = append(args, *input.Archived)
	}

	if len(updates) == 0 {
		return tr.FindByID(taskID)
	}

	updates = append(updates, "updated_at = ?")
	args = append(args, time.Now())
	args = append(args, taskID)

	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = ?", strings.Join(updates, ", "))

	_, err := tr.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return tr.FindByID(taskID)
}

// AddNote adds a note to a task.
func (tr *TaskRepo) AddNote(taskID string, body string) (*models.Note, error) {
	result, err := tr.db.Exec(
		"INSERT INTO task_notes (task_id, body) VALUES (?, ?)",
		taskID, body,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert note: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	note := &models.Note{
		ID:        int(id),
		TaskID:    taskID,
		Body:      body,
		CreatedAt: time.Now(),
	}

	return note, nil
}

// DeleteNote removes a note by ID.
func (tr *TaskRepo) DeleteNote(noteID int) error {
	_, err := tr.db.Exec("DELETE FROM task_notes WHERE id = ?", noteID)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}
	return nil
}

// AddTag adds a tag to a task.
func (tr *TaskRepo) AddTag(taskID string, tagName string) (*models.Tag, error) {
	tagName = strings.TrimSpace(tagName)
	if tagName == "" || len(tagName) > 64 {
		return nil, fmt.Errorf("invalid tag name")
	}

	// Insert or get tag
	var tagID int
	err := tr.db.QueryRow(
		"INSERT OR IGNORE INTO tags (name) VALUES (?); SELECT id FROM tags WHERE name = ?",
		tagName, tagName,
	).Scan(&tagID)

	if err != nil && err != sql.ErrNoRows {
		// Try separate approach if combined query fails
		result, err := tr.db.Exec("INSERT OR IGNORE INTO tags (name) VALUES (?)", tagName)
		if err != nil {
			return nil, fmt.Errorf("failed to insert tag: %w", err)
		}

		id, _ := result.LastInsertId()
		if id > 0 {
			tagID = int(id)
		} else {
			// Tag already exists, retrieve it
			err = tr.db.QueryRow("SELECT id FROM tags WHERE name = ?", tagName).Scan(&tagID)
			if err != nil {
				return nil, fmt.Errorf("failed to get tag id: %w", err)
			}
		}
	}

	// Add task_tag relation
	_, err = tr.db.Exec(
		"INSERT OR IGNORE INTO task_tags (task_id, tag_id) VALUES (?, ?)",
		taskID, tagID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add task tag: %w", err)
	}

	tag := &models.Tag{
		ID:   tagID,
		Name: tagName,
	}

	return tag, nil
}

// RemoveTag removes a tag from a task.
func (tr *TaskRepo) RemoveTag(taskID string, tagName string) error {
	_, err := tr.db.Exec(
		`DELETE FROM task_tags WHERE task_id = ? AND tag_id = (SELECT id FROM tags WHERE name = ?)`,
		taskID, tagName,
	)
	if err != nil {
		return fmt.Errorf("failed to remove tag: %w", err)
	}
	return nil
}

// MoveToState updates a task's workflow state.
func (tr *TaskRepo) MoveToState(taskID string, stateID int) (*models.Task, error) {
	_, err := tr.db.Exec(
		"UPDATE tasks SET state_id = ?, updated_at = ? WHERE id = ?",
		stateID, time.Now(), taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update state: %w", err)
	}
	return tr.FindByID(taskID)
}

// List retrieves tasks for a project with filters and ordering.
func (tr *TaskRepo) List(projectID string, filters models.Filters, sortBy models.SortOrder) ([]models.Task, error) {
	query := `SELECT id, project_id, title, description, state_id, priority, due_at, completed_at, session_id, agent_id, created_at, updated_at, archived
	          FROM tasks WHERE project_id = ?`

	args := []interface{}{projectID}

	// Apply filters
	if filters.StateID != nil {
		query += " AND state_id = ?"
		args = append(args, *filters.StateID)
	}

	if filters.IsActive {
		query += " AND archived = 0"
	}

	if filters.Priority != nil {
		query += " AND priority = ?"
		args = append(args, *filters.Priority)
	}

	if filters.DueBefore != nil {
		query += " AND due_at <= ?"
		args = append(args, *filters.DueBefore)
	}

	if filters.DueAfter != nil {
		query += " AND due_at >= ?"
		args = append(args, *filters.DueAfter)
	}

	if filters.HasNote {
		query += " AND EXISTS (SELECT 1 FROM task_notes WHERE task_notes.task_id = tasks.id)"
	}

	// Add tag filter
	if len(filters.Tags) > 0 {
		placeholders := strings.Repeat("?,", len(filters.Tags)-1) + "?"
		query += fmt.Sprintf(" AND id IN (SELECT tt.task_id FROM task_tags tt JOIN tags t ON tt.tag_id = t.id WHERE t.name IN (%s))", placeholders)
		for _, tag := range filters.Tags {
			args = append(args, tag)
		}
	}

	// Apply sorting
	switch sortBy {
	case models.SortByDue:
		query += " ORDER BY due_at ASC, created_at ASC"
	case models.SortByPriority:
		query += " ORDER BY priority DESC, created_at ASC"
	case models.SortByCreated:
		query += " ORDER BY created_at DESC"
	case models.SortByUpdated:
		query += " ORDER BY updated_at DESC"
	default:
		query += " ORDER BY created_at ASC"
	}

	rows, err := tr.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		task := models.Task{}
		if err := rows.Scan(&task.ID, &task.ProjectID, &task.Title, &task.Description, &task.StateID, &task.Priority, &task.DueAt, &task.CompletedAt, &task.SessionID, &task.AgentID, &task.CreatedAt, &task.UpdatedAt, &task.Archived); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		if err := tr.loadTaskRelations(&task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// Search performs full-text search on tasks using FTS5.
func (tr *TaskRepo) Search(projectID string, searchQuery string, limit, offset int) ([]models.SearchResult, error) {
	// Use CTE to get FTS5 results with rank, then join with tasks
	rows, err := tr.db.Query(
		`WITH fts_results AS (
			SELECT task_id, rank FROM task_search WHERE task_search MATCH ?
		)
		SELECT t.id, t.project_id, t.title, t.description, t.state_id, t.priority, t.due_at, t.completed_at, t.session_id, t.agent_id, t.created_at, t.updated_at, t.archived, fts_results.rank
		FROM fts_results
		JOIN tasks t ON t.id = fts_results.task_id
		WHERE t.project_id = ?
		ORDER BY fts_results.rank DESC
		LIMIT ? OFFSET ?`,
		searchQuery, projectID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}
	defer rows.Close()

	var results []models.SearchResult
	for rows.Next() {
		result := models.SearchResult{}
		task := models.Task{}
		var rank float64
		if err := rows.Scan(&task.ID, &task.ProjectID, &task.Title, &task.Description, &task.StateID, &task.Priority, &task.DueAt, &task.CompletedAt, &task.SessionID, &task.AgentID, &task.CreatedAt, &task.UpdatedAt, &task.Archived, &rank); err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		result.Rank = rank
		if err := tr.loadTaskRelations(&task); err != nil {
			return nil, err
		}
		result.Task = task
		results = append(results, result)
	}

	return results, rows.Err()
}

// Archive soft-deletes a task.
func (tr *TaskRepo) Archive(taskID string) (*models.Task, error) {
	_, err := tr.db.Exec(
		"UPDATE tasks SET archived = 1, updated_at = ? WHERE id = ?",
		time.Now(), taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to archive task: %w", err)
	}
	return tr.FindByID(taskID)
}

// HardDelete permanently deletes a task.
func (tr *TaskRepo) HardDelete(taskID string) error {
	_, err := tr.db.Exec("DELETE FROM tasks WHERE id = ?", taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}
