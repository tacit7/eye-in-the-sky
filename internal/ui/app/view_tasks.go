package app

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TaskStateType represents the async state of task loading
type TaskStateType int

const (
	TaskIdle TaskStateType = iota
	TaskLoading
	TaskLoaded
	TaskError
)

// TasksLoadedMsg contains the loaded tasks
type TasksLoadedMsg struct {
	Tasks []Task
}

// TasksErrorMsg contains task loading errors
type TasksErrorMsg struct {
	Error error
}

// renderAgentTasks renders the agent tasks tab in the agent detail view
// Shows tasks filtered by the current session
func (m *Model) renderAgentTasks() string {
	if m.selectedAgent == nil {
		return m.styles.Subtle.Render("No agent selected")
	}

	// Show loading state
	if m.taskState == TaskLoading {
		return m.styles.Subtle.Render("Loading tasks...")
	}

	// Show error state
	if m.taskState == TaskError {
		if m.taskError != nil {
			return m.styles.Failed.Render(fmt.Sprintf("Error: %v", m.taskError))
		}
		return m.styles.Failed.Render("Error loading tasks")
	}

	// Build status header showing task count
	var statusHeader string
	if m.taskState == TaskLoaded {
		statusHeader = m.styles.Completed.Render(fmt.Sprintf("✓ Loaded %d task(s)", len(m.tasks)))
	} else if len(m.tasks) == 0 {
		statusHeader = m.styles.Subtle.Render("No tasks")
	} else {
		statusHeader = m.styles.Primary.Render(fmt.Sprintf("%d task(s)", len(m.tasks)))
	}

	// Build item list for left pane
	items := make([]string, len(m.tasks))
	for i, task := range m.tasks {
		status := task.Status
		if len(status) > 10 {
			status = status[:10]
		}
		// Extract short UUID (first 8 chars) for ticket number display
		shortUUID := task.UUID
		if len(shortUUID) > 8 {
			shortUUID = shortUUID[:8]
		}
		items[i] = fmt.Sprintf("[%s] [%-10s] %s", shortUUID, status, truncate(task.Description, 40))
	}

	// Build detail content for right pane
	var detailContent string
	if m.tasksIndex >= 0 && m.tasksIndex < len(m.tasks) {
		detailContent = m.renderTaskDetails(m.tasks[m.tasksIndex])
	}

	// Footer
	footer := m.renderFooterWithKeys("[j/k] Move  [d] Mark Done  [r] Refresh  [q/esc] Back")

	splitView := m.renderSplitPaneView(
		"Tasks",
		items,
		m.tasksIndex,
		len(m.tasks),
		detailContent,
		footer,
	)

	// Prepend status header
	return fmt.Sprintf("%s\n%s", statusHeader, splitView)
}

