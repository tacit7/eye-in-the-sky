package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/tacit7/eye-in-the-sky/internal/database"
)

const usage = `eits-cli - Eye in the Sky CLI tool

Usage:
  eits-cli [options] <command> [command-options]

Commands:
  start-session <session-id>    Start a new agent session (returns agent UUID)
  get-latest-agent <session-id> Get most recent agent for session (returns agent UUID)
  update-status                 Update agent status
  log-action                    Log an agent action
  log-commits                   Log git commits
  log-compaction                Log conversation compaction event
  end-session                   End an agent session
  help                          Show this help message

Global Options:
  --db PATH        Database path (default: ~/.config/eye-in-the-sky/eits.db)
  --json           Output JSON instead of human-readable text
  --help           Show help

Examples:
  # Start session from hook
  AGENT_ID=$(eits-cli start-session "$SESSION_ID")

  # Get latest agent for session
  AGENT_ID=$(eits-cli get-latest-agent "$SESSION_ID")

  # Update status
  eits-cli update-status --agent-id "$AGENT_ID" --status working --task "Implementing auth"

  # Log action
  eits-cli log-action --agent-id "$AGENT_ID" --type task_start --desc "Starting new task"

  # Log compaction (PreCompact hook)
  eits-cli log-compaction --agent-id "$AGENT_ID" --session-id "$SESSION_ID" --transcript "$TRANSCRIPT_PATH" --trigger "auto"

For command-specific help:
  eits-cli <command> --help
`

type globalFlags struct {
	dbPath     string
	jsonOutput bool
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}

	// Parse global flags
	global := &globalFlags{}
	flagSet := flag.NewFlagSet("global", flag.ContinueOnError)
	flagSet.StringVar(&global.dbPath, "db", "", "Database path")
	flagSet.BoolVar(&global.jsonOutput, "json", false, "Output JSON")

	// Find the command position
	commandIdx := 1
	for i := 1; i < len(os.Args); i++ {
		if !strings.HasPrefix(os.Args[i], "-") {
			commandIdx = i
			break
		}
	}

	// Parse global flags before command
	if commandIdx > 1 {
		flagSet.Parse(os.Args[1:commandIdx])
	}

	// Get database path
	if global.dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fatal("Failed to get home directory: %v", err)
		}
		global.dbPath = filepath.Join(homeDir, ".config", "eye-in-the-sky", "eits.db")
	}

	// Initialize database
	db, err := database.New(global.dbPath)
	if err != nil {
		fatal("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Get command
	command := os.Args[commandIdx]
	args := os.Args[commandIdx+1:]

	// Route to command handler
	switch command {
	case "start-session":
		handleStartSession(db, args, global)
	case "get-latest-agent":
		handleGetLatestAgent(db, args, global)
	case "update-status":
		handleUpdateStatus(db, args, global)
	case "log-action":
		handleLogAction(db, args, global)
	case "log-commits":
		handleLogCommits(db, args, global)
	case "log-compaction":
		handleLogCompaction(db, args, global)
	case "end-session":
		handleEndSession(db, args, global)
	case "help", "--help", "-h":
		fmt.Println(usage)
		os.Exit(0)
	default:
		fatal("Unknown command: %s\n\n%s", command, usage)
	}
}

// Command: start-session <session-id>
func handleStartSession(db *database.DB, args []string, global *globalFlags) {
	fs := flag.NewFlagSet("start-session", flag.ExitOnError)
	description := fs.String("description", "Auto-started session", "Session description (optional)")
	worktreePath := fs.String("worktree", "", "Git worktree path (optional, defaults to cwd)")
	parentAgentID := fs.String("parent-agent-id", "", "Parent agent ID (optional)")
	projectName := fs.String("project", "", "Project name (optional)")

	fs.Parse(args)

	// Get session ID from positional argument
	if len(fs.Args()) < 1 {
		fatal("Usage: start-session <session-id> [options]")
	}
	sessionID := fs.Args()[0]

	// Get worktree path
	if *worktreePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fatal("Failed to get current directory: %v", err)
		}
		*worktreePath = cwd
	}

	// Generate agent ID
	agentID := uuid.New().String()

	// Create agent with pointer fields
	agent := &database.Agent{
		ID:                 agentID,
		Status:             "active",
		Source:             "worktree",
		GitWorktreePath:    worktreePath,
		FeatureDescription: description,
		SessionID:          &sessionID,
	}

	if *projectName != "" {
		agent.ProjectName = projectName
	}
	if *parentAgentID != "" {
		agent.ParentAgentID = parentAgentID
	}

	if err := db.CreateAgent(agent); err != nil {
		fatal("Failed to create agent: %v", err)
	}

	// Output: just UUID by default, JSON if --json flag
	if global.jsonOutput {
		result := map[string]interface{}{
			"success":    true,
			"message":    fmt.Sprintf("Session %s started for agent %s", sessionID, agentID),
			"agent_id":   agentID,
			"session_id": sessionID,
		}
		output(global, result)
	} else {
		fmt.Println(agentID)
	}
}

