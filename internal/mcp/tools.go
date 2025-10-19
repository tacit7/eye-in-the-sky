package mcp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/utils"
)

type Tools struct {
	db *database.DB
}

func NewTools(db *database.DB) *Tools {
	return &Tools{db: db}
}

// RegisterAgent implements the register_agent MCP tool
func (t *Tools) RegisterAgent(args RegisterAgentArgs) (RegisterAgentResult, error) {
	// Generate git-style agent ID if not provided
	var agentID string
	if args.AgentID == nil || *args.AgentID == "" {
		agentID = utils.GenerateGitStyleAgentID()
	} else {
		agentID = *args.AgentID
		// Validate provided ID
		if !utils.ValidateAgentID(agentID) {
			return RegisterAgentResult{Success: false, Message: "Agent ID must be exactly 8 hex characters"}, nil
		}
	}

	// Check if agent already exists
	existing, err := t.db.GetAgent(agentID)
	if err == nil && existing != nil {
		return RegisterAgentResult{Success: false, Message: fmt.Sprintf("Agent %s already exists", agentID)}, nil
	}

	// Create new agent and log registration action atomically
	agent := &database.Agent{
		ID:                 agentID,
		Status:             database.StatusActive,
		Source:             database.SourceWorktree,
		Description:        args.AgentDescription,
		GitWorktreePath:    args.WorktreePath,
		FeatureDescription: &args.Description,
		ProjectName:        args.ProjectName,
		LastActivityAt:     timePtr(time.Now()),
	}

	action := &database.Action{
		AgentID:     agentID,
		ActionType:  database.ActionStatusUpdate,
		Description: fmt.Sprintf("Agent registered: %s", args.Description),
	}

	if err := t.db.RegisterAgentWithAction(context.Background(), agent, action); err != nil {
		return RegisterAgentResult{Success: false, Message: fmt.Sprintf("Failed to register agent: %v", err)}, fmt.Errorf("database error: %w", err)
	}

	return RegisterAgentResult{Success: true, Message: fmt.Sprintf("Agent %s registered successfully", agentID)}, nil
}

// RegisterDesktopAgent implements the register_claude_desktop_agent MCP tool
func (t *Tools) RegisterDesktopAgent(args RegisterDesktopAgentArgs) (RegisterDesktopAgentResult, error) {
	// Generate git-style agent ID if not provided
	var agentID string
	if args.AgentID == nil || *args.AgentID == "" {
		agentID = utils.GenerateGitStyleAgentID()
	} else {
		agentID = *args.AgentID
		// Validate provided ID
		if !utils.ValidateAgentID(agentID) {
			return RegisterDesktopAgentResult{Success: false, Message: "Agent ID must be exactly 8 hex characters"}, nil
		}
	}

	// Check if agent already exists
	existing, err := t.db.GetAgent(agentID)
	if err == nil && existing != nil {
		return RegisterDesktopAgentResult{Success: false, Message: fmt.Sprintf("Agent %s already exists", agentID)}, nil
	}

	// Create new Claude Desktop agent and log registration action atomically
	agent := &database.Agent{
		ID:                 agentID,
		Status:             database.StatusActive,
		Source:             database.SourceDesktop,
		Description:        args.AgentDescription,
		GitWorktreePath:    nil, // Desktop agents don't have worktree paths
		FeatureDescription: &args.Description,
		ProjectName:        &args.ProjectName,
		LastActivityAt:     timePtr(time.Now()),
		WindowID:           args.WindowID,
	}

	action := &database.Action{
		AgentID:     agentID,
		ActionType:  database.ActionStatusUpdate,
		Description: fmt.Sprintf("Claude Desktop agent registered: %s (Project: %s)", args.Description, args.ProjectName),
	}

	if err := t.db.RegisterAgentWithAction(context.Background(), agent, action); err != nil {
		return RegisterDesktopAgentResult{Success: false, Message: fmt.Sprintf("Failed to register agent: %v", err)}, fmt.Errorf("database error: %w", err)
	}

	return RegisterDesktopAgentResult{Success: true, Message: fmt.Sprintf("Claude Desktop agent %s registered successfully", agentID)}, nil
}

