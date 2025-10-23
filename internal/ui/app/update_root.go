package app

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/tacit7/eye-in-the-sky/internal/ui/modal"
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
	// Modal gate: if modal is active, route all messages through modal
	if m.modalManager.IsActive() {
		switch msg := msg.(type) {
		case tea.KeyMsg, tea.WindowSizeMsg:
			cmd := m.modalManager.Update(msg)
			return m, cmd
		default:
			// Other messages pass through normally
			return m, nil
		}
	}

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

    // Update Claude viewport dimensions dynamically
    if m.currentView == ViewList && m.listTabs.ActiveIndex == 2 {
        const headerHeight = 2  // Header + tabs
        const footerHeight = 1  // Footer hints
        const leftPaneWidth = 30  // Approximate file list width

        available := msg.Height - headerHeight - footerHeight - 1
        if available < 10 {
            available = 10
        }

        rightPaneWidth := msg.Width - leftPaneWidth - 8  // Account for borders
        if rightPaneWidth < 40 {
            rightPaneWidth = 40
        }

        m.claudeViewport.Width = rightPaneWidth
        m.claudeViewport.Height = available
    }

    // Update Config tab viewport dimensions dynamically
    if m.currentView == ViewList && m.listTabs.ActiveIndex == 4 {
        const headerHeight = 2  // Header + tabs
        const footerHeight = 1  // Footer hints

        available := msg.Height - headerHeight - footerHeight - 1  // -1 for breathing room
        if available < 10 {
            available = 10
        }

        m.keybindingsViewport.Width = msg.Width - 4
        m.keybindingsViewport.Height = available
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

	// Claude tab messages
	case claudeFilesLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.statusMsg = "Failed to load Claude config files"
			return m, nil
		}
		m.claudeFiles = msg.files
		// Clamp selection to valid range after refresh
		if m.claudeSelectedIndex >= len(m.claudeFiles) && len(m.claudeFiles) > 0 {
			m.claudeSelectedIndex = len(m.claudeFiles) - 1
		}
		m.statusMsg = fmt.Sprintf("Loaded %d Claude config files", len(msg.files))
		return m, nil

	case claudeFileContentLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.statusMsg = "Failed to load file content"
			return m, nil
		}
		m.claudeContent = msg.content
		m.claudeViewport.SetContent(msg.content)
		m.claudeViewport.GotoTop()
		m.claudeValidStatus = "" // Clear validation status on new file load
		m.statusMsg = fmt.Sprintf("Loaded %s", msg.path)
		return m, nil

	case claudeEditorClosedMsg:
		if msg.err != nil {
			m.statusMsg = "Editor closed with error"
		} else {
			m.statusMsg = "File edited"
			// Reload the file content after editing
			if m.claudeSelectedIndex >= 0 && m.claudeSelectedIndex < len(m.claudeFiles) {
				selectedFile := m.claudeFiles[m.claudeSelectedIndex]
				return m, loadClaudeFileContentCmd(selectedFile.Path)
			}
		}
		return m, nil

	case claudeValidationResultMsg:
		m.claudeValidStatus = msg.status
		if msg.valid {
			m.statusMsg = "JSON is valid"
		} else {
			m.statusMsg = "JSON validation failed"
		}
		return m, nil

	case keybindingsSavedMsg:
		// Handle keybindings save result
		if msg.err != nil {
			m.err = msg.err
			m.statusMsg = fmt.Sprintf("Failed to save keybindings: %v", msg.err)
		} else {
			m.keybindingsYAML = m.keybindingsEditBuf
			m.keybindingsViewport.SetContent(m.keybindingsEditBuf)
			m.keybindingsEditing = false
			m.keybindingsEditBuf = ""
			m.keybindingsModified = false
			m.statusMsg = "Keybindings saved successfully"
		}
		return m, nil

	case keybindingsReloadedMsg:
		// Handle keybindings reload result
		if msg.err != nil {
			m.err = msg.err
			m.statusMsg = fmt.Sprintf("Failed to reload keybindings: %v", msg.err)
		} else {
			m.keybindingsYAML = msg.content
			m.keybindingsViewport.SetContent(msg.content)
			m.keybindingsViewport.GotoTop()
			m.keybindingsEditBuf = ""
			m.keybindingsModified = false
			m.statusMsg = "Keybindings reloaded from file"
		}
		return m, nil

	case modal.FormSubmitted:
		// Handle form submission from modal
		return m.handleFormSubmission(msg)
	}

	return m, nil
}

