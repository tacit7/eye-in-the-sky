package tabs

import (
	"fmt"
	"strings"
)

// RenderOverview renders the project overview tab content
// Pure function - takes context + styles, returns rendered string
func RenderOverview(ctx *DataContext, styles OverviewStyles) string {
	if ctx == nil || ctx.Project == nil {
		return "No project selected"
	}

	var sections []string

	// Project Information Section
	projectInfo := renderProjectInfo(ctx, styles)
	if projectInfo != "" {
		sections = append(sections, projectInfo)
	}

	// Summary Counts Section
	summarySection := renderSummaryCounts(ctx, styles)
	if summarySection != "" {
		sections = append(sections, summarySection)
	}

	// Join all sections with newlines
	return strings.Join(filterNonEmpty(sections), "\n\n")
}

// renderProjectInfo renders the project basic information
func renderProjectInfo(ctx *DataContext, styles OverviewStyles) string {
	if ctx.Project == nil {
		return ""
	}

	var b strings.Builder

	// Section header
	b.WriteString(styles.SectionTitle.Render("\uf07b Project Information")) // nf-fa-folder
	b.WriteString("\n\n")

	// Project name
	if ctx.Project.Name != "" {
		b.WriteString(renderLabelValue("Name:", ctx.Project.Name, styles.Value, styles.Label))
	}

	// Project path
	if ctx.Project.Path != nil && *ctx.Project.Path != "" {
		b.WriteString(renderLabelValue("Path:", *ctx.Project.Path, styles.Subtle, styles.Label))
	}

	// Git branch (if available)
	if ctx.Project.Branch != nil && *ctx.Project.Branch != "" {
		b.WriteString(renderLabelValue("Branch:", *ctx.Project.Branch, styles.Git, styles.Label))
	}

	// Remote URL (if available)
	if ctx.Project.RemoteURL != nil && *ctx.Project.RemoteURL != "" {
		b.WriteString(renderLabelValue("Remote:", *ctx.Project.RemoteURL, styles.Subtle, styles.Label))
	}

	return b.String()
}

// renderSummaryCounts renders summary counts for agents, notes, and tasks
func renderSummaryCounts(ctx *DataContext, styles OverviewStyles) string {
	var b strings.Builder

	// Section header
	b.WriteString(styles.SectionTitle.Render("\uf080 Summary")) // nf-fa-bar_chart
	b.WriteString("\n\n")

	// Active agents count
	activeAgents := len(ctx.Agents)
	b.WriteString(renderLabelValue("Active Agents:", fmt.Sprintf("%d", activeAgents), styles.Success, styles.Label))

	// Notes count
	notesCount := len(ctx.Notes)
	b.WriteString(renderLabelValue("Notes:", fmt.Sprintf("%d", notesCount), styles.Primary, styles.Label))

	// Tasks count
	tasksCount := len(ctx.Tasks)
	b.WriteString(renderLabelValue("Tasks:", fmt.Sprintf("%d", tasksCount), styles.Warning, styles.Label))

	return b.String()
}