// UpdateStatus implements the update_status MCP tool
func (t *Tools) UpdateStatus(args UpdateStatusArgs) (UpdateStatusResult, error) {
	validStatuses := map[string]bool{
		database.StatusActive: true, database.StatusIdle: true, database.StatusWorking: true,
		database.StatusCompleted: true, database.StatusFailed: true,
	}

	if !validStatuses[args.Status] {
		return UpdateStatusResult{Success: false, Message: fmt.Sprintf("Invalid status: %s", args.Status)}, nil
	}

	// Update status and log action atomically
	description := fmt.Sprintf("Status updated to: %s", args.Status)
	if args.CurrentTask != nil {
		description += fmt.Sprintf(" (Task: %s)", *args.CurrentTask)
	}

	action := &database.Action{
		AgentID:     args.AgentID,
		ActionType:  database.ActionStatusUpdate,
		Description: description,
	}

	if err := t.db.UpdateStatusWithAction(context.Background(), args.AgentID, args.Status, args.CurrentTask, action); err != nil {
		return UpdateStatusResult{Success: false, Message: fmt.Sprintf("Failed to update status: %v", err)}, nil
	}

	return UpdateStatusResult{Success: true, Message: fmt.Sprintf("Status updated to %s", args.Status)}, nil
}

// LogAction implements the log_action MCP tool
func (t *Tools) LogAction(args LogActionArgs) (LogActionResult, error) {
	validActionTypes := map[string]bool{
		database.ActionTaskStart: true, database.ActionFileOperation: true,
		database.ActionGitCommit: true, database.ActionStatusUpdate: true,
	}

	if !validActionTypes[args.ActionType] {
		return LogActionResult{Success: false, Message: fmt.Sprintf("Invalid action type: %s", args.ActionType)}, nil
	}
	action := &database.Action{
		AgentID:     args.AgentID,
		ActionType:  args.ActionType,
		Description: args.Description,
		Details:     args.Details,
	}

	if err := t.db.CreateAction(action); err != nil {
		return LogActionResult{Success: false, Message: fmt.Sprintf("Failed to log action: %v", err)}, nil
	}

	return LogActionResult{Success: true, Message: "Action logged successfully"}, nil
}

// LogCommits implements the log_commits MCP tool
func (t *Tools) LogCommits(args LogCommitsArgs) (LogCommitsResult, error) {
	var commitHashes []string
	var commitMessages []string
	var err error

	// If no hashes provided, try to get latest commits automatically
	if len(args.CommitHashes) == 0 {
		// Get agent to find worktree path
		agent, err := t.db.GetAgent(args.AgentID)
		if err != nil {
			return LogCommitsResult{Success: false, Message: fmt.Sprintf("Failed to get agent: %v", err)}, nil
		}

		workDir := "."
		if agent.GitWorktreePath != nil {
			workDir = *agent.GitWorktreePath
		}

		// Try to get latest 3 commits
		commitHashes, commitMessages, err = getLatestCommits(3, workDir)
		if err != nil {
			return LogCommitsResult{Success: false, Message: fmt.Sprintf("No commit hashes provided and failed to auto-detect: %v", err)}, nil
		}
	} else {
		// Use provided hashes and messages
		commitHashes = args.CommitHashes
		commitMessages = args.CommitMessages
	}

	// Use transaction-safe method
	err = t.db.WithTransaction(context.Background(), func(tx *database.Tx) error {
		return tx.CreateCommitsTx(args.AgentID, commitHashes, commitMessages)
	})

	if err != nil {
		return LogCommitsResult{Success: false, Message: fmt.Sprintf("Failed to log commits: %v", err)}, nil
	}

	autoDetected := len(args.CommitHashes) == 0
	message := fmt.Sprintf("Logged %d commit(s) successfully", len(commitHashes))
	if autoDetected {
		message += " (auto-detected)"
	}

	return LogCommitsResult{Success: true, Message: message}, nil
}

