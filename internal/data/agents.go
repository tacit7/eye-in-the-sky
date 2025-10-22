package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// agentStore implements SQL operations for agents
type agentStore struct {
	db *sql.DB
}

// NewAgentStore creates a new agent store
func NewAgentStore(db *sql.DB) *agentStore {
	return &agentStore{db: db}
}

// LoadAgents loads all active agents from the database
func (s *agentStore) LoadAgents(ctx context.Context) ([]domain.Agent, error) {
	query := `
		SELECT id, status, source, created_at, updated_at,
		       git_worktree_path, feature_description, current_task,
		       last_activity_at, window_id, terminal_application,
		       description, project_name, session_id, parent_agent_id
		FROM agents
		WHERE status IN ('active', 'working', 'idle', 'stale', 'unknown')
		ORDER BY
			CASE status
				WHEN 'active' THEN 0
				WHEN 'working' THEN 1
				WHEN 'idle' THEN 2
				WHEN 'stale' THEN 3
				WHEN 'unknown' THEN 999
				ELSE 4
			END,
			CASE WHEN parent_agent_id IS NULL THEN id ELSE parent_agent_id END,
			CASE WHEN parent_agent_id IS NULL THEN 0 ELSE 1 END,
			last_activity_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query agents: %w", err)
	}
	defer rows.Close()

	var agents []domain.Agent
	for rows.Next() {
		var a domain.Agent
		var gitPath, featureDesc, currentTask, windowID, terminalApp, desc, projectName, sessionID, parentAgentID sql.NullString
		var lastActivity sql.NullTime

		err := rows.Scan(
			&a.ID, &a.Status, &a.Source, &a.CreatedAt, &a.UpdatedAt,
			&gitPath, &featureDesc, &currentTask, &lastActivity,
			&windowID, &terminalApp, &desc, &projectName, &sessionID, &parentAgentID,
		)
		if err != nil {
			return nil, fmt.Errorf("scan agent: %w", err)
		}

		// Handle nullable fields
		if gitPath.Valid {
			a.GitWorktreePath = gitPath.String
		}
		if featureDesc.Valid {
			a.FeatureDesc = featureDesc.String
		}
		if currentTask.Valid {
			a.CurrentTask = currentTask.String
		}
		if lastActivity.Valid {
			a.LastActivityAt = lastActivity.Time
		}
		if windowID.Valid {
			a.WindowID = windowID.String
		}
		if terminalApp.Valid {
			a.TerminalApplication = terminalApp.String
		}
		if desc.Valid {
			a.AgentDescription = desc.String
		}
		if projectName.Valid {
			a.ProjectName = projectName.String
		}
		if sessionID.Valid {
			a.SessionID = sessionID.String
		}
		if parentAgentID.Valid {
			a.ParentAgentID = parentAgentID.String
		}

		agents = append(agents, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return agents, nil
}

// LoadAgent loads a single agent by ID
func (s *agentStore) LoadAgent(ctx context.Context, agentID domain.AgentID) (*domain.Agent, error) {
	query := `
		SELECT id, status, source, created_at, updated_at,
		       git_worktree_path, feature_description, current_task,
		       last_activity_at, window_id, terminal_application,
		       description, project_name, session_id, parent_agent_id
		FROM agents
		WHERE id = ?
	`

	var a domain.Agent
	var gitPath, featureDesc, currentTask, windowID, terminalApp, desc, projectName, sessionID, parentAgentID sql.NullString
	var lastActivity sql.NullTime

	err := s.db.QueryRowContext(ctx, query, string(agentID)).Scan(
		&a.ID, &a.Status, &a.Source, &a.CreatedAt, &a.UpdatedAt,
		&gitPath, &featureDesc, &currentTask, &lastActivity,
		&windowID, &terminalApp, &desc, &projectName, &sessionID, &parentAgentID,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}
	if err != nil {
		return nil, fmt.Errorf("query agent: %w", err)
	}

	// Handle nullable fields
	if gitPath.Valid {
		a.GitWorktreePath = gitPath.String
	}
	if featureDesc.Valid {
		a.FeatureDesc = featureDesc.String
	}
	if currentTask.Valid {
		a.CurrentTask = currentTask.String
	}
	if lastActivity.Valid {
		a.LastActivityAt = lastActivity.Time
	}
	if windowID.Valid {
		a.WindowID = windowID.String
	}
	if terminalApp.Valid {
		a.TerminalApplication = terminalApp.String
	}
	if desc.Valid {
		a.AgentDescription = desc.String
	}
	if projectName.Valid {
		a.ProjectName = projectName.String
	}
	if sessionID.Valid {
		a.SessionID = sessionID.String
	}
	if parentAgentID.Valid {
		a.ParentAgentID = parentAgentID.String
	}

	return &a, nil
}

// UpdateStatus updates an agent's status
func (s *agentStore) UpdateStatus(ctx context.Context, agentID domain.AgentID, status string) error {
	query := `UPDATE agents SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	result, err := s.db.ExecContext(ctx, query, status, string(agentID))
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	return nil
}