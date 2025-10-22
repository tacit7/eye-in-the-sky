package data

import (
	"context"
	"database/sql"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// LogsStore implements the LogsStore interface
type LogsStore struct {
	db *sql.DB
}

// NewLogsStore creates a new logs store
func NewLogsStore(db *sql.DB) *LogsStore {
	return &LogsStore{db: db}
}

// LoadBySession loads logs for a specific session
func (s *LogsStore) LoadBySession(ctx context.Context, sessionID string, limit int) ([]domain.Log, error) {
	query := `
		SELECT id, session_id, type, message, created_at
		FROM logs
		WHERE session_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.Log
	for rows.Next() {
		var log domain.Log
		err := rows.Scan(&log.ID, &log.SessionID, &log.Type, &log.Message, &log.Timestamp)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// Create creates a new log entry
func (s *LogsStore) Create(ctx context.Context, sessionID, logType, message string) error {
	query := `
		INSERT INTO logs (session_id, type, message, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err := s.db.ExecContext(ctx, query, sessionID, logType, message)
	return err
}