package app

import (
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/ui/util"
)

// ViewHandler is a function type that handles key presses for a specific view
type ViewHandler func(*Model, tea.KeyMsg) (tea.Model, tea.Cmd)

// viewHandlers maps views to their key handlers
var viewHandlers = map[ViewType]ViewHandler{
	ViewList:   (*Model).handleListKeys,
	ViewDetail: (*Model).handleDetailKeys,
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.MouseMsg:
		// Only handle left clicks
		if msg.Type != tea.MouseLeft {
			return m, nil
		}

		// Route to view-specific click handler
		if m.currentView == ViewList {
			return m.handleListClick(msg)
		} else if m.currentView == ViewDetail {
			return m.handleDetailClick(msg)
		}
		return m, nil

	case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    m.help.Width = msg.Width

    // Update usage viewport dimensions dynamically
    if m.currentView == ViewList && m.listTabs.ActiveIndex == 3 {
        const headerHeight = 2  // Header + tabs
        const footerHeight = 1  // Footer hints

        available := msg.Height - headerHeight - footerHeight - 1  // -1 for breathing room
        if available < 10 {
            available = 10
        }

        m.usageViewport.Width = msg.Width - 4
        m.usageViewport.Height = available

        debugf("Viewport: %dx%d | Terminal: %dx%d",
            m.usageViewport.Width, m.usageViewport.Height,
            msg.Width, msg.Height)
    }
		return m, nil

	case tickMsg:
		// Refresh data using command
		cmds := []tea.Cmd{
			loadAgentsCmd(m.data.Agents),
			m.tickCmd(), // Schedule next tick
		}

		// Refresh CCUsage data if needed (every 30 seconds)
		if m.ccusageDB != nil && time.Since(m.lastCCUsageSync) > 30*time.Second {
			if err := m.loadCCUsageData(); err != nil {
				// Log but don't fail
				log.Printf("Warning: Failed to reload ccusage data: %v\n", err)
			} else if m.currentView == ViewList && m.listTabs.ActiveIndex == 3 {
				// Trigger usage refresh if usage tab is active
				cmds = append(cmds, func() tea.Msg { return RefreshUsageMsg{} })
			}
		}

		return m, tea.Batch(cmds...)

	case initCCUsageMsg:
		// Sync complete, data reloaded - trigger usage refresh
		return m, tea.Batch(
			m.tickCmd(),
			func() tea.Msg { return RefreshUsageMsg{} },
		)

	case RefreshUsageMsg:
		// Resize viewport first (before setting content for correct scroll bounds)
		const headerHeight = 2
		const footerHeight = 1
		available := m.height - headerHeight - footerHeight - 1
		if available < 10 {
			available = 10
		}
		m.usageViewport.Width = m.width - 4
		m.usageViewport.Height = available

		// Refresh usage tab content
		summary := m.buildUsageSummary()
		content := m.renderUsageContent(summary)
		m.usageViewport.SetContent(content)
		m.cachedUsageRender = content
		m.usageDirty = false

		debugf("RefreshUsage: Viewport=%dx%d, Content length=%d",
			m.usageViewport.Width, m.usageViewport.Height, len(content))

		return m, nil

	case cmdResult:
		// Handle command execution results
		if msg.success {
			m.statusMsg = msg.message
		} else {
			m.err = msg.err
			m.statusMsg = msg.message
		}
		// Refresh agent list after command using command pattern
		return m, loadAgentsCmd(m.data.Agents)

	case util.FocusResult:
		// Handle window focus results
		if msg.Success {
			m.statusMsg = msg.Message
		} else {
			m.err = msg.Err
			m.statusMsg = msg.Message
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case AgentsLoadedMsg:
		// Update agents from command
		m.agents = msg.Agents
		// Load task counts after agents are loaded
		return m, loadTaskCountsCmd(m.data.Tasks, m.agents)

	case TaskCountsLoadedMsg:
		// Update task counts for each agent
		for i := range m.agents {
			for _, count := range msg.Counts {
				if m.agents[i].ID == count.AgentID {
					m.agents[i].TaskCount = count.Count
					break
				}
			}
		}
		return m, nil

	case AgentDetailsLoadedMsg:
		// Update agent details from command
		m.selectedAgent = msg.Agent
		m.actions = msg.Actions
		m.commits = msg.Commits
		m.notes = msg.Notes
		// Load metrics for the agent
		if m.selectedAgent != nil {
			return m, loadMetricsCmd(m.data.Metrics, m.selectedAgent.ID)
		}
		return m, nil

	case MetricsLoadedMsg:
		// Update metrics from command
		m.sessionMetrics = msg.Metrics
		return m, nil

	case TasksLoadedMsg, TasksErrorMsg:
		// Handle async task loading messages
		return m.handleTasksMessages(msg)
	}

	return m, nil
}

// errMsg is sent when an error occurs
type errMsg struct {
	err error
}

// handleKeyPress processes keyboard input by routing to view-specific or global handlers
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Try global keys first
	if newModel, cmd := m.handleGlobalKeys(msg); cmd != nil || newModel != m {
		return newModel, cmd
	}

	// Then route to view-specific handler
	if handler, ok := viewHandlers[m.currentView]; ok {
		return handler(m, msg)
	}

	return m, nil
}

// handleGlobalKeys handles keys that work across all views
func (m *Model) handleGlobalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}

	if Matches(msg, m.keys.Refresh) {
		m.statusMsg = "Refreshing..."
		return m, loadAgentsCmd(m.data.Agents)
	}

	if Matches(msg, m.keys.HelpToggle) {
		m.showHelp = !m.showHelp
		return m, nil
	}

	return m, nil
}
