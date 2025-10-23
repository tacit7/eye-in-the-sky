package database

import (
	"database/sql"
	"fmt"
	"strings"
)

// repeatPlaceholders creates a comma-separated string of ? placeholders
func repeatPlaceholders(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.Repeat("?,", count-1) + "?"
}

// CreateAgent inserts a new agent
func (db *DB) CreateAgent(agent *Agent) error {
	query := `
		INSERT INTO agents (id, status, source, git_worktree_path, feature_description, current_task, last_activity_at, window_id, terminal_application, project_name, session_id, parent_agent_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query, agent.ID, agent.Status, agent.Source, agent.GitWorktreePath,
		agent.FeatureDescription, agent.CurrentTask, agent.LastActivityAt, agent.WindowID, agent.TerminalApplication, agent.ProjectName, agent.SessionID, agent.ParentAgentID)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}
	return nil
}

// GetAgent retrieves an agent by ID
func (db *DB) GetAgent(id string) (*Agent, error) {
	// No longer validating agent ID format - accepting UUIDs now

	query := `
		SELECT id, status, source, description, created_at, updated_at, git_worktree_path, feature_description, current_task, last_activity_at, window_id, project_name, session_id, persona_id, parent_agent_id
		FROM agents WHERE id = ?
	`
	var agent Agent
	row := db.conn.QueryRow(query, id)
	err := row.Scan(&agent.ID, &agent.Status, &agent.Source, &agent.Description, &agent.CreatedAt, &agent.UpdatedAt,
		&agent.GitWorktreePath, &agent.FeatureDescription, &agent.CurrentTask, &agent.LastActivityAt, &agent.WindowID, &agent.ProjectName, &agent.SessionID, &agent.PersonaID, &agent.ParentAgentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, NewAgentError(id, "get", ErrAgentNotFound)
		}
		return nil, NewAgentError(id, "get", err)
	}
	return &agent, nil
}

// GetAgentBySessionID retrieves the most recent agent for a session ID
func (db *DB) GetAgentBySessionID(sessionID string) (*Agent, error) {
	query := `
		SELECT id, status, source, description, created_at, updated_at, git_worktree_path, feature_description, current_task, last_activity_at, window_id, project_name, session_id, persona_id, parent_agent_id
		FROM agents WHERE session_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`
	var agent Agent
	row := db.conn.QueryRow(query, sessionID)
	err := row.Scan(&agent.ID, &agent.Status, &agent.Source, &agent.Description, &agent.CreatedAt, &agent.UpdatedAt,
		&agent.GitWorktreePath, &agent.FeatureDescription, &agent.CurrentTask, &agent.LastActivityAt, &agent.WindowID, &agent.ProjectName, &agent.SessionID, &agent.PersonaID, &agent.ParentAgentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil without error if not found
		}
		return nil, fmt.Errorf("failed to get agent by session ID: %w", err)
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

// UpdateAgentDescription updates an agent's description
func (db *DB) UpdateAgentDescription(id string, description string) error {
	query := `
		UPDATE agents SET description = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	result, err := db.conn.Exec(query, description, id)
	if err != nil {
		return fmt.Errorf("failed to update agent description: %w", err)
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

// GetChildAgentIDs recursively retrieves all descendant agent IDs for a given agent
func (db *DB) GetChildAgentIDs(agentID string) ([]string, error) {
	query := `
		WITH RECURSIVE descendants AS (
			SELECT id FROM agents WHERE parent_agent_id = ?
			UNION ALL
			SELECT a.id FROM agents a
			INNER JOIN descendants d ON a.parent_agent_id = d.id
		)
		SELECT id FROM descendants
	`

	rows, err := db.conn.Query(query, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get child agent IDs: %w", err)
	}
	defer rows.Close()

	var childIDs []string
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to scan child agent ID: %w", err)
		}
		childIDs = append(childIDs, id)
	}

	return childIDs, nil
}