// renderProjectTickets renders all tickets for the current agent's project
func (m *Model) renderProjectTickets() string {
	if m.selectedAgent == nil {
		return m.styles.Subtle.Render("No agent selected")
	}

	// Check if agent has a project assigned
	if m.selectedAgent.ProjectName == "" {
		return m.styles.Subtle.Render("No project assigned to this agent")
	}

	// Build status header showing ticket count
	var statusHeader string
	if len(m.projectTickets) == 0 {
		statusHeader = m.styles.Subtle.Render("No tickets found for project")
	} else {
		statusHeader = m.styles.Primary.Render(fmt.Sprintf("%d ticket(s) in project '%s'", len(m.projectTickets), m.selectedAgent.ProjectName))
	}

	// Build item list for left pane
	items := make([]string, len(m.projectTickets))
	for i, task := range m.projectTickets {
		status := task.Status
		if len(status) > 10 {
			status = status[:10]
		}
		// Extract short UUID (first 8 chars) for ticket number display
		shortUUID := task.UUID
		if len(shortUUID) > 8 {
			shortUUID = shortUUID[:8]
		}
		items[i] = fmt.Sprintf("[%s] [%-10s] %s", shortUUID, status, truncate(task.Description, 40))
	}

	// Build detail content for right pane
	var detailContent string
	if m.projectTicketsIndex >= 0 && m.projectTicketsIndex < len(m.projectTickets) {
		detailContent = m.renderTaskDetails(m.projectTickets[m.projectTicketsIndex])
	}

	// Footer
	footer := m.renderFooterWithKeys("[j/k] Move  [r] Refresh  [q/esc] Back")

	splitView := m.renderSplitPaneView(
		"Project Tickets",
		items,
		m.projectTicketsIndex,
		len(m.projectTickets),
		detailContent,
		footer,
	)

	// Prepend status header
	return fmt.Sprintf("%s\n%s", statusHeader, splitView)
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

	// Determine if this is a subagent or parent agent
	isSubagent := m.selectedAgent.ParentAgentID != ""
	var searchTag string

	if isSubagent {
		// For subagents: filter by +subagent_<agent_id> tag (replace hyphens with underscores for TaskWarrior compatibility)
		searchTag = fmt.Sprintf("+subagent_%s", strings.ReplaceAll(m.selectedAgent.ID, "-", "_"))
		log.Printf("[TASKS] Loading tasks for subagent: %s (tag: %s)", m.selectedAgent.ID, searchTag)
	} else {
		// For parent agents: filter by +session:<session_id> tag
		if m.selectedAgent.SessionID == "" {
			// No session ID, return empty list
			m.tasks = make([]Task, 0)
			return nil
		}
		searchTag = fmt.Sprintf("+session_%s", strings.ReplaceAll(m.selectedAgent.SessionID, "-", "_"))
		log.Printf("[TASKS] Loading tasks for session: %s (tag: %s)", m.selectedAgent.SessionID, searchTag)
	}

	// Run task export and filter by appropriate tag in description
	cmd := exec.Command("task", "export")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("[TASKS] Task command failed: %v", err)
		// Return empty list if no tasks found (not an error)
		m.tasks = make([]Task, 0)
		return nil
	}
	log.Printf("[TASKS] Found %d bytes of task output", len(output))

	// Parse JSON output
	var tasks []map[string]interface{}
	if err := json.Unmarshal(output, &tasks); err != nil {
		return fmt.Errorf("failed to parse tasks: %w", err)
	}

	// Convert to Task structs, filtering by appropriate tag in description
	m.tasks = make([]Task, 0, len(tasks))
	agentSearchStr := fmt.Sprintf("+agent_%s", strings.ReplaceAll(m.selectedAgent.ID, "-", "_"))
	for _, taskData := range tasks {
		description := getString(taskData, "description")
		// Filter tasks that contain the appropriate tag (workaround for Taskwarrior tag parsing)
		if !strings.Contains(description, searchTag) {
			continue
		}
		// Clean up description by removing agent and session/subagent tags
		cleanDescription := strings.TrimSpace(description)
		cleanDescription = strings.ReplaceAll(cleanDescription, searchTag, "")
		cleanDescription = strings.ReplaceAll(cleanDescription, agentSearchStr, "")
		cleanDescription = strings.TrimSpace(cleanDescription)

		task := Task{
			UUID:        getString(taskData, "uuid"),
			Description: cleanDescription,
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

// loadTasksCmd returns an async tea.Cmd that loads tasks without blocking
func (m *Model) loadTasksCmd() tea.Cmd {
	return func() tea.Msg {
		// Replicate loadTasks logic but return message instead of modifying model
		if m.selectedAgent == nil {
			return TasksErrorMsg{Error: fmt.Errorf("no agent selected")}
		}

		// Check if taskwarrior is installed
		if _, err := exec.LookPath("task"); err != nil {
			return TasksErrorMsg{Error: err}
		}

		// Determine if this is a subagent or parent agent
		isSubagent := m.selectedAgent.ParentAgentID != ""
		var searchTag string

		if isSubagent {
			// For subagents: filter by +subagent:<agent_id> tag
			searchTag = fmt.Sprintf("+subagent:%s", strings.ReplaceAll(m.selectedAgent.ID, "-", "_"))
			log.Printf("[TASKS] Async loading tasks for subagent: %s (tag: %s)", m.selectedAgent.ID, searchTag)
		} else {
			// For parent agents: filter by +session:<session_id> tag
			if m.selectedAgent.SessionID == "" {
				return TasksLoadedMsg{Tasks: make([]Task, 0)}
			}
			searchTag = fmt.Sprintf("+session_%s", strings.ReplaceAll(m.selectedAgent.SessionID, "-", "_"))
			log.Printf("[TASKS] Async loading tasks for session: %s (tag: %s)", m.selectedAgent.SessionID, searchTag)
		}

		// Run task export
		cmd := exec.Command("task", "export")
		output, err := cmd.Output()
		if err != nil {
			log.Printf("[TASKS] Async task command failed: %v", err)
			// Return empty list instead of error (consistent with sync version)
			return TasksLoadedMsg{Tasks: make([]Task, 0)}
		}

		// Parse JSON
		var tasks []map[string]interface{}
		if err := json.Unmarshal(output, &tasks); err != nil {
			return TasksErrorMsg{Error: err}
		}

		// Convert to Task structs with filtering
		agentSearchStr := fmt.Sprintf("+agent:%s", strings.ReplaceAll(m.selectedAgent.ID, "-", "_"))
		filteredTasks := make([]Task, 0, len(tasks))

		for _, taskData := range tasks {
			description := getString(taskData, "description")
			if !strings.Contains(description, searchTag) {
				continue
			}

			cleanDescription := strings.TrimSpace(description)
			cleanDescription = strings.ReplaceAll(cleanDescription, searchTag, "")
			cleanDescription = strings.ReplaceAll(cleanDescription, agentSearchStr, "")
			cleanDescription = strings.TrimSpace(cleanDescription)

			task := Task{
				UUID:        getString(taskData, "uuid"),
				Description: cleanDescription,
				Status:      getString(taskData, "status"),
				Priority:    getString(taskData, "priority"),
				Project:     getString(taskData, "project"),
				Tags:        getTags(taskData, "tags"),
				Due:         getTime(taskData, "due"),
				Entry:       getTime(taskData, "entry"),
				Annotations: getAnnotations(taskData, "annotations"),
			}
			filteredTasks = append(filteredTasks, task)
		}

		// Sort before returning
		sortTasksByPriorityAndStatus(filteredTasks)

		return TasksLoadedMsg{Tasks: filteredTasks}
	}
}

// loadProjectTickets loads all tickets for the current agent's project
func (m *Model) loadProjectTickets() error {
	if m.selectedAgent == nil {
		return fmt.Errorf("no agent selected")
	}

	if m.selectedAgent.ProjectName == "" {
		m.projectTickets = make([]Task, 0)
		return nil
	}

	// Check if taskwarrior is installed
	if _, err := exec.LookPath("task"); err != nil {
		return fmt.Errorf("taskwarrior not installed")
	}

	// Run task export with project filter
	projectFilter := fmt.Sprintf("project:%s", m.selectedAgent.ProjectName)
	cmd := exec.Command("task", projectFilter, "export")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("[PROJECT TICKETS] Task command failed: %v", err)
		// Return empty list if no tasks found (not an error)
		m.projectTickets = make([]Task, 0)
		return nil
	}
	log.Printf("[PROJECT TICKETS] Found %d bytes of task output for project '%s'", len(output), m.selectedAgent.ProjectName)

	// Parse JSON output
	var tasks []map[string]interface{}
	if err := json.Unmarshal(output, &tasks); err != nil {
		return fmt.Errorf("failed to parse tasks: %w", err)
	}

	// Convert to Task structs
	m.projectTickets = make([]Task, 0, len(tasks))
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
		m.projectTickets = append(m.projectTickets, task)
	}

	// Sort by priority and status
	sortTasksByPriorityAndStatus(m.projectTickets)
	log.Printf("[PROJECT TICKETS] Loaded %d tickets for project '%s'", len(m.projectTickets), m.selectedAgent.ProjectName)

	return nil
}

