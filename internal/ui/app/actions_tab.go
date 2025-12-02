package app

import (
	"fmt"
	"strings"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// renderActionsTabView renders the actions tab with split-pane view
func (m *Model) renderActionsTabView() string {
	if len(m.actions) == 0 {
		return m.styles.Subtle.Render("No actions recorded")
	}

	// Build item list for left pane
	items := make([]string, len(m.actions))
	for i, action := range m.actions {
		timestamp := action.Timestamp.Format("15:04:05")
		actionType := truncate(action.ActionType, 15)
		items[i] = fmt.Sprintf("%s  %-15s", timestamp, actionType)
	}

	// Build detail content for right pane
	var detailContent string
	if m.actionsIndex >= 0 && m.actionsIndex < len(m.actions) {
		detailContent = renderActionDetails(m.actions[m.actionsIndex], m.styles)
	}

	return SplitPane{
		LeftItems:     items,
		SelectedIndex: m.actionsIndex,
		RightContent:  detailContent,
		Styles:        m.styles,
		Width:         m.width,
	}.View()
}

// renderActionDetails renders the full action details
func renderActionDetails(action domain.Action, styles Styles) string {
	var b strings.Builder

	// Action type
	b.WriteString(styles.Primary.Render("Action: "))
	b.WriteString(styles.Text.Render(action.ActionType))
	b.WriteString("\n")

	// Timestamp
	b.WriteString(styles.Primary.Render("Time: "))
	b.WriteString(styles.Text.Render(action.Timestamp.Format("2006-01-02 15:04:05")))
	b.WriteString("\n\n")

	// Description
	b.WriteString(styles.Title.Render("Description"))
	b.WriteString("\n")
	b.WriteString(styles.Text.Render(action.Description))
	b.WriteString("\n\n")

	// Details (if available)
	if action.Details != "" {
		b.WriteString(styles.Title.Render("Details"))
		b.WriteString("\n")
		b.WriteString(styles.Text.Render(action.Details))
		b.WriteString("\n")
	}

	return b.String()
}