// Command: get-latest-agent <session-id>
func handleGetLatestAgent(db *database.DB, args []string, global *globalFlags) {
	if len(args) < 1 {
		fatal("Usage: get-latest-agent <session-id>")
	}
	sessionID := args[0]

	agent, err := db.GetAgentBySessionID(sessionID)
	if err != nil {
		fatal("Failed to get agent: %v", err)
	}

	if agent == nil {
		fatal("No agent found for session: %s", sessionID)
	}

	// Output: just UUID by default, JSON if --json flag
	if global.jsonOutput {
		result := map[string]interface{}{
			"success":    true,
			"agent_id":   agent.ID,
			"session_id": sessionID,
			"status":     agent.Status,
		}
		output(global, result)
	} else {
		fmt.Println(agent.ID)
	}
}

// Command: update-status
func handleUpdateStatus(db *database.DB, args []string, global *globalFlags) {
	fs := flag.NewFlagSet("update-status", flag.ExitOnError)
	agentID := fs.String("agent-id", "", "Agent ID (required)")
	status := fs.String("status", "", "Status: active, working, idle, completed, failed (required)")
	currentTask := fs.String("task", "", "Current task description (optional)")

	fs.Parse(args)

	if *agentID == "" || *status == "" {
		fatal("--agent-id and --status are required")
	}

	// Update agent using UpdateAgentStatus
	var taskPtr *string
	if *currentTask != "" {
		taskPtr = currentTask
	}

	if err := db.UpdateAgentStatus(*agentID, *status, taskPtr); err != nil {
		fatal("Failed to update agent: %v", err)
	}

	result := map[string]interface{}{
		"success": true,
		"message": "Status updated successfully",
	}

	output(global, result)
}

// Command: log-action
func handleLogAction(db *database.DB, args []string, global *globalFlags) {
	fs := flag.NewFlagSet("log-action", flag.ExitOnError)
	agentID := fs.String("agent-id", "", "Agent ID (required)")
	actionType := fs.String("type", "", "Action type: task_start, file_operation, git_commit, status_update (required)")
	desc := fs.String("desc", "", "Action description (required)")
	details := fs.String("details", "", "Additional details as JSON string (optional)")

	fs.Parse(args)

	if *agentID == "" || *actionType == "" || *desc == "" {
		fatal("--agent-id, --type, and --desc are required")
	}

	action := &database.Action{
		AgentID:     *agentID,
		ActionType:  *actionType,
		Description: *desc,
	}

	if *details != "" {
		action.Details = details
	}

	if err := db.CreateAction(action); err != nil {
		fatal("Failed to log action: %v", err)
	}

	result := map[string]interface{}{
		"success": true,
		"message": "Action logged successfully",
	}

	output(global, result)
}

