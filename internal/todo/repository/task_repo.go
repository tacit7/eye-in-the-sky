package repository

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/todo/models"
)

// TaskRepo handles task-related database operations.
type TaskRepo struct {
	executor database.Executor
}

// NewTaskRepo creates a new TaskRepo.
func NewTaskRepo(database *database.DB) *TaskRepo {
	return &TaskRepo{executor: database}
}

// WithTx creates a new TaskRepo scoped to a transaction.
func (tr *TaskRepo) WithTx(tx *database.Tx) *TaskRepo {
	return &TaskRepo{executor: tx}
}

// CreateTask creates a new task in a project.
func (tr *TaskRepo) CreateTask(projectID int, input models.CreateTaskInput) (*models.Task, error) {
	// Validate input
	if strings.TrimSpace(input.Title) == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}

	// Generate UUID for task ID
	taskID := uuid.New().String()

	// Default to state_id = 1 (todo) if not provided
	stateID := 1
	if input.StateID != nil {
		stateID = *input.StateID
	}

	// Insert task (without session_id column)
	_, err := tr.executor.Exec(
		`INSERT INTO tasks (id, project_id, title, description, state_id, priority, due_at, agent_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		taskID, projectID, input.Title, input.Description, stateID, input.Priority, input.DueAt, input.AgentID, time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert task: %w", err)
	}

	// Insert into task_sessions junction table for each session ID
	for _, sessionID := range input.SessionIDs {
		_, err := tr.executor.Exec(
			`INSERT INTO task_sessions (task_id, session_id, created_at) VALUES (?, ?, ?)`,
			taskID, sessionID, time.Now(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert task_session: %w", err)
		}
	}

	return tr.FindByID(taskID)
}

// FindByID retrieves a task by ID with all related data.
func (tr *TaskRepo) FindByID(id string) (*models.Task, error) {
	task := &models.Task{}
	err := tr.executor.QueryRow(
		`SELECT id, project_id, title, description, state_id, COALESCE(priority, 0), due_at, completed_at, agent_id, created_at, updated_at, COALESCE(archived, 0)
		 FROM tasks WHERE id = ?`,
		id,
	).Scan(&task.ID, &task.ProjectID, &task.Title, &task.Description, &task.StateID, &task.Priority, &task.DueAt, &task.CompletedAt, &task.AgentID, &task.CreatedAt, &task.UpdatedAt, &task.Archived)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query task: %w", err)
	}

	// Load notes, tags, and sessions
	if err := tr.loadTaskRelations(task); err != nil {
		return nil, err
	}

	return task, nil
}

// loadTaskRelations loads notes, tags, and sessions for a task.
func (tr *TaskRepo) loadTaskRelations(task *models.Task) error {
	// Load notes
	rows, err := tr.executor.Query(
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
	rows, err = tr.executor.Query(
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

	// Load sessions from junction table
	rows, err = tr.executor.Query(
		`SELECT session_id FROM task_sessions WHERE task_id = ? ORDER BY created_at`,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	task.SessionIDs = []string{} // Initialize empty slice
	for rows.Next() {
		var sessionID string
		if err := rows.Scan(&sessionID); err != nil {
			return fmt.Errorf("failed to scan session_id: %w", err)
		}
		task.SessionIDs = append(task.SessionIDs, sessionID)
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

	_, err := tr.executor.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return tr.FindByID(taskID)
}

// AddNote adds a note to a task.
func (tr *TaskRepo) AddNote(taskID string, body string) (*models.Note, error) {
	// Get current user from environment
	author := getAuthor()

	result, err := tr.executor.Exec(
		"INSERT INTO task_notes (task_id, author, body) VALUES (?, ?, ?)",
		taskID, author, body,
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
		Author:    &author,
		Body:      body,
		CreatedAt: time.Now(),
	}

	return note, nil
}

// DeleteNote removes a note by ID.
func (tr *TaskRepo) DeleteNote(noteID int) error {
	_, err := tr.executor.Exec("DELETE FROM task_notes WHERE id = ?", noteID)
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
	err := tr.executor.QueryRow(
		"INSERT OR IGNORE INTO tags (name) VALUES (?); SELECT id FROM tags WHERE name = ?",
		tagName, tagName,
	).Scan(&tagID)

	if err != nil && err != sql.ErrNoRows {
		// Try separate approach if combined query fails
		result, err := tr.executor.Exec("INSERT OR IGNORE INTO tags (name) VALUES (?)", tagName)
		if err != nil {
			return nil, fmt.Errorf("failed to insert tag: %w", err)
		}

		id, _ := result.LastInsertId()
		if id > 0 {
			tagID = int(id)
		} else {
			// Tag already exists, retrieve it
			err = tr.executor.QueryRow("SELECT id FROM tags WHERE name = ?", tagName).Scan(&tagID)
			if err != nil {
				return nil, fmt.Errorf("failed to get tag id: %w", err)
			}
		}
	}

	// Add task_tag relation
	_, err = tr.executor.Exec(
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
	_, err := tr.executor.Exec(
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
	_, err := tr.executor.Exec(
		"UPDATE tasks SET state_id = ?, updated_at = ? WHERE id = ?",
		stateID, time.Now(), taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update state: %w", err)
	}
	return tr.FindByID(taskID)
}

// List retrieves tasks for a project with filters and ordering.
func (tr *TaskRepo) List(projectID int, filters models.Filters, sortBy models.SortOrder) ([]models.Task, error) {
	query := `SELECT id, project_id, title, description, state_id, COALESCE(priority, 0), due_at, completed_at, agent_id, created_at, updated_at, COALESCE(archived, 0)
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

	rows, err := tr.executor.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		task := models.Task{}
		if err := rows.Scan(&task.ID, &task.ProjectID, &task.Title, &task.Description, &task.StateID, &task.Priority, &task.DueAt, &task.CompletedAt, &task.AgentID, &task.CreatedAt, &task.UpdatedAt, &task.Archived); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		if err := tr.loadTaskRelations(&task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// ListByAgent retrieves tasks filtered by agent ID
func (tr *TaskRepo) ListByAgent(agentID string, projectID int, filters models.Filters, sortBy models.SortOrder) ([]models.Task, error) {
	query := `SELECT id, project_id, title, description, state_id, COALESCE(priority, 0), due_at, completed_at, agent_id, created_at, updated_at, COALESCE(archived, 0)
	          FROM tasks WHERE project_id = ? AND agent_id = ?`

	args := []interface{}{projectID, agentID}

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

	rows, err := tr.executor.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks by agent: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		task := models.Task{}
		if err := rows.Scan(&task.ID, &task.ProjectID, &task.Title, &task.Description, &task.StateID, &task.Priority, &task.DueAt, &task.CompletedAt, &task.AgentID, &task.CreatedAt, &task.UpdatedAt, &task.Archived); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		if err := tr.loadTaskRelations(&task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// ListBySession retrieves tasks filtered by session ID via task_sessions junction table
func (tr *TaskRepo) ListBySession(sessionID string, projectID int, filters models.Filters, sortBy models.SortOrder) ([]models.Task, error) {
	query := `SELECT t.id, t.project_id, t.title, t.description, t.state_id, COALESCE(t.priority, 0), t.due_at, t.completed_at, t.agent_id, t.created_at, t.updated_at, COALESCE(t.archived, 0)
	          FROM tasks t
	          INNER JOIN task_sessions ts ON t.id = ts.task_id
	          WHERE t.project_id = ? AND ts.session_id = ?`

	args := []interface{}{projectID, sessionID}

	// Apply filters (use table alias t. for tasks)
	if filters.StateID != nil {
		query += " AND t.state_id = ?"
		args = append(args, *filters.StateID)
	}

	if filters.IsActive {
		query += " AND t.archived = 0"
	}

	if filters.Priority != nil {
		query += " AND t.priority = ?"
		args = append(args, *filters.Priority)
	}

	if filters.DueBefore != nil {
		query += " AND t.due_at <= ?"
		args = append(args, *filters.DueBefore)
	}

	if filters.DueAfter != nil {
		query += " AND t.due_at >= ?"
		args = append(args, *filters.DueAfter)
	}

	if filters.HasNote {
		query += " AND EXISTS (SELECT 1 FROM task_notes WHERE task_notes.task_id = t.id)"
	}

	// Add tag filter
	if len(filters.Tags) > 0 {
		placeholders := strings.Repeat("?,", len(filters.Tags)-1) + "?"
		query += fmt.Sprintf(" AND t.id IN (SELECT tt.task_id FROM task_tags tt JOIN tags tg ON tt.tag_id = tg.id WHERE tg.name IN (%s))", placeholders)
		for _, tag := range filters.Tags {
			args = append(args, tag)
		}
	}

	// Apply sorting (use table alias t. for tasks)
	switch sortBy {
	case models.SortByDue:
		query += " ORDER BY t.due_at ASC, t.created_at ASC"
	case models.SortByPriority:
		query += " ORDER BY t.priority DESC, t.created_at ASC"
	case models.SortByCreated:
		query += " ORDER BY t.created_at DESC"
	case models.SortByUpdated:
		query += " ORDER BY t.updated_at DESC"
	default:
		query += " ORDER BY t.created_at ASC"
	}

	rows, err := tr.executor.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks by session: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		task := models.Task{}
		if err := rows.Scan(&task.ID, &task.ProjectID, &task.Title, &task.Description, &task.StateID, &task.Priority, &task.DueAt, &task.CompletedAt, &task.AgentID, &task.CreatedAt, &task.UpdatedAt, &task.Archived); err != nil {
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
func (tr *TaskRepo) Search(projectID int, searchQuery string, limit, offset int) ([]models.SearchResult, error) {
	// Use CTE to get FTS5 results with rank, then join with tasks
	rows, err := tr.executor.Query(
		`WITH fts_results AS (
			SELECT task_id, rank FROM task_search WHERE task_search MATCH ?
		)
		SELECT t.id, t.project_id, t.title, t.description, t.state_id, COALESCE(t.priority, 0), t.due_at, t.completed_at, t.agent_id, t.created_at, t.updated_at, COALESCE(t.archived, 0), fts_results.rank
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
		if err := rows.Scan(&task.ID, &task.ProjectID, &task.Title, &task.Description, &task.StateID, &task.Priority, &task.DueAt, &task.CompletedAt, &task.AgentID, &task.CreatedAt, &task.UpdatedAt, &task.Archived, &rank); err != nil {
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
	_, err := tr.executor.Exec(
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
	_, err := tr.executor.Exec("DELETE FROM tasks WHERE id = ?", taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// AddSession links a task to a session via the task_sessions junction table.
func (tr *TaskRepo) AddSession(taskID string, sessionID string) error {
	// Check if the relationship already exists
	var count int
	err := tr.executor.QueryRow(
		"SELECT COUNT(*) FROM task_sessions WHERE task_id = ? AND session_id = ?",
		taskID, sessionID,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing session: %w", err)
	}

	if count > 0 {
		// Already exists, nothing to do
		return nil
	}

	// Insert new task-session relationship
	_, err = tr.executor.Exec(
		"INSERT INTO task_sessions (task_id, session_id, created_at) VALUES (?, ?, ?)",
		taskID, sessionID, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to add session to task: %w", err)
	}

	return nil
}

// RemoveSession unlinks a task from a session via the task_sessions junction table.
func (tr *TaskRepo) RemoveSession(taskID string, sessionID string) error {
	result, err := tr.executor.Exec(
		"DELETE FROM task_sessions WHERE task_id = ? AND session_id = ?",
		taskID, sessionID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove session from task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not linked to task")
	}

	return nil
}

// BulkAddSessionToTasks adds targetSessionID to all tasks belonging to sourceSessionID.
// This allows you to "copy" all tasks from one session to another session.
func (tr *TaskRepo) BulkAddSessionToTasks(sourceSessionID string, targetSessionID string) (int, error) {
	// Get all unique task IDs for the source session
	rows, err := tr.executor.Query(
		"SELECT DISTINCT task_id FROM task_sessions WHERE session_id = ?",
		sourceSessionID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to query tasks for source session: %w", err)
	}
	defer rows.Close()

	var taskIDs []string
	for rows.Next() {
		var taskID string
		if err := rows.Scan(&taskID); err != nil {
			return 0, fmt.Errorf("failed to scan task_id: %w", err)
		}
		taskIDs = append(taskIDs, taskID)
	}

	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("error iterating task rows: %w", err)
	}

	// Add target session to each task (skip if already exists)
	addedCount := 0
	for _, taskID := range taskIDs {
		// Check if already exists
		var count int
		err := tr.executor.QueryRow(
			"SELECT COUNT(*) FROM task_sessions WHERE task_id = ? AND session_id = ?",
			taskID, targetSessionID,
		).Scan(&count)
		if err != nil {
			return addedCount, fmt.Errorf("failed to check existing session: %w", err)
		}

		if count > 0 {
			// Already linked, skip
			continue
		}

		// Insert new link
		_, err = tr.executor.Exec(
			"INSERT INTO task_sessions (task_id, session_id, created_at) VALUES (?, ?, ?)",
			taskID, targetSessionID, time.Now(),
		)
		if err != nil {
			return addedCount, fmt.Errorf("failed to add session to task %s: %w", taskID, err)
		}
		addedCount++
	}

	return addedCount, nil
}

// getAuthor returns the current user from environment
func getAuthor() string {
	// Try USER first (most Unix systems)
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	// Try USERNAME (Windows)
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	// Fallback
	return "user"
}
