package overview

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/app/messages"
)

// mockDataClient is a mock implementation of DataClient for testing
type mockDataClient struct {
	agents []domain.Agent
	err    error
}

func (m *mockDataClient) LoadAgents() ([]domain.Agent, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.agents, nil
}

func TestOverviewModel_Init(t *testing.T) {
	// Setup
	mockAgents := []domain.Agent{
		{
			ID:          "test-agent-1",
			Status:      "working",
			FeatureDesc: "Test feature 1",
		},
		{
			ID:          "test-agent-2",
			Status:      "active",
			FeatureDesc: "Test feature 2",
		},
	}

	dataClient := &mockDataClient{agents: mockAgents}
	model := New(dataClient, nil)

	// Execute Init()
	cmd := model.Init()

	// Verify command is returned
	if cmd == nil {
		t.Fatal("Init() should return a command to load agents")
	}

	// Execute the command to get the message
	msg := cmd()

	// Verify we get AgentsLoadedMsg
	loadedMsg, ok := msg.(messages.AgentsLoadedMsg)
	if !ok {
		t.Fatalf("Expected messages.AgentsLoadedMsg, got %T", msg)
	}

	// Verify agents are loaded
	if len(loadedMsg.Agents) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(loadedMsg.Agents))
	}

	if loadedMsg.Agents[0].ID != "test-agent-1" {
		t.Errorf("Expected first agent ID 'test-agent-1', got %s", loadedMsg.Agents[0].ID)
	}
}

func TestOverviewModel_Update_AgentsLoaded(t *testing.T) {
	// Setup
	dataClient := &mockDataClient{}
	model := New(dataClient, nil)

	agents := []domain.Agent{
		{
			ID:          "agent-1",
			Status:      "working",
			FeatureDesc: "Working on feature",
		},
	}

	// Send AgentsLoadedMsg
	msg := messages.AgentsLoadedMsg{Agents: agents}
	updatedModel, cmd := model.Update(msg)

	// Verify model was updated
	m := updatedModel.(Model)
	if len(m.agents) != 1 {
		t.Errorf("Expected 1 agent in model, got %d", len(m.agents))
	}

	if m.agents[0].ID != "agent-1" {
		t.Errorf("Expected agent ID 'agent-1', got %s", m.agents[0].ID)
	}

	if m.statusMsg != "Agents loaded" {
		t.Errorf("Expected status message 'Agents loaded', got '%s'", m.statusMsg)
	}

	if cmd != nil {
		t.Error("AgentsLoadedMsg should not return a command")
	}
}

func TestOverviewModel_Update_Error(t *testing.T) {
	// Setup
	dataClient := &mockDataClient{}
	model := New(dataClient, nil)

	// Send ErrMsg
	testErr := errors.New("test error")
	msg := ErrMsg{Error: testErr}
	updatedModel, cmd := model.Update(msg)

	// Verify error was handled
	m := updatedModel.(Model)
	if m.statusMsg != "Error: test error" {
		t.Errorf("Expected status message 'Error: test error', got '%s'", m.statusMsg)
	}

	if cmd != nil {
		t.Error("ErrMsg should not return a command")
	}
}

