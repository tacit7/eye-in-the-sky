package tabs

import (
	"fmt"
	"strings"
)

// RenderActions renders the actions tab content
// TODO: State migration in progress - needs actionsIndex and split-pane state
func RenderActions(ctx *DataContext, overviewStyles OverviewStyles) string {
	if len(ctx.Actions) == 0 {
		return overviewStyles.Subtle.Render("\n  No actions recorded for this agent\n")
	}

	var sb strings.Builder

	// Simple header
	sb.WriteString(overviewStyles.Primary.Render("Time      Action Type      Description"))
	sb.WriteString("\n")
	sb.WriteString(overviewStyles.Subtle.Render(strings.Repeat("─", 80)))
	sb.WriteString("\n")

	// Render actions in simple list format
	for _, action := range ctx.Actions {
		timestamp := action.Timestamp.Format("15:04:05")
		actionType := action.ActionType
		if len(actionType) > 15 {
			actionType = actionType[:12] + "..."
		}

		desc := action.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}

		line := fmt.Sprintf("%-9s %-16s %s\n", timestamp, actionType, desc)
		sb.WriteString(line)
	}

	return sb.String()
}

// TODO: Split-pane view with action details requires state migration
/*
// renderActionDetails renders the full action details
func renderActionDetails(action domain.Action, styles OverviewStyles) string {
	var b strings.Builder

	// Action type
	b.WriteString(styles.Primary.Render("Action: "))
	b.WriteString(action.ActionType)
	b.WriteString("\n")

	// Timestamp
	b.WriteString(styles.Primary.Render("Time: "))
	b.WriteString(action.Timestamp.Format("2006-01-02 15:04:05"))
	b.WriteString("\n\n")

	// Description
	b.WriteString(styles.Label.Render("Description"))
	b.WriteString("\n")
	b.WriteString(action.Description)
	b.WriteString("\n\n")

	// Details (if available)
	if action.Details != "" {
		b.WriteString(styles.Label.Render("Details"))
		b.WriteString("\n")
		b.WriteString(action.Details)
		b.WriteString("\n")
	}

	return b.String()
}
*/