// SyncCommits implements the sync_commits MCP tool to automatically sync recent git commits
func (t *Tools) SyncCommits(args SyncCommitsArgs) (SyncCommitsResult, error) {
	// Get agent to find worktree path
	agent, err := t.db.GetAgent(args.AgentID)
	if err != nil {
		return SyncCommitsResult{Success: false, Message: fmt.Sprintf("Failed to get agent: %v", err)}, nil
	}

	workDir := "."
	if agent.GitWorktreePath != nil {
		workDir = *agent.GitWorktreePath
	}

	// Default count is 5 if not specified
	count := 5
	if args.Count != nil {
		count = *args.Count
	}

	// Get latest commits
	commitHashes, commitMessages, err := getLatestCommits(count, workDir)
	if err != nil {
		return SyncCommitsResult{Success: false, Message: fmt.Sprintf("Failed to sync commits: %v", err)}, nil
	}

	// Use transaction-safe method
	err = t.db.WithTransaction(context.Background(), func(tx *database.Tx) error {
		return tx.CreateCommitsTx(args.AgentID, commitHashes, commitMessages)
	})

	if err != nil {
		return SyncCommitsResult{Success: false, Message: fmt.Sprintf("Failed to sync commits: %v", err)}, nil
	}

	return SyncCommitsResult{Success: true, Message: fmt.Sprintf("Synced %d recent commit(s) automatically", len(commitHashes))}, nil
}

// EndSession implements the end_session MCP tool
func (t *Tools) EndSession(args EndSessionArgs) (EndSessionResult, error) {
	finalStatus := database.StatusCompleted
	if args.FinalStatus != nil {
		finalStatus = *args.FinalStatus
	}

	validStatuses := map[string]bool{database.StatusCompleted: true, database.StatusFailed: true}
	if !validStatuses[finalStatus] {
		return EndSessionResult{Success: false, Message: fmt.Sprintf("Invalid final status: %s", finalStatus)}, nil
	}
	summary := ""
	if args.Summary != nil {
		summary = *args.Summary
	}

	if err := t.db.EndAgentSession(args.AgentID, summary, finalStatus); err != nil {
		return EndSessionResult{Success: false, Message: fmt.Sprintf("Failed to end session: %v", err)}, nil
	}

	message := fmt.Sprintf("Session ended with status: %s", finalStatus)
	if summary != "" {
		message += fmt.Sprintf(" (Summary: %s)", summary)
	}

	return EndSessionResult{Success: true, Message: message}, nil
}

// GetCurrentWindow implements the get_current_window MCP tool
func (t *Tools) GetCurrentWindow(args GetCurrentWindowArgs) (GetCurrentWindowResult, error) {
	if runtime.GOOS != "darwin" {
		return GetCurrentWindowResult{Success: false, Message: "Window detection only supported on macOS"}, nil
	}

	// Get frontmost application
	appCmd := exec.Command("osascript", "-e", `tell application "System Events" to get name of first application process whose frontmost is true`)
	appOutput, err := appCmd.Output()
	if err != nil {
		return GetCurrentWindowResult{Success: false, Message: fmt.Sprintf("Failed to get current application: %v", err)}, nil
	}
	app := strings.TrimSpace(string(appOutput))

	// Get window properties
	windowCmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "System Events" to tell process "%s" to get {name, position, size} of window 1`, app))
	windowOutput, err := windowCmd.Output()
	if err != nil {
		return GetCurrentWindowResult{Success: false, Message: fmt.Sprintf("Failed to get window properties: %v", err)}, nil
	}

	windowData := strings.TrimSpace(string(windowOutput))
	parts := strings.Split(windowData, ", ")

	windowTitle := "Unknown"
	position := "Unknown"
	size := "Unknown"

	if len(parts) >= 3 {
		windowTitle = parts[0]
		position = fmt.Sprintf("%s, %s", parts[1], parts[2])
		if len(parts) >= 5 {
			size = fmt.Sprintf("%s x %s", parts[3], parts[4])
		}
	}

	// Generate a simple window ID based on app + title
	windowID := fmt.Sprintf("%s_%s", app, strings.ReplaceAll(windowTitle, " ", "_"))

	return GetCurrentWindowResult{
		Success:     true,
		Message:     fmt.Sprintf("Current window detected: %s", app),
		Application: app,
		WindowTitle: windowTitle,
		WindowID:    windowID,
		Position:    position,
		Size:        size,
	}, nil
}

// BringWindowFront implements the bring_window_front MCP tool
func (t *Tools) BringWindowFront(args BringWindowFrontArgs) (BringWindowFrontResult, error) {
	if runtime.GOOS != "darwin" {
		return BringWindowFrontResult{Success: false, Message: "Window management only supported on macOS"}, nil
	}

	// Get agent to find window information
	agent, err := t.db.GetAgent(args.AgentID)
	if err != nil {
		return BringWindowFrontResult{Success: false, Message: fmt.Sprintf("Agent not found: %v", err)}, nil
	}

	// For desktop agents, we can use the window ID if available
	if agent.Source == database.SourceDesktop && agent.WindowID != nil {
		// Parse window ID to get application name
		windowID := *agent.WindowID
		parts := strings.Split(windowID, "_")
		if len(parts) > 0 {
			app := parts[0]

			// Bring application to front
			bringCmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "%s" to activate`, app))
			err := bringCmd.Run()
			if err != nil {
				return BringWindowFrontResult{Success: false, Message: fmt.Sprintf("Failed to bring window to front: %v", err)}, nil
			}

			return BringWindowFrontResult{
				Success: true,
				Message: fmt.Sprintf("Brought %s window to front", app),
			}, nil
		}
	}

	// Fallback: try to bring ghostty to front (since that's what you're using)
	bringCmd := exec.Command("osascript", "-e", `tell application "ghostty" to activate`)
	err = bringCmd.Run()
	if err != nil {
		return BringWindowFrontResult{Success: false, Message: fmt.Sprintf("Failed to bring terminal to front: %v", err)}, nil
	}

	return BringWindowFrontResult{
		Success: true,
		Message: "Brought terminal window to front",
	}, nil
}

