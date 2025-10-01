package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/database"
)

// SessionContext represents the complete state of an agent session
type SessionContext struct {
	AgentID           string                 `json:"agent_id"`
	SessionID         string                 `json:"session_id"`
	StartTime         time.Time              `json:"start_time"`
	LastCheckpoint    time.Time              `json:"last_checkpoint"`
	CurrentPhase      string                 `json:"current_phase"`
	Status            string                 `json:"status"`
	Description       string                 `json:"description"`
	ProjectName       string                 `json:"project_name,omitempty"`
	WorktreePath      string                 `json:"worktree_path,omitempty"`
	WindowID          string                 `json:"window_id,omitempty"`
	Progress          SessionProgress        `json:"progress"`
	Environment       map[string]interface{} `json:"environment"`
	NextActions       []string               `json:"next_actions"`
	CompletedTasks    []string               `json:"completed_tasks"`
	PendingTasks      []string               `json:"pending_tasks"`
	KeyDecisions      []SessionDecision      `json:"key_decisions"`
	ImportantFiles    []string               `json:"important_files"`
	Dependencies      []string               `json:"dependencies"`
	Notes             []SessionContextNote          `json:"notes"`
	Metrics           SessionMetrics         `json:"metrics"`
}

// SessionProgress tracks completion and milestones
type SessionProgress struct {
	OverallCompletion float32              `json:"overall_completion"`
	Milestones        []SessionMilestone   `json:"milestones"`
	CurrentGoals      []string             `json:"current_goals"`
	Blockers          []SessionBlocker     `json:"blockers"`
}

// SessionMilestone represents a significant achievement
type SessionMilestone struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CompletedAt time.Time `json:"completed_at"`
	Impact      string    `json:"impact"`
}

// SessionBlocker represents an obstacle or dependency
type SessionBlocker struct {
	Issue       string    `json:"issue"`
	Severity    string    `json:"severity"` // low, medium, high, critical
	CreatedAt   time.Time `json:"created_at"`
	Resolution  string    `json:"resolution,omitempty"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// SessionDecision captures important choices made during development
type SessionDecision struct {
	Decision    string    `json:"decision"`
	Rationale   string    `json:"rationale"`
	Timestamp   time.Time `json:"timestamp"`
	Impact      string    `json:"impact"`
	Reversible  bool      `json:"reversible"`
	Alternatives []string `json:"alternatives,omitempty"`
}

// SessionContextNote represents timestamped observations or thoughts for session context
type SessionContextNote struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // insight, reminder, warning, idea
	Content   string    `json:"content"`
	Priority  string    `json:"priority"` // low, medium, high
	Tags      []string  `json:"tags,omitempty"`
}

// SessionMetrics tracks quantitative progress
type SessionMetrics struct {
	FilesModified     int               `json:"files_modified"`
	LinesAdded        int               `json:"lines_added"`
	LinesRemoved      int               `json:"lines_removed"`
	CommitsCreated    int               `json:"commits_created"`
	TestsWritten      int               `json:"tests_written"`
	BugsFixed         int               `json:"bugs_fixed"`
	FeaturesAdded     int               `json:"features_added"`
	DocumentationDocs int               `json:"documentation_docs"`
	TimeSpent         time.Duration     `json:"time_spent"`
	CustomMetrics     map[string]interface{} `json:"custom_metrics"`
}

// SaveSessionContextArgs for saving session state
type SaveSessionContextArgs struct {
	AgentID        string                 `json:"agent_id"`
	CurrentPhase   string                 `json:"current_phase"`
	LearnedContext *string                `json:"learned_context,omitempty"` // Agent's accumulated knowledge/expertise
	Progress       *SessionProgress       `json:"progress,omitempty"`
	NextActions    []string               `json:"next_actions,omitempty"`
	CompletedTasks []string               `json:"completed_tasks,omitempty"`
	PendingTasks   []string               `json:"pending_tasks,omitempty"`
	KeyDecisions   []SessionDecision      `json:"key_decisions,omitempty"`
	ImportantFiles []string               `json:"important_files,omitempty"`
	Dependencies   []string               `json:"dependencies,omitempty"`
	Notes          []SessionContextNote          `json:"notes,omitempty"`
	Environment    map[string]interface{} `json:"environment,omitempty"`
	Metrics        *SessionMetrics        `json:"metrics,omitempty"`
	AutoSave       bool                   `json:"auto_save,omitempty"`
}

// SaveSessionContextResult response
type SaveSessionContextResult struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	SessionID  string `json:"session_id"`
	Checkpoint string `json:"checkpoint"`
}

// LoadSessionContextArgs for loading session state
type LoadSessionContextArgs struct {
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id,omitempty"` // If empty, loads latest
}

