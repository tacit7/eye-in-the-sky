package mcp

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/utils"
	"github.com/tacit7/eye-in-the-sky/internal/window"
	"gopkg.in/yaml.v3"
)

type Tools struct {
	db *database.DB
}

func NewTools(db *database.DB) *Tools {
	return &Tools{db: db}
}

// getGitRemoteURL extracts the git remote URL from a worktree path
func getGitRemoteURL(worktreePath string) (string, error) {
	if worktreePath == "" {
		return "", fmt.Errorf("worktree path is empty")
	}

	cmd := exec.Command("git", "-C", worktreePath, "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git remote URL: %w", err)
	}

	remoteURL := strings.TrimSpace(string(output))
	if remoteURL == "" {
		return "", fmt.Errorf("no remote URL found")
	}

	return remoteURL, nil
}

// getGitRoot finds the git repository root from a given path
func getGitRoot(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}

	cmd := exec.Command("git", "-C", path, "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository or git not found: %w", err)
	}

	gitRoot := strings.TrimSpace(string(output))
	if gitRoot == "" {
		return "", fmt.Errorf("could not determine git root")
	}

	return gitRoot, nil
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

At the start of the session, your session_id should be provided to you.

CRITICAL: Getting Your Session ID
- If you don't know your session_id, you MUST ask the user for it using /status command
- The user can provide their session_id which you'll use for registration
- If no session_id was given, use the agent_id as the session_id