// Instructions implements the i-instructions MCP tool
func (t *Tools) Instructions(args InstructionsArgs) (InstructionsResult, error) {
	instructions := `Eye in the Sky - Agent Lifecycle Management

═══════════════════════════════════════════════════════════════
INITIALIZATION - Do this FIRST on every new session
═══════════════════════════════════════════════════════════════

STEP 1: Check for marker file
Run: ls .claude/eye-in-the-sky/session-* 2>/dev/null

STEP 2: Parse agent_id and Claude Code session_id from filename
Format: session-{agentId}-{claudeCodeSessionId}
Example: session-a728dc53-a4ebc315-5577-4c81-b12e-1a7ae3fa907f
  → agent_id: a728dc53 (first 8 characters - REUSE THIS)
  → claude_session_id: a4ebc315-5577-4c81-b12e-1a7ae3fa907f (Claude Code UUID - for reference only)

STEP 3: Call i-start-session with the agent_id
  i-start-session({
    "agent_id": "a728dc53",  ← Use the parsed agent_id for continuity
    "agent_description": "Optional agent name/label",
    "description": "What you'll be working on",
    "project_name": "optional",
    "worktree_path": "optional"
  })

  NOTE: i-start-session will auto-generate a NEW Eye in the Sky session_id
  in format: {agentId}_{timestamp} (e.g., "a728dc53_1760832709")

STEP 4: Delete marker file after successful registration
Run: rm .claude/eye-in-the-sky/session-*

IMPORTANT:
- The agent_id provides continuity across multiple sessions
- Each session gets a fresh Eye in the Sky session_id
- If no marker file exists, omit agent_id and a new one will be auto-generated

═══════════════════════════════════════════════════════════════
WORKFLOW - During your session
═══════════════════════════════════════════════════════════════

i-status - Update your current status
  Statuses: active, working, idle, completed, failed
  Example: i-status({"agent_id": "a728dc53", "status": "working", "current_task": "Building TUI"})

i-action - Log significant activities
  Types: task_start, file_operation, git_commit, status_update
  Example: i-action({"agent_id": "a728dc53", "action_type": "file_operation", "description": "Created dashboard component"})

i-commits - Track git commits
  Example: i-commits({"agent_id": "a728dc53", "commit_hashes": ["abc123f"], "commit_messages": ["Add feature"]})

i-end - End session with summary
  Example: i-end({"agent_id": "a728dc53", "summary": "Completed TUI implementation", "final_status": "completed"})

═══════════════════════════════════════════════════════════════
COMPACTION TRACKING
═══════════════════════════════════════════════════════════════

When Claude detects a conversation compaction (indicated by system message
'This session is being continued from a previous conversation'), call:

i-log-compaction({
  "agent_id": "your-agent-id",
  "session_id": "current-session-id",
  "summary": "compaction summary text",
  "old_session_id": "previous-session-id (if known)"
})

This backs up the JSONL conversation file to data/compactions/ directory.

═══════════════════════════════════════════════════════════════
ADDITIONAL TOOLS
═══════════════════════════════════════════════════════════════

i-save-context - Save session state for resumption
i-load-context - Load previous session context
i-note - Add contextual notes to session
i-persona-get - Get persona details
i-persona-list - List available personas
i-snapshot-expertise - Save current expertise as persona

Dashboard: http://localhost:8080 (if web server running)
TUI: Run 'bin/dashboard' for terminal interface

═══════════════════════════════════════════════════════════════`

	return InstructionsResult{
		Success:      true,
		Message:      "Instructions retrieved successfully",
		Instructions: instructions,
	}, nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// getLatestCommits retrieves the latest git commits from the current directory
func getLatestCommits(count int, workDir string) ([]string, []string, error) {
	if workDir == "" {
		workDir = "."
	}

	// Get commit hashes
	hashCmd := exec.Command("git", "log", "--format=%H", fmt.Sprintf("-%d", count))
	hashCmd.Dir = workDir
	hashOutput, err := hashCmd.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get commit hashes: %w", err)
	}

	hashes := strings.Split(strings.TrimSpace(string(hashOutput)), "\n")
	if len(hashes) == 1 && hashes[0] == "" {
		return nil, nil, fmt.Errorf("no commits found")
	}

	// Get commit messages
	msgCmd := exec.Command("git", "log", "--format=%s", fmt.Sprintf("-%d", count))
	msgCmd.Dir = workDir
	msgOutput, err := msgCmd.Output()
	if err != nil {
		return hashes, nil, fmt.Errorf("failed to get commit messages: %w", err)
	}

	messages := strings.Split(strings.TrimSpace(string(msgOutput)), "\n")

	// Truncate hashes to 8 characters for display
	for i, hash := range hashes {
		if len(hash) > 8 {
			hashes[i] = hash[:8]
		}
	}

	return hashes, messages, nil
}

