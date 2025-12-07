package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// ClaudeSpawnConfig holds configuration for spawning Claude processes
type ClaudeSpawnConfig struct {
	ClaudePath  string
	ProjectPath string
	Prompt      string
	Model       string
}

// ClaudeSessionInfo holds information about a spawned Claude session
type ClaudeSessionInfo struct {
	SessionID string
	PID       int
	Prompt    string
	Model     string
}

// FindClaudeBinary locates the Claude binary in standard locations
func FindClaudeBinary() (string, error) {
	// Check common installation paths for Claude CLI
	commonPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".cargo", "bin", "claude"),
		"/usr/local/bin/claude",
		"/usr/bin/claude",
		filepath.Join(os.Getenv("HOME"), ".local", "bin", "claude"),
	}

	// Also check PATH
	if claudePath, err := exec.LookPath("claude"); err == nil {
		return claudePath, nil
	}

	// Check common paths
	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("claude binary not found in standard paths or PATH")
}

// CreateSystemCommand builds an os/exec.Cmd with environment variables, arguments, and piped I/O
func CreateSystemCommand(claudePath string, args []string, projectPath string) *exec.Cmd {
	cmd := exec.Command(claudePath, args...)

	// Set up environment variables (inherit parent environment)
	cmd.Env = os.Environ()

	// Set working directory to project path
	cmd.Dir = projectPath

	// Pipe stdout and stderr
	cmd.Stdout = nil // We'll read this with a pipe
	cmd.Stderr = nil // We'll read this with a pipe

	return cmd
}

// SpawnClaudeProcess spawns a Claude Code process and handles its output
func SpawnClaudeProcess(ctx context.Context, config ClaudeSpawnConfig, natsPublish func(subject, message string) error) (*ClaudeSessionInfo, error) {
	// Create command
	args := []string{
		"-p", config.Prompt,
		"--model", config.Model,
		"--output-format", "stream-json",
		"--verbose",
		"--dangerously-skip-permissions",
	}

	cmd := CreateSystemCommand(config.ClaudePath, args, config.ProjectPath)

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to spawn Claude process: %w", err)
	}

	sessionInfo := &ClaudeSessionInfo{
		PID:    cmd.Process.Pid,
		Prompt: config.Prompt,
		Model:  config.Model,
	}

	// Mutex for thread-safe session ID setting
	sessionIDMutex := &sync.Mutex{}
	sessionIDExtracted := make(chan struct{})

	// Read stdout in a goroutine
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()

			// Try to parse as JSON
			var msg map[string]interface{}
			if err := json.Unmarshal([]byte(line), &msg); err == nil {
				// Check for init message with session_id
				if msgType, ok := msg["type"].(string); ok && msgType == "system" {
					if subtype, ok := msg["subtype"].(string); ok && subtype == "init" {
						if sessionID, ok := msg["session_id"].(string); ok {
							sessionIDMutex.Lock()
							if sessionInfo.SessionID == "" {
								sessionInfo.SessionID = sessionID
								sessionIDMutex.Unlock()
								close(sessionIDExtracted)
								// Publish session started event
								if natsPublish != nil {
									natsPublish("events.claude-spawn.session-started", fmt.Sprintf(`{
										"session_id": "%s",
										"pid": %d,
										"prompt": "%s",
										"model": "%s"
									}`, sessionID, sessionInfo.PID, config.Prompt, config.Model))
								}
								return
							}
							sessionIDMutex.Unlock()
						}
					}
				}
			}

			// Emit per-session event if we have the session ID
			sessionIDMutex.Lock()
			if sessionInfo.SessionID != "" {
				subject := fmt.Sprintf("events.claude-output:%s", sessionInfo.SessionID)
				if natsPublish != nil {
					natsPublish(subject, line)
				}
			}
			sessionIDMutex.Unlock()

			// Always emit generic event
			if natsPublish != nil {
				natsPublish("events.claude-output", line)
			}
		}
	}()

	// Read stderr in a goroutine
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()

			// Emit per-session error event if we have the session ID
			sessionIDMutex.Lock()
			if sessionInfo.SessionID != "" {
				subject := fmt.Sprintf("events.claude-error:%s", sessionInfo.SessionID)
				if natsPublish != nil {
					natsPublish(subject, line)
				}
			}
			sessionIDMutex.Unlock()

			// Always emit generic event
			if natsPublish != nil {
				natsPublish("events.claude-error", line)
			}
		}
	}()

	// Wait for process completion and emit completion event
	go func() {
		// Wait for session ID extraction or timeout
		select {
		case <-sessionIDExtracted:
		case <-time.After(5 * time.Second):
		}

		// Wait for process to complete
		if err := cmd.Wait(); err != nil {
			// Process exited with error
			sessionIDMutex.Lock()
			sessionID := sessionInfo.SessionID
			sessionIDMutex.Unlock()

			if sessionID != "" && natsPublish != nil {
				natsPublish(fmt.Sprintf("events.claude-complete:%s", sessionID), fmt.Sprintf(`{"status": "failed", "error": "%s"}`, err.Error()))
			}
			if natsPublish != nil {
				natsPublish("events.claude-complete", fmt.Sprintf(`{"status": "failed", "error": "%s"}`, err.Error()))
			}
		} else {
			// Process exited successfully
			sessionIDMutex.Lock()
			sessionID := sessionInfo.SessionID
			sessionIDMutex.Unlock()

			if sessionID != "" && natsPublish != nil {
				natsPublish(fmt.Sprintf("events.claude-complete:%s", sessionID), `{"status": "success"}`)
			}
			if natsPublish != nil {
				natsPublish("events.claude-complete", `{"status": "success"}`)
			}
		}
	}()

	return sessionInfo, nil
}
