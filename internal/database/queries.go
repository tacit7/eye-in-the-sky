package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// Agent CRUD operations

// CreateAgent inserts a new agent into the database
func (db *DB) CreateAgent(agent *Agent) error {
	query := `
		INSERT INTO agents (id, status, git_worktree_path, feature_description, current_task, last_activity_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(query,
		agent.ID,
		agent.Status,
		agent.GitWorktreePath,
		agent.FeatureDescription,
		agent.CurrentTask,
		agent.LastActivityAt,
	)

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

	err := row.Scan(
		&agent.ID,
		&agent.Status,
		&agent.CreatedAt,
		&agent.UpdatedAt,
		&agent.GitWorktreePath,
		&agent.FeatureDescription,
		&agent.CurrentTask,
		&agent.LastActivityAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("agent not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	return &agent, nil
}

// UpdateAgentStatus updates an agent's status and current task
func (db *DB) UpdateAgentStatus(id, status string, currentTask *string) error {
	query := `
		UPDATE agents
		SET status = ?, current_task = ?, updated_at = CURRENT_TIMESTAMP, last_activity_at = CURRENT_TIMESTAMP
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

// ListAgents retrieves all agents, optionally filtered by status
func (db *DB) ListAgents(status string) ([]*Agent, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `
			SELECT id, status, created_at, updated_at, git_worktree_path, feature_description, current_task, last_activity_at
			FROM agents WHERE status = ?
			ORDER BY updated_at DESC
		`
		args = append(args, status)
	} else {
		query = `
			SELECT id, status, created_at, updated_at, git_worktree_path, feature_description, current_task, last_activity_at
			FROM agents
			ORDER BY updated_at DESC
		`
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer rows.Close()

	var agents []*Agent
	for rows.Next() {
		var agent Agent
		err := rows.Scan(
			&agent.ID,
			&agent.Status,
			&agent.CreatedAt,
			&agent.UpdatedAt,
			&agent.GitWorktreePath,
			&agent.FeatureDescription,
			&agent.CurrentTask,
			&agent.LastActivityAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}
		agents = append(agents, &agent)
	}

	return agents, nil
}

// Action CRUD operations

// CreateAction logs a new action for an agent
func (db *DB) CreateAction(action *Action) error {
	query := `
		INSERT INTO actions (agent_id, action_type, description, details)
		VALUES (?, ?, ?, ?)
	`

	_, err := db.conn.Exec(query,
		action.AgentID,
		action.ActionType,
		action.Description,
		action.Details,
	)

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

// GetActionsForAgent retrieves all actions for a specific agent
func (db *DB) GetActionsForAgent(agentID string, limit int) ([]*Action, error) {
	query := `
		SELECT id, agent_id, timestamp, action_type, description, details
		FROM actions
		WHERE agent_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := db.conn.Query(query, agentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions: %w", err)
	}
	defer rows.Close()

	var actions []*Action
	for rows.Next() {
		var action Action
		err := rows.Scan(
			&action.ID,
			&action.AgentID,
			&action.Timestamp,
			&action.ActionType,
			&action.Description,
			&action.Details,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan action: %w", err)
		}
		actions = append(actions, &action)
	}

	return actions, nil
}

// Commit CRUD operations

// CreateCommits logs multiple git commits for an agent
func (db *DB) CreateCommits(agentID string, commitHashes []string, commitMessages []string) error {
	// Start transaction
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

// GetCommitsForAgent retrieves all commits for a specific agent
func (db *DB) GetCommitsForAgent(agentID string) ([]*Commit, error) {
	query := `
		SELECT id, agent_id, commit_hash, commit_message, timestamp
		FROM commits
		WHERE agent_id = ?
		ORDER BY timestamp DESC
	`

	rows, err := db.conn.Query(query, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}
	defer rows.Close()

	var commits []*Commit
	for rows.Next() {
		var commit Commit
		err := rows.Scan(
			&commit.ID,
			&commit.AgentID,
			&commit.CommitHash,
			&commit.CommitMessage,
			&commit.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan commit: %w", err)
		}
		commits = append(commits, &commit)
	}

	return commits, nil
}

// EndAgentSession marks an agent as completed and logs final action
func (db *DB) EndAgentSession(agentID, summary string, finalStatus string) error {
	// Start transaction
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update agent status
	status := finalStatus
	if status == "" {
		status = StatusCompleted
	}

	updateQuery := `
		UPDATE agents
		SET status = ?, updated_at = CURRENT_TIMESTAMP, last_activity_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err = tx.Exec(updateQuery, status, agentID)
	if err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}

	// Log session end action
	if summary != "" {
		details := map[string]string{"summary": summary}
		detailsJSON, _ := json.Marshal(details)

		actionQuery := `
			INSERT INTO actions (agent_id, action_type, description, details)
			VALUES (?, ?, ?, ?)
		`
		_, err = tx.Exec(actionQuery, agentID, ActionStatusUpdate, "Session ended", string(detailsJSON))
		if err != nil {
			return fmt.Errorf("failed to log session end: %w", err)
		}
	}

	return tx.Commit()
}

// GetAgentStats returns basic statistics about agents
func (db *DB) GetAgentStats() (map[string]int, error) {
	query := `
		SELECT status, COUNT(*) as count
		FROM agents
		GROUP BY status
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		err := rows.Scan(&status, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}
		stats[status] = count
	}

	return stats, nil
}