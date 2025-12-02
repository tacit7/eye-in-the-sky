package mcp

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tacit7/eye-in-the-sky/internal/database"
)

// openEditorForInput opens the user's preferred editor to capture input
func openEditorForInput() (string, error) {
	// Get editor from environment
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim" // fallback to vim
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "eits-note-*.md")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Open editor
	cmd := exec.Command(editor, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	// Read the file content
	content, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to read temp file: %w", err)
	}

	return string(content), nil
}

// HandleCLI handles command-line interface mode
func HandleCLI(args []string, db *database.DB) int {
	if len(args) < 2 {
		printCLIUsage()
		return 1
	}

	command := args[1]
	tools := NewTools(db)

	switch command {
	case "i-speak":
		return handleSpeakCLI(args[2:], tools)
	case "i-start-session":
		return handleStartSessionCLI(args[2:], tools)
	case "i-get-ids":
		return handleGetIDsCLI(args[2:], tools)
	case "i-note", "i-note-add":
		return handleNoteCLI(args[2:], tools)
	case "i-log":
		return handleLogCLI(args[2:], tools)
	case "i-commits":
		return handleCommitsCLI(args[2:], tools)
	case "i-end", "i-end-session":
		return handleEndSessionCLI(args[2:], tools)
	case "i-update-status", "i-status":
		return handleUpdateStatusCLI(args[2:], tools)
	case "i-action":
		return handleActionCLI(args[2:], tools)
	case "i-context-set":
		return handleContextSetCLI(args[2:], tools)
	case "i-update-description":
		return handleUpdateDescriptionCLI(args[2:], tools)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printCLIUsage()
		return 1
	}
}

func printCLIUsage() {
	fmt.Println("Eye in the Sky - CLI Mode")
	fmt.Println("\nUsage: eye-in-the-sky <command> [flags]")
	fmt.Println("\nCommands:")
	fmt.Println("  i-speak              Speak a message aloud")
	fmt.Println("  i-start-session      Start a new agent session")
	fmt.Println("  i-get-ids            Get session_id and agent_id from SESSION env var")
	fmt.Println("  i-update-status      Update agent status")
	fmt.Println("  i-action             Log an agent action")
	fmt.Println("  i-note               Add a note")
	fmt.Println("  i-log                Add a log entry")
	fmt.Println("  i-commits            Log git commits")
	fmt.Println("  i-context-set        Set session context")
	fmt.Println("  i-update-description Update agent description")
	fmt.Println("  i-end                End an agent session")
	fmt.Println("\nRun without arguments to start MCP stdio server")
}

