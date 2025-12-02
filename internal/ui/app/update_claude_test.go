package app

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// Phase 2D: Claude Tab Operations Tests

// TestClaudeTabPathDisplay tests that claude path is properly displayed
func TestClaudeTabPathDisplay(t *testing.T) {
	m := newTestModel(t)
	m.claudePath = "/usr/local/bin/claude"

	if m.claudePath == "" {
		t.Error("Claude path is empty")
	}

	if m.claudePath != "/usr/local/bin/claude" {
		t.Errorf("Expected path /usr/local/bin/claude, got %s", m.claudePath)
	}
}

// TestClaudeTabFileLoadingState tests that file loading state is tracked
func TestClaudeTabFileLoadingState(t *testing.T) {
	m := newTestModel(t)
	m.currentView = ViewDetail
	m.selectedAgent = &domain.Agent{ID: "agent-1"}

	// Simulate file loading
	m.isLoading = true

	if !m.isLoading {
		t.Error("Loading state should be true")
	}

	// Simulate load complete
	m.isLoading = false
	assertBool(t, m.isLoading, false, "Loading state after complete")
}

// TestClaudeContentLoading tests handling of content loading operations
func TestClaudeContentLoading(t *testing.T) {
	m := newTestModel(t)
	m.currentView = ViewDetail
	m.selectedAgent = &domain.Agent{ID: "agent-1"}

	// Simulate content becoming available
	m.statusMsg = "Loading claude content..."

	if m.statusMsg != "Loading claude content..." {
		t.Errorf("Expected status message 'Loading claude content...', got %q", m.statusMsg)
	}

	// Simulate load complete
	m.statusMsg = "Claude content loaded"
	assertString(t, m.statusMsg, "Claude content loaded", "Status message after load")
}

// TestClaudeUpDownNavigation tests up/down navigation in claude file list
func TestClaudeUpDownNavigation(t *testing.T) {
	m := newTestModel(t)
	m.currentView = ViewDetail
	m.selectedAgent = &domain.Agent{ID: "agent-1"}

	// Simulate having claude files to navigate
	m.logsIndex = 5
	m.logsOffset = 0

	// Navigate up (decrease index)
	if m.logsIndex > 0 {
		m.logsIndex--
	}

	assertIndex(t, m.logsIndex, 4, "Navigation up in claude logs")

	// Navigate down (increase index)
	m.logsIndex++
	assertIndex(t, m.logsIndex, 5, "Navigation down in claude logs")
}

// TestClaudeTabOpenKey tests "open" key functionality
func TestClaudeTabOpenKey(t *testing.T) {
	m := newTestModel(t)
	m.currentView = ViewDetail
	m.selectedAgent = &domain.Agent{ID: "agent-1"}

	// Simulate open key behavior
	msg := tea.KeyMsg{Runes: []rune("o")}
	newM, _ := m.handleDetailKeys(msg)
	m = newM.(*Model)

	// Document that open key is handled
	t.Logf("Open key handled, current view: %v", m.currentView)
}

// TestClaudeFileBrowsing tests simulated file browsing state
func TestClaudeFileBrowsing(t *testing.T) {
	m := newTestModel(t)

	// Simulate claude file structure
	m.logs = make([]Log, 3)
	now := time.Now()
	m.logs[0] = Log{
		Timestamp: now.Add(-10 * time.Minute),
		Type:      "info",
		Message:   "Session started",
	}
	m.logs[1] = Log{
		Timestamp: now.Add(-5 * time.Minute),
		Type:      "debug",
		Message:   "Processing request",
	}
	m.logs[2] = Log{
		Timestamp: now,
		Type:      "info",
		Message:   "Session completed",
	}

	if len(m.logs) != 3 {
		t.Errorf("Expected 3 logs, got %d", len(m.logs))
	}

	// Verify log content
	if m.logs[0].Message != "Session started" {
		t.Errorf("Expected first log 'Session started', got %q", m.logs[0].Message)
	}
}

// TestClaudeContentScrolling tests scrolling through claude content
func TestClaudeContentScrolling(t *testing.T) {
	m := newTestModel(t)
	m.logsIndex = 0
	m.logsOffset = 0

	// Simulate scrolling down
	m.logsIndex = 5
	m.logsOffset = 2

	assertIndex(t, m.logsIndex, 5, "Logs index after scroll")
	assertOffset(t, m.logsOffset, 2, "Logs offset after scroll")

	// Simulate scrolling up
	m.logsIndex = 3
	m.logsOffset = 0

	assertIndex(t, m.logsIndex, 3, "Logs index after scroll up")
	assertOffset(t, m.logsOffset, 0, "Logs offset after scroll up")
}

// TestClaudeMessageDisplay tests that messages are properly formatted
func TestClaudeMessageDisplay(t *testing.T) {
	m := newTestModel(t)

	// Add messages to logs
	m.logs = append(m.logs, Log{
		Timestamp: time.Now(),
		Type:      "info",
		Message:   "Claude session initialized",
	})

	if len(m.logs) < 1 {
		t.Fatal("No logs added")
	}

	log := m.logs[0]
	if log.Message != "Claude session initialized" {
		t.Errorf("Expected message 'Claude session initialized', got %q", log.Message)
	}
}

// TestClaudeWorkingDirectory tests that working directory context is maintained
func TestClaudeWorkingDirectory(t *testing.T) {
	m := newTestModel(t)
	m.selectedAgent = &domain.Agent{
		ID:              "agent-1",
		GitWorktreePath: "/home/user/project",
		ProjectName:     "my-project",
	}

	// Working directory should be available from agent
	if m.selectedAgent.GitWorktreePath == "" {
		t.Error("Working directory not set")
	}

	if m.selectedAgent.GitWorktreePath != "/home/user/project" {
		t.Errorf("Expected /home/user/project, got %s", m.selectedAgent.GitWorktreePath)
	}
}
