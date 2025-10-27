package data

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

// LoadByAgent loads notes for a specific agent and its session (polymorphic)
func (s *notesStore) LoadByAgent(ctx context.Context, agentID domain.AgentID, sessionID string) ([]domain.Note, error) {
	query := `
		SELECT id, parent_id, parent_type, body, created_at
		FROM notes
		WHERE (parent_id = ? AND parent_type = 'agents')
		   OR (parent_id = ? AND parent_type = 'sessions')
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, string(agentID), sessionID)
	if err != nil {
		return nil, fmt.Errorf("query notes: %w", err)
	}
	defer rows.Close()

	var notes []domain.Note
	for rows.Next() {
		var n domain.Note
		var idStr string
		err := rows.Scan(&idStr, &n.ParentID, &n.ParentType, &n.Body, &n.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		n.ID = domain.NoteID(idStr)
		n.AgentID = agentID
		// Set legacy fields for backward compatibility
		n.Content = n.Body
		n.Title = extractTitle(n.Body)
		notes = append(notes, n)
	}

	return notes, nil
}

// extractTitle extracts the first two words from content as a title
func extractTitle(content string) string {
	words := strings.Fields(content)
	if len(words) == 0 {
		return "Untitled"
	}
	if len(words) == 1 {
		return words[0]
	}
	return strings.Join(words[:2], " ")
}

// Create creates a new note (agentID parameter kept for interface compatibility, uses sessionID in query)
func (s *notesStore) Create(ctx context.Context, agentID domain.AgentID, content string) error {
	title := extractTitle(content)
	query := `
		INSERT INTO notes (session_id, title, content, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err := s.db.ExecContext(ctx, query, string(agentID), title, content)
	if err != nil {
		return fmt.Errorf("insert note: %w", err)
	}

	return nil
}

// Delete deletes a note
func (s *notesStore) Delete(ctx context.Context, noteID domain.NoteID) error {
	query := `DELETE FROM notes WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, string(noteID))
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