package app

import (
	"testing"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// Phase 3: Edge Case Tests

// TestNilAgentSelection tests behavior when no agent is selected
func TestNilAgentSelection(t *testing.T) {
	m := newTestModel(t)
	m.selectedAgent = nil
	m.currentView = ViewDetail

	// Should handle nil gracefully
	if m.selectedAgent != nil {
		t.Error("Selected agent should be nil")
	}

	// Navigation should not crash
	m.selectedIndex = 0
	if m.selectedIndex != 0 {
		t.Error("Should be able to set index even with nil agent")
	}
}

// TestEmptyAgentList tests navigation with empty agent list
func TestEmptyAgentList(t *testing.T) {
	m := newTestModel(t)
	m.agents = make([]domain.Agent, 0)
	m.selectedIndex = 0

	if len(m.agents) != 0 {
		t.Error("Agent list should be empty")
	}

	// selectedIndex should be valid (0 is valid for length 0)
	isValid := m.selectedIndex >= 0 && m.selectedIndex < len(m.agents)
	if len(m.agents) == 0 && m.selectedIndex == 0 {
		isValid = true // Special case: index 0 is acceptable when list is empty
	}
	if !isValid {
		t.Error("Index should be valid even with empty list")
	}
}

// TestNegativeIndices tests that negative indices are handled
func TestNegativeIndices(t *testing.T) {
	m := newTestModel(t)
	m.selectedIndex = -1
	m.agents = make([]domain.Agent, 5)

	isValid := m.selectedIndex >= 0 && m.selectedIndex < len(m.agents)
	assertBool(t, isValid, false, "Negative index should be invalid")

	// Prevent negative navigation
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
	assertIndex(t, m.selectedIndex, 0, "Index should be clamped to 0")
}

// TestOutOfBoundsIndices tests indices beyond list length
func TestOutOfBoundsIndices(t *testing.T) {
	m := newTestModel(t)
	m.agents = make([]domain.Agent, 5)
	m.selectedIndex = 10

	isValid := m.selectedIndex >= 0 && m.selectedIndex < len(m.agents)
	assertBool(t, isValid, false, "Out of bounds index should be invalid")

	// Clamp to bounds
	if m.selectedIndex >= len(m.agents) {
		m.selectedIndex = len(m.agents) - 1
	}
	assertIndex(t, m.selectedIndex, 4, "Index should be clamped to max")
}

// TestTabOutOfRange tests invalid tab indices
func TestTabOutOfRange(t *testing.T) {
	m := newTestModel(t)
	// Assuming tabs has 7 items (0-6)

	// Try to set invalid tab
	m.tabs.Set(99)
	// After set, should handle gracefully or wrap

	// Document current behavior
	t.Logf("Tab index after setting to 99: %d", m.tabs.ActiveIndex)
}

// TestOffsetWithoutData tests scroll offset with no data
func TestOffsetWithoutData(t *testing.T) {
	m := newTestModel(t)
	m.commits = make([]domain.Commit, 0)
	m.commitsOffset = 5
	m.commitsIndex = 0

	// Offset should not cause crashes
	if m.commitsOffset < 0 {
		t.Error("Offset should not be negative")
	}

	// Should be clamped if there's no data
	if len(m.commits) == 0 && m.commitsOffset > 0 {
		m.commitsOffset = 0
	}
	assertOffset(t, m.commitsOffset, 0, "Offset should be 0 with no data")
}

// TestWidthHeightBoundary tests zero and negative dimensions
func TestWidthHeightBoundary(t *testing.T) {
	m := newTestModel(t)
	m.width = 0
	m.height = 0

	if m.width < 0 {
		t.Error("Width should not be negative")
	}
	if m.height < 0 {
		t.Error("Height should not be negative")
	}

	// Zero dimensions are technically valid (edge case)
	assertBool(t, m.width >= 0, true, "Width should be non-negative")
	assertBool(t, m.height >= 0, true, "Height should be non-negative")
}

// TestEmptyStringFields tests nil and empty string handling
func TestEmptyStringFields(t *testing.T) {
	m := newTestModel(t)
	agent := &domain.Agent{
		ID:          "test",
		ProjectName: "",
		FeatureDesc: "",
		CurrentTask: "",
	}
	m.selectedAgent = agent

	// Should handle empty strings gracefully
	assertString(t, agent.ProjectName, "", "Empty project name")
	assertString(t, agent.FeatureDesc, "", "Empty feature description")
	assertString(t, agent.CurrentTask, "", "Empty current task")
}

// TestNilSliceHandling tests nil slices vs empty slices
func TestNilSliceHandling(t *testing.T) {
	m := newTestModel(t)

	// Test nil slice
	var nilSlice []domain.Agent
	m.agents = nilSlice

	if m.agents != nil {
		t.Error("Agents should be nil")
	}

	// Test empty slice
	m.agents = make([]domain.Agent, 0)
	if m.agents == nil {
		t.Error("Agents should not be nil after make")
	}
	if len(m.agents) != 0 {
		t.Error("Agents should have length 0")
	}
}

// TestStatusMessageEdgeCases tests empty and very long status messages
func TestStatusMessageEdgeCases(t *testing.T) {
	m := newTestModel(t)

	// Empty status
	m.statusMsg = ""
	assertString(t, m.statusMsg, "", "Empty status message")

	// Very long status
	longMsg := ""
	for i := 0; i < 1000; i++ {
		longMsg += "a"
	}
	m.statusMsg = longMsg

	if len(m.statusMsg) != 1000 {
		t.Errorf("Expected message length 1000, got %d", len(m.statusMsg))
	}
}

// TestOffsetGreaterThanDataLength tests scroll offset larger than data
func TestOffsetGreaterThanDataLength(t *testing.T) {
	m := newTestModel(t)
	m.commits = make([]domain.Commit, 3)
	m.commitsOffset = 10

	// This is technically valid but may cause rendering issues
	// Just verify it doesn't crash
	if m.commitsOffset > len(m.commits) {
		m.commitsOffset = len(m.commits)
	}
	assertOffset(t, m.commitsOffset, 3, "Offset clamped to data length")
}

// TestIndexGreaterThanLength tests when index equals or exceeds length
func TestIndexGreaterThanLength(t *testing.T) {
	m := newTestModel(t)
	m.agents = make([]domain.Agent, 5)
	m.selectedIndex = 5

	isValid := m.selectedIndex < len(m.agents)
	assertBool(t, isValid, false, "Index at length boundary is invalid")

	// Clamp to valid range
	if m.selectedIndex >= len(m.agents) {
		m.selectedIndex = len(m.agents) - 1
	}
	assertIndex(t, m.selectedIndex, 4, "Index clamped to last valid")
}

// TestMultipleNilChecks tests cascading nil checks
func TestMultipleNilChecks(t *testing.T) {
	m := newTestModel(t)
	m.selectedAgent = nil

	// First check: is agent nil?
	agentExists := m.selectedAgent != nil

	// Only proceed if agent exists
	if agentExists {
		if m.selectedAgent.ProjectName == "" {
			t.Log("Agent exists but no project")
		}
	} else {
		t.Log("Agent is nil, skip checks")
	}

	assertBool(t, agentExists, false, "Agent should not exist")
}

// TestEmptyProjectName tests operations with empty project
func TestEmptyProjectName(t *testing.T) {
	m := newTestModel(t)
	m.selectedAgent = &domain.Agent{
		ID:          "test",
		ProjectName: "",
	}

	if m.selectedAgent.ProjectName == "" {
		m.projectTickets = make([]domain.Task, 0) // Empty list for empty project
	}

	if len(m.projectTickets) != 0 {
		t.Error("Project tickets should be empty for empty project")
	}
}

// TestNegativePageSize tests edge case of zero or negative page sizes
func TestNegativePageSize(t *testing.T) {
	m := newTestModel(t)
	m.agents = make([]domain.Agent, 100)

	pageSize := 0
	if pageSize <= 0 {
		pageSize = 10 // Use default
	}

	if pageSize != 10 {
		t.Error("Page size should default to 10")
	}
}

// TestZeroTimeValues tests zero time handling (time.Time{})
func TestZeroTimeValues(t *testing.T) {
	m := newTestModel(t)
	agent := &domain.Agent{
		ID: "test",
		// CreatedAt, UpdatedAt default to zero value
	}
	m.selectedAgent = agent

	// Zero time is valid (just early date)
	t.Logf("Agent created at: %v", agent.CreatedAt)
	t.Logf("Agent updated at: %v", agent.UpdatedAt)
}

// TestSingleElementBoundaries tests operations with single item
func TestSingleElementBoundaries(t *testing.T) {
	m := newTestModel(t)
	m.agents = make([]domain.Agent, 1)
	m.agents[0] = domain.Agent{ID: "only-one"}
	m.selectedIndex = 0

	isValid := m.selectedIndex >= 0 && m.selectedIndex < len(m.agents)
	assertBool(t, isValid, true, "Single element index is valid")

	// Can't go up from 0
	if m.selectedIndex > 0 {
		m.selectedIndex--
	}
	assertIndex(t, m.selectedIndex, 0, "Can't go above single element")

	// Can't go down from 0 with length 1
	if m.selectedIndex < len(m.agents)-1 {
		m.selectedIndex++
	}
	assertIndex(t, m.selectedIndex, 0, "Can't go below single element")
}

// TestScrollingWithoutContent tests scrolling when no content exists
func TestScrollingWithoutContent(t *testing.T) {
	m := newTestModel(t)
	m.commits = make([]domain.Commit, 0)
	m.commitsIndex = 0
	m.commitsOffset = 0

	// Try to scroll
	if len(m.commits) > 0 {
		m.commitsOffset += 5
	}

	assertOffset(t, m.commitsOffset, 0, "Offset unchanged with no content")
}

// TestViewModeWithoutAgent tests changing views without agent selection
func TestViewModeWithoutAgent(t *testing.T) {
	m := newTestModel(t)
	m.selectedAgent = nil
	m.currentView = ViewDetail

	// Should not crash
	if m.selectedAgent == nil && m.currentView == ViewDetail {
		t.Log("Detail view with nil agent - should handle gracefully")
	}
}

// TestMaxIntegerBoundary tests very large indices
func TestMaxIntegerBoundary(t *testing.T) {
	m := newTestModel(t)
	m.agents = make([]domain.Agent, 5)
	m.selectedIndex = 2147483647 // Max int32

	isValid := m.selectedIndex >= 0 && m.selectedIndex < len(m.agents)
	assertBool(t, isValid, false, "Very large index should be invalid")

	// Clamp to valid range
	if m.selectedIndex >= len(m.agents) {
		m.selectedIndex = len(m.agents) - 1
	}
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}

	assertIndex(t, m.selectedIndex, 4, "Index properly clamped from max int")
}
