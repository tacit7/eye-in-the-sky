package app

import (
	"strings"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// renderDetail renders the agent detail view
func (m *Model) renderDetail() string {
	if m.selectedAgent == nil {
		return "No agent selected"
	}

	var view strings.Builder

	// Header
	view.WriteString(m.renderHeader())
	view.WriteString("\n\n")

	// Tabs (with back arrow)
	view.WriteString(m.tabs.Render())
	view.WriteString("\n")

	// Agent info header
	view.WriteString(NewAgentHeader(m.selectedAgent, m.styles, m.width).View())
	view.WriteString("\n")

	// Content based on active tab
	var content string
	if m.tabs.ActiveIndex == 0 {
		// Back arrow - shouldn't render content
		content = m.styles.Subtle.Render("Press Enter or Esc to go back")
	} else {
		// Use registry to get renderer
		if renderer, ok := DetailTabRenderers[m.tabs.ActiveIndex-1]; ok {
			content = renderer(m)
		} else {
			content = "Unknown tab"
		}
	}

	// Calculate content height
	contentHeight := m.height - 10 // Account for header, tabs, agent info, footer
	if contentHeight < 5 {
		contentHeight = 5
	}

	// Create scrollable content box
	contentBox := m.styles.ContentBox.
		Width(m.width - 4).
		Height(contentHeight).
		Render(content)
	view.WriteString(contentBox)
	view.WriteString("\n")

	// Footer
	view.WriteString(m.renderFooter())

	// Use unified error overlay if needed
	return m.renderErrorOverlay(view.String())
}

// NewAgentHeader creates a new AgentHeader for rendering
func NewAgentHeader(agent *domain.Agent, styles Styles, width int) AgentHeader {
	return AgentHeader{
		Agent:  *agent,
		Styles: styles,
		Width:  width,
	}
}
