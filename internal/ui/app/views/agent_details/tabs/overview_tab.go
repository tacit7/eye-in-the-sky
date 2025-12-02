package tabs

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/domain"
	"github.com/tacit7/eye-in-the-sky/internal/ui/presenters"
)

const (
	maxCommitsToShow = 5
	maxActionsToShow = 3
)

// DataContext wraps all data needed by tabs (avoids circular import)
type DataContext struct {
	Agent              *domain.Agent
	Commits            []domain.Commit
	Notes              []domain.Note
	Tasks              []domain.Task
	TaskNotes          []domain.TaskNote // Annotations for selected task
	Actions            []domain.Action
	Logs               []domain.Log
	SessionContexts    []domain.SessionContext // Saved session contexts
	SelectedTaskIndex  int                     // Index of selected task in tasks list
	NotesIndex         int                     // Index of selected note in notes list
	CommitsIndex       int                     // Index of selected commit in commits list
	LogsListIndex      int                     // Viewport scroll position for logs tab
	LastFetchedAt      time.Time               // Last timestamp for incremental log fetching
	Width                int                     // Terminal width for responsive layout
	MarkdownRenderer     MarkdownRenderer        // Markdown renderer for formatting
	CommitDetailViewport viewport.Model         // Viewport for scrollable commit details
	TaskDetailViewport   viewport.Model         // Viewport for scrollable task details
	LogsViewport         viewport.Model         // Viewport for scrollable logs
}

// MarkdownRenderer is an interface for rendering markdown
type MarkdownRenderer interface {
	Render(in string) (string, error)
}

// OverviewStyles is temporarily kept here to avoid cycles
// TODO: Extract to shared location after full migration
type OverviewStyles struct {
	SectionTitle lipgloss.Style
	Label        lipgloss.Style
	Value        lipgloss.Style
	Subtle       lipgloss.Style
	Success      lipgloss.Style
	Warning      lipgloss.Style
	Primary      lipgloss.Style
	Secondary    lipgloss.Style
	Git          lipgloss.Style
	Error        lipgloss.Style
}

// getStatusStyleFromOverview converts OverviewStyles to Styles for GetStatusStyle
// This is a temporary helper to bridge the style systems
func getStatusStyleFromOverview(status string, overviewStyles OverviewStyles) lipgloss.Style {
	switch status {
	case "active":
		return overviewStyles.Success
	case "working":
		return overviewStyles.Warning
	case "idle":
		return overviewStyles.Subtle
	case "failed":
		return overviewStyles.Error
	case "completed":
		return overviewStyles.Primary
	default:
		return overviewStyles.Subtle
	}
}

// RenderOverview renders the overview tab content
// Pure function - takes context + styles, returns rendered string
func RenderOverview(ctx *DataContext, overviewStyles OverviewStyles) string {
	if ctx == nil || ctx.Agent == nil {
		return "No agent selected"
	}

	// Build preprocessed data using presenter layer
	data := presenters.BuildOverviewData(
		ctx.Agent,
		ctx.Commits,
		ctx.Notes,
		ctx.Tasks,
		ctx.Actions,
	)

	// Render each section with clean data and centralized styles
	sections := []string{
		renderAgentInfo(data.AgentInfo, overviewStyles),
		renderTiming(data.AgentInfo, overviewStyles),
		renderCommitsSection(data.Commits, overviewStyles),
		renderNotesSection(data.Notes, overviewStyles),
		renderTasksSection(data.TaskCounts, overviewStyles),
		renderActionsSection(data.Actions, overviewStyles),
	}

	// Return joined output
	return strings.Join(filterNonEmpty(sections), "\n")
}

