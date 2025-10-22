package data

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// taskStore implements TaskWarrior integration
type taskStore struct{}

// NewTaskStore creates a new TaskWarrior store
func NewTaskStore() *taskStore {
	return &taskStore{}
}

// twTask represents TaskWarrior JSON structure
type twTask struct {
	UUID        string   `json:"uuid"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	Priority    string   `json:"priority"`
	Project     string   `json:"project"`
	Tags        []string `json:"tags"`
	Due         string   `json:"due"`   // RFC3339 string
	Entry       string   `json:"entry"` // RFC3339 string
	Annotations []struct {
		Entry       string `json:"entry"`
		Description string `json:"description"`
	} `json:"annotations,omitempty"`
}

// LoadCountsByAgent returns task counts for multiple agents
func (s *taskStore) LoadCountsByAgent(ctx context.Context, agents []domain.Agent) ([]struct {
	AgentID domain.AgentID
	Count   int
}, error) {
	// Export all tasks once
	cmd := exec.CommandContext(ctx, "task", "export")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start task export: %w", err)
	}

	// Stream decode tasks
	dec := json.NewDecoder(stdout)
	var allTasks []twTask

	for dec.More() {
		var t twTask
		if err := dec.Decode(&t); err != nil {
			// Continue on decode error
			continue
		}
		// Only count pending tasks
		if t.Status == "pending" {
			allTasks = append(allTasks, t)
		}
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("task export wait: %w", err)
	}

	// Count tasks for each agent
	result := make([]struct {
		AgentID domain.AgentID
		Count   int
	}, 0, len(agents))

	for _, agent := range agents {
		count := 0
		searchTag := s.getSearchTag(agent)
		if searchTag == "" {
			continue
		}

		// Remove + prefix for comparison
		cleanTag := strings.TrimPrefix(searchTag, "+")

		for _, task := range allTasks {
			for _, tag := range task.Tags {
				if tag == cleanTag {
					count++
					break
				}
			}
		}

		result = append(result, struct {
			AgentID domain.AgentID
			Count   int
		}{
			AgentID: agent.ID,
			Count:   count,
		})
	}

	return result, nil
}

// LoadByAgent returns tasks for a specific agent with pagination
func (s *taskStore) LoadByAgent(ctx context.Context, agentID domain.AgentID, limit, offset int) ([]domain.Task, error) {
	// For now, we need to get the full agent to determine the search tag
	// In a real implementation, you'd pass the agent or have a way to lookup
	// For this example, we'll export all and filter
	cmd := exec.CommandContext(ctx, "task", "export")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start task export: %w", err)
	}

	dec := json.NewDecoder(stdout)
	var tasks []domain.Task
	skipped := 0

	// We need to know if this is a subagent or parent agent
	// For now, we'll look for both patterns
	searchTags := []string{
		fmt.Sprintf("subagent_%s", strings.ReplaceAll(string(agentID), "-", "_")),
		fmt.Sprintf("session_%s", strings.ReplaceAll(string(agentID), "-", "_")),
	}

	for dec.More() {
		var t twTask
		if err := dec.Decode(&t); err != nil {
			continue
		}

		// Skip non-pending tasks
		if t.Status != "pending" && t.Status != "deleted" {
			continue
		}

		// Check if task belongs to this agent
		belongs := false
		for _, searchTag := range searchTags {
			for _, tag := range t.Tags {
				if tag == searchTag {
					belongs = true
					break
				}
			}
			if belongs {
				break
			}
		}

		if !belongs {
			continue
		}

		// Handle offset
		if offset > 0 && skipped < offset {
			skipped++
			continue
		}

		tasks = append(tasks, s.toDomainTask(t))

		// Handle limit
		if limit > 0 && len(tasks) >= limit {
			break
		}
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("task export wait: %w", err)
	}

	return tasks, nil
}

// LoadRecentByAgent returns the most recent tasks for an agent
func (s *taskStore) LoadRecentByAgent(ctx context.Context, agentID domain.AgentID, limit int) ([]domain.Task, error) {
	tasks, err := s.LoadByAgent(ctx, agentID, 0, 0) // Get all tasks first
	if err != nil {
		return nil, err
	}

	// Sort by entry date (newest first)
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Entry.After(tasks[j].Entry)
	})

	// Return only the requested limit
	if limit > 0 && len(tasks) > limit {
		return tasks[:limit], nil
	}

	return tasks, nil
}

// MarkDone marks a task as completed
func (s *taskStore) MarkDone(ctx context.Context, taskID domain.TaskID) error {
	cmd := exec.CommandContext(ctx, "task", string(taskID), "done")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mark task done: %w", err)
	}
	return nil
}

// LoadByProject loads tasks for a specific project
func (s *taskStore) LoadByProject(ctx context.Context, projectName string, limit int) ([]domain.Task, error) {
	// Use TaskWarrior to get tasks with project tag
	projectTag := fmt.Sprintf("project:%s", projectName)

	var args []string
	args = append(args, projectTag, "export")

	cmd := exec.CommandContext(ctx, "task", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("task export: %w", err)
	}

	// Parse the JSON output
	var rawTasks []twTask
	if err := json.Unmarshal(output, &rawTasks); err != nil {
		return nil, fmt.Errorf("parse task json: %w", err)
	}

	// Convert to domain tasks
	var tasks []domain.Task
	for _, raw := range rawTasks {
		task := s.toDomainTask(raw)
		tasks = append(tasks, task)

		if limit > 0 && len(tasks) >= limit {
			break
		}
	}

	return tasks, nil
}

// getSearchTag determines the appropriate tag for an agent
func (s *taskStore) getSearchTag(agent domain.Agent) string {
	if agent.ParentAgentID != "" {
		// Subagent: use +subagent_<agent_id>
		return fmt.Sprintf("+subagent_%s", strings.ReplaceAll(string(agent.ID), "-", "_"))
	} else if agent.SessionID != "" {
		// Parent agent: use +session_<session_id>
		return fmt.Sprintf("+session_%s", strings.ReplaceAll(agent.SessionID, "-", "_"))
	}
	return ""
}

// toDomainTask converts TaskWarrior task to domain task
func (s *taskStore) toDomainTask(t twTask) domain.Task {
	task := domain.Task{
		ID:          domain.TaskID(t.UUID),
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		Project:     t.Project,
		Tags:        t.Tags,
	}

	// Parse dates
	if t.Due != "" {
		if due, err := time.Parse(time.RFC3339, t.Due); err == nil {
			task.Due = due
		}
	}
	if t.Entry != "" {
		if entry, err := time.Parse(time.RFC3339, t.Entry); err == nil {
			task.Entry = entry
		}
	}

	// Convert annotations
	for _, ann := range t.Annotations {
		annotation := domain.TaskAnnotation{
			Description: ann.Description,
		}
		if entry, err := time.Parse(time.RFC3339, ann.Entry); err == nil {
			annotation.Entry = entry
		}
		task.Annotations = append(task.Annotations, annotation)
	}

	// Extract workflow status from tags
	task.WorkflowStatus = s.extractWorkflowStatus(t.Tags)

	return task
}

// extractWorkflowStatus finds workflow status tags
func (s *taskStore) extractWorkflowStatus(tags []string) string {
	workflowStates := []string{
		"ready", "working", "testing", "debugging", "review",
		"revision", "qa", "blocked", "waiting", "hold",
		"merged", "deployed", "verified",
	}

	for _, tag := range tags {
		for _, state := range workflowStates {
			if tag == state {
				return state
			}
		}
	}

	return ""
}