To register yourself, call i-start-session with your session ID:

  i-start-session({
    "session_id": "your_provided_session_id",  // Ask user via /status if not provided
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
CLAUDE CODE HOOKS - Manual Mapping File Update Required
═══════════════════════════════════════════════════════════════

CRITICAL: If you're using Claude Code hooks for logging tool execution:

After calling i-start-session and receiving your agent_id, you MUST manually update
the session-to-agent mapping file so hooks can look up your agent ID.

STEP 1: Read the current mapping file:
  Read: .claude/hooks/session_agent_map.json

STEP 2: Add your session→agent mapping:
  Edit the file to add a new entry:
  {
    "existing-session-1": "existing-agent-1",
    "YOUR_SESSION_ID": "YOUR_AGENT_ID_FROM_START_SESSION"
  }

STEP 3: Save the file

Why this is needed:
- Claude Code provides session_id to hooks via JSON stdin
- Eye-in-the-Sky uses agent_id (UUID) to track agents
- Hooks need to translate session_id → agent_id
- The mapping file bridges these two systems
- JSON file lookup is 6.9% faster than SQLite query

Example mapping file:
  {
    "97c212de-24e0-4c8d-b16a-876e406b7c14": "2743c649-18ab-422d-bdbe-bb3b59d2c431",
    "1a398965-1f97-4329-802d-0bbe0d923564": "354ab567-5a8b-4259-8f25-9ed68d96e7d8"
  }

Without this mapping:
- Hooks will fire but agent_id will be "unknown"
- Logs will be written but not linked to your agent

═══════════════════════════════════════════════════════════════
iSUBAGENTS - Creating and Managing Child Agents
═══════════════════════════════════════════════════════════════

iSubagents are child agents spawned by a parent agent to handle specific tasks.
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
TASK MANAGEMENT - REQUIRED: Use i-todo for ALL task logging
═══════════════════════════════════════════════════════════════

CRITICAL: All agents MUST use i-todo tools for task tracking and logging.
This is the primary system for tracking work progress across sessions.

Core Task Workflow:

1. CREATE tasks at the start of work:
   i-todo-create({
     "project_id": 1,
     "title": "Implement user authentication",
     "description": "Add JWT-based auth with refresh tokens",
     "priority": 1,           // 1=high, 2=medium, 3=low
     "tags": ["backend", "security"],
     "session_ids": ["your-session-id"],
     "agent_id": "your-agent-id"
   })

2. START working on a task:
   i-todo-start({"task_id": "task-uuid"})
   // Moves task to "doing" state

3. ANNOTATE progress as you work:
   i-todo-annotate({
     "task_id": "task-uuid",
     "body": "Completed JWT token generation, now working on refresh logic"
   })
   // Add notes throughout development to track progress

4. COMPLETE tasks when done:
   i-todo-done({"task_id": "task-uuid"})
   // Moves task to "done" state

Additional Task Operations:

i-todo-status - Move task to any workflow state
  Example: i-todo-status({"task_id": "task-uuid", "state_id": 2})

i-todo-tag - Add/remove tags from task
  Example: i-todo-tag({"task_id": "task-uuid", "add": ["tested"], "remove": ["wip"]})

i-todo-list - List all tasks for a project
  Example: i-todo-list({"project_id": 1, "filters": {"state_id": 2, "priority": 1}})

i-todo-list-agent - List tasks for specific agent
  Example: i-todo-list-agent({"agent_id": "your-uuid", "project_id": 1})

i-todo-list-session - List tasks for current session
  Example: i-todo-list-session({"session_id": "your-session", "project_id": 1})

i-todo-search - Full-text search across tasks
  Example: i-todo-search({"project_id": 1, "query": "authentication"})

i-todo-add-session - Link task to session
  Example: i-todo-add-session({"task_id": "task-uuid", "session_id": "session-uuid"})

i-todo-remove-session - Unlink task from session
  Example: i-todo-remove-session({"task_id": "task-uuid", "session_id": "session-uuid"})

i-todo-delete - Permanently delete task
  Example: i-todo-delete({"task_id": "task-uuid"})

Best Practices:
- Create tasks BEFORE starting work to establish clear goals
- Use i-todo-annotate frequently to document progress and decisions
- Link tasks to sessions using session_ids array in i-todo-create
- Use priority (1=high, 2=medium, 3=low) to indicate urgency
- Tag tasks appropriately for filtering and organization
- Move tasks through workflow states (todo → doing → done)

═══════════════════════════════════════════════════════════════
ADDITIONAL TOOLS
═══════════════════════════════════════════════════════════════

i-save-context - Save session state for resumption
i-note-add - Add contextual notes to session
i-project-add - Create new project for tracking agents and tasks

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
	// Check if session already exists first (idempotent behavior)
	sessionID := args.SessionID
	existingSession, err := t.db.GetSession(sessionID)
	if err == nil && existingSession != nil {
		// Session exists - return existing agent_id instead of failing
		message := fmt.Sprintf("Session %s already exists, using existing agent %s",
			existingSession.ID, existingSession.AgentID)
		fmt.Fprintf(os.Stderr, "[DEBUG] %s\n", message)

		return StartSessionResult{
			Success:   true,
			Message:   message,
			AgentID:   existingSession.AgentID,
			SessionID: existingSession.ID,
		}, nil
	}

	// Session doesn't exist - create new agent and session
	// Always generate a new UUID agent ID
	agentID := utils.GenerateGitStyleAgentID()

	// Auto-detect worktree path if not provided
	if args.WorktreePath == nil || *args.WorktreePath == "" {
		// Get current working directory
		cwd, err := os.Getwd()
		if err == nil {
			// Try to find git root from cwd
			gitRoot, err := getGitRoot(cwd)
			if err == nil {
				args.WorktreePath = &gitRoot
				fmt.Fprintf(os.Stderr, "[DEBUG] Auto-detected worktree path: %s\n", gitRoot)
			} else {
				fmt.Fprintf(os.Stderr, "[DEBUG] Could not auto-detect git root: %v\n", err)
			}
		}
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

	// Look up project by git remote URL or path if worktree path is provided
	var projectID *int
	var foundProject *database.Project
	if args.WorktreePath != nil && *args.WorktreePath != "" {
		// First try by remote URL
		remoteURL, err := getGitRemoteURL(*args.WorktreePath)
		if err == nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] Got remote URL: %s\n", remoteURL)
			project, err := t.db.GetProjectByRemoteURL(remoteURL)
			if err == nil {
				fmt.Fprintf(os.Stderr, "[DEBUG] Found project by remote URL: %s\n", project.ID)
				foundProject = project
				// Convert string ID to int
				projID, err := strconv.Atoi(project.ID)
				if err == nil {
					projectID = &projID
				}
			} else {
				fmt.Fprintf(os.Stderr, "[DEBUG] Remote URL lookup failed: %v\n", err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "[DEBUG] Failed to get remote URL: %v\n", err)
		}

		// If remote URL lookup failed, try by path
		if projectID == nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] Trying path lookup: %s\n", *args.WorktreePath)
			project, err := t.db.GetProjectByPath(*args.WorktreePath)
			if err == nil {
				fmt.Fprintf(os.Stderr, "[DEBUG] Found project by path: %s\n", project.ID)
				foundProject = project
				// Convert string ID to int
				projID, err := strconv.Atoi(project.ID)
				if err == nil {
					projectID = &projID
				} else {
					fmt.Fprintf(os.Stderr, "[DEBUG] Failed to convert project ID to int: %v\n", err)
				}
			} else {
				fmt.Fprintf(os.Stderr, "[DEBUG] Path lookup failed: %v\n", err)
				// Project not found by either method - return helpful error
				return StartSessionResult{}, fmt.Errorf(
					"project not found for worktree path '%s'. Please create a project entry first. "+
					"You can do this by adding a row to the projects table with the path or remote_url set.",
					*args.WorktreePath,
				)
			}
		}
	}

	// Auto-populate project name from found project if not provided
	projectName := args.ProjectName
	if (projectName == nil || *projectName == "") && foundProject != nil {
		projectName = &foundProject.Name
		fmt.Fprintf(os.Stderr, "[DEBUG] Auto-populated project name: %s\n", *projectName)
	}

	// Create agent with window tracking
	agent := &database.Agent{
		ID:                  agentID,
		Status:              database.StatusActive,
		Source:              database.SourceWorktree, // Always worktree now
		Description:         args.AgentDescription,
		GitWorktreePath:     args.WorktreePath,
		FeatureDescription:  &args.Description,
		ProjectName:         projectName, // Use auto-populated project name
		ProjectID:           projectID,
		WindowID:            windowID,
		TerminalApplication: terminalApp,
		ParentAgentID:       args.ParentAgentID,
		LastActivityAt:      timePtr(now()),
	}

	if err := t.db.CreateAgent(agent); err != nil {
		return StartSessionResult{}, fmt.Errorf("failed to create agent: %w", err)
	}

	// Use name if provided, otherwise use description as session name
	sessionName := args.Name
	if sessionName == nil || *sessionName == "" {
		sessionName = &args.Description
	}

	session := &database.Session{
		ID:        sessionID,
		AgentID:   agentID,
		Name:      sessionName,
		StartedAt: now(),
	}

	if err := t.db.CreateSession(session); err != nil {
		return StartSessionResult{}, fmt.Errorf("failed to create session: %w", err)
	}

	// Log initial entry
	logMessage := fmt.Sprintf("Started session: %s", args.Description)

	log := &database.Log{
		SessionID: sessionID,
		Type:      "info",
		Message:   logMessage,
		Timestamp: now(),
	}

	if err := t.db.CreateLog(log); err != nil {
		return StartSessionResult{}, fmt.Errorf("failed to create initial log: %w", err)
	}

	// Update agent's current session
	if err := t.db.UpdateAgentCurrentSession(agentID, sessionID); err != nil {
		return StartSessionResult{}, fmt.Errorf("failed to update current session: %w", err)
	}

	message := fmt.Sprintf("Session %s started for agent %s", sessionID, agentID)

	return StartSessionResult{
		Success:   true,
		Message:   message,
		AgentID:   agentID,
		SessionID: sessionID,
	}, nil
}