// handleFormSubmission handles form submissions from modals
func (m *Model) handleFormSubmission(msg modal.FormSubmitted) (tea.Model, tea.Cmd) {
	switch msg.Source {
	case "new-session":
		// Create new Claude Code session with description
		description := msg.Data["description"]
		if description == "" {
			m.statusMsg = "Session description required"
			return m, nil
		}
		return m, m.createNewSessionCmd(description)

	case "new-ticket":
		// Create new Taskwarrior ticket
		description := msg.Data["description"]
		project := msg.Data["project"]
		tags := msg.Data["tags"]
		if description == "" {
			m.statusMsg = "Ticket description required"
			return m, nil
		}
		return m, m.createNewTicketCmd(description, project, tags)

	default:
		m.statusMsg = fmt.Sprintf("Unknown form source: %s", msg.Source)
		return m, nil
	}
}

// createNewSessionCmd creates a command to spawn a new Claude Code session
func (m *Model) createNewSessionCmd(description string) tea.Cmd {
	return func() tea.Msg {
		// Generate new session ID
		sessionID := uuid.New().String()

		// Run AppleScript to create session in new Claude Code window
		script := fmt.Sprintf(`tell application "Terminal" to do script "%s --session-id %s '%s'"`,
			m.claudePath, sessionID, description)

		cmd := exec.Command("osascript", "-e", script)
		if err := cmd.Run(); err != nil {
			return errMsg{err: fmt.Errorf("failed to create session: %w", err)}
		}

		m.statusMsg = fmt.Sprintf("New session created: %s", truncateID(sessionID, 8))
		return cmdResult{
			success: true,
			message: fmt.Sprintf("New session created: %s", description),
		}
	}
}

// createNewTicketCmd creates a command to create a new Taskwarrior ticket
func (m *Model) createNewTicketCmd(description string, project string, tags string) tea.Cmd {
	return func() tea.Msg {
		// Build task command arguments
		args := []string{"add", description}

		if project != "" {
			args = append(args, fmt.Sprintf("project:%s", project))
		}

		if tags != "" {
			args = append(args, tags)
		}

		// Execute task command
		cmd := exec.Command("task", args...)
		if err := cmd.Run(); err != nil {
			return errMsg{err: fmt.Errorf("failed to create ticket: %w", err)}
		}

		m.statusMsg = fmt.Sprintf("Ticket created: %s", description)
		return cmdResult{
			success: true,
			message: fmt.Sprintf("Ticket created: %s", description),
		}
	}
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
		m.openContextualHelp()
		return m, nil
	}

	return m, nil
}

// openContextualHelp opens the help modal with keybindings for the current context
func (m *Model) openContextualHelp() {
	if m.keybindResolver == nil {
		return
	}

	// Determine the current scope based on view and tab
	var scope string

	if m.currentView == ViewList {
		// Determine which list tab is active
		if m.listTabs.ActiveIndex < len(m.listTabs.Titles) {
			tabName := m.listTabs.Titles[m.listTabs.ActiveIndex]
			// Map tab names to scope names
			switch {
			case tabName == "[O]verview" || tabName == "Overview":
				scope = "list_overview"
			case tabName == "[P]roject" || tabName == "Project":
				scope = "list_project"
			case tabName == "[C]laude" || tabName == "Claude":
				scope = "list_claude"
			case tabName == "[T]oken Usage" || tabName == "Token Usage":
				scope = "list_usage"
			default:
				scope = "list_overview"
			}
		}
	} else if m.currentView == ViewDetail {
		// Determine which detail tab is active
		if m.tabs.ActiveIndex < len(m.tabs.Titles) {
			tabName := m.tabs.Titles[m.tabs.ActiveIndex]
			// Map tab names to scope names
			switch {
			case tabName == "[A]gent View" || tabName == "Agent View":
				scope = "detail_overview"
			case tabName == "[C]ommits" || tabName == "Commits":
				scope = "detail_commits"
			case tabName == "[L]ogs" || tabName == "Logs":
				scope = "detail_logs"
			case tabName == "[N]otes" || tabName == "Notes":
				scope = "detail_notes"
			case tabName == "[A]ctions" || tabName == "Actions":
				scope = "detail_actions"
			case tabName == "[T]asks" || tabName == "Tasks":
				scope = "detail_tasks"
			default:
				scope = "detail_overview"
			}
		}
	} else {
		scope = "global"
	}

	// Get keybindings for the scope and open help modal
	bindings := m.keybindResolver.GetScopeKeybindings(scope)

	// Also include global keybindings
	globalBindings := m.keybindResolver.GetScopeKeybindings("global")
	for action, keys := range globalBindings {
		if _, exists := bindings[action]; !exists {
			bindings[action] = keys
		}
	}

	m.modalManager.OpenHelpForScope(scope, bindings)
}
