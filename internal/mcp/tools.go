package mcp

import (
	"context"
	"fmt"
	"os/exec"
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


// Help implements the help MCP tool
func (t *Tools) Help(args HelpArgs) (HelpResult, error) {
	if args.Tool != nil {
		// Return detailed help for specific tool
		toolHelp := t.getToolHelp(*args.Tool)
		if toolHelp == nil {
			return HelpResult{Success: false, Message: fmt.Sprintf("Tool '%s' not found", *args.Tool)}, nil
		}
		return HelpResult{
			Success: true,
			Message: fmt.Sprintf("Help for tool '%s'", *args.Tool),
			Tools:   []Tool{*toolHelp},
		}, nil
	}

	// Return help for all tools
	tools := t.getAllToolsWithHelp()
	return HelpResult{
		Success: true,
		Message: "Eye in the Sky - Claude Code Multi-Agent Management System\n\nAvailable tools for tracking and managing Claude Code agent activities:",
		Tools:   tools,
	}, nil
}

// getToolHelp returns detailed help for a specific tool
func (t *Tools) getToolHelp(toolName string) *Tool {
	tools := t.getAllToolsWithHelp()
	for _, tool := range tools {
		if tool.Name == toolName {
			return &tool
		}
	}
	return nil
}

// getAllToolsWithHelp returns all tools with comprehensive documentation
func (t *Tools) getAllToolsWithHelp() []Tool {
	return []Tool{
		{
			Name:        "register_agent",
			Description: "Register a new Claude Code agent to start tracking activities",
			Instructions: `Register a new agent before starting any work. If no agent_id is provided, a git-style 8-character hash will be auto-generated.

This tool:
- Creates a new agent record in the database
- Auto-generates git-style hash ID if not provided
- Sets initial status to 'active'
- Logs the registration action
- Enables tracking for all subsequent activities

Required before using any other tools for this agent.`,
			Parameters: map[string]string{
				"agent_id":      "Unique 8-character hex identifier (optional - auto-generated if not provided)",
				"description":   "Brief description of what the agent will work on (required)",
				"worktree_path": "Path to the git repository (optional)",
				"project_name":  "Name of the project being worked on (optional)",
			},
			Examples: []string{
				`{"description": "Working on user authentication system"}`,
				`{"agent_id": "abc123de", "description": "Working on user authentication system", "project_name": "MyApp"}`,
				`{"description": "Frontend dashboard development", "worktree_path": "/path/to/project", "project_name": "Eye in the Sky"}`,
			},
		},
		{
			Name:        "register_claude_desktop_agent",
			Description: "Register a new Claude Desktop agent to start tracking activities",
			Instructions: `Register a new Claude Desktop agent before starting any work. This tool is specifically for agents running in Claude Desktop (not git worktrees). If no agent_id is provided, a git-style 8-character hash will be auto-generated.

This tool:
- Creates a new agent record with 'desktop' source
- Auto-generates git-style hash ID if not provided
- Sets initial status to 'active'
- Logs the registration action with project context
- Does not require git worktree paths`,
			Parameters: map[string]string{
				"agent_id":     "Unique 8-character hex identifier (optional - auto-generated if not provided)",
				"description":  "Brief description of what the agent will work on (required)",
				"project_name": "Name of the project being worked on (required)",
				"window_id":    "Claude Desktop window identifier for window management (optional)",
			},
			Examples: []string{
				`{"description": "Building user authentication system", "project_name": "MyApp"}`,
				`{"agent_id": "desk1234", "description": "Building user authentication system", "project_name": "MyApp"}`,
				`{"description": "Frontend component development", "project_name": "Dashboard"}`,
				`{"description": "Working on API endpoints", "project_name": "Backend", "window_id": "win_abc123"}`,
			},
		},
		{
			Name:        "update_status",
			Description: "Update agent status and current task being worked on",
			Instructions: `Update the agent's current status and what they're working on. This helps track progress and current focus.

Valid statuses:
- 'active': Agent is available and ready to work
- 'working': Agent is actively working on a task
- 'idle': Agent is paused or waiting
- 'completed': Agent has finished all work
- 'failed': Agent encountered an error

Updates the last activity timestamp automatically.`,
			Parameters: map[string]string{
				"agent_id":     "8-character agent identifier (required)",
				"status":       "One of: active, working, idle, completed, failed (required)",
				"current_task": "Description of current task (optional)",
			},
			Examples: []string{
				`{"agent_id": "abc123de", "status": "working", "current_task": "Implementing JWT validation"}`,
				`{"agent_id": "web45678", "status": "completed"}`,
			},
		},
		{
			Name:        "log_action",
			Description: "Log agent activities and actions for audit trail",
			Instructions: `Log important activities performed by the agent. This creates an audit trail of all work done.

Action types:
- 'task_start': Starting a new task or phase
- 'file_operation': Creating, editing, or deleting files
- 'git_commit': Making git commits (use log_commits for detailed commit tracking)
- 'status_update': Changing status or current task

The details field can contain JSON for structured information.`,
			Parameters: map[string]string{
				"agent_id":     "8-character agent identifier (required)",
				"action_type":  "One of: task_start, file_operation, git_commit, status_update (required)",
				"description":  "Human-readable description of the action (required)",
				"details":      "Additional structured information as JSON string (optional)",
			},
			Examples: []string{
				`{"agent_id": "abc123de", "action_type": "task_start", "description": "Started implementing user authentication"}`,
				`{"agent_id": "web45678", "action_type": "file_operation", "description": "Created login component", "details": "{\"file\": \"src/components/Login.tsx\", \"lines\": 45}"}`,
			},
		},
		{
			Name:        "log_commits",
			Description: "Track git commits made by the agent",
			Instructions: `Record git commits for tracking code changes. This provides a detailed history of code modifications.

You can log multiple commits at once. Commit messages are optional but recommended for better tracking.

This automatically updates the agent's last activity timestamp.`,
			Parameters: map[string]string{
				"agent_id":        "8-character agent identifier (required)",
				"commit_hashes":   "Array of git commit hash strings (required)",
				"commit_messages": "Array of commit messages, same order as hashes (optional)",
			},
			Examples: []string{
				`{"agent_id": "abc123de", "commit_hashes": ["a1b2c3d"], "commit_messages": ["Add JWT middleware"]}`,
				`{"agent_id": "web45678", "commit_hashes": ["e4f5g6h", "i7j8k9l"], "commit_messages": ["Fix login bug", "Add error handling"]}`,
			},
		},
		{
			Name:        "end_session",
			Description: "Complete agent session with summary and final status",
			Instructions: `End the agent's work session. This should be called when all work is complete or if the agent encounters a fatal error.

The final status should reflect the outcome:
- 'completed': All work finished successfully
- 'failed': Work stopped due to errors

Provide a summary of what was accomplished for better tracking.`,
			Parameters: map[string]string{
				"agent_id":     "8-character agent identifier (required)",
				"summary":      "Summary of work completed (optional but recommended)",
				"final_status": "Either 'completed' or 'failed' (optional, defaults to 'completed')",
			},
			Examples: []string{
				`{"agent_id": "abc123de", "summary": "Successfully implemented JWT authentication with tests", "final_status": "completed"}`,
				`{"agent_id": "web45678", "summary": "Failed to deploy due to configuration issues", "final_status": "failed"}`,
			},
		},
		{
			Name:        "help",
			Description: "Get help information about available tools",
			Instructions: `Get help and documentation for the Eye in the Sky MCP tools.

Call without arguments to get help for all tools, or specify a tool name to get detailed help for that specific tool.

This tool provides comprehensive documentation including parameters, examples, and usage instructions.`,
			Parameters: map[string]string{
				"tool": "Name of specific tool to get help for (optional)",
			},
			Examples: []string{
				`{}`,
				`{"tool": "register_agent"}`,
				`{"tool": "log_action"}`,
			},
		},
		{
			Name:        "save_session_context",
			Description: "Save current session state for resumption later",
			Instructions: `Save the current session context to enable resuming work later. This captures:
- Current progress and phase
- Completed and pending tasks
- Important decisions and notes
- Key files and dependencies
- Metrics and environment state

This enables pausing and resuming sessions across different Claude Code instances.`,
			Parameters: map[string]string{
				"agent_id":         "8-character agent identifier (required)",
				"current_phase":    "Current work phase or milestone (required)",
				"progress":         "Progress information with completion percentage and goals (optional)",
				"next_actions":     "Array of next steps to take (optional)",
				"completed_tasks":  "Array of completed tasks (optional)",
				"pending_tasks":    "Array of remaining tasks (optional)",
				"key_decisions":    "Array of important decisions made (optional)",
				"important_files":  "Array of key files modified or created (optional)",
				"dependencies":     "Array of dependencies or blockers (optional)",
				"notes":           "Array of contextual notes and observations (optional)",
				"environment":     "Key-value pairs of environment info (optional)",
				"metrics":         "Performance and progress metrics (optional)",
				"auto_save":       "Whether to automatically update agent status to idle (optional)",
			},
			Examples: []string{
				`{"agent_id": "abc123de", "current_phase": "Authentication implementation", "next_actions": ["Implement JWT validation", "Add password hashing"]}`,
				`{"agent_id": "web45678", "current_phase": "Dashboard frontend", "completed_tasks": ["User login component", "Navigation bar"], "pending_tasks": ["User profile page"], "auto_save": true}`,
			},
		},
		{
			Name:        "load_session_context",
			Description: "Load previous session state to resume work",
			Instructions: `Load the most recent session context for an agent to resume work where it was left off. This retrieves:
- Progress and current phase
- Completed and pending tasks
- Important decisions and notes
- Key files and dependencies
- Previous metrics and environment

Enables seamless session resumption across Claude Code instances.`,
			Parameters: map[string]string{
				"agent_id":   "8-character agent identifier (required)",
				"session_id": "Specific session ID to load (optional - loads latest if not provided)",
			},
			Examples: []string{
				`{"agent_id": "abc123de"}`,
				`{"agent_id": "web45678", "session_id": "web45678_1696847200"}`,
			},
		},
		{
			Name:        "add_session_note",
			Description: "Add contextual notes to the current session",
			Instructions: `Add timestamped notes to the current session context. Useful for:
- Recording insights and observations
- Noting important decisions
- Leaving reminders for later
- Documenting blockers or issues

Note types: insight, reminder, warning, idea
Priority levels: low, medium, high`,
			Parameters: map[string]string{
				"agent_id": "8-character agent identifier (required)",
				"type":     "Note type: insight, reminder, warning, idea (required)",
				"content":  "Note content and description (required)",
				"priority": "Priority level: low, medium, high (optional, defaults to medium)",
				"tags":     "Array of tags for categorization (optional)",
			},
			Examples: []string{
				`{"agent_id": "abc123de", "type": "insight", "content": "JWT token validation works better with async/await pattern"}`,
				`{"agent_id": "web45678", "type": "reminder", "content": "Need to add error handling for API calls", "priority": "high"}`,
				`{"agent_id": "api67890", "type": "warning", "content": "Database migration needed before deploy", "priority": "high", "tags": ["deployment", "database"]}`,
			},
		},
		{
			Name:        "get_current_window",
			Description: "Get information about the current active window (macOS only)",
			Instructions: `Detect the currently active window and return information about it. This tool helps identify which window/application is currently in focus.

This tool:
- Detects the frontmost application
- Gets window title, position, and size
- Generates a window ID for tracking
- Only works on macOS systems

Useful for registering agents with accurate window information.`,
			Parameters: map[string]string{
				// No parameters needed
			},
			Examples: []string{
				`{}`,
			},
		},
		{
			Name:        "bring_window_front",
			Description: "Bring an agent's window to the front (macOS only)",
			Instructions: `Bring the specified agent's window to the front, making it the active window. This is useful for quickly switching focus to a specific agent's workspace.

This tool:
- Looks up the agent's window information
- Uses the stored window ID to identify the application
- Brings the application/window to the front
- Falls back to bringing terminal (ghostty) to front if no specific window ID

Only works on macOS systems with AppleScript support.`,
			Parameters: map[string]string{
				"agent_id": "8-character agent identifier (required)",
			},
			Examples: []string{
				`{"agent_id": "534002f0"}`,
				`{"agent_id": "abc123de"}`,
			},
		},
	}
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
