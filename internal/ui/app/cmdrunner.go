package app

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

// truncateID safely truncates an ID string to specified length
func truncateID(id string, maxLen int) string {
	if len(id) <= maxLen {
		return id
	}
	return id[:maxLen]
}

// ResolveClaudePath resolves the path to the claude binary
// Priority: config → $PATH → error
func ResolveClaudePath(cfgPath string) (string, error) {
	// If config specifies an explicit path, use it
	if cfgPath != "" {
		if _, err := os.Stat(cfgPath); err == nil {
			return cfgPath, nil
		}
		return "", fmt.Errorf("claude binary not found at configured path: %s", cfgPath)
	}

	// Try to find in $PATH
	path, err := exec.LookPath("claude")
	if err != nil {
		return "", fmt.Errorf("claude binary not found in $PATH")
	}

	return path, nil
}

// cmdResult represents the result of a command execution
type cmdResult struct {
	success bool
	err     error
	message string
}

// ResumeSession spawns claude --resume for the given session ID
func ResumeSession(claudePath, sessionID string) tea.Cmd {
	return func() tea.Msg {
		if sessionID == "" {
			return cmdResult{success: false, message: "No session ID", err: fmt.Errorf("empty session ID")}
		}

		cmd := exec.Command(claudePath, "--resume", sessionID)
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil

		if err := cmd.Start(); err != nil {
			return cmdResult{success: false, message: "Failed to resume session", err: err}
		}

		// Detach - don't wait
		go cmd.Wait()

		return cmdResult{success: true, message: fmt.Sprintf("Resumed session %s", truncateID(sessionID, 8))}
	}
}

// StartSession spawns claude -s for the given session ID
func StartSession(claudePath, sessionID string) tea.Cmd {
	return func() tea.Msg {
		if sessionID == "" {
			return cmdResult{success: false, message: "No session ID", err: fmt.Errorf("empty session ID")}
		}

		cmd := exec.Command(claudePath, "-s", sessionID)
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil

		if err := cmd.Start(); err != nil {
			return cmdResult{success: false, message: "Failed to start session", err: err}
		}

		// Detach - don't wait
		go cmd.Wait()

		return cmdResult{success: true, message: fmt.Sprintf("Started session %s", truncateID(sessionID, 8))}
	}
}

// NewSession spawns claude --session-id with a new UUID and agent-id in a new terminal window
func NewSession(claudePath, terminal string) tea.Cmd {
	return func() tea.Msg {
		// Generate session ID and agent ID
		sessionID := uuid.New().String()
		agentID := uuid.New().String()

		// Build the claude command
		claudeCmd := fmt.Sprintf(`%s --session-id %s "agent-id: %s  session-id: %s"`,
			claudePath, sessionID, agentID, sessionID)

		log.Printf("Starting new session in %s: session=%s agent=%s",
			terminal, truncateID(sessionID, 8), truncateID(agentID, 8))

		var cmd *exec.Cmd

		// Build command based on terminal type
		switch terminal {
		case "iterm":
			// iTerm2
			script := fmt.Sprintf(`tell application "iTerm2"
    set newWindow to (create window with default profile)
    tell current session of newWindow
        write text "%s"
    end tell
end tell`, claudeCmd)
			cmd = exec.Command("osascript", "-e", script)
		case "warp":
			// Warp
			cmd = exec.Command("open", "-a", "Warp", "--args", claudeCmd)
		case "kitty":
			// Kitty
			cmd = exec.Command("kitty", "--", "sh", "-c", claudeCmd)
		case "alacritty":
			// Alacritty
			cmd = exec.Command("alacritty", "-e", "sh", "-c", claudeCmd)
		default:
			// Default: macOS Terminal
			cmd = exec.Command("osascript", "-e",
				fmt.Sprintf(`tell application "Terminal" to do script "%s"`, claudeCmd))
		}

		if err := cmd.Start(); err != nil {
			log.Printf("Error starting session: %v", err)
			return cmdResult{success: false, message: "Failed to create new session", err: err}
		}

		// Detach - don't wait
		go cmd.Wait()

		log.Printf("Session started successfully")
		return cmdResult{
			success: true,
			message: fmt.Sprintf("Created session %s with agent %s", truncateID(sessionID, 8), truncateID(agentID, 8)),
		}
	}
}