// POA Spec Tools - Session Management

// StartSession implements the i-start-session tool
func (t *Tools) StartSession(args StartSessionArgs) (StartSessionResult, error) {
	// Generate agent ID if not provided
	var agentID string
	if args.AgentID == nil || *args.AgentID == "" {
		agentID = utils.GenerateGitStyleAgentID()
	} else {
		agentID = *args.AgentID
	}

	// Load persona if provided
	var initialContext string
	var personaID *string
	if args.PersonaID != nil && *args.PersonaID != "" {
		persona, err := t.db.GetPersona(*args.PersonaID)
		if err != nil {
			return StartSessionResult{}, fmt.Errorf("failed to load persona %s: %w", *args.PersonaID, err)
		}
		initialContext = persona.InitialContext
		personaID = args.PersonaID
	}

	// Create agent
	agent := &database.Agent{
		ID:                 agentID,
		Status:             database.StatusActive,
		Source:             database.SourceWorktree,
		Description:        args.AgentDescription,
		GitWorktreePath:    args.WorktreePath,
		FeatureDescription: &args.Description,
		ProjectName:        args.ProjectName,
		PersonaID:          personaID,
		LastActivityAt:     timePtr(time.Now()),
	}

	if err := t.db.CreateAgent(agent); err != nil {
		return StartSessionResult{}, fmt.Errorf("failed to create agent: %w", err)
	}

	// Create session
	sessionID := fmt.Sprintf("%s_%d", agentID, time.Now().Unix())
	session := &database.Session{
		ID:        sessionID,
		AgentID:   agentID,
		Name:      args.Name,
		StartedAt: time.Now(),
	}

	if err := t.db.CreateSession(session); err != nil {
		return StartSessionResult{}, fmt.Errorf("failed to create session: %w", err)
	}

	// Log initial entry
	var logMessage string
	if personaID != nil {
		logMessage = fmt.Sprintf("Started session with persona '%s': %s", *personaID, args.Description)
	} else {
		logMessage = fmt.Sprintf("Started session: %s", args.Description)
	}

	log := &database.Log{
		SessionID: sessionID,
		Type:      "info",
		Message:   logMessage,
		Timestamp: time.Now(),
	}

	if err := t.db.CreateLog(log); err != nil {
		return StartSessionResult{}, fmt.Errorf("failed to create initial log: %w", err)
	}

	// Update agent's current session
	if err := t.db.UpdateAgentCurrentSession(agentID, sessionID); err != nil {
		return StartSessionResult{}, fmt.Errorf("failed to update current session: %w", err)
	}

	var message string
	if personaID != nil {
		message = fmt.Sprintf("Session %s started for agent %s with persona '%s' loaded", sessionID, agentID, *personaID)
	} else {
		message = fmt.Sprintf("Session %s started for agent %s", sessionID, agentID)
	}

	return StartSessionResult{
		Success:        true,
		Message:        message,
		AgentID:        agentID,
		SessionID:      sessionID,
		InitialContext: initialContext,
	}, nil
}

