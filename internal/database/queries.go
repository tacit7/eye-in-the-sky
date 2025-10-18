package database

import (
	"database/sql"
	"fmt"
)

// CreateAgent inserts a new agent
func (db *DB) CreateAgent(agent *Agent) error {
	query := `
		INSERT INTO agents (id, status, source, git_worktree_path, feature_description, current_task, last_activity_at, window_id, project_name)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query, agent.ID, agent.Status, agent.Source, agent.GitWorktreePath,
		agent.FeatureDescription, agent.CurrentTask, agent.LastActivityAt, agent.WindowID, agent.ProjectName)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}
	return nil
}

// GetAgent retrieves an agent by ID
func (db *DB) GetAgent(id string) (*Agent, error) {
	if len(id) != 8 {
		return nil, NewValidationError("agent_id", id, ErrInvalidAgentID)
	}

	query := `
		SELECT id, status, source, created_at, updated_at, git_worktree_path, feature_description, current_task, last_activity_at, window_id, project_name, current_session_id, persona_id
		FROM agents WHERE id = ?
	`
	var agent Agent
	row := db.conn.QueryRow(query, id)
	err := row.Scan(&agent.ID, &agent.Status, &agent.Source, &agent.CreatedAt, &agent.UpdatedAt,
		&agent.GitWorktreePath, &agent.FeatureDescription, &agent.CurrentTask, &agent.LastActivityAt, &agent.WindowID, &agent.ProjectName, &agent.CurrentSessionID, &agent.PersonaID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, NewAgentError(id, "get", ErrAgentNotFound)
		}
		return nil, NewAgentError(id, "get", err)
	}
	return &agent, nil
}

// UpdateAgentStatus updates an agent's status
func (db *DB) UpdateAgentStatus(id, status string, currentTask *string) error {
	query := `
		UPDATE agents SET status = ?, current_task = ?, last_activity_at = CURRENT_TIMESTAMP
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

	query := `UPDATE agents SET status = ?, last_activity_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.conn.Exec(query, status, agentID)
	if err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}

	return nil
}

// GetActionsForAgent retrieves all actions for a specific agent with limit
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
// ListAgents retrieves all agents, optionally filtered by status
func (db *DB) ListAgents(status string) ([]*Agent, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `
			SELECT id, status, source, created_at, updated_at, git_worktree_path, feature_description, current_task, last_activity_at, window_id, project_name
			FROM agents WHERE status = ?
			ORDER BY updated_at DESC
		`
		args = append(args, status)
	} else {
		query = `
			SELECT id, status, source, created_at, updated_at, git_worktree_path, feature_description, current_task, last_activity_at, window_id, project_name
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
			&agent.Source,
			&agent.CreatedAt,
			&agent.UpdatedAt,
			&agent.GitWorktreePath,
			&agent.FeatureDescription,
			&agent.CurrentTask,
			&agent.LastActivityAt,
			&agent.WindowID,
			&agent.ProjectName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}
		agents = append(agents, &agent)
	}

	return agents, nil
}

// ListActions retrieves all actions for a specific agent
func (db *DB) ListActions(agentID string) ([]*Action, error) {
	query := `
		SELECT id, agent_id, action_type, description, details, timestamp
		FROM actions
		WHERE agent_id = ?
		ORDER BY timestamp DESC
	`
	rows, err := db.conn.Query(query, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list actions: %w", err)
	}
	defer rows.Close()

	var actions []*Action
	for rows.Next() {
		var action Action
		err := rows.Scan(
			&action.ID,
			&action.AgentID,
			&action.ActionType,
			&action.Description,
			&action.Details,
			&action.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan action: %w", err)
		}
		actions = append(actions, &action)
	}

	return actions, nil
}

