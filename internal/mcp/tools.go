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
	"github.com/tacit7/eye-in-the-sky/internal/window"
)

type Tools struct {
	db *database.DB
}

func NewTools(db *database.DB) *Tools {
	return &Tools{db: db}
}

// Removed RegisterAgent and RegisterDesktopAgent - now using only StartSession

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

	// Get agent first to obtain full UUID for database operations
	agent, err := t.db.GetAgent(args.AgentID)
	if err != nil {
		return LogCommitsResult{Success: false, Message: fmt.Sprintf("Failed to get agent: %v", err)}, nil
	}

	// If no hashes provided, try to get latest commits automatically
	if len(args.CommitHashes) == 0 {
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

	// Use transaction-safe method with full agent UUID
	err = t.db.WithTransaction(context.Background(), func(tx *database.Tx) error {
		return tx.CreateCommitsTx(agent.ID, commitHashes, commitMessages)
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

// LogSessionCost implements the i-log-session-cost MCP tool
func (t *Tools) LogSessionCost(args LogSessionCostArgs) (LogSessionCostResult, error) {
	// Validate agent exists
	agent, err := t.db.GetAgent(args.AgentID)
	if err != nil {
		return LogSessionCostResult{Success: false, Message: fmt.Sprintf("Agent not found: %v", err)}, nil
	}

	// Create session metrics record
	metrics := &database.SessionMetrics{
		AgentID:          args.AgentID,
		SessionID:        args.SessionID,
		TokensUsed:       args.TokensUsed,
		TokensBudget:     args.TokensBudget,
		TokensRemaining:  args.TokensRemaining,
		InputTokens:      args.InputTokens,
		OutputTokens:     args.OutputTokens,
		EstimatedCostUSD: args.EstimatedCostUSD,
		ModelName:        args.ModelName,
		Notes:            args.Notes,
	}

	// Log to database
	if err := t.db.LogSessionMetrics(metrics); err != nil {
		return LogSessionCostResult{Success: false, Message: fmt.Sprintf("Failed to log session cost: %v", err)}, nil
	}

	// Build success message
	message := fmt.Sprintf("Session cost logged for agent %s: %d/%d tokens used",
		args.AgentID, args.TokensUsed, args.TokensBudget)

	if args.EstimatedCostUSD != nil {
		message += fmt.Sprintf(" (~$%.4f)", *args.EstimatedCostUSD)
	}

	// Also log as action for visibility in timeline
	action := &database.Action{
		AgentID:     args.AgentID,
		ActionType:  database.ActionStatusUpdate,
		Description: message,
	}

	if err := t.db.CreateAction(action); err != nil {
		// Don't fail the whole operation if action logging fails
		fmt.Fprintf(os.Stderr, "Warning: Failed to log cost action: %v\n", err)
	}

	_ = agent // Suppress unused variable warning
	return LogSessionCostResult{Success: true, Message: message}, nil
}

// UpdateFeatureDescription implements the i-update-description MCP tool
func (t *Tools) UpdateFeatureDescription(args UpdateFeatureDescriptionArgs) (UpdateFeatureDescriptionResult, error) {
	// Validate agent exists
	_, err := t.db.GetAgent(args.AgentID)
	if err != nil {
		return UpdateFeatureDescriptionResult{Success: false, Message: fmt.Sprintf("Agent not found: %v", err)}, nil
	}

	// Update feature description
	if err := t.db.UpdateAgentFeatureDescription(args.AgentID, args.FeatureDescription); err != nil {
		return UpdateFeatureDescriptionResult{Success: false, Message: fmt.Sprintf("Failed to update feature description: %v", err)}, nil
	}

	// Log the update as an action
	action := &database.Action{
		AgentID:     args.AgentID,
		ActionType:  database.ActionStatusUpdate,
		Description: fmt.Sprintf("Feature description updated to: %s", args.FeatureDescription),
	}

	if err := t.db.CreateAction(action); err != nil {
		// Don't fail the whole operation if action logging fails
		fmt.Fprintf(os.Stderr, "Warning: Failed to log description update action: %v\n", err)
	}

	return UpdateFeatureDescriptionResult{Success: true, Message: "Feature description updated successfully"}, nil
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

	// Use the window ID if available
	if agent.WindowID != nil {
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

	// Fallback: try to bring terminal to front
	if agent.TerminalApplication != nil {
		bringCmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "%s" to activate`, *agent.TerminalApplication))
		err = bringCmd.Run()
		if err == nil {
			return BringWindowFrontResult{
				Success: true,
				Message: fmt.Sprintf("Brought %s window to front", *agent.TerminalApplication),
			}, nil
		}
	}

	return BringWindowFrontResult{
		Success: false,
		Message: "No window information available for this agent",
	}, nil
}

// Instructions implements the i-instructions MCP tool
func (t *Tools) Instructions(args InstructionsArgs) (InstructionsResult, error) {
	instructions := `Eye in the Sky - Agent Lifecycle Management

═══════════════════════════════════════════════════════════════
INITIALIZATION - CRITICAL: Do this FIRST on EVERY new session
═══════════════════════════════════════════════════════════════

At the start of the session, your session_id will be provided to you.
Use this ID when starting your session.

To register yourself, call i-start-session with your session ID:

  i-start-session({
    "session_id": "your_provided_session_id",
    "description": "What you'll be working on",
    "worktree_path": "optional",
    "project_name": "optional",
    "parent_agent_id": "optional",
    "parent_session_id": "optional"
  })

IMPORTANT: The system will return a response containing:
  {
    "agent_id": "generated-uuid-for-your-agent",
    "session_id": "your-session-id",
    "success": true,
    "message": "Session started for agent UUID"
  }

You MUST extract and use this returned agent_id for all subsequent MCP calls.
The agent_id is auto-generated as a UUID and uniquely identifies your agent instance.

═══════════════════════════════════════════════════════════════
SUBAGENTS - Creating and Managing Child Agents
═══════════════════════════════════════════════════════════════

Subagents are child agents spawned by a parent agent to handle specific tasks.
They maintain a hierarchical relationship with their parent for tracking.

Creating a Subagent:
  When your main agent needs to spawn a subagent (e.g., using Task tool), the subagent should:

  1. Generate its own new session_id (UUID)
  2. Call i-start-session with parent tracking:

  i-start-session({
    "session_id": "new-uuid-for-subagent-session",
    "description": "Specific task for subagent",
    "parent_agent_id": "your-current-agent-uuid",      // Links to parent agent
    "parent_session_id": "your-current-session-uuid"   // Links to parent session
  })

  3. The subagent receives its own agent_id in response
  4. Subagent operates independently but is tracked as child of parent

Parent-Child Relationships:
  - parent_agent_id: Links this agent to its parent agent (UUID)
  - parent_session_id: Links this session to parent session (UUID)
  - Both help maintain hierarchy for multi-agent workflows
  - Dashboard/TUI shows subagents indented under parents

Example Workflow:
  Main Agent (agent: abc-123, session: def-456) needs to analyze code
  └─> Spawns Subagent for analysis:
      i-start-session({
        "session_id": "ghi-789",  // New session ID
        "description": "Analyze authentication module",
        "parent_agent_id": "abc-123",
        "parent_session_id": "def-456"
      })
      └─> Subagent (agent: jkl-012, session: ghi-789) works independently
          but is tracked as child of Main Agent

When to Use Subagents:
  - Delegating specific subtasks to specialized agents
  - Parallel processing of independent tasks
  - Isolating complex operations
  - When using Task tool with subagent_type parameter

═══════════════════════════════════════════════════════════════
WORKFLOW - During your session
═══════════════════════════════════════════════════════════════

After receiving your agent_id from i-start-session, use it in all subsequent calls:

i-status - Update your current status
  Statuses: active, working, idle, completed, failed
  Example: i-status({"agent_id": "your-uuid-from-start-session", "status": "working", "current_task": "Building TUI"})

i-action - Log significant activities
  Types: task_start, file_operation, git_commit, status_update
  Example: i-action({"agent_id": "your-uuid-from-start-session", "action_type": "file_operation", "description": "Created dashboard component"})

i-commits - Track git commits
  Example: i-commits({"agent_id": "your-uuid-from-start-session", "commit_hashes": ["abc123f"], "commit_messages": ["Add feature"]})

i-end - End session with summary
  Example: i-end({"agent_id": "your-uuid-from-start-session", "summary": "Completed TUI implementation", "final_status": "completed"})

═══════════════════════════════════════════════════════════════
COMPACTION TRACKING
═══════════════════════════════════════════════════════════════

When Claude detects a conversation compaction (indicated by system message
'This session is being continued from a previous conversation'), call:

i-log-compaction({
  "agent_id": "your-agent-id",
  "session_id": "session-id-that-was-compacted",
  "summary": "compaction summary text (optional)"
})

This backs up the JSONL conversation file to data/compactions/ directory.
The session continues with the same ID, just with compacted (summarized) messages.

═══════════════════════════════════════════════════════════════
ADDITIONAL TOOLS
═══════════════════════════════════════════════════════════════

i-save-context - Save session state for resumption
i-load-context - Load previous session context
i-note - Add contextual notes to session
i-persona-get - Get persona details
i-persona-list - List available personas
i-snapshot-expertise - Save current expertise as persona

═══════════════════════════════════════════════════════════════
DASHBOARD & TUI
═══════════════════════════════════════════════════════════════

Web Dashboard: http://localhost:8080 (if web server running)
TUI: Run 'bin/dashboard' for terminal interface

TUI KEYBINDINGS:
Agent List View (Overview Page):
  [q] Quit          [r] Refresh         [a] Toggle filter (active/all)
  [j/k] Navigate    [Enter] View details
  [n] New session   [c] Continue        [s] Start session   [w] Window
  [L] Logs          [D] Archive

Agent Detail View (Individual Agent Page):
  [q] Back to list  [r] Refresh         [j/k] Scroll
  [s] Start session [w] Go to window    [L] View all logs

  Tabs (navigate with letter keys):
  [O] Overview      [C] Commits         [L] Logs
  [N] Notes         [A] Actions         [T] Tasks
  [P] Project tickets (Note: Known issue - may not work when pressed)

STATUS INDICATORS:
  Active Sessions (shown by default):
    ● ACTIVE   - Ready for work (green)
    ● WORKING  - Currently working (blue)
    ● IDLE     - Waiting for next task (yellow)
    ● STALE    - Inactive 30min-1hr (gray) - needs attention
    ? UNKNOWN  - Inactive >1hr (gray) - possibly dead/disconnected

  Completed Sessions (shown with 'a' toggle):
    ✓ COMPLETE - Finished successfully (cyan)
    ✗ FAILED   - Ended with error (red)

ACTIVE FILTER:
  Default view shows: active, working, idle, stale, unknown
  Press 'a' to toggle between active sessions and all sessions (including completed/failed/archived)

COMMANDS:
  'n' New Session   - Creates new session with marker file, runs 'claude --session-id <id>' and exits
  's' Start Session - Runs 'claude -s <session-id>' and exits dashboard
  'c' Continue      - Runs 'claude --resume <session-id>' and returns to dashboard

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
	// Always generate a new UUID agent ID
	agentID := utils.GenerateGitStyleAgentID()

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

	// Detect window ID on macOS
	var windowID *string
	var terminalApp *string
	if runtime.GOOS == "darwin" {
		// Auto-detect current window on macOS
		wm := window.NewManager()
		winInfo, err := wm.GetCurrentWindowID("current", "")
		if err == nil && winInfo != nil {
			// Format: "Application:WindowID"
			detectedWindowID := fmt.Sprintf("%s:%s", winInfo.Application, winInfo.ID)
			windowID = &detectedWindowID
			// Store terminal application separately
			terminalApp = &winInfo.Application
		}
		// If detection fails, just continue without window ID
	}

	// Create agent with window tracking
	agent := &database.Agent{
		ID:                  agentID,
		Status:              database.StatusActive,
		Source:              database.SourceWorktree, // Always worktree now
		Description:         args.AgentDescription,
		GitWorktreePath:     args.WorktreePath,
		FeatureDescription:  &args.Description,
		ProjectName:         args.ProjectName,
		PersonaID:           personaID,
		WindowID:            windowID,
		TerminalApplication: terminalApp,
		ParentAgentID:       args.ParentAgentID,
		LastActivityAt:      timePtr(time.Now()),
	}

	if err := t.db.CreateAgent(agent); err != nil {
		return StartSessionResult{}, fmt.Errorf("failed to create agent: %w", err)
	}

	// Use the mandatory session_id provided
	sessionID := args.SessionID

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
		ID:         fmt.Sprintf("%d", time.Now().UnixNano()),
		ParentID:   args.ParentID,
		ParentType: args.ParentType,
		Body:       args.Body,
		CreatedAt:  time.Now(),
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

	// Get notes for this session
	notes, err := t.db.GetNotes(args.SessionID, "sessions")
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
			Content:   n.Body,
			Timestamp: n.CreatedAt.Format(time.RFC3339),
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

	// Convert project path to Claude's format (replace / with - and prepend -)
	projectDirName := "-" + strings.ReplaceAll(strings.TrimPrefix(projectPath, "/"), "/", "-")
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
		SessionID:     args.SessionID,
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

// ISpeak implements the i-speak tool for text-to-speech output
func (t *Tools) ISpeak(args ISpeakArgs) (ISpeakResult, error) {
	// Premium voices available (must include " (Premium)" suffix)
	premiumVoices := map[string]bool{
		"Ava (Premium)":    true, // en_US
		"Isha (Premium)":   true, // en_IN
		"Lee (Premium)":    true, // en_AU
		"Jamie (Premium)":  true, // en_GB
		"Serena (Premium)": true, // en_GB
	}

	// Default to Ava (Premium) if not specified
	voice := "Ava (Premium)"
	if args.Voice != nil && *args.Voice != "" {
		requestedVoice := *args.Voice
		if premiumVoices[requestedVoice] {
			voice = requestedVoice
		}
	}

	// Default rate is 200 words per minute
	rate := 200
	if args.Rate != nil {
		// Clamp rate between 90 and 450
		if *args.Rate >= 90 && *args.Rate <= 450 {
			rate = *args.Rate
		}
	}

	// Execute ls && say command in background with rate parameter
	cmd := fmt.Sprintf("ls && say -v \"%s\" -r %d \"%s\" &", voice, rate, args.Message)
	output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return ISpeakResult{
			Success:   false,
			Message:   fmt.Sprintf("Failed to execute say command: %v - %s", err, string(output)),
			VoiceUsed: voice,
		}, nil
	}

	return ISpeakResult{
		Success:   true,
		Message:   "Speech command executed",
		VoiceUsed: voice,
	}, nil
}