// AddLog implements the i-log tool
func (t *Tools) AddLog(args AddLogArgs) (AddLogResult, error) {
	log := &database.Log{
		SessionID: args.SessionID,
		Type:      args.Type,
		Message:   args.Message,
		Timestamp: time.Now(),
	}

	if err := t.db.CreateLog(log); err != nil {
		return AddLogResult{}, fmt.Errorf("failed to create log: %w", err)
	}

	return AddLogResult{
		Success: true,
		Message: "Log entry added",
	}, nil
}

// AddNote implements the i-note-add tool
func (t *Tools) AddNote(args AddNoteArgs) (AddNoteResult, error) {
	note := &database.Note{
		SessionID: args.SessionID,
		Content:   args.Content,
		Timestamp: time.Now(),
	}

	if err := t.db.CreateNote(note); err != nil {
		return AddNoteResult{}, fmt.Errorf("failed to create note: %w", err)
	}

	return AddNoteResult{
		Success: true,
		Message: "Note added",
	}, nil
}

// SetContext implements the i-context-set tool
func (t *Tools) SetContext(args SetContextArgs) (SetContextResult, error) {
	if err := t.db.SetContext(args.SessionID, args.Key, args.Value); err != nil {
		return SetContextResult{}, fmt.Errorf("failed to set context: %w", err)
	}

	return SetContextResult{
		Success: true,
		Message: "Context updated",
	}, nil
}

// GetSession implements the i-session-get tool
func (t *Tools) GetSession(args GetSessionArgs) (GetSessionResult, error) {
	// Get session
	session, err := t.db.GetSession(args.SessionID)
	if err != nil {
		return GetSessionResult{}, fmt.Errorf("failed to get session: %w", err)
	}

	// Get logs
	logs, err := t.db.GetLogs(args.SessionID)
	if err != nil {
		return GetSessionResult{}, fmt.Errorf("failed to get logs: %w", err)
	}

	// Get notes
	notes, err := t.db.GetNotes(args.SessionID)
	if err != nil {
		return GetSessionResult{}, fmt.Errorf("failed to get notes: %w", err)
	}

	// Get context (already returns a map)
	contextMap, err := t.db.GetContext(args.SessionID)
	if err != nil {
		return GetSessionResult{}, fmt.Errorf("failed to get context: %w", err)
	}

	// Convert logs
	logResults := make([]SessionLog, len(logs))
	for i, l := range logs {
		logResults[i] = SessionLog{
			Type:      l.Type,
			Message:   l.Message,
			Timestamp: l.Timestamp.Format(time.RFC3339),
		}
	}

	// Convert notes
	noteResults := make([]SessionNote, len(notes))
	for i, n := range notes {
		noteResults[i] = SessionNote{
			Content:   n.Content,
			Timestamp: n.Timestamp.Format(time.RFC3339),
		}
	}

	return GetSessionResult{
		Success:   true,
		Message:   "Session retrieved successfully",
		SessionID: session.ID,
		AgentID:   session.AgentID,
		Logs:      logResults,
		Notes:     noteResults,
		Context:   contextMap,
	}, nil
}

// CreatePersona implements the i-persona-create MCP tool
func (t *Tools) CreatePersona(args CreatePersonaArgs) (CreatePersonaResult, error) {
	// Check if persona already exists
	existing, _ := t.db.GetPersona(args.ID)
	if existing != nil {
		return CreatePersonaResult{
			Success: false,
			Message: fmt.Sprintf("Persona %s already exists", args.ID),
		}, nil
	}

	persona := &database.Persona{
		ID:             args.ID,
		Name:           args.Name,
		Description:    args.Description,
		Expertise:      args.Expertise,
		InitialContext: args.InitialContext,
		PreferredTools: args.PreferredTools,
		Specialization: args.Specialization,
	}

	if err := t.db.CreatePersona(persona); err != nil {
		return CreatePersonaResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create persona: %v", err),
		}, fmt.Errorf("database error: %w", err)
	}

	return CreatePersonaResult{
		Success: true,
		Message: fmt.Sprintf("Persona %s created successfully", args.ID),
	}, nil
}

