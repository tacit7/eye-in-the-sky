package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// renderListView renders the agent list view using the new component system
func (m *Model) renderListView() string {
	// Calculate layout dynamically
	layout := LayoutForModel(m)

	// Build the view components
	var view strings.Builder

	// Header
	view.WriteString(m.renderHeader())
	view.WriteString("\n\n")

	// Tabs
	tabsDisplay := m.listTabs.Render()
	view.WriteString(tabsDisplay)
	view.WriteString("\n")

	// Render content based on active tab
	content := m.renderTabContent(layout)

	// Use Lipgloss for proper content placement
	contentBox := m.styles.ContentBox.
		Width(m.width - 4).
		Height(layout.ContentH).
		Render(content)
	view.WriteString(contentBox)
	view.WriteString("\n")

	// Footer
	view.WriteString(m.renderFooter())

	// Use unified error overlay if needed
	return m.renderErrorOverlay(view.String())
}

// renderTabContent renders the content for the active tab
func (m *Model) renderTabContent(layout Layout) string {
	switch m.listTabs.ActiveIndex {
	case 1: // Project tab
		return m.renderProjectTab()
	case 2: // Claude tab
		return m.renderClaudeTab()
	case 3: // Usage tab
		return m.renderUsageTab()
	case 4: // Keybindings/Config tab
		return m.renderConfigTab()
	default: // Overview tab
		return m.renderAgentTableContent(layout)
	}
}

// renderAgentTableContent renders the agent table using new components
func (m *Model) renderAgentTableContent(layout Layout) string {
	// Prepare agent data
	agents := GetVisibleAgents(m)

	if len(agents) == 0 {
		return m.styles.Subtle.Render("  No agents found")
	}

	// Create viewport for future virtualization
	viewport := NewViewport(layout.ContentH, len(agents))
	viewport.Offset = m.listOffset

	// Get visible range
	start, end := viewport.VisibleRange()
	visibleAgents := agents[start:end]

	// Configure table
	config := TableConfig{
		ShowIcons:   true,
		ShowStatus:  true,
		ShowID:      true,
		ShowTask:    true,
		ShowSource:  true,
		ShowSession: true,
	}

	return m.renderAgentTable(visibleAgents, config)
}

// TableConfig defines which columns to show in the agent table
type TableConfig struct {
	ShowIcons   bool
	ShowStatus  bool
	ShowID      bool
	ShowTask    bool
	ShowSource  bool
	ShowSession bool
}

// renderAgentTable renders agents as a table using the new table system
func (m *Model) renderAgentTable(agents []domain.Agent, config TableConfig) string {
	// Create agent renderer
	renderer := NewAgentLineRenderer(&m.styles)

	// Configure columns
	renderer.ConfigureColumns(ColumnConfig{
		Icon:    config.ShowIcons,
		Status:  config.ShowStatus,
		ID:      config.ShowID,
		Task:    config.ShowTask,
		Source:  config.ShowSource,
		Session: config.ShowSession,
	})

	// Create table builder
	tb := renderer.CreateTableBuilder(&m.styles)

	// Set row styling for selection
	tb.SetRowStyleFunc(func(rowIndex int) lipgloss.Style {
		actualIndex := m.listOffset + rowIndex
		if actualIndex == m.selectedIndex {
			return m.styles.Selected
		}
		// Alternating row colors for better readability
		if rowIndex%2 == 1 {
			return lipgloss.NewStyle().Background(lipgloss.Color("233"))
		}
		return lipgloss.NewStyle()
	})

	// Convert agents to table rows
	var rows [][]string
	for _, agent := range agents {
		row := renderer.RenderAgent(agent, false)
		rows = append(rows, row)
	}

	// Render the table
	result, _ := tb.RenderTable(rows)
	return result
}

// renderAgentList is deprecated - use renderAgentTable instead
func (m *Model) renderAgentList() string {
	layout := LayoutForModel(m)
	return m.renderAgentTableContent(layout)
}

// renderConfigTab renders the keybindings configuration tab
func (m *Model) renderConfigTab() string {
	if m.keybindingsYAML == "" {
		return m.styles.Subtle.Render("  Loading keybindings...")
	}

	var b strings.Builder

	// Title and mode indicator
	title := "⚙️  Keybindings Configuration"
	if m.keybindingsEditing {
		title += " [EDIT MODE]"
	}
	b.WriteString(m.styles.SectionTitle.Render(title))
	b.WriteString("\n")

	// Display content or edit buffer
	if m.keybindingsEditing {
		// In edit mode, show the buffer
		b.WriteString(m.styles.Subtle.Render("Editing keybindings. Press Ctrl+S to save, Esc to cancel.\n"))
		b.WriteString(m.keybindingsEditBuf)
		if m.keybindingsModified {
			b.WriteString("\n")
			b.WriteString(m.styles.Warning.Render("(modified)"))
		}
	} else {
		// In view mode, show with scrolling
		b.WriteString(m.styles.Subtle.Render("Press 'e' to edit, 'r' to reload, j/k to scroll.\n"))
		b.WriteString(m.keybindingsYAML)
	}

	return b.String()
}