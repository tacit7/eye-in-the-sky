package tabs

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RenderSessionContext renders the session context tab content
func RenderSessionContext(ctx *DataContext, overviewStyles OverviewStyles) string {
	if len(ctx.SessionContexts) == 0 {
		return overviewStyles.Subtle.Render("\n  No session context saved for this agent\n")
	}

	var sb strings.Builder

	// Display most recent context
	context := ctx.SessionContexts[0]

	// Header
	sb.WriteString(overviewStyles.Primary.Render("Session Context"))
	sb.WriteString("\n")
	sb.WriteString(overviewStyles.Subtle.Render(strings.Repeat("─", ctx.Width-4)))
	sb.WriteString("\n\n")

	// Current Phase
	if context.CurrentPhase != "" {
		sb.WriteString(overviewStyles.Primary.Render("Current Phase: "))
		sb.WriteString(context.CurrentPhase)
		sb.WriteString("\n\n")
	}

	// Overall Progress
	if context.OverallProgress > 0 {
		sb.WriteString(overviewStyles.Primary.Render("Overall Progress: "))
		progressBar := renderProgressBar(context.OverallProgress, 30)
		sb.WriteString(progressBar)
		sb.WriteString(fmt.Sprintf(" %.1f%%\n\n", context.OverallProgress*100))
	}

	// Current Goals
	if context.CurrentGoals != "" {
		sb.WriteString(overviewStyles.Primary.Render("Current Goals"))
		sb.WriteString("\n")
		renderJSONList(&sb, context.CurrentGoals, overviewStyles)
		sb.WriteString("\n")
	}

	// Pending Tasks
	if context.PendingTasks != "" {
		sb.WriteString(overviewStyles.Primary.Render("Pending Tasks"))
		sb.WriteString("\n")
		renderJSONList(&sb, context.PendingTasks, overviewStyles)
		sb.WriteString("\n")
	}

	// Completed Tasks
	if context.CompletedTasks != "" {
		sb.WriteString(overviewStyles.Primary.Render("Completed Tasks"))
		sb.WriteString("\n")
		renderJSONList(&sb, context.CompletedTasks, overviewStyles)
		sb.WriteString("\n")
	}

	// Next Actions
	if context.NextActions != "" {
		sb.WriteString(overviewStyles.Primary.Render("Next Actions"))
		sb.WriteString("\n")
		renderJSONList(&sb, context.NextActions, overviewStyles)
		sb.WriteString("\n")
	}

	// Blockers
	if context.Blockers != "" {
		sb.WriteString(overviewStyles.Primary.Render("Blockers"))
		sb.WriteString("\n")
		renderJSONList(&sb, context.Blockers, overviewStyles)
		sb.WriteString("\n")
	}

	// Milestones
	if context.Milestones != "" {
		sb.WriteString(overviewStyles.Primary.Render("Milestones"))
		sb.WriteString("\n")
		renderJSONList(&sb, context.Milestones, overviewStyles)
		sb.WriteString("\n")
	}

	// Important Files
	if context.ImportantFiles != "" {
		sb.WriteString(overviewStyles.Primary.Render("Important Files"))
		sb.WriteString("\n")
		renderJSONList(&sb, context.ImportantFiles, overviewStyles)
		sb.WriteString("\n")
	}

	// Key Decisions
	if context.KeyDecisions != "" {
		sb.WriteString(overviewStyles.Primary.Render("Key Decisions"))
		sb.WriteString("\n")
		renderJSONList(&sb, context.KeyDecisions, overviewStyles)
		sb.WriteString("\n")
	}

	// Learned Context
	if context.LearnedContext != "" {
		sb.WriteString(overviewStyles.Primary.Render("Learned Context"))
		sb.WriteString("\n")
		sb.WriteString(overviewStyles.Subtle.Render("  " + context.LearnedContext))
		sb.WriteString("\n\n")
	}

	// Metadata
	sb.WriteString(overviewStyles.Subtle.Render(strings.Repeat("─", ctx.Width-4)))
	sb.WriteString("\n")
	sb.WriteString(overviewStyles.Subtle.Render(fmt.Sprintf("Last updated: %s", context.UpdatedAt.Format("2006-01-02 15:04:05"))))
	sb.WriteString("\n")

	return sb.String()
}

// renderProgressBar renders a visual progress bar
func renderProgressBar(progress float32, width int) string {
	filled := int(progress * float32(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return bar
}

// renderJSONList parses and renders a JSON array as a bulleted list
func renderJSONList(sb *strings.Builder, jsonStr string, overviewStyles OverviewStyles) {
	if jsonStr == "" || jsonStr == "null" {
		sb.WriteString(overviewStyles.Subtle.Render("  (none)"))
		sb.WriteString("\n")
		return
	}

	var items []string
	err := json.Unmarshal([]byte(jsonStr), &items)
	if err != nil {
		// If not an array, try as single string
		var single string
		if err := json.Unmarshal([]byte(jsonStr), &single); err == nil {
			sb.WriteString("  • " + single)
			sb.WriteString("\n")
			return
		}
		// If parsing fails, show raw
		sb.WriteString(overviewStyles.Subtle.Render("  " + jsonStr))
		sb.WriteString("\n")
		return
	}

	if len(items) == 0 {
		sb.WriteString(overviewStyles.Subtle.Render("  (none)"))
		sb.WriteString("\n")
		return
	}

	for _, item := range items {
		sb.WriteString("  • " + item)
		sb.WriteString("\n")
	}
}