// LoadSessionContextResult response
type LoadSessionContextResult struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Context *SessionContext `json:"context,omitempty"`
}

// AddSessionNoteArgs for adding contextual notes
type AddSessionNoteArgs struct {
	AgentID  string   `json:"agent_id"`
	Type     string   `json:"type"`
	Content  string   `json:"content"`
	Priority string   `json:"priority,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

// AddSessionNoteResult response
type AddSessionNoteResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	NoteID  string `json:"note_id"`
}

// SaveSessionContext saves the current session state for resumption
func (t *Tools) SaveSessionContext(args SaveSessionContextArgs) (SaveSessionContextResult, error) {
	// Validate agent exists
	agent, err := t.db.GetAgent(args.AgentID)
	if err != nil {
		return SaveSessionContextResult{
			Success: false,
			Message: fmt.Sprintf("Agent not found: %s", args.AgentID),
		}, nil
	}

	// Generate session ID and checkpoint
	sessionID := fmt.Sprintf("%s_%d", args.AgentID, time.Now().Unix())
	checkpoint := time.Now().Format("2006-01-02_15:04:05")

	// Load existing context or create new
	var context *SessionContext
	existing, err := t.loadLatestSessionContext(args.AgentID)
	if err == nil && existing != nil {
		context = existing
	} else {
		// Create new session context
		context = &SessionContext{
			AgentID:     args.AgentID,
			StartTime:   time.Now(),
			Status:      agent.Status,
			Description: stringValue(agent.FeatureDescription),
			ProjectName: "", // Will be set based on agent type
			WorktreePath: stringValue(agent.GitWorktreePath),
			WindowID:     stringValue(agent.WindowID),
			Environment:    make(map[string]interface{}),
			NextActions:    []string{},
			CompletedTasks: []string{},
			PendingTasks:   []string{},
			KeyDecisions:   []SessionDecision{},
			ImportantFiles: []string{},
			Dependencies:   []string{},
			Notes:          []SessionContextNote{},
			Progress: SessionProgress{
				OverallCompletion: 0.0,
				Milestones:        []SessionMilestone{},
				CurrentGoals:      []string{},
				Blockers:          []SessionBlocker{},
			},
			Metrics: SessionMetrics{
				CustomMetrics: make(map[string]interface{}),
			},
		}
	}

	// Update context with new information
	context.SessionID = sessionID
	context.LastCheckpoint = time.Now()
	context.CurrentPhase = args.CurrentPhase

	if args.Progress != nil {
		context.Progress = *args.Progress
	}
	if args.NextActions != nil {
		context.NextActions = args.NextActions
	}
	if args.CompletedTasks != nil {
		context.CompletedTasks = args.CompletedTasks
	}
	if args.PendingTasks != nil {
		context.PendingTasks = args.PendingTasks
	}
	if args.KeyDecisions != nil {
		context.KeyDecisions = append(context.KeyDecisions, args.KeyDecisions...)
	}
	if args.ImportantFiles != nil {
		context.ImportantFiles = mergeStringSlices(context.ImportantFiles, args.ImportantFiles)
	}
	if args.Dependencies != nil {
		context.Dependencies = mergeStringSlices(context.Dependencies, args.Dependencies)
	}
	if args.Notes != nil {
		context.Notes = append(context.Notes, args.Notes...)
	}
	if args.Environment != nil {
		for k, v := range args.Environment {
			context.Environment[k] = v
		}
	}
	if args.Metrics != nil {
		context.Metrics = *args.Metrics
	}

	// Calculate time spent
	context.Metrics.TimeSpent = time.Since(context.StartTime)

	// Serialize context to JSON
	contextJSON, err := json.Marshal(context)
	if err != nil {
		return SaveSessionContextResult{
			Success: false,
			Message: fmt.Sprintf("Failed to serialize context: %v", err),
		}, nil
	}

	// Save to database
	action := &database.Action{
		AgentID:     args.AgentID,
		ActionType:  "session_checkpoint",
		Description: fmt.Sprintf("Session context saved - Phase: %s", args.CurrentPhase),
		Details:     stringPtr(string(contextJSON)),
	}

	err = t.db.CreateAction(action)
	if err != nil {
		return SaveSessionContextResult{
			Success: false,
			Message: fmt.Sprintf("Failed to save context: %v", err),
		}, nil
	}

	// Update agent status if auto-save
	if args.AutoSave {
		err = t.db.UpdateAgentStatus(args.AgentID, "idle",
			stringPtr(fmt.Sprintf("Session paused at: %s", args.CurrentPhase)))
		if err != nil {
			// Log warning but don't fail
			fmt.Fprintf(os.Stderr, "Warning: Failed to update agent status during auto-save: %v\n", err)
		}
	}

	return SaveSessionContextResult{
		Success:    true,
		Message:    fmt.Sprintf("Session context saved successfully for phase: %s", args.CurrentPhase),
		SessionID:  sessionID,
		Checkpoint: checkpoint,
	}, nil
}

// LoadSessionContext loads the most recent session state for resumption
func (t *Tools) LoadSessionContext(args LoadSessionContextArgs) (LoadSessionContextResult, error) {
	context, err := t.loadLatestSessionContext(args.AgentID)
	if err != nil {
		return LoadSessionContextResult{
			Success: false,
			Message: fmt.Sprintf("Failed to load session context: %v", err),
		}, nil
	}

	if context == nil {
		return LoadSessionContextResult{
			Success: false,
			Message: "No session context found for agent",
		}, nil
	}

	return LoadSessionContextResult{
		Success: true,
		Message: fmt.Sprintf("Session context loaded - Last checkpoint: %s",
			context.LastCheckpoint.Format("2006-01-02 15:04:05")),
		Context: context,
	}, nil
}

// AddSessionNote adds a contextual note to the current session
func (t *Tools) AddSessionNote(args AddSessionNoteArgs) (AddSessionNoteResult, error) {
	note := SessionContextNote{
		Timestamp: time.Now(),
		Type:      args.Type,
		Content:   args.Content,
		Priority:  args.Priority,
		Tags:      args.Tags,
	}

	if note.Priority == "" {
		note.Priority = "medium"
	}

	// Load existing context
	context, err := t.loadLatestSessionContext(args.AgentID)
	if context == nil {
		return AddSessionNoteResult{
			Success: false,
			Message: "No active session context found",
		}, nil
	}

	// Add note to context
	context.Notes = append(context.Notes, note)
	context.LastCheckpoint = time.Now()

	// Save updated context
	saveArgs := SaveSessionContextArgs{
		AgentID:     args.AgentID,
		CurrentPhase: context.CurrentPhase,
		Notes:       []SessionContextNote{note},
		AutoSave:    true,
	}

	result, err := t.SaveSessionContext(saveArgs)
	if err != nil || !result.Success {
		return AddSessionNoteResult{
			Success: false,
			Message: "Failed to save note to session context",
		}, nil
	}

	noteID := fmt.Sprintf("%s_%d", args.AgentID, time.Now().UnixNano())
	return AddSessionNoteResult{
		Success: true,
		Message: fmt.Sprintf("Note added to session context: %s", args.Type),
		NoteID:  noteID,
	}, nil
}

// Helper function to load latest session context from database
func (t *Tools) loadLatestSessionContext(agentID string) (*SessionContext, error) {
	actions, err := t.db.GetActionsForAgent(agentID, 50)
	if err != nil {
		return nil, err
	}

	// Find most recent session checkpoint
	for _, action := range actions {
		if action.ActionType == "session_checkpoint" && action.Details != nil {
			var context SessionContext
			err := json.Unmarshal([]byte(*action.Details), &context)
			if err != nil {
				continue // Skip malformed contexts
			}
			return &context, nil
		}
	}

	return nil, nil // No context found
}

// Helper function to merge string slices without duplicates
func mergeStringSlices(existing, new []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(existing)+len(new))

	// Add existing items
	for _, item := range existing {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	// Add new items
	for _, item := range new {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// Helper function to safely get string value from pointer
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}