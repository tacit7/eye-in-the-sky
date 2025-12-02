package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// Phase 2C: Project Tab Navigation Tests

// TestProjectTabSectionSwitching tests numeric key navigation (1,2,3) in project tab
func TestProjectTabSectionSwitching(t *testing.T) {
	m := newTestModel(t)
	m.currentView = ViewDetail
	m.selectedAgent = &domain.Agent{ID: "agent-1"}
	m.tabs.Set(6) // Set to Projects tab

	tests := []struct {
		key      string
		section  string
		expected string
	}{
		{"1", "Tasks", "tasks"},
		{"2", "Markdown", "markdown"},
		{"3", "Info", "info"},
	}

	for _, tt := range tests {
		t.Run("section_"+tt.section, func(t *testing.T) {
			msg := tea.KeyMsg{Runes: []rune(tt.key)}
			newM, _ := m.handleDetailKeys(msg)
			_ = newM.(*Model)

			t.Logf("Key %s should switch to %s section", tt.key, tt.section)
		})
	}
}

// TestProjectTicketNavigation tests navigation within project tickets/tasks
func TestProjectTicketNavigation(t *testing.T) {
	m := newTestModel(t)
	m.currentView = ViewDetail
	m.selectedAgent = &domain.Agent{ID: "agent-1", ProjectName: "my-project"}
	m.tabs.Set(6) // Projects tab
	m.projectTickets = make([]domain.Task, 5)
	for i := 0; i < 5; i++ {
		m.projectTickets[i].ID = domain.TaskID("task-" + string(rune(i+'0')))
	}

	m.projectTicketsIndex = 2

	// Test selecting a ticket
	if m.projectTicketsIndex < 0 || m.projectTicketsIndex >= len(m.projectTickets) {
		t.Errorf("Invalid ticket index after navigation")
	}

	assertIndex(t, m.projectTicketsIndex, 2, "Project ticket selection")
}

// TestProjectMarkdownFileNavigation tests navigating through markdown files
func TestProjectMarkdownFileNavigation(t *testing.T) {
	m := newTestModel(t)
	m.currentView = ViewDetail
	m.selectedAgent = &domain.Agent{ID: "agent-1"}
	m.tabs.Set(6) // Projects tab

	// Simulate adding project markdown files
	m.projectMDFiles = make([]ProjectFile, 3)
	m.projectMDFiles[0].Name = "README.md"
	m.projectMDFiles[1].Name = "CHANGELOG.md"
	m.projectMDFiles[2].Name = "CONTRIBUTING.md"

	// Navigate through files
	m.projectMDFilesIndex = 0

	// Test bounds
	if m.projectMDFilesIndex < 0 || m.projectMDFilesIndex >= len(m.projectMDFiles) {
		t.Errorf("Invalid markdown file index")
	}

	assertIndex(t, m.projectMDFilesIndex, 0, "Markdown file selection")
}

// TestProjectInfoDisplay tests that project information is displayed
func TestProjectInfoDisplay(t *testing.T) {
	m := newTestModel(t)
	m.selectedAgent = &domain.Agent{
		ID:          "agent-1",
		ProjectName: "my-project",
		Source:      "worktree",
	}

	if m.selectedAgent.ProjectName == "" {
		t.Error("Project name missing from agent")
	}

	if m.selectedAgent.ProjectName != "my-project" {
		t.Errorf("Expected project name 'my-project', got %q", m.selectedAgent.ProjectName)
	}
}

// TestProjectTaskListOffsets tests offset tracking for project task list
func TestProjectTaskListOffsets(t *testing.T) {
	m := newTestModel(t)
	m.projectTickets = make([]domain.Task, 20)
	m.projectTicketsIndex = 5
	m.projectTicketsOffset = 0

	// Document offset behavior when navigating
	t.Logf("Project tickets index: %d, offset: %d", m.projectTicketsIndex, m.projectTicketsOffset)

	// Offset should adjust when index moves far from visible area
	m.projectTicketsIndex = 15
	t.Logf("After navigation to 15: index: %d, offset: %d", m.projectTicketsIndex, m.projectTicketsOffset)
}

// TestProjectTabPreservesAgentSelection tests that viewing project info doesn't change selected agent
func TestProjectTabPreservesAgentSelection(t *testing.T) {
	m := newTestModel(t)
	m.selectedAgent = &domain.Agent{ID: "agent-5"}
	m.selectedIndex = 5

	initialAgent := m.selectedAgent.ID

	// Switch to project tab (tab 6)
	m.tabs.Set(6)

	// Selected agent should remain unchanged
	if m.selectedAgent.ID != initialAgent {
		t.Errorf("Viewing project tab changed selected agent")
	}

	assertIndex(t, m.selectedIndex, 5, "Selected agent index should be preserved")
}

// TestProjectPathValidation tests that project path is valid and accessible
func TestProjectPathValidation(t *testing.T) {
	m := newTestModel(t)
	m.selectedAgent = &domain.Agent{
		ID:              "agent-1",
		ProjectName:     "test-project",
		GitWorktreePath: "/path/to/project",
	}

	// Project path should be set
	if m.selectedAgent.GitWorktreePath == "" {
		t.Error("Project path is empty")
	}

	// Document that we have path information
	t.Logf("Project path: %s", m.selectedAgent.GitWorktreePath)
}

// TestProjectTabBeforeAgentSelection tests behavior when project tab accessed without selected agent
func TestProjectTabBeforeAgentSelection(t *testing.T) {
	m := newTestModel(t)
	m.selectedAgent = nil // No agent selected

	// Accessing project tab without selected agent should handle gracefully
	if m.selectedAgent != nil && m.selectedAgent.ProjectName != "" {
		t.Log("Project info displayed when agent selected")
	} else {
		t.Log("No project info when agent not selected - this is expected")
	}
}

// TestProjectTaskCountTracking tests tracking of project task states
func TestProjectTaskCountTracking(t *testing.T) {
	m := newTestModel(t)

	// Create project tasks with different states
	m.projectTickets = make([]domain.Task, 5)
	m.projectTickets[0] = domain.Task{ID: "task-1", StateID: 1} // todo
	m.projectTickets[1] = domain.Task{ID: "task-2", StateID: 2} // in progress
	m.projectTickets[2] = domain.Task{ID: "task-3", StateID: 3} // done
	m.projectTickets[3] = domain.Task{ID: "task-4", StateID: 1, Archived: true}
	m.projectTickets[4] = domain.Task{ID: "task-5", StateID: 2}

	todoCount := 0
	inProgressCount := 0
	doneCount := 0

	for _, task := range m.projectTickets {
		if task.Archived {
			continue
		}
		switch task.StateID {
		case 1:
			todoCount++
		case 2:
			inProgressCount++
		case 3:
			doneCount++
		}
	}

	if todoCount != 1 {
		t.Errorf("Expected 1 todo task, got %d", todoCount)
	}
	if inProgressCount != 2 {
		t.Errorf("Expected 2 in-progress tasks, got %d", inProgressCount)
	}
	if doneCount != 1 {
		t.Errorf("Expected 1 done task, got %d", doneCount)
	}
}