// renderAgentInfo renders agent identification information
// Pure function - no Model dependency
func renderAgentInfo(agent domain.Agent, styles OverviewStyles) string {
	var sb strings.Builder
	sb.WriteString(styles.SectionTitle.Render("\uf05a Agent Information"))
	sb.WriteString("\n\n")

	sb.WriteString(renderLabelValue("Agent ID:", truncateID(string(agent.ID), 8), styles.Value, styles.Label))
	sb.WriteString(renderLabelValue("Status:", agent.Status, getStatusStyleFromOverview(agent.Status, styles), styles.Label))

	if agent.FeatureDesc != "" {
		sb.WriteString(renderLabelValue("Description:", agent.FeatureDesc, styles.Value, styles.Label))
	}
	if agent.CurrentTask != "" {
		sb.WriteString(renderLabelValue("Current Task:", agent.CurrentTask, styles.Value, styles.Label))
	}

	sb.WriteString(renderLabelValue("Source:", agent.Source, styles.Value, styles.Label))

	if agent.GitWorktreePath != "" {
		sb.WriteString(renderLabelValue("Worktree:", agent.GitWorktreePath, styles.Value, styles.Label))
	}
	if agent.SessionID != "" {
		sb.WriteString(renderLabelValue("Session ID:", truncateID(agent.SessionID, 8), styles.Value, styles.Label))
	}
	if agent.ProjectName != "" {
		sb.WriteString(renderLabelValue("Project:", agent.ProjectName, styles.Value, styles.Label))
	}
	if agent.ParentSessionID != "" {
		sb.WriteString(renderLabelValue("Parent Session:", truncateID(agent.ParentSessionID, 8), styles.Value, styles.Label))
	}
	if agent.ParentAgentID != "" {
		sb.WriteString(renderLabelValue("Parent Agent:", truncateID(string(agent.ParentAgentID), 8), styles.Value, styles.Label))
	}
	if agent.WindowID != "" {
		sb.WriteString(renderLabelValue("Window ID:", agent.WindowID, styles.Value, styles.Label))
	}

	return sb.String()
}

// renderTiming renders timing information section
// Pure function - no Model dependency
func renderTiming(agent domain.Agent, styles OverviewStyles) string {
	var sb strings.Builder
	sb.WriteString(styles.SectionTitle.Render("\uf017 Timing"))
	sb.WriteString("\n\n")

	sb.WriteString(renderLabelValue("Created:", formatTimestamp(agent.CreatedAt), styles.Value, styles.Label))
	sb.WriteString(renderLabelValue("Updated:", formatTimestamp(agent.UpdatedAt), styles.Value, styles.Label))

	if !agent.LastActivityAt.IsZero() {
		elapsed := time.Since(agent.LastActivityAt)
		activityStr := fmt.Sprintf("%s (%s ago)",
			formatTime(agent.LastActivityAt),
			presenters.FormatDuration(elapsed))
		style := activityStyle(elapsed, styles)
		sb.WriteString(renderLabelValue("Last Activity:", activityStr, style, styles.Label))
	}

	sb.WriteString(renderDuration(agent, styles))

	return sb.String()
}