// GetCommitsForAgentHierarchy retrieves commits for an agent and all its descendants
func (db *DB) GetCommitsForAgentHierarchy(agentID string, limit int) ([]*Commit, error) {
	// Get all child agent IDs
	childIDs, err := db.GetChildAgentIDs(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get child agents: %w", err)
	}

	// Build list of all agent IDs (parent + children)
	allAgentIDs := []string{agentID}
	allAgentIDs = append(allAgentIDs, childIDs...)

	// Build the query with placeholders for all agent IDs
	query := `
		SELECT id, agent_id, commit_hash, commit_message, timestamp
		FROM commits
		WHERE agent_id IN (` + repeatPlaceholders(len(allAgentIDs)) + `)
		ORDER BY timestamp DESC`

	if limit > 0 {
		query += ` LIMIT ?`
	}

	// Convert agent IDs to interface slice for variadic argument
	args := make([]interface{}, len(allAgentIDs))
	for i, id := range allAgentIDs {
		args[i] = id
	}
	if limit > 0 {
		args = append(args, limit)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits for agent hierarchy: %w", err)
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
		// Special handling for "active" filter - show all active sessions
		if status == "active" {
			query = `
				SELECT id, status, source, description, created_at, updated_at, git_worktree_path, feature_description, current_task, last_activity_at, window_id, project_name, session_id, persona_id
				FROM agents WHERE status IN ('active', 'working', 'idle')
				ORDER BY updated_at DESC
			`
		} else {
			query = `
				SELECT id, status, source, description, created_at, updated_at, git_worktree_path, feature_description, current_task, last_activity_at, window_id, project_name, session_id, persona_id
				FROM agents WHERE status = ?
				ORDER BY updated_at DESC
			`
			args = append(args, status)
		}
	} else {
		query = `
			SELECT id, status, source, description, created_at, updated_at, git_worktree_path, feature_description, current_task, last_activity_at, window_id, project_name, session_id, persona_id
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
			&agent.Description,
			&agent.CreatedAt,
			&agent.UpdatedAt,
			&agent.GitWorktreePath,
			&agent.FeatureDescription,
			&agent.CurrentTask,
			&agent.LastActivityAt,
			&agent.WindowID,
			&agent.ProjectName,
			&agent.SessionID,
			&agent.PersonaID,
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
	query := `UPDATE agents SET session_id = ? WHERE id = ?`
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

	if agent.SessionID == nil || *agent.SessionID == "" {
		return nil, fmt.Errorf("no active session for agent %s", agentID)
	}

	return db.GetSession(*agent.SessionID)
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

// CreateCompaction logs a conversation compaction event
func (db *DB) CreateCompaction(compaction *Compaction) error {
	query := `
		INSERT INTO compactions (agent_id, session_id, summary, jsonl_file_path, jsonl_file_size, message_count)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query,
		compaction.AgentID,
		compaction.SessionID,
		compaction.Summary,
		compaction.JsonlFilePath,
		compaction.JsonlFileSize,
		compaction.MessageCount,
	)
	if err != nil {
		return fmt.Errorf("failed to create compaction: %w", err)
	}
	return nil
}

// GetCompactionsForAgent retrieves all compactions for a specific agent
func (db *DB) GetCompactionsForAgent(agentID string) ([]*Compaction, error) {
	query := `
		SELECT id, agent_id, session_id, compacted_at, summary, jsonl_file_path, jsonl_file_size, message_count
		FROM compactions
		WHERE agent_id = ?
		ORDER BY compacted_at DESC
	`

	rows, err := db.conn.Query(query, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get compactions: %w", err)
	}
	defer rows.Close()

	var compactions []*Compaction
	for rows.Next() {
		var c Compaction
		err := rows.Scan(
			&c.ID,
			&c.AgentID,
			&c.SessionID,
			&c.CompactedAt,
			&c.Summary,
			&c.JsonlFilePath,
			&c.JsonlFileSize,
			&c.MessageCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan compaction: %w", err)
		}
		compactions = append(compactions, &c)
	}

	return compactions, nil
}