// ListCommits retrieves all commits for a specific agent
func (db *DB) ListCommits(agentID string) ([]*Commit, error) {
	query := `
		SELECT id, agent_id, commit_hash, commit_message, timestamp
		FROM commits
		WHERE agent_id = ?
		ORDER BY timestamp DESC
	`
	rows, err := db.conn.Query(query, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list commits: %w", err)
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

// CreateSession inserts a new session
func (db *DB) CreateSession(session *Session) error {
	query := `INSERT INTO sessions (id, agent_id, name, started_at) VALUES (?, ?, ?, ?)`
	_, err := db.conn.Exec(query, session.ID, session.AgentID, session.Name, session.StartedAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// GetSession retrieves a session by ID
func (db *DB) GetSession(id string) (*Session, error) {
	query := `SELECT id, agent_id, name, started_at, ended_at FROM sessions WHERE id = ?`
	var session Session
	row := db.conn.QueryRow(query, id)
	err := row.Scan(&session.ID, &session.AgentID, &session.Name, &session.StartedAt, &session.EndedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &session, nil
}

// CreateLog inserts a new log entry
func (db *DB) CreateLog(log *Log) error {
	query := `INSERT INTO logs (session_id, type, message, timestamp) VALUES (?, ?, ?, ?)`
	_, err := db.conn.Exec(query, log.SessionID, log.Type, log.Message, log.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to create log: %w", err)
	}
	return nil
}

// GetLogs retrieves all logs for a session
func (db *DB) GetLogs(sessionID string) ([]*Log, error) {
	query := `SELECT id, session_id, type, message, timestamp FROM logs WHERE session_id = ? ORDER BY timestamp ASC`
	rows, err := db.conn.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	defer rows.Close()

	var logs []*Log
	for rows.Next() {
		var log Log
		err := rows.Scan(&log.ID, &log.SessionID, &log.Type, &log.Message, &log.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log: %w", err)
		}
		logs = append(logs, &log)
	}
	return logs, nil
}

// CreateNote inserts a new note
func (db *DB) CreateNote(note *Note) error {
	query := `INSERT INTO notes (session_id, content, created_at) VALUES (?, ?, ?)`
	_, err := db.conn.Exec(query, note.SessionID, note.Content, note.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}
	return nil
}

// GetNotes retrieves all notes for a session
func (db *DB) GetNotes(sessionID string) ([]*Note, error) {
	query := `SELECT id, session_id, content, created_at FROM notes WHERE session_id = ? ORDER BY created_at ASC`
	rows, err := db.conn.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notes: %w", err)
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.SessionID, &note.Content, &note.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		notes = append(notes, &note)
	}
	return notes, nil
}

// SetContext sets or updates a context key-value pair
func (db *DB) SetContext(sessionID, key, value string) error {
	query := `INSERT OR REPLACE INTO context (session_id, key, value) VALUES (?, ?, ?)`
	_, err := db.conn.Exec(query, sessionID, key, value)
	if err != nil {
		return fmt.Errorf("failed to set context: %w", err)
	}
	return nil
}

// GetContext retrieves all context key-value pairs for a session
func (db *DB) GetContext(sessionID string) (map[string]string, error) {
	query := `SELECT key, value FROM context WHERE session_id = ?`
	rows, err := db.conn.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get context: %w", err)
	}
	defer rows.Close()

	context := make(map[string]string)
	for rows.Next() {
		var key, value string
		err := rows.Scan(&key, &value)
		if err != nil {
			return nil, fmt.Errorf("failed to scan context: %w", err)
		}
		context[key] = value
	}
	return context, nil
}

// UpdateAgentCurrentSession sets the current session for an agent
func (db *DB) UpdateAgentCurrentSession(agentID, sessionID string) error {
	query := `UPDATE agents SET current_session_id = ? WHERE id = ?`
	_, err := db.conn.Exec(query, sessionID, agentID)
	if err != nil {
		return fmt.Errorf("failed to update current session: %w", err)
	}
	return nil
}

// GetCurrentSessionForAgent retrieves the current session for an agent
func (db *DB) GetCurrentSessionForAgent(agentID string) (*Session, error) {
	agent, err := db.GetAgent(agentID)
	if err != nil {
		return nil, err
	}

	if agent.CurrentSessionID == nil || *agent.CurrentSessionID == "" {
		return nil, fmt.Errorf("no active session for agent %s", agentID)
	}

	return db.GetSession(*agent.CurrentSessionID)
}

// CreatePersona inserts a new persona
func (db *DB) CreatePersona(persona *Persona) error {
	query := `
		INSERT INTO personas (id, name, description, expertise, initial_context, preferred_tools, specialization)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query, persona.ID, persona.Name, persona.Description,
		persona.Expertise, persona.InitialContext, persona.PreferredTools, persona.Specialization)
	if err != nil {
		return fmt.Errorf("failed to create persona: %w", err)
	}
	return nil
}

// GetPersona retrieves a persona by ID
func (db *DB) GetPersona(id string) (*Persona, error) {
	query := `
		SELECT id, name, description, expertise, initial_context, preferred_tools, specialization, created_at, updated_at
		FROM personas WHERE id = ?
	`
	var persona Persona
	row := db.conn.QueryRow(query, id)
	err := row.Scan(&persona.ID, &persona.Name, &persona.Description, &persona.Expertise,
		&persona.InitialContext, &persona.PreferredTools, &persona.Specialization,
		&persona.CreatedAt, &persona.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("persona not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get persona: %w", err)
	}
	return &persona, nil
}

// ListPersonas retrieves all personas, optionally filtered by specialization
func (db *DB) ListPersonas(specialization string) ([]*Persona, error) {
	var query string
	var args []interface{}

	if specialization != "" {
		query = `
			SELECT id, name, description, expertise, initial_context, preferred_tools, specialization, created_at, updated_at
			FROM personas WHERE specialization = ?
			ORDER BY name ASC
		`
		args = append(args, specialization)
	} else {
		query = `
			SELECT id, name, description, expertise, initial_context, preferred_tools, specialization, created_at, updated_at
			FROM personas
			ORDER BY name ASC
		`
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list personas: %w", err)
	}
	defer rows.Close()

	var personas []*Persona
	for rows.Next() {
		var persona Persona
		err := rows.Scan(
			&persona.ID,
			&persona.Name,
			&persona.Description,
			&persona.Expertise,
			&persona.InitialContext,
			&persona.PreferredTools,
			&persona.Specialization,
			&persona.CreatedAt,
			&persona.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan persona: %w", err)
		}
		personas = append(personas, &persona)
	}

	return personas, nil
}

// UpdatePersona updates an existing persona
func (db *DB) UpdatePersona(persona *Persona) error {
	query := `
		UPDATE personas
		SET name = ?, description = ?, expertise = ?, initial_context = ?,
		    preferred_tools = ?, specialization = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	result, err := db.conn.Exec(query, persona.Name, persona.Description, persona.Expertise,
		persona.InitialContext, persona.PreferredTools, persona.Specialization, persona.ID)
	if err != nil {
		return fmt.Errorf("failed to update persona: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("persona not found: %s", persona.ID)
	}

	return nil
}

// DeletePersona removes a persona
func (db *DB) DeletePersona(id string) error {
	query := `DELETE FROM personas WHERE id = ?`
	result, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete persona: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("persona not found: %s", id)
	}

	return nil
}

// UpdateAgentPersona associates a persona with an agent
func (db *DB) UpdateAgentPersona(agentID string, personaID string) error {
	query := `UPDATE agents SET persona_id = ? WHERE id = ?`
	result, err := db.conn.Exec(query, personaID, agentID)
	if err != nil {
		return fmt.Errorf("failed to update agent persona: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	return nil
}

// ListSessions retrieves all sessions with their associated agent information
func (db *DB) ListSessions(agentID string, activeOnly bool) ([]*SessionWithAgent, error) {
	var query string
	var args []interface{}

	if agentID != "" {
		if activeOnly {
			query = `
				SELECT s.id, s.agent_id, s.name, s.started_at, s.ended_at,
				       a.status, a.feature_description, a.current_task, a.project_name
				FROM sessions s
				JOIN agents a ON s.agent_id = a.id
				WHERE s.agent_id = ? AND s.ended_at IS NULL
				ORDER BY s.started_at DESC
			`
		} else {
			query = `
				SELECT s.id, s.agent_id, s.name, s.started_at, s.ended_at,
				       a.status, a.feature_description, a.current_task, a.project_name
				FROM sessions s
				JOIN agents a ON s.agent_id = a.id
				WHERE s.agent_id = ?
				ORDER BY s.started_at DESC
			`
		}
		args = append(args, agentID)
	} else {
		if activeOnly {
			query = `
				SELECT s.id, s.agent_id, s.name, s.started_at, s.ended_at,
				       a.status, a.feature_description, a.current_task, a.project_name
				FROM sessions s
				JOIN agents a ON s.agent_id = a.id
				WHERE s.ended_at IS NULL
				ORDER BY s.started_at DESC
			`
		} else {
			query = `
				SELECT s.id, s.agent_id, s.name, s.started_at, s.ended_at,
				       a.status, a.feature_description, a.current_task, a.project_name
				FROM sessions s
				JOIN agents a ON s.agent_id = a.id
				ORDER BY s.started_at DESC
			`
		}
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*SessionWithAgent
	for rows.Next() {
		var session SessionWithAgent
		err := rows.Scan(
			&session.ID,
			&session.AgentID,
			&session.Name,
			&session.StartedAt,
			&session.EndedAt,
			&session.AgentStatus,
			&session.FeatureDescription,
			&session.CurrentTask,
			&session.ProjectName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}
