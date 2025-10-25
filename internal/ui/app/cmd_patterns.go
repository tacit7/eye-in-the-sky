package app

import (
	"context"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// Message types for Bubble Tea

// AgentsLoadedMsg is sent when agents are loaded
type AgentsLoadedMsg struct {
	Agents []domain.Agent
}

// TasksLoadedMsg is sent when tasks are loaded
type TasksLoadedMsg struct {
	Tasks []domain.Task
}

// TaskCountsLoadedMsg is sent when task counts are loaded
type TaskCountsLoadedMsg struct {
	Counts []struct {
		AgentID domain.AgentID
		Count   int
	}
}

// MetricsLoadedMsg is sent when metrics are loaded
type MetricsLoadedMsg struct {
	Metrics []domain.SessionMetric
}

// loadMetricsCmd loads metrics for a specific agent
func loadMetricsCmd(store MetricsStore, agentID domain.AgentID) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := withTimeout()
		defer cancel()

		metrics, err := store.LoadByAgent(ctx, agentID, 10)
		if err != nil {
			return ErrMsg{Error: err}
		}
		return MetricsLoadedMsg{Metrics: metrics}
	}
}

// NotesLoadedMsg is sent when notes are loaded
type NotesLoadedMsg struct {
	Notes []domain.Note
}

// CommitsLoadedMsg is sent when commits are loaded
type CommitsLoadedMsg struct {
	Commits []domain.Commit
}

// ActionsLoadedMsg is sent when actions are loaded
type ActionsLoadedMsg struct {
	Actions []domain.Action
}

// ErrMsg is sent when an error occurs
type ErrMsg struct {
	Error error
}

// withTimeout creates a context with standard timeout
func withTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// loadAgentsCmd loads all agents
func loadAgentsCmd(store AgentStore) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := withTimeout()
		defer cancel()

		agents, err := store.LoadAgents(ctx)
		if err != nil {
			return ErrMsg{Error: err}
		}
		return AgentsLoadedMsg{Agents: agents}
	}
}

// AgentDetailsLoadedMsg is sent when agent details are loaded
type AgentDetailsLoadedMsg struct {
	Agent   *domain.Agent
	Actions []domain.Action
	Commits []domain.Commit
	Notes   []domain.Note
}

// loadAgentDetailsCmd loads details for a specific agent
func loadAgentDetailsCmd(client *DataClient, agentID domain.AgentID) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := withTimeout()
		defer cancel()

		// Load agent
		agent, err := client.Agents.LoadAgent(ctx, agentID)
		if err != nil {
			return ErrMsg{Error: err}
		}

		// Load related data in parallel
		actionsCh := make(chan []domain.Action, 1)
		commitsCh := make(chan []domain.Commit, 1)
		notesCh := make(chan []domain.Note, 1)

		go func() {
			actions, _ := client.Actions.LoadByAgent(ctx, agentID, 50)
			actionsCh <- actions
		}()

		go func() {
			commits, _ := client.Commits.LoadByAgentHierarchy(ctx, agentID, 20)
			commitsCh <- commits
		}()

		go func() {
			notes, _ := client.Notes.LoadByAgent(ctx, agentID, agent.SessionID)
			notesCh <- notes
		}()

		// Collect results
		actions := <-actionsCh
		commits := <-commitsCh
		notes := <-notesCh

		return AgentDetailsLoadedMsg{
			Agent:   agent,
			Actions: actions,
			Commits: commits,
			Notes:   notes,
		}
	}
}

// loadTasksCmd loads tasks for a specific agent
func loadTasksCmd(store TaskStore, agentID domain.AgentID, limit, offset int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := withTimeout()
		defer cancel()

		cmdStart := time.Now()
		startTime := cmdStart.Format("15:04:05.000")
		log.Printf("[PERF][%s] Starting task load for agent %s", startTime, agentID)

		start := time.Now()
		tasks, err := store.LoadByAgent(ctx, agentID, limit, offset)
		elapsed := time.Since(start)
		endTime := time.Now().Format("15:04:05.000")
		log.Printf("[PERF][%s] LoadByAgent total for agent %s (limit=%d, offset=%d) took %v (loaded %d tasks)", endTime, agentID, limit, offset, elapsed, len(tasks))

		if err != nil {
			return ErrMsg{Error: err}
		}

		cmdElapsed := time.Since(cmdStart)
		cmdEndTime := time.Now().Format("15:04:05.000")
		log.Printf("[PERF][%s] Task load command complete: total %v", cmdEndTime, cmdElapsed)
		return TasksLoadedMsg{Tasks: tasks}
	}
}