// AddLog implements the i-log tool
func (t *Tools) AddLog(args AddLogArgs) (AddLogResult, error) {
	log := &database.Log{
		SessionID: args.SessionID,
		Type:      args.Type,
		Message:   args.Message,
		Timestamp: now(),
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
		CreatedAt:  now(),
	}

	if err := t.db.CreateNote(note); err != nil {
		return AddNoteResult{}, fmt.Errorf("failed to create note: %w", err)
	}

	return AddNoteResult{
		Success: true,
		Message: "Note added",
	}, nil
}

// GetNote implements the i-note-get tool
func (t *Tools) GetNote(args GetNoteArgs) (GetNoteResult, error) {
	query := `SELECT id, parent_id, parent_type, body, created_at
	          FROM notes WHERE id = ?`

	var result GetNoteResult
	var noteID, parentID int64
	err := t.db.QueryRow(query, args.NoteID).Scan(
		&noteID,
		&parentID,
		&result.ParentType,
		&result.Body,
		&result.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return GetNoteResult{}, fmt.Errorf("note not found: %s", args.NoteID)
		}
		return GetNoteResult{}, fmt.Errorf("failed to retrieve note: %w", err)
	}

	result.NoteID = fmt.Sprintf("%d", noteID)
	result.ParentID = fmt.Sprintf("%d", parentID)

	return result, nil
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

