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

// LoadByAgentHierarchy loads commits for an agent and all its child agents
func (s *commitsStore) LoadByAgentHierarchy(ctx context.Context, agentID domain.AgentID, limit int) ([]domain.Commit, error) {
	query := `
		WITH RECURSIVE agent_hierarchy AS (
			SELECT id FROM agents WHERE id = ?
			UNION ALL
			SELECT a.id FROM agents a
			INNER JOIN agent_hierarchy ah ON a.parent_agent_id = ah.id
		)
		SELECT c.id, c.agent_id, c.hash, c.message, c.timestamp
		FROM commits c
		WHERE c.agent_id IN (SELECT id FROM agent_hierarchy)
		ORDER BY c.timestamp DESC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, string(agentID), limit)
	if err != nil {
		return nil, fmt.Errorf("query commits hierarchy: %w", err)
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