func handleSpeakCLI(args []string, tools *Tools) int {
	fs := flag.NewFlagSet("i-speak", flag.ExitOnError)
	message := fs.String("message", "", "Message to speak")
	voice := fs.String("voice", "Ava (Premium)", "Voice to use")
	fs.Parse(args)

	if *message == "" {
		fmt.Fprintln(os.Stderr, "Error: --message is required")
		return 1
	}

	result, err := tools.ISpeak(ISpeakArgs{
		Message: *message,
		Voice:   voice,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if result.Success {
		fmt.Printf("✓ %s (voice: %s)\n", result.Message, result.VoiceUsed)
	} else {
		fmt.Printf("✗ %s\n", result.Message)
		return 1
	}
	return 0
}

func handleStartSessionCLI(args []string, tools *Tools) int {
	fs := flag.NewFlagSet("i-start-session", flag.ExitOnError)
	sessionID := fs.String("session-id", "", "Session ID (UUID)")
	description := fs.String("description", "", "Session description")
	worktreePath := fs.String("worktree-path", "", "Git worktree path (optional)")
	fs.Parse(args)

	if *sessionID == "" {
		fmt.Fprintln(os.Stderr, "Error: --session-id is required")
		return 1
	}

	result, err := tools.StartSession(StartSessionArgs{
		SessionID:      *sessionID,
		Description:    *description,
		WorktreePath:   worktreePath,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Printf("✓ Session started\n")
	fmt.Printf("  Agent ID: %s\n", result.AgentID)
	fmt.Printf("  Session ID: %s\n", *sessionID)
	return 0
}

func handleNoteCLI(args []string, tools *Tools) int {
	fs := flag.NewFlagSet("i-note", flag.ExitOnError)
	parentID := fs.String("parent-id", "", "Parent entity ID")
	parentType := fs.String("parent-type", "", "Parent entity type (sessions, agents, projects, global)")
	body := fs.String("body", "", "Note content (optional - will open EDITOR if not provided)")
	fs.Parse(args)

	if *parentID == "" || *parentType == "" {
		fmt.Fprintln(os.Stderr, "Error: --parent-id and --parent-type are required")
		return 1
	}

	// If no body provided, open editor
	noteBody := *body
	if noteBody == "" {
		content, err := openEditorForInput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening editor: %v\n", err)
			return 1
		}
		noteBody = content
	}

	// Don't create empty notes
	if strings.TrimSpace(noteBody) == "" {
		fmt.Fprintln(os.Stderr, "Error: note body cannot be empty")
		return 1
	}

	result, err := tools.AddNote(AddNoteArgs{
		ParentID:   *parentID,
		ParentType: *parentType,
		Body:       noteBody,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Printf("✓ %s\n", result.Message)
	return 0
}

func handleLogCLI(args []string, tools *Tools) int {
	fs := flag.NewFlagSet("i-log", flag.ExitOnError)
	sessionID := fs.String("session-id", "", "Session ID")
	logType := fs.String("type", "info", "Log type")
	message := fs.String("message", "", "Log message")
	fs.Parse(args)

	if *sessionID == "" || *message == "" {
		fmt.Fprintln(os.Stderr, "Error: --session-id and --message are required")
		return 1
	}

	result, err := tools.AddLog(AddLogArgs{
		SessionID: *sessionID,
		Type:      *logType,
		Message:   *message,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Printf("✓ %s\n", result.Message)
	return 0
}

func handleCommitsCLI(args []string, tools *Tools) int {
	fs := flag.NewFlagSet("i-commits", flag.ExitOnError)
	agentID := fs.String("agent-id", "", "Agent ID")
	hashes := fs.String("commit-hashes", "", "Comma-separated commit hashes")
	messages := fs.String("commit-messages", "", "Comma-separated commit messages (optional)")
	fs.Parse(args)

	if *agentID == "" || *hashes == "" {
		fmt.Fprintln(os.Stderr, "Error: --agent-id and --commit-hashes are required")
		return 1
	}

	hashList := strings.Split(*hashes, ",")
	var messageList []string
	if *messages != "" {
		messageList = strings.Split(*messages, ",")
	}

	result, err := tools.LogCommits(LogCommitsArgs{
		AgentID:        *agentID,
		CommitHashes:   hashList,
		CommitMessages: messageList,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Printf("✓ %s\n", result.Message)
	return 0
}

func handleEndSessionCLI(args []string, tools *Tools) int {
	fs := flag.NewFlagSet("i-end", flag.ExitOnError)
	agentID := fs.String("agent-id", "", "Agent ID")
	summary := fs.String("summary", "", "Session summary (optional)")
	finalStatus := fs.String("final-status", "completed", "Final status (completed/failed)")
	fs.Parse(args)

	if *agentID == "" {
		fmt.Fprintln(os.Stderr, "Error: --agent-id is required")
		return 1
	}

	result, err := tools.EndSession(EndSessionArgs{
		AgentID:     *agentID,
		Summary:     summary,
		FinalStatus: finalStatus,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Printf("✓ %s\n", result.Message)
	return 0
}

func handleUpdateStatusCLI(args []string, tools *Tools) int {
	fs := flag.NewFlagSet("i-update-status", flag.ExitOnError)
	agentID := fs.String("agent-id", "", "Agent ID")
	status := fs.String("status", "", "Status (active/working/idle/completed/failed)")
	currentTask := fs.String("current-task", "", "Current task description (optional)")
	fs.Parse(args)

	if *agentID == "" || *status == "" {
		fmt.Fprintln(os.Stderr, "Error: --agent-id and --status are required")
		return 1
	}

	result, err := tools.UpdateStatus(UpdateStatusArgs{
		AgentID:     *agentID,
		Status:      *status,
		CurrentTask: currentTask,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Printf("✓ %s\n", result.Message)
	return 0
}

func handleActionCLI(args []string, tools *Tools) int {
	fs := flag.NewFlagSet("i-action", flag.ExitOnError)
	agentID := fs.String("agent-id", "", "Agent ID")
	actionType := fs.String("action-type", "", "Action type")
	description := fs.String("description", "", "Action description")
	details := fs.String("details", "", "Additional details (optional)")
	fs.Parse(args)

	if *agentID == "" || *actionType == "" || *description == "" {
		fmt.Fprintln(os.Stderr, "Error: --agent-id, --action-type, and --description are required")
		return 1
	}

	result, err := tools.LogAction(LogActionArgs{
		AgentID:     *agentID,
		ActionType:  *actionType,
		Description: *description,
		Details:     details,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Printf("✓ %s\n", result.Message)
	return 0
}

func handleContextSetCLI(args []string, tools *Tools) int {
	fs := flag.NewFlagSet("i-context-set", flag.ExitOnError)
	sessionID := fs.String("session-id", "", "Session ID")
	key := fs.String("key", "", "Context key")
	value := fs.String("value", "", "Context value")
	fs.Parse(args)

	if *sessionID == "" || *key == "" || *value == "" {
		fmt.Fprintln(os.Stderr, "Error: --session-id, --key, and --value are required")
		return 1
	}

	result, err := tools.SetContext(SetContextArgs{
		SessionID: *sessionID,
		Key:       *key,
		Value:     *value,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Printf("✓ %s\n", result.Message)
	return 0
}

func handleUpdateDescriptionCLI(args []string, tools *Tools) int {
	fs := flag.NewFlagSet("i-update-description", flag.ExitOnError)
	agentID := fs.String("agent-id", "", "Agent ID")
	description := fs.String("description", "", "New description")
	fs.Parse(args)

	if *agentID == "" || *description == "" {
		fmt.Fprintln(os.Stderr, "Error: --agent-id and --description are required")
		return 1
	}

	result, err := tools.UpdateFeatureDescription(UpdateFeatureDescriptionArgs{
		AgentID:            *agentID,
		FeatureDescription: *description,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Printf("✓ %s\n", result.Message)
	return 0
}

func handleGetIDsCLI(args []string, tools *Tools) int {
	// Read SESSION env var
	sessionID := os.Getenv("SESSION")
	if sessionID == "" {
		fmt.Fprintln(os.Stderr, "Error: SESSION environment variable not set")
		return 1
	}

	// Query database for agent with this session_id
	query := `SELECT id FROM agents WHERE session_id = ? LIMIT 1`
	var agentID string
	err := tools.db.QueryRow(query, sessionID).Scan(&agentID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: No agent found for session %s: %v\n", sessionID, err)
		return 1
	}

	// Output in lowercase
	fmt.Printf("session_id=%s\n", strings.ToLower(sessionID))
	fmt.Printf("agent_id=%s\n", strings.ToLower(agentID))
	return 0
}
