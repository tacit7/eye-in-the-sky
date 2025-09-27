package database

import (
	"database/sql"
	"fmt"
)

// CreateAgent inserts a new agent
func (db *DB) CreateAgent(agent *Agent) error {
	query := `
		INSERT INTO agents (id, status, git_worktree_path, feature_description, current_task, last_activity_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query, agent.ID, agent.Status, agent.GitWorktreePath,
		agent.FeatureDescription, agent.CurrentTask, agent.LastActivityAt)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}
	return nil
}

// GetAgent retrieves an agent by ID
func (db *DB) GetAgent(id string) (*Agent, error) {
	query := `
		SELECT id, status, created_at, updated_at, git_worktree_path, feature_description, current_task, last_activity_at
		FROM agents WHERE id = ?
	`
	var agent Agent
	row := db.conn.QueryRow(query, id)
	err := row.Scan(&agent.ID, &agent.Status, &agent.CreatedAt, &agent.UpdatedAt,
		&agent.GitWorktreePath, &agent.FeatureDescription, &agent.CurrentTask, &agent.LastActivityAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("agent not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	return &agent, nil
}

// UpdateAgentStatus updates an agent's status
func (db *DB) UpdateAgentStatus(id, status string, currentTask *string) error {
	query := `
		UPDATE agents SET status = ?, current_task = ?, updated_at = CURRENT_TIMESTAMP, last_activity_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	result, err := db.conn.Exec(query, status, currentTask, id)
	if err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("agent not found: %s", id)
	}

	return nil
}

// CreateAction logs a new action
func (db *DB) CreateAction(action *Action) error {
	query := `INSERT INTO actions (agent_id, action_type, description, details) VALUES (?, ?, ?, ?)`
	_, err := db.conn.Exec(query, action.AgentID, action.ActionType, action.Description, action.Details)
	if err != nil {
		return fmt.Errorf("failed to create action: %w", err)
	}

	// Update agent's last activity
	updateQuery := `UPDATE agents SET last_activity_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err = db.conn.Exec(updateQuery, action.AgentID)
	if err != nil {
		return fmt.Errorf("failed to update agent activity: %w", err)
	}

	return nil
}

// CreateCommits logs git commits
func (db *DB) CreateCommits(agentID string, commitHashes []string, commitMessages []string) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO commits (agent_id, commit_hash, commit_message) VALUES (?, ?, ?)`
	for i, hash := range commitHashes {
		var message *string
		if i < len(commitMessages) && commitMessages[i] != "" {
			message = &commitMessages[i]
		}
		_, err := tx.Exec(query, agentID, hash, message)
		if err != nil {
			return fmt.Errorf("failed to insert commit %s: %w", hash, err)
		}
	}

	// Update agent's last activity
	updateQuery := `UPDATE agents SET last_activity_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err = tx.Exec(updateQuery, agentID)
	if err != nil {
		return fmt.Errorf("failed to update agent activity: %w", err)
	}

	return tx.Commit()
}

// EndAgentSession marks an agent as completed
func (db *DB) EndAgentSession(agentID, summary string, finalStatus string) error {
	status := finalStatus
	if status == "" {
		status = StatusCompleted
	}

	query := `UPDATE agents SET status = ?, updated_at = CURRENT_TIMESTAMP, last_activity_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.conn.Exec(query, status, agentID)
	if err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}

	return nil
}