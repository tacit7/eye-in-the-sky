package database

import (
	"context"
	"database/sql"
	"fmt"
)

// Tx wraps a database transaction with our custom methods
type Tx struct {
	tx *sql.Tx
}

// BeginTx starts a new database transaction
func (db *DB) BeginTx(ctx context.Context) (*Tx, error) {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &Tx{tx: tx}, nil
}

// Commit commits the transaction
func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

// Rollback rolls back the transaction
func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}

// Exec executes a statement within the transaction without returning rows.
func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.tx.Exec(query, args...)
}

// Query executes a query within the transaction that returns rows.
func (tx *Tx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tx.tx.Query(query, args...)
}

// QueryRow executes a query within the transaction that returns a single row.
func (tx *Tx) QueryRow(query string, args ...interface{}) *sql.Row {
	return tx.tx.QueryRow(query, args...)
}

// CreateAgentTx creates an agent within a transaction
func (tx *Tx) CreateAgentTx(agent *Agent) error {
	query := `
		INSERT INTO agents (id, status, source, git_worktree_path, feature_description, current_task, last_activity_at, window_id)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)
	`
	_, err := tx.tx.Exec(query, agent.ID, agent.Status, agent.Source, agent.GitWorktreePath,
		agent.FeatureDescription, agent.CurrentTask, agent.WindowID)
	if err != nil {
		return fmt.Errorf("failed to create agent in transaction: %w", err)
	}
	return nil
}

// CreateActionTx creates an action within a transaction
func (tx *Tx) CreateActionTx(action *Action) error {
	query := `INSERT INTO actions (agent_id, action_type, description, details) VALUES (?, ?, ?, ?)`
	_, err := tx.tx.Exec(query, action.AgentID, action.ActionType, action.Description, action.Details)
	if err != nil {
		return fmt.Errorf("failed to create action in transaction: %w", err)
	}

	// Update agent's last activity within the same transaction
	updateQuery := `UPDATE agents SET last_activity_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err = tx.tx.Exec(updateQuery, action.AgentID)
	if err != nil {
		return fmt.Errorf("failed to update agent activity in transaction: %w", err)
	}

	return nil
}

// UpdateAgentStatusTx updates agent status within a transaction
func (tx *Tx) UpdateAgentStatusTx(id, status string, currentTask *string) error {
	query := `
		UPDATE agents SET status = ?, current_task = ?, last_activity_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	result, err := tx.tx.Exec(query, status, currentTask, id)
	if err != nil {
		return fmt.Errorf("failed to update agent status in transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected in transaction: %w", err)
	}

	if rowsAffected == 0 {
		return NewAgentError(id, "update", ErrAgentNotFound)
	}

	return nil
}

// CreateCommitsTx creates commits within a transaction
func (tx *Tx) CreateCommitsTx(agentID string, commitHashes []string, commitMessages []string) error {
	query := `INSERT INTO commits (agent_id, commit_hash, commit_message) VALUES (?, ?, ?)`
	for i, hash := range commitHashes {
		var message *string
		if i < len(commitMessages) && commitMessages[i] != "" {
			message = &commitMessages[i]
		}
		_, err := tx.tx.Exec(query, agentID, hash, message)
		if err != nil {
			return fmt.Errorf("failed to insert commit %s in transaction: %w", hash, err)
		}
	}

	// Update agent's last activity within the same transaction
	updateQuery := `UPDATE agents SET last_activity_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := tx.tx.Exec(updateQuery, agentID)
	if err != nil {
		return fmt.Errorf("failed to update agent activity in transaction: %w", err)
	}

	return nil
}

// WithTransaction executes a function within a database transaction
// Automatically handles commit/rollback
func (db *DB) WithTransaction(ctx context.Context, fn func(*Tx) error) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Safe to call even after commit

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

// RegisterAgentWithAction atomically registers an agent and logs initial action
func (db *DB) RegisterAgentWithAction(ctx context.Context, agent *Agent, initialAction *Action) error {
	return db.WithTransaction(ctx, func(tx *Tx) error {
		// Create agent first
		if err := tx.CreateAgentTx(agent); err != nil {
			return err
		}

		// Log initial action
		if err := tx.CreateActionTx(initialAction); err != nil {
			return err
		}

		return nil
	})
}

// UpdateStatusWithAction atomically updates agent status and logs action
func (db *DB) UpdateStatusWithAction(ctx context.Context, agentID, status string, currentTask *string, action *Action) error {
	return db.WithTransaction(ctx, func(tx *Tx) error {
		// Update agent status first
		if err := tx.UpdateAgentStatusTx(agentID, status, currentTask); err != nil {
			return err
		}

		// Log action
		if err := tx.CreateActionTx(action); err != nil {
			return err
		}

		return nil
	})
}