// GetPersona implements the i-persona-get MCP tool
func (t *Tools) GetPersona(args GetPersonaArgs) (GetPersonaResult, error) {
	persona, err := t.db.GetPersona(args.ID)
	if err != nil {
		return GetPersonaResult{
			Success: false,
			Message: fmt.Sprintf("Persona not found: %s", args.ID),
		}, nil
	}

	return GetPersonaResult{
		Success:        true,
		Message:        "Persona retrieved successfully",
		ID:             persona.ID,
		Name:           persona.Name,
		Description:    persona.Description,
		Expertise:      persona.Expertise,
		InitialContext: persona.InitialContext,
		PreferredTools: persona.PreferredTools,
		Specialization: persona.Specialization,
	}, nil
}

// ListPersonas implements the i-persona-list MCP tool
func (t *Tools) ListPersonas(args ListPersonasArgs) (ListPersonasResult, error) {
	var specialization string
	if args.Specialization != nil {
		specialization = *args.Specialization
	}

	personas, err := t.db.ListPersonas(specialization)
	if err != nil {
		return ListPersonasResult{
			Success: false,
			Message: fmt.Sprintf("Failed to list personas: %v", err),
		}, fmt.Errorf("database error: %w", err)
	}

	summaries := make([]PersonaSummary, len(personas))
	for i, p := range personas {
		summaries[i] = PersonaSummary{
			ID:             p.ID,
			Name:           p.Name,
			Description:    p.Description,
			Specialization: p.Specialization,
		}
	}

	return ListPersonasResult{
		Success:  true,
		Message:  fmt.Sprintf("Found %d personas", len(personas)),
		Personas: summaries,
	}, nil
}

// SnapshotExpertise implements the i-snapshot-expertise MCP tool
// The agent provides its current learned context/expertise to create a reusable persona
func (t *Tools) SnapshotExpertise(args SnapshotExpertiseArgs) (SnapshotExpertiseResult, error) {
	// Check if persona already exists
	existing, _ := t.db.GetPersona(args.PersonaID)
	if existing != nil {
		return SnapshotExpertiseResult{
			Success: false,
			Message: fmt.Sprintf("Persona %s already exists", args.PersonaID),
		}, nil
	}

	description := fmt.Sprintf("Expert persona created from learned context on %s",
		time.Now().Format("2006-01-02"))

	persona := &database.Persona{
		ID:             args.PersonaID,
		Name:           args.PersonaName,
		Description:    description,
		Expertise:      args.ExpertiseAreas,
		InitialContext: args.CurrentContext,
		PreferredTools: args.PreferredTools,
		Specialization: args.Specialization,
	}

	if err := t.db.CreatePersona(persona); err != nil {
		return SnapshotExpertiseResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create persona: %v", err),
		}, fmt.Errorf("database error: %w", err)
	}

	return SnapshotExpertiseResult{
		Success:   true,
		Message:   fmt.Sprintf("Persona %s created successfully. Use persona_id='%s' when starting new sessions to load this expertise.", args.PersonaName, args.PersonaID),
		PersonaID: args.PersonaID,
	}, nil
}

