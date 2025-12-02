package app

import (
	"fmt"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// AgentHeader encapsulates the agent info section rendered above tabs
type AgentHeader struct {
	Agent  domain.Agent
	Styles Styles
	Width  int
}

// View renders the agent information header
func (h AgentHeader) View() string {
	// Format task/feature info
	taskInfo := h.Agent.CurrentTask
	if taskInfo == "" {
		taskInfo = h.Agent.FeatureDesc
	}
	if taskInfo == "" {
		taskInfo = h.Styles.Subtle.Render("No task description")
	}

	// Build header content
	header := fmt.Sprintf(" %s %s  %s  %s",
		h.Styles.Primary.Bold(true).Render("Agent:"),
		h.Styles.Bold.Render(truncateID(string(h.Agent.ID), 8)),
		GetStatusStyle(h.Agent.Status, h.Styles).Render(fmt.Sprintf("[%s]", h.Agent.Status)),
		taskInfo,
	)

	return h.Styles.InfoBox.
		Width(h.Width - 4).
		Render(header)
}
