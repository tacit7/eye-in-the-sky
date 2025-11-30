package tabs

import (
	"fmt"
)

// RenderProjects renders the project tickets tab content
// TODO: State migration in progress - needs projectTicketsIndex and split-pane state
// Note: DataContext doesn't have ProjectTickets yet, needs to be added
func RenderProjects(ctx *DataContext, overviewStyles OverviewStyles) string {
	// TODO: Add ProjectTickets and ProjectName fields to DataContext
	if ctx.Agent == nil {
		return overviewStyles.Subtle.Render("\n  No agent selected\n")
	}

	if ctx.Agent.ProjectName == "" {
		return overviewStyles.Subtle.Render("\n  No project associated with this agent\n")
	}

	// Stub for now - will implement after adding ProjectTickets to DataContext
	return overviewStyles.Subtle.Render(fmt.Sprintf("\n  Project tickets tab - state migration in progress\n  Project: %s\n", ctx.Agent.ProjectName))
}

// TODO: Split-pane view with project ticket details requires state migration
/*
// renderProjectTicketDetails renders the full ticket details in the right pane
func renderProjectTicketDetails(ticket domain.Task, styles OverviewStyles) string {
	var b strings.Builder

	// Ticket ID and Priority
	b.WriteString(styles.Primary.Render("Ticket: "))
	b.WriteString(string(ticket.ID))
	b.WriteString("  ")

	var priText string
	switch ticket.Priority {
	case 5:
		priText = "[CRITICAL]"
	case 4, 3:
		priText = "[HIGH]"
	case 2:
		priText = "[MEDIUM]"
	default:
		priText = "[LOW]"
	}
	b.WriteString(priText)
	b.WriteString("\n")

	// Status
	b.WriteString(styles.Primary.Render("Status: "))
	statusText := getTaskStateName(ticket.StateID)
	b.WriteString(statusText)
	b.WriteString("\n\n")

	// Description
	b.WriteString(styles.Label.Render("Description"))
	b.WriteString("\n")
	b.WriteString(ticket.Description)
	b.WriteString("\n\n")

	// Project
	if ticket.Project != "" {
		b.WriteString(styles.Primary.Render("Project: "))
		b.WriteString(ticket.Project)
		b.WriteString("\n")
	}

	// Dates
	b.WriteString(styles.Primary.Render("Created: "))
	b.WriteString(ticket.CreatedAt.Format("2006-01-02 15:04"))
	b.WriteString("\n")

	if !ticket.UpdatedAt.IsZero() && !ticket.UpdatedAt.Equal(ticket.CreatedAt) {
		b.WriteString(styles.Primary.Render("Updated: "))
		b.WriteString(ticket.UpdatedAt.Format("2006-01-02 15:04"))
		b.WriteString("\n")
	}

	// Tags
	if len(ticket.Tags) > 0 {
		b.WriteString("\n")
		b.WriteString(styles.Primary.Render("Tags: "))
		for i, tag := range ticket.Tags {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(tag)
		}
		b.WriteString("\n")
	}

	// Archived status
	if ticket.Archived {
		b.WriteString("\n")
		b.WriteString(styles.Error.Render("\uf071 ARCHIVED")) // nf-fa-exclamation_triangle
		b.WriteString("\n")
	}

	return b.String()
}

// Helper functions for task state rendering
func getTaskStateIcon(stateID int) string {
	switch stateID {
	case 1:
		return "\uf096" // nf-fa-square_o (todo)
	case 2:
		return "\uf04b" // nf-fa-play (in progress)
	case 3:
		return "\uf00c" // nf-fa-check (done)
	default:
		return "\uf128" // nf-fa-question
	}
}

func getTaskStateName(stateID int) string {
	switch stateID {
	case 1:
		return "TODO"
	case 2:
		return "IN PROGRESS"
	case 3:
		return "DONE"
	default:
		return "UNKNOWN"
	}
}
*/
