package app

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// renderAgentTasks renders the agent tasks tab in the agent detail view
// Shows tasks filtered by the current session
func (m *Model) renderAgentTasks() string {
	if m.selectedAgent == nil {
		return m.styles.Subtle.Render("No agent selected")
	}

	// Build item list for left pane
	items := make([]string, len(m.tasks))
	for i, task := range m.tasks {
		status := task.Status
		if len(status) > 10 {
			status = status[:10]
		}
		items[i] = fmt.Sprintf("[%-10s] %s", status, truncate(task.Description, 50))
	}

	// Build detail content for right pane
	var detailContent string
	if m.tasksIndex >= 0 && m.tasksIndex < len(m.tasks) {
		detailContent = m.renderTaskDetails(m.tasks[m.tasksIndex])
	}

	// Footer
	footer := m.renderFooterWithKeys("[j/k] Move  [d] Mark Done  [r] Refresh  [q/esc] Back")

	return m.renderSplitPaneView(
		"Tasks",
		items,
		m.tasksIndex,
		len(m.tasks),
		detailContent,
		footer,
	)
}

// renderTaskDetails renders the details of a single task
func (m *Model) renderTaskDetails(task Task) string {
	var b strings.Builder

	// UUID and Description
	b.WriteString(m.styles.Primary.Render("UUID: "))
	b.WriteString(m.styles.Text.Render(task.UUID))
	b.WriteString("\n\n")

	b.WriteString(m.styles.Primary.Render("Description: "))
	b.WriteString(m.styles.Text.Render(task.Description))
	b.WriteString("\n\n")

	// Status
	b.WriteString(m.styles.Primary.Render("Status: "))
	statusStyle := m.styles.Text
	switch task.Status {
	case "pending":
		statusStyle = m.styles.Idle
	case "completed":
		statusStyle = m.styles.Completed
	case "deleted":
		statusStyle = m.styles.Failed
	}
	b.WriteString(statusStyle.Render(strings.ToUpper(task.Status)))
	b.WriteString("\n\n")

	// Priority
	if task.Priority != "" {
		b.WriteString(m.styles.Primary.Render("Priority: "))
		b.WriteString(m.styles.Text.Render(strings.ToUpper(task.Priority)))
		b.WriteString("\n\n")
	}

	// Project
	if task.Project != "" {
		b.WriteString(m.styles.Primary.Render("Project: "))
		b.WriteString(m.styles.Text.Render(task.Project))
		b.WriteString("\n\n")
	}

	// Tags
	if len(task.Tags) > 0 {
		b.WriteString(m.styles.Primary.Render("Tags: "))
		b.WriteString(m.styles.Text.Render(strings.Join(task.Tags, ", ")))
		b.WriteString("\n\n")
	}

	// Due date
	if !task.Due.IsZero() {
		b.WriteString(m.styles.Primary.Render("Due: "))
		b.WriteString(m.styles.Text.Render(task.Due.Format("2006-01-02 15:04")))
		b.WriteString("\n\n")
	}

	// Entry date
	if !task.Entry.IsZero() {
		b.WriteString(m.styles.Primary.Render("Created: "))
		b.WriteString(m.styles.Text.Render(task.Entry.Format("2006-01-02 15:04")))
		b.WriteString("\n\n")
	}

	// Annotations
	if len(task.Annotations) > 0 {
		b.WriteString(m.styles.Title.Render("Annotations"))
		b.WriteString("\n")
		for _, ann := range task.Annotations {
			line := fmt.Sprintf("  %s  %s",
				ann.Entry.Format("2006-01-02 15:04"),
				ann.Description,
			)
			b.WriteString(m.styles.Text.Render(line))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// loadTasks loads tasks for the current session from Taskwarrior
func (m *Model) loadTasks() error {
	if m.selectedAgent == nil {
		return fmt.Errorf("no agent selected")
	}

	// Check if taskwarrior is installed
	if _, err := exec.LookPath("task"); err != nil {
		return fmt.Errorf("taskwarrior not installed")
	}

	// Build task filter for current session
	// Filter by session tag with underscores: session:45045287_a68c_4ec5_833f_5c46535be414
	// Show all tasks (pending, completed, deleted, etc.)
	sessionID := strings.ReplaceAll(m.selectedAgent.CurrentSessionID, "-", "_")
	sessionTag := fmt.Sprintf("+session:%s", sessionID)

	// Run task export - shows all statuses for this session
	cmd := exec.Command("task", sessionTag, "export")
	output, err := cmd.Output()
	if err != nil {
		// Return empty list if no tasks found (not an error)
		m.tasks = make([]Task, 0)
		return nil
	}

	// Parse JSON output
	var tasks []map[string]interface{}
	if err := json.Unmarshal(output, &tasks); err != nil {
		return fmt.Errorf("failed to parse tasks: %w", err)
	}

	// Convert to Task structs
	m.tasks = make([]Task, 0, len(tasks))
	for _, taskData := range tasks {
		task := Task{
			UUID:        getString(taskData, "uuid"),
			Description: getString(taskData, "description"),
			Status:      getString(taskData, "status"),
			Priority:    getString(taskData, "priority"),
			Project:     getString(taskData, "project"),
			Tags:        getTags(taskData, "tags"),
			Due:         getTime(taskData, "due"),
			Entry:       getTime(taskData, "entry"),
			Annotations: getAnnotations(taskData, "annotations"),
		}
		m.tasks = append(m.tasks, task)
	}

	// Sort tasks: by priority first, deleted tasks at bottom
	sortTasksByPriorityAndStatus(m.tasks)

	// Reset scroll position
	if len(m.tasks) > 0 && m.tasksIndex >= len(m.tasks) {
		m.tasksIndex = len(m.tasks) - 1
	}

	return nil
}

// sortTasksByPriorityAndStatus sorts tasks by priority, with deleted tasks at the bottom
func sortTasksByPriorityAndStatus(tasks []Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		// Deleted tasks always go to the bottom
		if tasks[i].Status == "deleted" && tasks[j].Status != "deleted" {
			return false
		}
		if tasks[i].Status != "deleted" && tasks[j].Status == "deleted" {
			return true
		}

		// Priority order: H > M > L > (empty)
		priorityOrder := map[string]int{"H": 3, "M": 2, "L": 1, "": 0}
		priorityI := priorityOrder[tasks[i].Priority]
		priorityJ := priorityOrder[tasks[j].Priority]

		if priorityI != priorityJ {
			return priorityI > priorityJ
		}

		// If priorities are equal, sort by entry time (oldest first)
		return tasks[i].Entry.Before(tasks[j].Entry)
	})
}

// markTaskDone marks the currently selected task as done
func (m *Model) markTaskDone() error {
	if m.tasksIndex < 0 || m.tasksIndex >= len(m.tasks) {
		return fmt.Errorf("no task selected")
	}

	task := m.tasks[m.tasksIndex]
	cmd := exec.Command("task", task.UUID, "done")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to mark task done: %w", err)
	}

	// Reload tasks
	return m.loadTasks()
}

// Helper functions for parsing taskwarrior JSON

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getTags(data map[string]interface{}, key string) []string {
	if val, ok := data[key].([]interface{}); ok {
		tags := make([]string, 0, len(val))
		for _, tag := range val {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
		return tags
	}
	return nil
}

func getTime(data map[string]interface{}, key string) time.Time {
	if val, ok := data[key].(string); ok {
		// Taskwarrior uses ISO 8601 format
		t, err := time.Parse("20060102T150405Z", val)
		if err != nil {
			return time.Time{}
		}
		return t
	}
	return time.Time{}
}

func getAnnotations(data map[string]interface{}, key string) []TaskAnnotation {
	if val, ok := data[key].([]interface{}); ok {
		annotations := make([]TaskAnnotation, 0, len(val))
		for _, ann := range val {
			if annMap, ok := ann.(map[string]interface{}); ok {
				annotation := TaskAnnotation{
					Entry:       getTime(annMap, "entry"),
					Description: getString(annMap, "description"),
				}
				annotations = append(annotations, annotation)
			}
		}
		return annotations
	}
	return nil
}
