package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/todo/db"
	"github.com/tacit7/eye-in-the-sky/internal/todo/models"
)

// TaskRepo handles task-related database operations.
type TaskRepo struct {
	db *db.DB
}

// NewTaskRepo creates a new TaskRepo.
func NewTaskRepo(database *db.DB) *TaskRepo {
	return &TaskRepo{db: database}
}

// CreateTask creates a new task in a project.
func (tr *TaskRepo) CreateTask(projectID int, input models.CreateTaskInput) (*models.Task, error) {
	// Validate input
	if strings.TrimSpace(input.Description) == "" {
		return nil, fmt.Errorf("description cannot be empty")
	}

	result, err := tr.db.Exec(
		`INSERT INTO tasks (project_id, description, parent_id, state_code, priority, weight, session_id, agent_id, position)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, (SELECT COALESCE(MAX(position), 0) + 1 FROM tasks WHERE project_id = ?))`,
		projectID, input.Description, input.ParentID, input.StateCode, input.Priority, input.Weight, input.SessionID, input.AgentID, projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return tr.FindByID(int(id))
}

// FindByID retrieves a task by ID with all related data.
func (tr *TaskRepo) FindByID(id int) (*models.Task, error) {
	task := &models.Task{}
	err := tr.db.QueryRow(
		`SELECT id, project_id, description, state_code, parent_id, priority, weight, position, due_date, session_id, agent_id, created_at, updated_at, archived_at
		 FROM tasks WHERE id = ?`,
		id,
	).Scan(&task.ID, &task.ProjectID, &task.Description, &task.StateCode, &task.ParentID, &task.Priority, &task.Weight, &task.Position, &task.DueDate, &task.SessionID, &task.AgentID, &task.CreatedAt, &task.UpdatedAt, &task.ArchivedAt)

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
		"SELECT id, task_id, body_markdown, created_at FROM task_notes WHERE task_id = ? ORDER BY created_at DESC",
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to query notes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		note := models.Note{}
		if err := rows.Scan(&note.ID, &note.TaskID, &note.BodyMarkdown, &note.CreatedAt); err != nil {
			return fmt.Errorf("failed to scan note: %w", err)
		}
		task.Notes = append(task.Notes, note)
	}

	// Load tags
	rows, err = tr.db.Query(
		`SELECT t.id, t.name, t.created_at FROM tags t
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
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.CreatedAt); err != nil {
			return fmt.Errorf("failed to scan tag: %w", err)
		}
		task.Tags = append(task.Tags, tag)
	}

	return nil
}

// UpdateDescription updates a task's description.
func (tr *TaskRepo) UpdateDescription(taskID int, description string) (*models.Task, error) {
	if strings.TrimSpace(description) == "" {
		return nil, fmt.Errorf("description cannot be empty")
	}

	_, err := tr.db.Exec(
		"UPDATE tasks SET description = ? WHERE id = ?",
		description, taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update description: %w", err)
	}

	return tr.FindByID(taskID)
}

// AddNote adds a markdown note to a task.
func (tr *TaskRepo) AddNote(taskID int, bodyMarkdown string) (*models.Note, error) {
	result, err := tr.db.Exec(
		"INSERT INTO task_notes (task_id, body_markdown) VALUES (?, ?)",
		taskID, bodyMarkdown,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert note: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	note := &models.Note{
		ID:           int(id),
		TaskID:       taskID,
		BodyMarkdown: bodyMarkdown,
		CreatedAt:    time.Now(),
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
func (tr *TaskRepo) AddTag(taskID int, tagName string) (*models.Tag, error) {
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
		ID:        tagID,
		Name:      tagName,
		CreatedAt: time.Now(),
	}

	return tag, nil
}

// RemoveTag removes a tag from a task.
func (tr *TaskRepo) RemoveTag(taskID int, tagName string) error {
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
func (tr *TaskRepo) MoveToState(taskID int, stateCode string) (*models.Task, error) {
	_, err := tr.db.Exec(
		"UPDATE tasks SET state_code = ? WHERE id = ?",
		stateCode, taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update state: %w", err)
	}
	return tr.FindByID(taskID)
}

// SetDue sets the due date for a task.
func (tr *TaskRepo) SetDue(taskID int, due *time.Time) (*models.Task, error) {
	_, err := tr.db.Exec(
		"UPDATE tasks SET due_date = ? WHERE id = ?",
		due, taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set due date: %w", err)
	}
	return tr.FindByID(taskID)
}

// SetPriority sets the priority (1-5) for a task.
func (tr *TaskRepo) SetPriority(taskID int, priority *int) (*models.Task, error) {
	if priority != nil && (*priority < 1 || *priority > 5) {
		return nil, fmt.Errorf("priority must be between 1 and 5")
	}

	_, err := tr.db.Exec(
		"UPDATE tasks SET priority = ? WHERE id = ?",
		priority, taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set priority: %w", err)
	}
	return tr.FindByID(taskID)
}

// SetWeight sets the weight for a task.
func (tr *TaskRepo) SetWeight(taskID int, weight *int) (*models.Task, error) {
	_, err := tr.db.Exec(
		"UPDATE tasks SET weight = ? WHERE id = ?",
		weight, taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set weight: %w", err)
	}
	return tr.FindByID(taskID)
}

// SetParent sets the parent task, preventing cycles and depth limits.
func (tr *TaskRepo) SetParent(childID int, parentID *int) (*models.Task, error) {
	if parentID != nil && *parentID == childID {
		return nil, fmt.Errorf("cannot set task as its own parent")
	}

	// Check depth to prevent deep nesting
	if parentID != nil {
		depth := 0
		currentID := *parentID
		for currentID != 0 && depth < 32 {
			var nextParentID *int
			err := tr.db.QueryRow(
				"SELECT parent_id FROM tasks WHERE id = ?",
				currentID,
			).Scan(&nextParentID)

			if err != nil {
				return nil, fmt.Errorf("failed to check depth: %w", err)
			}

			if nextParentID == nil {
				break
			}
			currentID = *nextParentID
			depth++
		}

		if depth >= 32 {
			return nil, fmt.Errorf("parent depth exceeds maximum of 32")
		}
	}

	_, err := tr.db.Exec(
		"UPDATE tasks SET parent_id = ? WHERE id = ?",
		parentID, childID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set parent: %w", err)
	}
	return tr.FindByID(childID)
}

// Reorder changes the position of a task.
func (tr *TaskRepo) Reorder(taskID int, newPosition int) (*models.Task, error) {
	_, err := tr.db.Exec(
		"UPDATE tasks SET position = ? WHERE id = ?",
		newPosition, taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to reorder: %w", err)
	}
	return tr.FindByID(taskID)
}

// List retrieves tasks for a project with filters and ordering.
func (tr *TaskRepo) List(projectID int, filters models.Filters, sortBy models.SortOrder) ([]models.Task, error) {
	query := `SELECT id, project_id, description, state_code, parent_id, priority, weight, position, due_date, session_id, agent_id, created_at, updated_at, archived_at
	          FROM tasks WHERE project_id = ?`

	args := []interface{}{projectID}

	// Apply filters
	if filters.StateCode != nil {
		query += " AND state_code = ?"
		args = append(args, *filters.StateCode)
	}

	if filters.IsActive {
		query += " AND archived_at IS NULL"
	}

	if filters.Priority != nil {
		query += " AND priority = ?"
		args = append(args, *filters.Priority)
	}

	if filters.DueBefore != nil {
		query += " AND due_date <= ?"
		args = append(args, *filters.DueBefore)
	}

	if filters.DueAfter != nil {
		query += " AND due_date >= ?"
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
		query += " ORDER BY due_date ASC, position ASC"
	case models.SortByPriority:
		query += " ORDER BY priority DESC, position ASC"
	case models.SortByCreated:
		query += " ORDER BY created_at DESC"
	case models.SortByUpdated:
		query += " ORDER BY updated_at DESC"
	default:
		query += " ORDER BY position ASC"
	}

	rows, err := tr.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		task := models.Task{}
		if err := rows.Scan(&task.ID, &task.ProjectID, &task.Description, &task.StateCode, &task.ParentID, &task.Priority, &task.Weight, &task.Position, &task.DueDate, &task.SessionID, &task.AgentID, &task.CreatedAt, &task.UpdatedAt, &task.ArchivedAt); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		if err := tr.loadTaskRelations(&task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// Search performs full-text search on tasks.
func (tr *TaskRepo) Search(projectID int, query string, limit, offset int) ([]models.SearchResult, error) {
	rows, err := tr.db.Query(
		`SELECT t.id, t.project_id, t.description, t.state_code, t.parent_id, t.priority, t.weight, t.position, t.due_date, t.session_id, t.agent_id, t.created_at, t.updated_at, t.archived_at, s.rank
		 FROM task_search s
		 JOIN tasks t ON s.rowid = t.id
		 WHERE t.project_id = ? AND s MATCH ?
		 ORDER BY s.rank ASC
		 LIMIT ? OFFSET ?`,
		projectID, query, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}
	defer rows.Close()

	var results []models.SearchResult
	for rows.Next() {
		result := models.SearchResult{}
		task := models.Task{}
		if err := rows.Scan(&task.ID, &task.ProjectID, &task.Description, &task.StateCode, &task.ParentID, &task.Priority, &task.Weight, &task.Position, &task.DueDate, &task.SessionID, &task.AgentID, &task.CreatedAt, &task.UpdatedAt, &task.ArchivedAt, &result.Rank); err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		if err := tr.loadTaskRelations(&task); err != nil {
			return nil, err
		}
		result.Task = task
		results = append(results, result)
	}

	return results, rows.Err()
}

// Delete soft-deletes a task and all its children recursively.
func (tr *TaskRepo) Delete(taskID int) error {
	// Get all child IDs recursively
	var childIDs []int
	if err := tr.getChildTaskIDs(taskID, &childIDs); err != nil {
		return err
	}

	// Add the task itself
	childIDs = append(childIDs, taskID)

	// Soft delete all (could also use hard delete if preferred)
	for _, id := range childIDs {
		_, err := tr.db.Exec(
			"UPDATE tasks SET archived_at = CURRENT_TIMESTAMP WHERE id = ?",
			id,
		)
		if err != nil {
			return fmt.Errorf("failed to delete task: %w", err)
		}
	}

	return nil
}

// getChildTaskIDs recursively gets all child task IDs.
func (tr *TaskRepo) getChildTaskIDs(parentID int, childIDs *[]int) error {
	rows, err := tr.db.Query(
		"SELECT id FROM tasks WHERE parent_id = ?",
		parentID,
	)
	if err != nil {
		return fmt.Errorf("failed to query child tasks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var childID int
		if err := rows.Scan(&childID); err != nil {
			return fmt.Errorf("failed to scan child id: %w", err)
		}
		*childIDs = append(*childIDs, childID)
		if err := tr.getChildTaskIDs(childID, childIDs); err != nil {
			return err
		}
	}

	return nil
}