// CreatePrompt implements the i-prompt-create MCP tool
func (t *Tools) CreateSubagentPrompt(args CreatePromptArgs) (PromptResult, error) {
	prompt := &database.SubagentPrompt{
		Name:        args.Name,
		Slug:        args.Slug,
		Description: args.Description,
		PromptText:  args.PromptText,
		ProjectID:   args.ProjectID,
		Active:      true,
		Tags:        args.Tags,
		CreatedBy:   args.CreatedBy,
	}

	if err := t.db.CreateSubagentPrompt(prompt); err != nil{
		return PromptResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create prompt: %v", err),
		}, nil
	}

	return PromptResult{
		Success: true,
		Message: fmt.Sprintf("Prompt created with ID: %s", prompt.ID),
		Prompt:  prompt,
	}, nil
}

// GetPrompt implements the i-prompt-get MCP tool
func (t *Tools) GetSubagentPrompt(args GetPromptArgs) (PromptResult, error) {
	var prompt *database.SubagentPrompt
	var err error

	if args.ID != "" {
		prompt, err = t.db.GetSubagentPromptByID(args.ID)
	} else if args.Slug != "" {
		prompt, err = t.db.GetSubagentPromptBySlug(args.Slug, args.ProjectID)
	} else {
		return PromptResult{
			Success: false,
			Message: "Either id or slug must be provided",
		}, nil
	}

	if err != nil {
		return PromptResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get prompt: %v", err),
		}, nil
	}

	// Optionally exclude prompt_text
	if !args.IncludeText {
		prompt.PromptText = ""
	}

	return PromptResult{
		Success: true,
		Message: "Prompt retrieved successfully",
		Prompt:  prompt,
	}, nil
}

// ListPrompts implements the i-prompt-list MCP tool
func (t *Tools) ListSubagentPrompts(args ListPromptsArgs) (ListPromptsResult, error) {
	opts := database.ListSubagentPromptsOptions{
		ProjectID:   args.ProjectID,
		Active:      args.Active,
		Tags:        args.Tags,
		Resolve:     args.Resolve,
		IncludeText: args.IncludeText,
		Limit:       args.Limit,
		Offset:      args.Offset,
	}

	prompts, err := t.db.ListSubagentPrompts(opts)
	if err != nil {
		return ListPromptsResult{
			Success: false,
			Message: fmt.Sprintf("Failed to list prompts: %v", err),
		}, nil
	}

	return ListPromptsResult{
		Success: true,
		Message: fmt.Sprintf("Found %d prompts", len(prompts)),
		Prompts: prompts,
		Count:   len(prompts),
	}, nil
}