// ListSessions implements the i-list-sessions MCP tool
func (t *Tools) ListSessions(args ListSessionsArgs) (ListSessionsResult, error) {
	var agentID string
	activeOnly := false

	if args.AgentID != nil {
		agentID = *args.AgentID
	}
	if args.ActiveOnly != nil {
		activeOnly = *args.ActiveOnly
	}

	sessions, err := t.db.ListSessions(agentID, activeOnly)
	if err != nil {
		return ListSessionsResult{}, fmt.Errorf("failed to list sessions: %w", err)
	}

	summaries := make([]SessionSummary, len(sessions))
	for i, s := range sessions {
		var endedAt *string
		if s.EndedAt != nil {
			formatted := s.EndedAt.Format(time.RFC3339)
			endedAt = &formatted
		}

		summaries[i] = SessionSummary{
			ID:                 s.ID,
			AgentID:            s.AgentID,
			Name:               s.Name,
			StartedAt:          s.StartedAt.Format(time.RFC3339),
			EndedAt:            endedAt,
			AgentStatus:        s.AgentStatus,
			FeatureDescription: s.FeatureDescription,
			CurrentTask:        s.CurrentTask,
			ProjectName:        s.ProjectName,
			IsActive:           s.EndedAt == nil,
		}
	}

	var message string
	if activeOnly {
		message = fmt.Sprintf("Found %d active sessions", len(sessions))
	} else if agentID != "" {
		message = fmt.Sprintf("Found %d sessions for agent %s", len(sessions), agentID)
	} else {
		message = fmt.Sprintf("Found %d sessions", len(sessions))
	}

	return ListSessionsResult{
		Success:  true,
		Message:  message,
		Sessions: summaries,
	}, nil
}

// Helper function to safely get string from pointer
func strPtrOrEmpty(s *string) string {
	if s == nil {
		return "N/A"
	}
	return *s
}

// LogCompaction logs a conversation compaction event and backs up the JSONL file
func (t *Tools) LogCompaction(args LogCompactionArgs) (LogCompactionResult, error) {
	// Validate agent exists
	agent, err := t.db.GetAgent(args.AgentID)
	if err != nil {
		return LogCompactionResult{Success: false, Message: fmt.Sprintf("Agent not found: %s", args.AgentID)}, nil
	}

	// Get project path from agent
	var projectPath string
	if agent.GitWorktreePath != nil {
		projectPath = *agent.GitWorktreePath
	} else {
		return LogCompactionResult{Success: false, Message: "Agent does not have a project path - cannot locate JSONL file"}, nil
	}

	// Construct path to JSONL file
	// Pattern: ~/.claude/projects/-Users-...-<project-name>/<session-id>.jsonl
	homeDir := os.Getenv("HOME")
	claudeProjectsDir := filepath.Join(homeDir, ".claude", "projects")

	// Convert project path to Claude's format (replace / with -)
	projectDirName := strings.ReplaceAll(strings.TrimPrefix(projectPath, "/"), "/", "-")
	jsonlPath := filepath.Join(claudeProjectsDir, projectDirName, args.SessionID+".jsonl")

	// Check if JSONL file exists
	fileInfo, err := os.Stat(jsonlPath)
	if err != nil {
		return LogCompactionResult{Success: false, Message: fmt.Sprintf("JSONL file not found: %s", jsonlPath)}, nil
	}

	// Create compactions directory if it doesn't exist
	compactionsDir := filepath.Join(homeDir, "projects", "eye-in-the-sky", "data", "compactions")
	if err := os.MkdirAll(compactionsDir, 0755); err != nil {
		return LogCompactionResult{Success: false, Message: fmt.Sprintf("Failed to create compactions directory: %v", err)}, nil
	}

	// Copy JSONL file to compactions directory
	backupPath := filepath.Join(compactionsDir, args.SessionID+".jsonl")
	if err := copyFile(jsonlPath, backupPath); err != nil {
		return LogCompactionResult{Success: false, Message: fmt.Sprintf("Failed to copy JSONL file: %v", err)}, nil
	}

	// Count messages in JSONL file (optional, for metadata)
	messageCount, _ := countJSONLLines(backupPath)
	fileSize := fileInfo.Size()

	// Create compaction record
	compaction := &database.Compaction{
		AgentID:       args.AgentID,
		OldSessionID:  args.OldSessionID,
		NewSessionID:  args.SessionID,
		Summary:       args.Summary,
		JsonlFilePath: &backupPath,
		JsonlFileSize: &fileSize,
		MessageCount:  &messageCount,
	}

	if err := t.db.CreateCompaction(compaction); err != nil {
		return LogCompactionResult{Success: false, Message: fmt.Sprintf("Failed to log compaction: %v", err)}, nil
	}

	return LogCompactionResult{
		Success:         true,
		Message:         fmt.Sprintf("Compaction logged successfully for session %s", args.SessionID),
		JsonlBackupPath: backupPath,
	}, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// countJSONLLines counts lines in a JSONL file
func countJSONLLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		count++
	}
	return count, scanner.Err()
}
