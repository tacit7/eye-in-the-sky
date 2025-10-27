package overview

import (
	"fmt"
	"strings"

	"github.com/tacit7/eye-in-the-sky/internal/ui/app/views/overview/tabs"
)

// View renders the OverviewView
func (m Model) View() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n\n")

	// Tabs
	b.WriteString(m.tabs.Render())
	b.WriteString("\n\n")

	// Content based on active tab
	b.WriteString(m.renderTabContent())
	b.WriteString("\n")

	// Footer
	b.WriteString(m.renderFooter())

	return b.String()
}

// renderHeader renders the view header
func (m *Model) renderHeader() string {
	return fmt.Sprintf("Eye in the Sky - Agent Management Dashboard")
}

// renderFooter renders the view footer with status
func (m *Model) renderFooter() string {
	status := m.statusMsg
	if status == "" {
		status = "Ready"
	}
	return fmt.Sprintf("Status: %s | Tab: %s", status, m.tabs.Titles[m.tabs.ActiveIndex])
}

// renderTabContent renders content based on active tab
func (m *Model) renderTabContent() string {
	switch m.tabs.ActiveIndex {
	case 0:
		return m.renderAgentsTab()
	case 1:
		return m.renderProjectTab()
	case 2:
		return m.renderClaudeTab()
	case 3:
		return m.renderUsageTab()
	case 4:
		return m.renderConfigTab()
	default:
		return "Unknown tab"
	}
}

// renderAgentsTab renders the agents list
func (m *Model) renderAgentsTab() string {
	agents := m.GetVisibleAgents()

	// Calculate layout (using old viewport calculation)
	layout := tabs.LayoutInfo{
		Width:    m.width,
		Height:   m.height,
		ContentH: m.height - 8, // Reserve space for header (3 lines) + title bar + footer + borders
	}

	// Pass styles directly - it implements components.Styles interface
	return tabs.RenderAgentsTab(agents, m.selectedIndex, m.listOffset, m.styles, layout)
}

// renderProjectTab renders the project tab
func (m *Model) renderProjectTab() string {
	if m.viewports == nil {
		return m.styles.RenderSubtle("Loading project data...")
	}
	viewport := m.viewports.GetProjectViewport()
	return tabs.RenderProjectTab(viewport, m.styles)
}

// renderClaudeTab renders the Claude config tab
func (m *Model) renderClaudeTab() string {
	if m.viewports == nil {
		return m.styles.RenderSubtle("Loading Claude config...")
	}
	viewport := m.viewports.GetClaudeViewport()
	return tabs.RenderClaudeTab(viewport, m.styles)
}

// renderUsageTab renders the token usage tab
func (m *Model) renderUsageTab() string {
	if m.viewports == nil {
		return m.styles.RenderSubtle("Loading usage data...")
	}
	viewport := m.viewports.GetUsageViewport()
	return tabs.RenderUsageTab(viewport, m.styles)
}

// renderConfigTab renders the keybindings config tab
func (m *Model) renderConfigTab() string {
	if m.viewports == nil {
		return m.styles.RenderSubtle("Loading keybindings...")
	}
	viewport := m.viewports.GetKeybindingsViewport()
	return tabs.RenderConfigTab(viewport, m.styles)
}