// UpdatePrompt implements the i-prompt-update MCP tool
func (t *Tools) UpdateSubagentPrompt(args UpdatePromptArgs) (PromptResult, error) {
	if args.ID == "" {
		return PromptResult{
			Success: false,
			Message: "id is required",
		}, nil
	}

	if args.ExpectedVersion == 0 {
		return PromptResult{
			Success: false,
			Message: "expected_version is required for optimistic locking",
		}, nil
	}

	updates := make(map[string]interface{})
	if args.Name != nil {
		updates["name"] = *args.Name
	}
	if args.Slug != nil {
		updates["slug"] = *args.Slug
	}
	if args.Description != nil {
		updates["description"] = *args.Description
	}
	if args.PromptText != nil {
		updates["prompt_text"] = *args.PromptText
	}
	if args.ProjectID != nil {
		updates["project_id"] = *args.ProjectID
	}
	if args.Tags != nil {
		updates["tags"] = *args.Tags
	}
	if args.Active != nil {
		updates["active"] = *args.Active
	}

	if len(updates) == 0 {
		return PromptResult{
			Success: false,
			Message: "No fields to update",
		}, nil
	}

	if err := t.db.UpdateSubagentPrompt(args.ID, updates, args.ExpectedVersion); err != nil {
		return PromptResult{
			Success: false,
			Message: fmt.Sprintf("Failed to update prompt: %v", err),
		}, nil
	}

	// Fetch updated prompt
	prompt, err := t.db.GetSubagentPromptByID(args.ID)
	if err != nil {
		return PromptResult{
			Success: true,
			Message: "Prompt updated but failed to fetch updated version",
		}, nil
	}

	return PromptResult{
		Success: true,
		Message: "Prompt updated successfully",
		Prompt:  prompt,
	}, nil
}

// DeletePrompt implements the i-prompt-delete MCP tool
func (t *Tools) DeleteSubagentPrompt(args DeletePromptArgs) (PromptResult, error) {
	if args.ID == "" {
		return PromptResult{
			Success: false,
			Message: "id is required",
		}, nil
	}

	var err error
	if args.HardDelete {
		err = t.db.DeleteSubagentPrompt(args.ID)
	} else {
		err = t.db.DeactivateSubagentPrompt(args.ID)
	}

	if err != nil {
		return PromptResult{
			Success: false,
			Message: fmt.Sprintf("Failed to delete prompt: %v", err),
		}, nil
	}

	deleteType := "deactivated"
	if args.HardDelete {
		deleteType = "deleted"
	}

	return PromptResult{
		Success: true,
		Message: fmt.Sprintf("Prompt %s successfully", deleteType),
	}, nil
}

// AgentFrontmatter represents the YAML frontmatter from .claude/agents/*.md files
type AgentFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Tools       string `yaml:"tools"`
	Model       string `yaml:"model"`
	Color       string `yaml:"color"`
}