// loadTaskCountsCmd loads task counts for multiple agents
func loadTaskCountsCmd(store TaskStore, agents []domain.Agent) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := withTimeout()
		defer cancel()

		counts, err := store.LoadCountsByAgent(ctx, agents)
		if err != nil {
			return ErrMsg{Error: err}
		}
		return TaskCountsLoadedMsg{Counts: counts}
	}
}

// markTaskDoneCmd marks a task as completed
func markTaskDoneCmd(store TaskStore, taskID domain.TaskID) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := withTimeout()
		defer cancel()

		err := store.MarkDone(ctx, taskID)
		if err != nil {
			return ErrMsg{Error: err}
		}
		// After marking done, reload tasks
		// This would need the agent ID in real implementation
		return nil
	}
}

// loadAllMetricsCmd loads all metrics
func loadAllMetricsCmd(store MetricsStore) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := withTimeout()
		defer cancel()

		metrics, err := store.LoadAll(ctx)
		if err != nil {
			return ErrMsg{Error: err}
		}
		return MetricsLoadedMsg{Metrics: metrics}
	}
}

// loadMetricsByAgentCmd loads metrics for a specific agent
func loadMetricsByAgentCmd(store MetricsStore, agentID domain.AgentID, limit int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := withTimeout()
		defer cancel()

		metrics, err := store.LoadByAgent(ctx, agentID, limit)
		if err != nil {
			return ErrMsg{Error: err}
		}
		return MetricsLoadedMsg{Metrics: metrics}
	}
}

// loadNotesCmd loads notes for an agent's session
func loadNotesCmd(store NotesStore, agentID domain.AgentID, sessionID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := withTimeout()
		defer cancel()

		notes, err := store.LoadByAgent(ctx, agentID, sessionID)
		if err != nil {
			return ErrMsg{Error: err}
		}
		return NotesLoadedMsg{Notes: notes}
	}
}

// createNoteCmd creates a new note
func createNoteCmd(store NotesStore, agentID domain.AgentID, sessionID string, content string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := withTimeout()
		defer cancel()

		err := store.Create(ctx, agentID, content)
		if err != nil {
			return ErrMsg{Error: err}
		}
		// After creating, reload notes
		return loadNotesCmd(store, agentID, sessionID)()
	}
}

// loadCommitsCmd loads commits for an agent
func loadCommitsCmd(store CommitsStore, agentID domain.AgentID) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := withTimeout()
		defer cancel()

		commits, err := store.LoadByAgent(ctx, agentID)
		if err != nil {
			return ErrMsg{Error: err}
		}
		return CommitsLoadedMsg{Commits: commits}
	}
}

// loadRecentCommitsCmd loads recent commits across all agents
func loadRecentCommitsCmd(store CommitsStore, limit int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := withTimeout()
		defer cancel()

		commits, err := store.LoadRecent(ctx, limit)
		if err != nil {
			return ErrMsg{Error: err}
		}
		return CommitsLoadedMsg{Commits: commits}
	}
}

// loadActionsCmd loads actions for an agent
func loadActionsCmd(store ActionsStore, agentID domain.AgentID) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := withTimeout()
		defer cancel()

		actions, err := store.LoadByAgent(ctx, agentID, 50) // Default limit of 50
		if err != nil {
			return ErrMsg{Error: err}
		}
		return ActionsLoadedMsg{Actions: actions}
	}
}

// createActionCmd creates a new action
func createActionCmd(store ActionsStore, action domain.Action) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := withTimeout()
		defer cancel()

		err := store.Create(ctx, action)
		if err != nil {
			return ErrMsg{Error: err}
		}
		// After creating, reload actions
		return loadActionsCmd(store, action.AgentID)()
	}
}