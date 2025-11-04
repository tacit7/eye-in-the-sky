package data

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

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

// LoadAgents loads agents from the database with optional filtering
func (s *agentStore) LoadAgents(ctx context.Context, showAll bool) ([]domain.Agent, error) {
	query := `
		SELECT id, status, source, created_at, updated_at,
		       git_worktree_path, feature_description, current_task,
		       last_activity_at, window_id, terminal_application,
		       description, project_name, session_id, parent_session_id, parent_agent_id
		FROM agents`

	if !showAll {
		query += `
		WHERE status IN ('active', 'working', 'idle', 'stale', 'unknown')`
	}

	query += `
		ORDER BY
			substr(session_id, 1, 8),
			CASE WHEN parent_agent_id IS NULL THEN id ELSE parent_agent_id END,
			CASE WHEN parent_agent_id IS NULL THEN 0 ELSE 1 END
	`

	log.Printf("[AGENTS] Executing LoadAgents query (showAll=%v)", showAll)
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("[AGENTS] Query error: %v", err)
		return nil, fmt.Errorf("query agents: %w", err)
	}
	defer rows.Close()
	log.Printf("[AGENTS] Query executed successfully")

	var agents []domain.Agent
	for rows.Next() {
		var a domain.Agent
		var gitPath, featureDesc, currentTask, windowID, terminalApp, desc, projectName, sessionID, parentSessionID, parentAgentID sql.NullString
		var lastActivity sql.NullTime

		err := rows.Scan(
			&a.ID, &a.Status, &a.Source, &a.CreatedAt, &a.UpdatedAt,
			&gitPath, &featureDesc, &currentTask, &lastActivity,
			&windowID, &terminalApp, &desc, &projectName, &sessionID, &parentSessionID, &parentAgentID,
		)
		if err != nil {
			log.Printf("[AGENTS] Scan error: %v", err)
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
		if parentSessionID.Valid {
			a.ParentSessionID = parentSessionID.String
		}
		if parentAgentID.Valid {
			a.ParentAgentID = parentAgentID.String
		}
		// Bookmarked and LastLog not loaded in base query
		a.Bookmarked = false
		a.LastLog = ""

		agents = append(agents, a)
	}

	if err := rows.Err(); err != nil {
		log.Printf("[AGENTS] Rows error: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	log.Printf("[AGENTS] Successfully loaded %d agents", len(agents))

	// Load last log timestamp for each agent (separate query)
	logQuery := `SELECT timestamp FROM logs WHERE session_id = ? ORDER BY timestamp DESC LIMIT 1`
	for i := range agents {
		if agents[i].SessionID != "" {
			var lastLogTime sql.NullTime
			err := s.db.QueryRowContext(ctx, logQuery, agents[i].SessionID).Scan(&lastLogTime)
			if err == nil && lastLogTime.Valid {
				// Format timestamp as relative time (e.g., "2m ago", "5h ago")
				agents[i].LastLog = formatRelativeTime(lastLogTime.Time)
			}
		}
	}
	log.Printf("[AGENTS] Loaded last log timestamps for agents")

	return agents, nil
}

// LoadAgent loads a single agent by ID
func (s *agentStore) LoadAgent(ctx context.Context, agentID domain.AgentID) (*domain.Agent, error) {
	query := `
		SELECT id, status, source, created_at, updated_at,
		       git_worktree_path, feature_description, current_task,
		       last_activity_at, window_id, terminal_application,
		       description, project_name, session_id, parent_session_id, parent_agent_id
		FROM agents
		WHERE id = ?
	`

	var a domain.Agent
	var gitPath, featureDesc, currentTask, windowID, terminalApp, desc, projectName, sessionID, parentSessionID, parentAgentID sql.NullString
	var lastActivity sql.NullTime

	err := s.db.QueryRowContext(ctx, query, string(agentID)).Scan(
		&a.ID, &a.Status, &a.Source, &a.CreatedAt, &a.UpdatedAt,
		&gitPath, &featureDesc, &currentTask, &lastActivity,
		&windowID, &terminalApp, &desc, &projectName, &sessionID, &parentSessionID, &parentAgentID,
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
	if parentSessionID.Valid {
		a.ParentSessionID = parentSessionID.String
	}
	if parentAgentID.Valid {
		a.ParentAgentID = parentAgentID.String
	}
	// Bookmarked and LastLog not loaded in base query
	a.Bookmarked = false
	a.LastLog = ""

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

// formatRelativeTime formats a timestamp as relative time (e.g., "2m ago", "5h ago")
func formatRelativeTime(t time.Time) string {
	elapsed := time.Since(t)

	if elapsed < time.Minute {
		return "just now"
	} else if elapsed < time.Hour {
		return fmt.Sprintf("%dm ago", int(elapsed.Minutes()))
	} else if elapsed < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(elapsed.Hours()))
	} else {
		return fmt.Sprintf("%dd ago", int(elapsed.Hours()/24))
	}
}