// Command: log-commits
func handleLogCommits(db *database.DB, args []string, global *globalFlags) {
	fs := flag.NewFlagSet("log-commits", flag.ExitOnError)
	agentID := fs.String("agent-id", "", "Agent ID (required)")
	hashes := fs.String("hashes", "", "Comma-separated commit hashes (required)")
	messages := fs.String("messages", "", "Comma-separated commit messages (optional)")

	fs.Parse(args)

	if *agentID == "" || *hashes == "" {
		fatal("--agent-id and --hashes are required")
	}

	hashList := strings.Split(*hashes, ",")
	for i := range hashList {
		hashList[i] = strings.TrimSpace(hashList[i])
	}

	var messageList []string
	if *messages != "" {
		messageList = strings.Split(*messages, ",")
		for i := range messageList {
			messageList[i] = strings.TrimSpace(messageList[i])
		}
	}

	if err := db.CreateCommits(*agentID, hashList, messageList); err != nil {
		fatal("Failed to log commits: %v", err)
	}

	result := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Logged %d commit(s)", len(hashList)),
	}

	output(global, result)
}

// Command: log-compaction
func handleLogCompaction(db *database.DB, args []string, global *globalFlags) {
	fs := flag.NewFlagSet("log-compaction", flag.ExitOnError)
	agentID := fs.String("agent-id", "", "Agent ID (required)")
	sessionID := fs.String("session-id", "", "Session ID (required)")
	transcript := fs.String("transcript", "", "Path to transcript JSONL file (optional)")
	trigger := fs.String("trigger", "", "Trigger type: manual or auto (optional)")
	summary := fs.String("summary", "", "Compaction summary (optional)")

	fs.Parse(args)

	if *agentID == "" || *sessionID == "" {
		fatal("--agent-id and --session-id are required")
	}

	// Build summary text from trigger and summary
	var summaryText string
	if *trigger != "" {
		summaryText = fmt.Sprintf("Trigger: %s", *trigger)
		if *summary != "" {
			summaryText += fmt.Sprintf(", %s", *summary)
		}
	} else if *summary != "" {
		summaryText = *summary
	}

	// Get file size and message count if transcript provided
	var fileSize *int64
	var messageCount *int
	if *transcript != "" {
		info, err := os.Stat(*transcript)
		if err == nil {
			size := info.Size()
			fileSize = &size

			// Count lines in JSONL file (each line is a message)
			file, err := os.Open(*transcript)
			if err == nil {
				scanner := bufio.NewScanner(file)
				count := 0
				for scanner.Scan() {
					count++
				}
				file.Close()
				messageCount = &count
			}
		}
	}

	compaction := &database.Compaction{
		AgentID:   *agentID,
		SessionID: *sessionID,
	}

	if summaryText != "" {
		compaction.Summary = &summaryText
	}
	if *transcript != "" {
		compaction.JsonlFilePath = transcript
	}
	if fileSize != nil {
		compaction.JsonlFileSize = fileSize
	}
	if messageCount != nil {
		compaction.MessageCount = messageCount
	}

	if err := db.CreateCompaction(compaction); err != nil {
		fatal("Failed to log compaction: %v", err)
	}

	result := map[string]interface{}{
		"success": true,
		"message": "Compaction logged successfully",
	}

	output(global, result)
}

// Command: end-session
func handleEndSession(db *database.DB, args []string, global *globalFlags) {
	fs := flag.NewFlagSet("end-session", flag.ExitOnError)
	agentID := fs.String("agent-id", "", "Agent ID (required)")
	finalStatus := fs.String("status", "completed", "Final status: completed or failed")
	summary := fs.String("summary", "", "Session summary (optional)")

	fs.Parse(args)

	if *agentID == "" {
		fatal("--agent-id is required")
	}

	if err := db.EndAgentSession(*agentID, *summary, *finalStatus); err != nil {
		fatal("Failed to end session: %v", err)
	}

	result := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Session ended with status: %s", *finalStatus),
	}

	output(global, result)
}

// Utility functions
func output(global *globalFlags, result map[string]interface{}) {
	if global.jsonOutput {
		json.NewEncoder(os.Stdout).Encode(result)
	} else {
		fmt.Println(result["message"])
	}
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
