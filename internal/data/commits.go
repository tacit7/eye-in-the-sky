package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// commitsStore implements SQL operations for commits
type commitsStore struct {
	db *sql.DB
}

// NewCommitsStore creates a new commits store
func NewCommitsStore(db *sql.DB) *commitsStore {
	return &commitsStore{db: db}
}

// LoadByAgent loads commits for a specific agent
func (s *commitsStore) LoadByAgent(ctx context.Context, agentID domain.AgentID) ([]domain.Commit, error) {
	query := `
		SELECT id, agent_id, hash, message, timestamp
		FROM commits
		WHERE agent_id = ?
		ORDER BY timestamp DESC
	`

	rows, err := s.db.QueryContext(ctx, query, string(agentID))
	if err != nil {
		return nil, fmt.Errorf("query commits: %w", err)
	}
	defer rows.Close()

	var commits []domain.Commit
	for rows.Next() {
		var c domain.Commit
		err := rows.Scan(&c.ID, &c.AgentID, &c.Hash, &c.Message, &c.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("scan commit: %w", err)
		}
		commits = append(commits, c)
	}

	return commits, nil
}

// LoadRecent loads the most recent commits across all agents
func (s *commitsStore) LoadRecent(ctx context.Context, limit int) ([]domain.Commit, error) {
	query := `
		SELECT id, agent_id, hash, message, timestamp
		FROM commits
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query commits: %w", err)
	}
	defer rows.Close()

	var commits []domain.Commit
	for rows.Next() {
		var c domain.Commit
		err := rows.Scan(&c.ID, &c.AgentID, &c.Hash, &c.Message, &c.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("scan commit: %w", err)
		}
		commits = append(commits, c)
	}

	return commits, nil
}