// loadProjectTicketsCmd returns an async tea.Cmd that loads project tickets without blocking
func (m *Model) loadProjectTicketsCmd() tea.Cmd {
	return func() tea.Msg {
		if m.selectedAgent == nil {
			return TasksErrorMsg{Error: fmt.Errorf("no agent selected")}
		}

		if m.selectedAgent.ProjectName == "" {
			// Return success message with empty list
			return TasksLoadedMsg{Tasks: make([]Task, 0)}
		}

		// Check if taskwarrior is installed
		if _, err := exec.LookPath("task"); err != nil {
			return TasksErrorMsg{Error: err}
		}

		// Run task export with project filter
		projectFilter := fmt.Sprintf("project:%s", m.selectedAgent.ProjectName)
		cmd := exec.Command("task", projectFilter, "export")
		output, err := cmd.Output()
		if err != nil {
			log.Printf("[PROJECT TICKETS] Async task command failed: %v", err)
			// Return empty list instead of error
			return TasksLoadedMsg{Tasks: make([]Task, 0)}
		}

		// Parse JSON
		var tasks []map[string]interface{}
		if err := json.Unmarshal(output, &tasks); err != nil {
			return TasksErrorMsg{Error: err}
		}

		// Convert to Task structs
		filteredTasks := make([]Task, 0, len(tasks))
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
			filteredTasks = append(filteredTasks, task)
		}

		// Sort by priority and status
		sortTasksByPriorityAndStatus(filteredTasks)

		return TasksLoadedMsg{Tasks: filteredTasks}
	}
}
