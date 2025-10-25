package app

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

func TestRenderAgentInfo(t *testing.T) {
	m := &Model{
		styles: defaultTestStyles(),
	}

	agent := domain.Agent{
		ID:              "test-agent-123",
		Status:          "active",
		FeatureDesc:     "Test feature",
		ProjectName:     "test-project",
		CurrentTask:     "Implement feature",
		Source:          "worktree",
		GitWorktreePath: "/path/to/worktree",
		SessionID:       "session-123",
		ParentSessionID: "parent-session-456",
		ParentAgentID:   "parent-agent-789",
		WindowID:        "window-001",
	}

	output := m.renderAgentInfo(agent)

	// Verify section header
	if !strings.Contains(output, "Agent Information") {
		t.Error("Missing 'Agent Information' header")
	}

	// Verify all fields are present
	expectedFields := []string{
		"Agent ID:",
		"Status:",
		"Description:",
		"Project:",
		"Current Task:",
		"Source:",
		"Worktree:",
		"Session ID:",
		"Parent Session:",
		"Parent Agent:",
		"Window ID:",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Missing field: %s", field)
		}
	}

	// Verify values are present
	if !strings.Contains(output, "test-agent-123") {
		t.Error("Agent ID value missing")
	}
	if !strings.Contains(output, "Test feature") {
		t.Error("Feature description missing")
	}
}

func TestRenderAgentInfoOptionalFields(t *testing.T) {
	m := &Model{
		styles: defaultTestStyles(),
	}

	agent := domain.Agent{
		ID:     "test-agent",
		Status: "active",
		Source: "worktree",
	}

	output := m.renderAgentInfo(agent)

	// Verify optional fields are not included
	if strings.Contains(output, "Description:") {
		t.Error("Optional description should not appear when empty")
	}
	if strings.Contains(output, "Project:") {
		t.Error("Optional project should not appear when empty")
	}
}

func TestRenderTiming(t *testing.T) {
	m := &Model{
		styles: defaultTestStyles(),
	}

	now := time.Now()
	agent := domain.Agent{
		CreatedAt:      now.Add(-1 * time.Hour),
		UpdatedAt:      now.Add(-10 * time.Minute),
		LastActivityAt: now.Add(-2 * time.Minute),
		Status:         "active",
	}

	output := m.renderTiming(agent)

	// Verify section header
	if !strings.Contains(output, "Timing") {
		t.Error("Missing 'Timing' header")
	}

	// Verify timing fields
	if !strings.Contains(output, "Created:") {
		t.Error("Missing Created field")
	}
	if !strings.Contains(output, "Updated:") {
		t.Error("Missing Updated field")
	}
	if !strings.Contains(output, "Last Activity:") {
		t.Error("Missing Last Activity field")
	}
	if !strings.Contains(output, "Session Duration:") {
		t.Error("Missing Session Duration field")
	}
}

func TestRenderTimingCompleted(t *testing.T) {
	m := &Model{
		styles: defaultTestStyles(),
	}

	completedTime := time.Now()
	agent := domain.Agent{
		CreatedAt:   completedTime.Add(-2 * time.Hour),
		UpdatedAt:   completedTime,
		CompletedAt: &completedTime,
		Status:      "completed",
	}

	output := m.renderTiming(agent)

	// Completed agents should show total duration, not session duration
	if !strings.Contains(output, "Total Duration:") {
		t.Error("Completed agents should show 'Total Duration'")
	}
	if strings.Contains(output, "Session Duration:") {
		t.Error("Completed agents should not show 'Session Duration'")
	}
}

func TestRenderCommitsSection(t *testing.T) {
	m := &Model{
		styles: defaultTestStyles(),
		commits: []domain.Commit{
			{
				Hash:      "abc123def456",
				Message:   "First commit",
				Timestamp: time.Now().Add(-1 * time.Hour),
			},
			{
				Hash:      "xyz789uvw012",
				Message:   "Second commit",
				Timestamp: time.Now().Add(-30 * time.Minute),
			},
		},
	}

	output := m.renderCommitsSection()

	// Verify section header
	if !strings.Contains(output, "Recent Commits") {
		t.Error("Missing 'Recent Commits' header")
	}

	// Verify commits are shown
	if !strings.Contains(output, "abc123de") {
		t.Error("First commit hash not shown")
	}
	if !strings.Contains(output, "First commit") {
		t.Error("First commit message not shown")
	}
}

func TestRenderCommitsSectionEmpty(t *testing.T) {
	m := &Model{
		styles:  defaultTestStyles(),
		commits: []domain.Commit{},
	}

	output := m.renderCommitsSection()

	if strings.TrimSpace(output) != "" {
		t.Error("Empty commits should return empty string")
	}
}