func TestOverviewModel_Update_WindowSize(t *testing.T) {
	// Setup
	dataClient := &mockDataClient{}
	model := New(dataClient, nil)

	// Send WindowSizeMsg
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(msg)

	// Verify dimensions were updated
	m := updatedModel.(Model)
	if m.width != 100 {
		t.Errorf("Expected width 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("Expected height 50, got %d", m.height)
	}
}

func TestOverviewModel_GetVisibleAgents_ShowAll(t *testing.T) {
	// Setup
	dataClient := &mockDataClient{}
	model := New(dataClient, nil)
	model.showAll = true
	model.agents = []domain.Agent{
		{ID: "agent-1", Status: "working"},
		{ID: "agent-2", Status: "completed"},
		{ID: "agent-3", Status: "active"},
	}

	// Get visible agents
	visible := model.GetVisibleAgents()

	// Should show all agents
	if len(visible) != 3 {
		t.Errorf("Expected 3 visible agents (showAll=true), got %d", len(visible))
	}
}

func TestOverviewModel_GetVisibleAgents_FilterActive(t *testing.T) {
	// Setup
	dataClient := &mockDataClient{}
	model := New(dataClient, nil)
	model.showAll = false // Filter to active/working only
	model.agents = []domain.Agent{
		{ID: "agent-1", Status: "working"},
		{ID: "agent-2", Status: "completed"},
		{ID: "agent-3", Status: "active"},
		{ID: "agent-4", Status: "failed"},
	}

	// Get visible agents
	visible := model.GetVisibleAgents()

	// Should only show active and working agents
	if len(visible) != 2 {
		t.Errorf("Expected 2 visible agents (active/working only), got %d", len(visible))
	}

	// Verify correct agents are shown
	for _, agent := range visible {
		if agent.Status != "active" && agent.Status != "working" {
			t.Errorf("Unexpected agent status in filtered list: %s", agent.Status)
		}
	}
}

func TestOverviewModel_SelectedAgent(t *testing.T) {
	// Setup
	dataClient := &mockDataClient{}
	model := New(dataClient, nil)
	model.agents = []domain.Agent{
		{ID: "agent-1", Status: "working"},
		{ID: "agent-2", Status: "active"},
	}

	// Test valid selection
	model.selectedIndex = 1
	selected := model.SelectedAgent()
	if selected == nil {
		t.Fatal("Expected selected agent, got nil")
	}
	if selected.ID != "agent-2" {
		t.Errorf("Expected selected agent 'agent-2', got %s", selected.ID)
	}

	// Test out of bounds selection
	model.selectedIndex = 10
	selected = model.SelectedAgent()
	if selected != nil {
		t.Error("Expected nil for out of bounds selection")
	}

	// Test negative selection
	model.selectedIndex = -1
	selected = model.SelectedAgent()
	if selected != nil {
		t.Error("Expected nil for negative selection")
	}
}

func TestOverviewModel_Navigation(t *testing.T) {
	// Setup
	dataClient := &mockDataClient{}
	model := New(dataClient, nil)
	model.agents = []domain.Agent{
		{ID: "agent-1", Status: "working"},
		{ID: "agent-2", Status: "active"},
		{ID: "agent-3", Status: "idle"},
	}
	model.selectedIndex = 0
	model.height = 20 // Set height for scroll calculations
	model.tabs.Set(0) // Ensure we're on the agents tab (index 0)

	// Test down navigation (use arrow key to avoid 'j'/'k' tab switching conflict)
	msg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)
	if m.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1 after down arrow, got %d", m.selectedIndex)
	}

	// Test up navigation
	msg = tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0 after up arrow, got %d", m.selectedIndex)
	}

	// Test up at top boundary (should stay at 0)
	msg = tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to stay at 0 at top boundary, got %d", m.selectedIndex)
	}

	// Test down to bottom
	m.selectedIndex = 2
	msg = tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)
	if m.selectedIndex != 2 {
		t.Errorf("Expected selectedIndex to stay at 2 at bottom boundary, got %d", m.selectedIndex)
	}
}

func TestOverviewModel_SelectAgent(t *testing.T) {
	// Setup
	dataClient := &mockDataClient{}
	model := New(dataClient, nil)
	model.agents = []domain.Agent{
		{ID: "agent-1", Status: "working", FeatureDesc: "Test"},
	}
	model.selectedIndex = 0

	// Press enter to select
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := model.Update(msg)

	// Should return a command that produces SelectAgentMsg
	if cmd == nil {
		t.Fatal("Expected command to be returned for agent selection")
	}

	cmdMsg := cmd()
	selectMsg, ok := cmdMsg.(SelectAgentMsg)
	if !ok {
		t.Fatalf("Expected SelectAgentMsg, got %T", cmdMsg)
	}

	if selectMsg.Agent.ID != "agent-1" {
		t.Errorf("Expected selected agent 'agent-1', got %s", selectMsg.Agent.ID)
	}
}

func TestOverviewModel_ToggleShowAll(t *testing.T) {
	// Setup
	dataClient := &mockDataClient{}
	model := New(dataClient, nil)
	model.showAll = false

	// Press 'a' to toggle
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	updatedModel, cmd := model.Update(msg)

	// Should toggle showAll and reload agents
	m := updatedModel.(Model)
	if !m.showAll {
		t.Error("Expected showAll to be true after toggle")
	}

	if cmd == nil {
		t.Error("Expected command to reload agents after toggle")
	}
}

func TestOverviewModel_LoadAgentsError(t *testing.T) {
	// Setup with error
	testErr := errors.New("database connection failed")
	dataClient := &mockDataClient{err: testErr}
	model := New(dataClient, nil)

	// Execute Init()
	cmd := model.Init()
	msg := cmd()

	// Should return ErrMsg
	errMsg, ok := msg.(ErrMsg)
	if !ok {
		t.Fatalf("Expected ErrMsg for load error, got %T", msg)
	}

	if errMsg.Error != testErr {
		t.Errorf("Expected error '%v', got '%v'", testErr, errMsg.Error)
	}
}

func TestOverviewModel_DataClientAdapter(t *testing.T) {
	// Test that the adapter correctly wraps AgentStore
	mockStore := &mockAgentStore{
		agents: []domain.Agent{
			{ID: "test-1", Status: "working"},
		},
	}

	adapter := NewDataClientAdapter(mockStore)
	agents, err := adapter.LoadAgents()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(agents))
	}

	if agents[0].ID != "test-1" {
		t.Errorf("Expected agent ID 'test-1', got %s", agents[0].ID)
	}
}

// mockAgentStore implements the AgentStore interface for adapter testing
type mockAgentStore struct {
	agents []domain.Agent
	err    error
}

func (m *mockAgentStore) LoadAgents(ctx context.Context) ([]domain.Agent, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.agents, nil
}
