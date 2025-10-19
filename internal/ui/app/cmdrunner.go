package app

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

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

		return cmdResult{success: true, message: fmt.Sprintf("Resumed session %s", sessionID[:8])}
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

		return cmdResult{success: true, message: fmt.Sprintf("Started session %s", sessionID[:8])}
	}
}

// NewSession spawns claude --session-id with a new UUID and agent-id
func NewSession(claudePath string) tea.Cmd {
	return func() tea.Msg {
		// Generate session ID and agent ID
		sessionID := uuid.New().String()
		agentID := uuid.New().String()

		// Run: claude --session-id "$CCSESSION" "agent-id: $CCAGENT"
		cmd := exec.Command(claudePath, "--session-id", sessionID, fmt.Sprintf("agent-id: %s", agentID))
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil

		if err := cmd.Start(); err != nil {
			return cmdResult{success: false, message: "Failed to create new session", err: err}
		}

		// Detach - don't wait
		go cmd.Wait()

		return cmdResult{
			success: true,
			message: fmt.Sprintf("Created session %s with agent %s", sessionID[:8], agentID[:8]),
		}
	}
}