// GetActionsByType retrieves all actions for a specific agent filtered by action type
func (db *DB) GetActionsByType(agentID string, actionType string) ([]*Action, error) {
	query := `
		SELECT id, agent_id, action_type, description, details, timestamp
		FROM actions
		WHERE agent_id = ? AND action_type = ?
		ORDER BY timestamp ASC
	`
	rows, err := db.conn.Query(query, agentID, actionType)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions by type: %w", err)
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

// GetSessionContextsForAgent retrieves all saved session contexts for a specific agent
func (db *DB) GetSessionContextsForAgent(agentID string) ([]*SessionContext, error) {
	query := `
		SELECT id, agent_id, session_id, created_at, updated_at, current_phase,
		       overall_progress, pending_tasks, completed_tasks, next_actions,
		       dependencies, important_files, milestones, current_goals, blockers,
		       key_decisions, environment, metrics, auto_save, learned_context
		FROM session_context
		WHERE agent_id = ?
		ORDER BY created_at DESC
	`

	rows, err := db.conn.Query(query, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session contexts: %w", err)
	}
	defer rows.Close()

	var contexts []*SessionContext
	for rows.Next() {
		var ctx SessionContext
		err := rows.Scan(
			&ctx.ID,
			&ctx.AgentID,
			&ctx.SessionID,
			&ctx.CreatedAt,
			&ctx.UpdatedAt,
			&ctx.CurrentPhase,
			&ctx.OverallProgress,
			&ctx.PendingTasks,
			&ctx.CompletedTasks,
			&ctx.NextActions,
			&ctx.Dependencies,
			&ctx.ImportantFiles,
			&ctx.Milestones,
			&ctx.CurrentGoals,
			&ctx.Blockers,
			&ctx.KeyDecisions,
			&ctx.Environment,
			&ctx.Metrics,
			&ctx.AutoSave,
			&ctx.LearnedContext,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session context: %w", err)
		}
		contexts = append(contexts, &ctx)
	}

	return contexts, nil
}

// CreateSessionContext inserts or updates a session context checkpoint
func (db *DB) CreateSessionContext(ctx *SessionContext) error {
	query := `
		INSERT INTO session_context (
			agent_id, session_id, current_phase, overall_progress,
			pending_tasks, completed_tasks, next_actions, dependencies,
			important_files, milestones, current_goals, blockers,
			key_decisions, environment, metrics, auto_save, learned_context
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(query,
		ctx.AgentID,
		ctx.SessionID,
		ctx.CurrentPhase,
		ctx.OverallProgress,
		ctx.PendingTasks,
		ctx.CompletedTasks,
		ctx.NextActions,
		ctx.Dependencies,
		ctx.ImportantFiles,
		ctx.Milestones,
		ctx.CurrentGoals,
		ctx.Blockers,
		ctx.KeyDecisions,
		ctx.Environment,
		ctx.Metrics,
		ctx.AutoSave,
		ctx.LearnedContext,
	)

	if err != nil {
		return fmt.Errorf("failed to create session context: %w", err)
	}

	return nil
}

// LogSessionMetrics logs token usage and cost information for a session
func (db *DB) LogSessionMetrics(metrics *SessionMetrics) error {
	query := `
		INSERT INTO session_metrics (
			agent_id, session_id, tokens_used, tokens_budget, tokens_remaining,
			input_tokens, output_tokens, estimated_cost_usd, model_name, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(
		query,
		metrics.AgentID,
		metrics.SessionID,
		metrics.TokensUsed,
		metrics.TokensBudget,
		metrics.TokensRemaining,
		metrics.InputTokens,
		metrics.OutputTokens,
		metrics.EstimatedCostUSD,
		metrics.ModelName,
		metrics.Notes,
	)

	if err != nil {
		return fmt.Errorf("failed to log session metrics: %w", err)
	}

	return nil
}

// GetSessionMetrics retrieves all metrics for a specific agent
func (db *DB) GetSessionMetrics(agentID string, limit int) ([]*SessionMetrics, error) {
	query := `
		SELECT id, agent_id, session_id, tokens_used, tokens_budget, tokens_remaining,
		       input_tokens, output_tokens, estimated_cost_usd, model_name, timestamp, created_at, notes
		FROM session_metrics
		WHERE agent_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := db.conn.Query(query, agentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get session metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*SessionMetrics
	for rows.Next() {
		var m SessionMetrics
		err := rows.Scan(
			&m.ID,
			&m.AgentID,
			&m.SessionID,
			&m.TokensUsed,
			&m.TokensBudget,
			&m.TokensRemaining,
			&m.InputTokens,
			&m.OutputTokens,
			&m.EstimatedCostUSD,
			&m.ModelName,
			&m.Timestamp,
			&m.CreatedAt,
			&m.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session metrics: %w", err)
		}
		metrics = append(metrics, &m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating session metrics: %w", err)
	}

	return metrics, nil
}

// GetMonthlyCosts retrieves all session metrics from the current month with timestamps
func (db *DB) GetMonthlyCosts() ([]*SessionMetrics, error) {
	query := `
		SELECT id, agent_id, session_id, tokens_used, tokens_budget, tokens_remaining,
		       input_tokens, output_tokens, estimated_cost_usd, model_name, timestamp, created_at, notes
		FROM session_metrics
		WHERE strftime('%Y-%m', timestamp) = strftime('%Y-%m', 'now')
		ORDER BY timestamp DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly costs: %w", err)
	}
	defer rows.Close()

	var metrics []*SessionMetrics
	for rows.Next() {
		var m SessionMetrics
		var sessionID, modelName, notes sql.NullString
		var inputTokens, outputTokens sql.NullInt64
		var estimatedCost sql.NullFloat64

		err := rows.Scan(
			&m.ID,
			&m.AgentID,
			&sessionID,
			&m.TokensUsed,
			&m.TokensBudget,
			&m.TokensRemaining,
			&inputTokens,
			&outputTokens,
			&estimatedCost,
			&modelName,
			&m.Timestamp,
			&m.CreatedAt,
			&notes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monthly costs: %w", err)
		}

		if sessionID.Valid {
			m.SessionID = &sessionID.String
		}
		if inputTokens.Valid {
			val := int(inputTokens.Int64)
			m.InputTokens = &val
		}
		if outputTokens.Valid {
			val := int(outputTokens.Int64)
			m.OutputTokens = &val
		}
		if estimatedCost.Valid {
			m.EstimatedCostUSD = &estimatedCost.Float64
		}
		if modelName.Valid {
			m.ModelName = &modelName.String
		}
		if notes.Valid {
			m.Notes = &notes.String
		}

		metrics = append(metrics, &m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating monthly costs: %w", err)
	}

	return metrics, nil
}
