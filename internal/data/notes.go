package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// notesStore implements SQL operations for notes
type notesStore struct {
	db *sql.DB
}

// NewNotesStore creates a new notes store
func NewNotesStore(db *sql.DB) *notesStore {
	return &notesStore{db: db}
}

// LoadByAgent loads notes for a specific agent's session
func (s *notesStore) LoadByAgent(ctx context.Context, agentID domain.AgentID, sessionID string) ([]domain.Note, error) {
	query := `
		SELECT id, session_id, content, created_at
		FROM notes
		WHERE session_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("query notes: %w", err)
	}
	defer rows.Close()

	var notes []domain.Note
	for rows.Next() {
		var n domain.Note
		var sessionID string
		err := rows.Scan(&n.ID, &sessionID, &n.Content, &n.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		n.AgentID = agentID
		notes = append(notes, n)
	}

	return notes, nil
}

// Create creates a new note (agentID parameter kept for interface compatibility, uses sessionID in query)
func (s *notesStore) Create(ctx context.Context, agentID domain.AgentID, content string) error {
	query := `
		INSERT INTO notes (session_id, content, created_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
	`

	_, err := s.db.ExecContext(ctx, query, string(agentID), content)
	if err != nil {
		return fmt.Errorf("insert note: %w", err)
	}

	return nil
}

// Delete deletes a note
func (s *notesStore) Delete(ctx context.Context, noteID domain.NoteID) error {
	query := `DELETE FROM notes WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, int(noteID))
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("note not found: %d", noteID)
	}

	return nil
}