func TestRenderNotesSection(t *testing.T) {
	m := &Model{
		styles: defaultTestStyles(),
		notes: []domain.Note{
			{
				Content:   "This is a test note",
				CreatedAt: time.Now(),
			},
		},
	}

	output := m.renderNotesSection()

	// Verify section header
	if !strings.Contains(output, "Notes Summary") {
		t.Error("Missing 'Notes Summary' header")
	}

	// Verify note content
	if !strings.Contains(output, "This is a test note") {
		t.Error("Note content not shown")
	}
}

func TestRenderNotesSectionEmpty(t *testing.T) {
	m := &Model{
		styles: defaultTestStyles(),
		notes:  []domain.Note{},
	}

	output := m.renderNotesSection()

	if strings.TrimSpace(output) != "" {
		t.Error("Empty notes should return empty string")
	}
}

func TestRenderTasksSection(t *testing.T) {
	m := &Model{
		styles: defaultTestStyles(),
		tasks: []domain.Task{
			{ID: "task-1", Description: "Todo task", StateID: 1, Archived: false},
			{ID: "task-2", Description: "In progress", StateID: 2, Archived: false},
			{ID: "task-3", Description: "Done task", StateID: 3, Archived: false},
			{ID: "task-4", Description: "Archived", StateID: 1, Archived: true},
		},
	}

	output := m.renderTasksSection()

	// Verify section header
	if !strings.Contains(output, "Tasks Summary") {
		t.Error("Missing 'Tasks Summary' header")
	}

	// Verify task counts appear
	if !strings.Contains(output, "Todo:") {
		t.Error("Missing Todo count")
	}
}

func TestRenderTasksSectionEmpty(t *testing.T) {
	m := &Model{
		styles: defaultTestStyles(),
		tasks:  []domain.Task{},
	}

	output := m.renderTasksSection()

	if strings.TrimSpace(output) != "" {
		t.Error("Empty tasks should return empty string")
	}
}

func TestRenderActionsSection(t *testing.T) {
	m := &Model{
		styles: defaultTestStyles(),
		actions: []domain.Action{
			{
				Description: "Started session",
				ActionType:  "task_start",
				Timestamp:   time.Now().Add(-5 * time.Minute),
			},
			{
				Description: "Committed changes",
				ActionType:  "git_commit",
				Timestamp:   time.Now().Add(-3 * time.Minute),
			},
		},
	}

	output := m.renderActionsSection()

	// Verify section header
	if !strings.Contains(output, "Recent Actions") {
		t.Error("Missing 'Recent Actions' header")
	}

	// Verify actions are shown
	if !strings.Contains(output, "Started session") {
		t.Error("First action not shown")
	}
}

func TestRenderActionsSectionEmpty(t *testing.T) {
	m := &Model{
		styles:  defaultTestStyles(),
		actions: []domain.Action{},
	}

	output := m.renderActionsSection()

	if strings.TrimSpace(output) != "" {
		t.Error("Empty actions should return empty string")
	}
}

func TestRenderOverviewTabNoAgent(t *testing.T) {
	m := &Model{
		styles:        defaultTestStyles(),
		selectedAgent: nil,
	}

	output := m.renderOverviewTab()

	if !strings.Contains(output, "No agent selected") {
		t.Error("Should show 'No agent selected' message")
	}
}

func TestRenderDuration(t *testing.T) {
	m := &Model{
		styles: defaultTestStyles(),
	}

	// Active agent should show session duration
	activeAgent := domain.Agent{
		CreatedAt: time.Now().Add(-2 * time.Hour),
		Status:    "active",
	}

	output := m.renderDuration(activeAgent)
	if !strings.Contains(output, "Session Duration:") {
		t.Error("Active agent should show session duration")
	}

	// Completed agent should show total duration
	completedTime := time.Now()
	completedAgent := domain.Agent{
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		CompletedAt: &completedTime,
		Status:      "completed",
	}

	output = m.renderDuration(completedAgent)
	if !strings.Contains(output, "Total Duration:") {
		t.Error("Completed agent should show total duration")
	}
}

// Helper function to create default test styles
func defaultTestStyles() Styles {
	return Styles{
		SectionTitle: lipgloss.NewStyle(),
		Label:        lipgloss.NewStyle(),
		Value:        lipgloss.NewStyle(),
		Primary:      lipgloss.NewStyle(),
		Warning:      lipgloss.NewStyle(),
		Success:      lipgloss.NewStyle(),
		Error:        lipgloss.NewStyle(),
		Subtle:       lipgloss.NewStyle(),
		Border:       lipgloss.NewStyle(),
		Selected:     lipgloss.NewStyle(),
		Git:          lipgloss.NewStyle(),
		Text:         lipgloss.NewStyle(),
		Highlight:    lipgloss.NewStyle(),
	}
}
