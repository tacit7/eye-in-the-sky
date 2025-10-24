package repository

import (
	"database/sql"
	"fmt"

	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/todo/models"
)

// NoteRepo handles note-related database operations.
type NoteRepo struct {
	db *database.DB
}

// NewNoteRepo creates a new NoteRepo.
func NewNoteRepo(database *database.DB) *NoteRepo {
	return &NoteRepo{db: database}
}

// GetNoteByID retrieves a note by ID.
func (nr *NoteRepo) GetNoteByID(id int) (*models.Note, error) {
	note := &models.Note{}
	err := nr.db.QueryRow(
		"SELECT id, task_id, author, body, created_at FROM task_notes WHERE id = ?",
		id,
	).Scan(&note.ID, &note.TaskID, &note.Author, &note.Body, &note.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("note not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query note: %w", err)
	}

	return note, nil
}

// GetNotesByTaskID retrieves all notes for a task, ordered newest first.
func (nr *NoteRepo) GetNotesByTaskID(taskID string) ([]models.Note, error) {
	rows, err := nr.db.Query(
		"SELECT id, task_id, author, body, created_at FROM task_notes WHERE task_id = ? ORDER BY created_at DESC",
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query notes: %w", err)
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		note := models.Note{}
		if err := rows.Scan(&note.ID, &note.TaskID, &note.Author, &note.Body, &note.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		notes = append(notes, note)
	}

	return notes, rows.Err()
}

// GetLatestNoteByTaskID retrieves the most recent note for a task.
func (nr *NoteRepo) GetLatestNoteByTaskID(taskID string) (*models.Note, error) {
	note := &models.Note{}
	err := nr.db.QueryRow(
		"SELECT id, task_id, author, body, created_at FROM task_notes WHERE task_id = ? ORDER BY created_at DESC LIMIT 1",
		taskID,
	).Scan(&note.ID, &note.TaskID, &note.Author, &note.Body, &note.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil // No notes is valid
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query note: %w", err)
	}

	return note, nil
}

// GetNotesByProjectID retrieves all notes in a project across all tasks.
func (nr *NoteRepo) GetNotesByProjectID(projectID string) ([]models.Note, error) {
	rows, err := nr.db.Query(
		`SELECT tn.id, tn.task_id, tn.author, tn.body, tn.created_at
		 FROM task_notes tn
		 JOIN tasks t ON tn.task_id = t.id
		 WHERE t.project_id = ?
		 ORDER BY tn.created_at DESC`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query notes: %w", err)
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		note := models.Note{}
		if err := rows.Scan(&note.ID, &note.TaskID, &note.Author, &note.Body, &note.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		notes = append(notes, note)
	}

	return notes, rows.Err()
}

// UpdateNoteContent updates the content of a note.
func (nr *NoteRepo) UpdateNoteContent(noteID int, body string) (*models.Note, error) {
	_, err := nr.db.Exec(
		"UPDATE task_notes SET body = ? WHERE id = ?",
		body, noteID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	return nr.GetNoteByID(noteID)
}

// DeleteNote removes a note by ID.
func (nr *NoteRepo) DeleteNote(noteID int) error {
	_, err := nr.db.Exec("DELETE FROM task_notes WHERE id = ?", noteID)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}
	return nil
}

// DeleteNotesByTaskID removes all notes for a task (called when task is deleted).
func (nr *NoteRepo) DeleteNotesByTaskID(taskID string) error {
	_, err := nr.db.Exec("DELETE FROM task_notes WHERE task_id = ?", taskID)
	if err != nil {
		return fmt.Errorf("failed to delete notes: %w", err)
	}
	return nil
}

// CountNotesByTaskID returns the number of notes for a task.
func (nr *NoteRepo) CountNotesByTaskID(taskID string) (int, error) {
	var count int
	err := nr.db.QueryRow(
		"SELECT COUNT(*) FROM task_notes WHERE task_id = ?",
		taskID,
	).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count notes: %w", err)
	}

	return count, nil
}

// GetNoteStats returns statistics for notes in a project.
type NoteStats struct {
	TotalNotes     int
	TasksWithNotes int
}

// GetProjectNoteStats returns note statistics for a project.
func (nr *NoteRepo) GetProjectNoteStats(projectID string) (*NoteStats, error) {
	stats := &NoteStats{}

	// Count total notes
	err := nr.db.QueryRow(
		`SELECT COUNT(*) FROM task_notes tn
		 JOIN tasks t ON tn.task_id = t.id
		 WHERE t.project_id = ?`,
		projectID,
	).Scan(&stats.TotalNotes)

	if err != nil {
		return nil, fmt.Errorf("failed to count notes: %w", err)
	}

	// Count tasks with notes
	err = nr.db.QueryRow(
		`SELECT COUNT(DISTINCT tn.task_id) FROM task_notes tn
		 JOIN tasks t ON tn.task_id = t.id
		 WHERE t.project_id = ?`,
		projectID,
	).Scan(&stats.TasksWithNotes)

	if err != nil {
		return nil, fmt.Errorf("failed to count tasks with notes: %w", err)
	}

	return stats, nil
}

// SearchNotes performs full-text search on note content within a project.
func (nr *NoteRepo) SearchNotes(projectID string, query string, limit, offset int) ([]models.Note, error) {
	rows, err := nr.db.Query(
		`SELECT tn.id, tn.task_id, tn.author, tn.body, tn.created_at
		 FROM task_notes tn
		 JOIN tasks t ON tn.task_id = t.id
		 JOIN task_search ts ON t.id = ts.rowid
		 WHERE t.project_id = ? AND ts MATCH ?
		 ORDER BY tn.created_at DESC
		 LIMIT ? OFFSET ?`,
		projectID, query, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search notes: %w", err)
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		note := models.Note{}
		if err := rows.Scan(&note.ID, &note.TaskID, &note.Author, &note.Body, &note.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		notes = append(notes, note)
	}

	return notes, rows.Err()
}