// ImportAgents implements the i-agent-import MCP tool
func (t *Tools) ImportAgents(args ImportAgentsArgs) (ImportAgentsResult, error) {
	// Default agents directory
	agentsDir := args.AgentsDir
	if agentsDir == "" {
		agentsDir = ".claude/agents"
	}

	// Check if directory exists
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		return ImportAgentsResult{
			Success: false,
			Message: fmt.Sprintf("Agents directory not found: %s", agentsDir),
		}, nil
	}

	// Read all .md files from agents directory
	files, err := filepath.Glob(filepath.Join(agentsDir, "*.md"))
	if err != nil {
		return ImportAgentsResult{
			Success: false,
			Message: fmt.Sprintf("Failed to read agents directory: %v", err),
		}, nil
	}

	if len(files) == 0 {
		return ImportAgentsResult{
			Success: true,
			Message: "No agent files found in directory",
			Count:   0,
		}, nil
	}

	var imported []string
	var skipped []string
	var errors []string

	// Regex to split frontmatter from body
	frontmatterRegex := regexp.MustCompile(`(?s)^---\n(.*?)\n---\n(.*)$`)

	for _, file := range files {
		filename := filepath.Base(file)

		// Read file content
		content, err := os.ReadFile(file)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to read file: %v", filename, err))
			continue
		}

		// Parse frontmatter and body
		matches := frontmatterRegex.FindStringSubmatch(string(content))
		if matches == nil || len(matches) < 3 {
			errors = append(errors, fmt.Sprintf("%s: invalid format (missing YAML frontmatter)", filename))
			continue
		}

		frontmatterYAML := matches[1]
		promptBody := strings.TrimSpace(matches[2])

		// Parse YAML frontmatter
		var fm AgentFrontmatter
		if err := yaml.Unmarshal([]byte(frontmatterYAML), &fm); err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to parse YAML: %v", filename, err))
			continue
		}

		// Validate required fields
		if fm.Name == "" {
			errors = append(errors, fmt.Sprintf("%s: missing 'name' in frontmatter", filename))
			continue
		}

		// Use name as slug (already in kebab-case format)
		slug := fm.Name

		// Check if prompt already exists
		existing, err := t.db.GetSubagentPromptBySlug(slug, args.ProjectID)
		if err == nil && existing != nil && !args.Overwrite {
			skipped = append(skipped, fmt.Sprintf("%s (already exists with slug: %s)", filename, slug))
			continue
		}

		// Build tags from model, color, and tools metadata
		tags := []string{}
		if fm.Model != "" {
			tags = append(tags, "model:"+fm.Model)
		}
		if fm.Color != "" {
			tags = append(tags, "color:"+fm.Color)
		}
		if fm.Tools != "" {
			tags = append(tags, "source:claude-agents")
		}
		tagsStr := strings.Join(tags, ",")

		// Create or update the prompt
		if existing != nil && args.Overwrite {
			// Update existing prompt
			updates := map[string]interface{}{
				"description": fm.Description,
				"prompt_text": promptBody,
				"tags":        tagsStr,
			}
			if err := t.db.UpdateSubagentPrompt(existing.ID, updates, existing.Version); err != nil {
				errors = append(errors, fmt.Sprintf("%s: failed to update: %v", filename, err))
				continue
			}
			imported = append(imported, fmt.Sprintf("%s (updated slug: %s)", filename, slug))
		} else {
			// Create new prompt
			prompt := &database.SubagentPrompt{
				Name:        fm.Name,
				Slug:        slug,
				Description: fm.Description,
				PromptText:  promptBody,
				ProjectID:   args.ProjectID,
				Active:      true,
				Tags:        tagsStr,
				CreatedBy:   args.CreatedBy,
			}

			if err := t.db.CreateSubagentPrompt(prompt); err != nil {
				errors = append(errors, fmt.Sprintf("%s: failed to create: %v", filename, err))
				continue
			}
			imported = append(imported, fmt.Sprintf("%s (slug: %s, id: %s)", filename, slug, prompt.ID))
		}
	}

	// Build result message
	message := fmt.Sprintf("Processed %d agent files: %d imported, %d skipped, %d errors",
		len(files), len(imported), len(skipped), len(errors))

	return ImportAgentsResult{
		Success:  len(errors) == 0,
		Message:  message,
		Imported: imported,
		Skipped:  skipped,
		Errors:   errors,
		Count:    len(imported),
	}, nil
}

// ChatSend implements the i-chat-send tool
func (t *Tools) ChatSend(args ChatSendArgs) (ChatSendResult, error) {
	// Generate message ID
	messageID := utils.GenerateGitStyleAgentID()

	// Set defaults
	senderRole := "agent"
	if args.SenderRole != nil {
		senderRole = *args.SenderRole
	}

	recipientRole := "user"
	if args.RecipientRole != nil {
		recipientRole = *args.RecipientRole
	}

	provider := "claude"
	if args.Provider != nil {
		provider = *args.Provider
	}

	// Determine direction based on roles
	direction := "outbound"
	if senderRole == "user" {
		direction = "outbound"
	} else if senderRole == "agent" || senderRole == "system" {
		direction = "inbound"
	}

	// Insert message into database
	// Use UTC timestamp to match Phoenix format (Z suffix instead of timezone offset)
	now := time.Now().UTC().Format(time.RFC3339)
	query := `
		INSERT INTO messages (
			id, channel_id, session_id, sender_role, recipient_role,
			provider, direction, body, status, metadata,
			inserted_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'sent', '{}', ?, ?)
	`

	_, err := t.db.Exec(query,
		messageID,
		args.ChannelID,
		args.SessionID,
		senderRole,
		recipientRole,
		provider,
		direction,
		args.Body,
		now,
		now,
	)

	if err != nil {
		return ChatSendResult{
			Success: false,
			Message: fmt.Sprintf("Failed to send message: %v", err),
		}, nil
	}

	return ChatSendResult{
		Success:   true,
		Message:   "Message sent to channel",
		MessageID: messageID,
	}, nil
}