// renderCommitsSection renders recent commits section
// Pure function - no Model dependency
func renderCommitsSection(commits []domain.Commit, styles OverviewStyles) string {
	if len(commits) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(styles.SectionTitle.Render("\uf1d3 Recent Commits"))
	sb.WriteString("\n\n")

	maxCommits := minInt(len(commits), maxCommitsToShow)
	for i := 0; i < maxCommits; i++ {
		commit := commits[i]
		elapsed := time.Since(commit.Timestamp)
		commitLine := fmt.Sprintf("  %s %s - %s\n",
			styles.Git.Render(truncateID(string(commit.Hash), 8)),
			styles.Subtle.Render(fmt.Sprintf("(%s ago)", presenters.FormatDuration(elapsed))),
			commit.Message)
		sb.WriteString(commitLine)
	}

	if len(commits) > maxCommits {
		sb.WriteString(styles.Subtle.Render(
			fmt.Sprintf("  ... and %d more commits", len(commits)-maxCommits)))
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderNotesSection renders notes summary section with timestamp, title, and content
// Pure function - no Model dependency
func renderNotesSection(notes []domain.Note, styles OverviewStyles) string {
	if len(notes) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(styles.SectionTitle.Render("\uf249 Notes Summary"))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("  %s %d notes\n",
		styles.Label.Render("Total:"),
		len(notes)))

	mostRecent := notes[0]

	// Show timestamp
	timestamp := mostRecent.CreatedAt.Format("2006-01-02 15:04:05")
	sb.WriteString(fmt.Sprintf("  %s %s\n",
		styles.Label.Render("Latest:"),
		styles.Subtle.Render(timestamp)))

	// Show title
	sb.WriteString(fmt.Sprintf("  %s %s\n",
		styles.Label.Render("Title:"),
		styles.Primary.Render(mostRecent.Title)))

	// Show content preview
	preview := mostRecent.Content
	if len(preview) > 80 {
		preview = preview[:77] + "..."
	}
	sb.WriteString(fmt.Sprintf("  %s %s\n",
		styles.Label.Render("Content:"),
		styles.Value.Render(preview)))

	return sb.String()
}

// renderTasksSection renders task summary section
// Pure function - no Model dependency
func renderTasksSection(taskCounts map[string]int, styles OverviewStyles) string {
	if len(taskCounts) == 0 || (taskCounts["todo"] == 0 && taskCounts["inProgress"] == 0 && taskCounts["completed"] == 0 && taskCounts["archived"] == 0) {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(styles.SectionTitle.Render("\uf0ae Tasks Summary"))
	sb.WriteString("\n\n")

	if taskCounts["todo"] > 0 {
		sb.WriteString(fmt.Sprintf("  %s %d tasks\n",
			styles.Primary.Render("Todo:"),
			taskCounts["todo"]))
	}
	if taskCounts["inProgress"] > 0 {
		sb.WriteString(fmt.Sprintf("  %s %d tasks\n",
			styles.Warning.Render("In Progress:"),
			taskCounts["inProgress"]))
	}
	if taskCounts["completed"] > 0 {
		sb.WriteString(fmt.Sprintf("  %s %d tasks\n",
			styles.Success.Render("Done:"),
			taskCounts["completed"]))
	}
	if taskCounts["archived"] > 0 {
		sb.WriteString(fmt.Sprintf("  %s %d tasks\n",
			styles.Subtle.Render("Archived:"),
			taskCounts["archived"]))
	}

	return sb.String()
}

// renderActionsSection renders recent actions section
// Pure function - no Model dependency
func renderActionsSection(actions []domain.Action, styles OverviewStyles) string {
	if len(actions) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(styles.SectionTitle.Render("⚡ Recent Actions"))
	sb.WriteString("\n\n")

	maxActions := minInt(len(actions), maxActionsToShow)
	for i := 0; i < maxActions; i++ {
		action := actions[i]
		elapsed := time.Since(action.Timestamp)
		actionLine := fmt.Sprintf("  %s %s - %s\n",
			GetActionIcon(action.ActionType),
			styles.Subtle.Render(fmt.Sprintf("(%s ago)", presenters.FormatDuration(elapsed))),
			action.Description)
		sb.WriteString(actionLine)
	}

	if len(actions) > maxActions {
		sb.WriteString(styles.Subtle.Render(
			fmt.Sprintf("  ... and %d more actions", len(actions)-maxActions)))
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderDuration renders session or total duration
// Pure function - no Model dependency
func renderDuration(agent domain.Agent, styles OverviewStyles) string {
	if agent.Status != "completed" && agent.Status != "failed" {
		duration := time.Since(agent.CreatedAt)
		return renderLabelValue("Session Duration:", presenters.FormatDuration(duration), styles.Value, styles.Label)
	} else if agent.CompletedAt != nil {
		duration := agent.CompletedAt.Sub(agent.CreatedAt)
		return renderLabelValue("Total Duration:", presenters.FormatDuration(duration), styles.Value, styles.Label)
	}
	return ""
}