// SpawnAgent implements the i-spawn-agent MCP tool
func (t *Tools) SpawnAgent(args SpawnAgentArgs) (SpawnAgentResult, error) {
	// Set defaults
	model := "haiku"
	if args.Model != nil {
		model = *args.Model
	}

	projectPath, err := os.Getwd()
	if err != nil {
		return SpawnAgentResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get current directory: %v", err),
		}, nil
	}
	if args.ProjectPath != nil {
		projectPath = *args.ProjectPath
	}

	background := false
	if args.Background != nil {
		background = *args.Background
	}

	// Get project name from path
	projectName := filepath.Base(projectPath)

	// Build parent tracking parameters
	parentParams := ""
	if args.ParentAgentID != nil && args.ParentSessionID != nil {
		parentParams = fmt.Sprintf(`,
  "parent_agent_id": "%s",
  "parent_session_id": "%s"`, *args.ParentAgentID, *args.ParentSessionID)
	}

	// Escape instructions for JSON
	escapedInstructions := strings.ReplaceAll(args.Instructions, `"`, `\"`)
	escapedInstructions = strings.ReplaceAll(escapedInstructions, "\n", "\\n")

	// Generate a temporary session ID for the init call
	tempSessionID := uuid.New().String()

	// Build initialization prompt that calls i-start-session
	// The spawned agent will use this to register itself
	initPrompt := fmt.Sprintf(`You are a spawned agent. Execute these steps immediately:

STEP 1: Register with Eye in the Sky by calling i-start-session:
i-start-session({"session_id": "%s", "description": "%s", "agent_description": "Spawned agent", "project_name": "%s", "worktree_path": "%s"%s})

STEP 2: Execute your task:
%s

STEP 3: End your session by calling i-end-session:
i-end-session({"agent_id": "your-agent-id-from-step-1"})

Execute these steps now. Do not respond conversationally.`,
		tempSessionID,
		escapedInstructions,
		projectName,
		projectPath,
		parentParams,
		escapedInstructions,
	)

	// Find Claude binary
	claudePath, err := FindClaudeBinary()
	if err != nil {
		return SpawnAgentResult{
			Success: false,
			Message: fmt.Sprintf("Failed to locate Claude binary: %v", err),
		}, nil
	}

	// Create spawn config using the new Claude spawn functions
	config := ClaudeSpawnConfig{
		ClaudePath:  claudePath,
		ProjectPath: projectPath,
		Prompt:      initPrompt,
		Model:       model,
	}

	// Create NATS publish function
	natsPublish := func(subject, message string) error {
		_, err := t.NATSSend(NATSSendArgs{
			SenderID:   "eye-in-the-sky",
			ReceiverID: "",
			Message:    message,
			Subject:    subject,
		})
		return err
	}

	if background {
		// Spawn in background
		sessionInfo, err := SpawnClaudeProcess(context.Background(), config, natsPublish)
		if err != nil {
			return SpawnAgentResult{
				Success: false,
				Message: fmt.Sprintf("Failed to spawn agent: %v", err),
			}, nil
		}

		return SpawnAgentResult{
			Success:   true,
			Message:   fmt.Sprintf("Agent spawned in background. Awaiting session registration (temp session: %s)", tempSessionID),
			SessionID: sessionInfo.SessionID, // Will be populated once init message is received
		}, nil
	}

	// Spawn in foreground (blocking)
	sessionInfo, err := SpawnClaudeProcess(context.Background(), config, natsPublish)
	if err != nil {
		return SpawnAgentResult{
			Success: false,
			Message: fmt.Sprintf("Failed to spawn agent: %v", err),
		}, nil
	}

	return SpawnAgentResult{
		Success:   true,
		Message:   fmt.Sprintf("Agent completed. Session ID: %s", sessionInfo.SessionID),
		SessionID: sessionInfo.SessionID,
	}, nil
}

// AddProject implements the i-project-add MCP tool
func (t *Tools) AddProject(args ProjectAddArgs) (ProjectAddResult, error) {
	// Generate project ID
	projectID := uuid.New().String()

	// Set default active value
	active := true
	if args.Active != nil {
		active = *args.Active
	}

	// Insert project into database
	query := `
		INSERT INTO projects (
			id, name, slug, path, remote_url, git_remote, repo_url,
			branch, created_at, updated_at, active
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := nowString()
	_, err := t.db.Exec(query,
		projectID,
		args.Name,
		args.Slug,
		args.Path,
		args.RemoteURL,
		args.GitRemote,
		args.RepoURL,
		args.Branch,
		now,
		now,
		active,
	)

	if err != nil {
		return ProjectAddResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create project: %v", err),
		}, nil
	}

	return ProjectAddResult{
		Success:   true,
		Message:   fmt.Sprintf("Project '%s' created successfully", args.Name),
		ProjectID: projectID,
	}, nil
}

// SpawnClaude implements the Claude Code process spawning MCP tool
func (t *Tools) SpawnClaude(args SpawnClaudeArgs) (SpawnClaudeResult, error) {
	// Validate required arguments
	if args.Prompt == "" {
		return SpawnClaudeResult{
			Success: false,
			Message: "prompt is required",
		}, nil
	}

	if args.Model == "" {
		return SpawnClaudeResult{
			Success: false,
			Message: "model is required",
		}, nil
	}

	// Determine project path
	projectPath := "."
	if args.ProjectPath != nil && *args.ProjectPath != "" {
		projectPath = *args.ProjectPath
	}

	// Find Claude binary
	claudePath, err := FindClaudeBinary()
	if err != nil {
		return SpawnClaudeResult{
			Success: false,
			Message: fmt.Sprintf("Failed to locate Claude binary: %v", err),
		}, nil
	}

	// Create spawn config
	config := ClaudeSpawnConfig{
		ClaudePath:  claudePath,
		ProjectPath: projectPath,
		Prompt:      args.Prompt,
		Model:       args.Model,
	}

	// Create NATS publish function using NATSSend
	natsPublish := func(subject, message string) error {
		_, err := t.NATSSend(NATSSendArgs{
			SenderID:   "eye-in-the-sky",
			ReceiverID: "",
			Message:    message,
			Subject:    subject,
		})
		return err
	}

	// Spawn the process
	sessionInfo, err := SpawnClaudeProcess(context.Background(), config, natsPublish)
	if err != nil {
		return SpawnClaudeResult{
			Success: false,
			Message: fmt.Sprintf("Failed to spawn Claude process: %v", err),
		}, nil
	}

	return SpawnClaudeResult{
		Success:   true,
		Message:   fmt.Sprintf("Claude Code process spawned with PID %d", sessionInfo.PID),
		SessionID: sessionInfo.SessionID,
		PID:       sessionInfo.PID,
	}, nil
}

// now returns the current time for use in Go structs.
// Returns UTC time without timezone or monotonic clock reading.
func now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

// nowString returns the current time formatted for SQLite/Ecto compatibility.
// Returns UTC time as string without timezone or monotonic clock reading.
// Format: "2006-01-02 15:04:05.999999" (naive datetime)
func nowString() string {
	return now().Format("2006-01-02 15:04:05.999999")
}
