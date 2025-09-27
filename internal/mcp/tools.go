package mcp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/database"
)

// Tools implements the MCP tool handlers
type Tools struct {
	db *database.DB
}

// NewTools creates a new Tools instance
func NewTools(db *database.DB) *Tools {
	return &Tools{db: db}
}

// RegisterAgent implements the register_agent MCP tool
func (t *Tools) RegisterAgent(args RegisterAgentArgs) (RegisterAgentResult, error) {
	// Validate agent ID (should be 8 characters)
	if len(args.AgentID) != 8 {
		return RegisterAgentResult{
			Success: false,
			Message: "Agent ID must be exactly 8 characters",
		}, nil
	}

	// Check if agent already exists
	existing, err := t.db.GetAgent(args.AgentID)
	if err == nil && existing != nil {
		return RegisterAgentResult{
			Success: false,
			Message: fmt.Sprintf("Agent %s already exists", args.AgentID),
		}, nil
	}

	// Create new agent
	agent := &database.Agent{
		ID:                 args.AgentID,
		Status:             database.StatusActive,
		GitWorktreePath:    args.WorktreePath,
		FeatureDescription: &args.Description,
		LastActivityAt:     timePtr(time.Now()),
	}

	if err := t.db.CreateAgent(agent); err != nil {
		return RegisterAgentResult{
			Success: false,
			Message: fmt.Sprintf("Failed to register agent: %v", err),
		}, fmt.Errorf("database error: %w", err)
	}

	// Log registration action
	action := &database.Action{
		AgentID:     args.AgentID,
		ActionType:  database.ActionStatusUpdate,
		Description: fmt.Sprintf("Agent registered: %s", args.Description),
		Details:     nil,
	}

	if err := t.db.CreateAction(action); err != nil {
		// Don't fail the registration, just log the error
		fmt.Printf("Warning: Failed to log registration action: %v\n", err)
	}

	return RegisterAgentResult{
		Success: true,
		Message: fmt.Sprintf("Agent %s registered successfully", args.AgentID),
	}, nil
}

// UpdateStatus implements the update_status MCP tool
func (t *Tools) UpdateStatus(args UpdateStatusArgs) (UpdateStatusResult, error) {
	// Validate status
	validStatuses := map[string]bool{
		database.StatusActive:    true,
		database.StatusIdle:      true,
		database.StatusWorking:   true,
		database.StatusCompleted: true,
		database.StatusFailed:    true,
	}

	if !validStatuses[args.Status] {
		return UpdateStatusResult{
			Success: false,
			Message: fmt.Sprintf("Invalid status: %s", args.Status),
		}, nil
	}

	// Update agent status
	if err := t.db.UpdateAgentStatus(args.AgentID, args.Status, args.CurrentTask); err != nil {
		return UpdateStatusResult{
			Success: false,
			Message: fmt.Sprintf("Failed to update status: %v", err),
		}, nil
	}

	// Log status update action
	description := fmt.Sprintf("Status updated to: %s", args.Status)
	if args.CurrentTask != nil {
		description += fmt.Sprintf(" (Task: %s)", *args.CurrentTask)
	}

	action := &database.Action{
		AgentID:     args.AgentID,
		ActionType:  database.ActionStatusUpdate,
		Description: description,
		Details:     nil,
	}

	if err := t.db.CreateAction(action); err != nil {
		fmt.Printf("Warning: Failed to log status update action: %v\n", err)
	}

	return UpdateStatusResult{
		Success: true,
		Message: fmt.Sprintf("Status updated to %s", args.Status),
	}, nil
}

// LogAction implements the log_action MCP tool
func (t *Tools) LogAction(args LogActionArgs) (LogActionResult, error) {
	// Validate action type
	validActionTypes := map[string]bool{
		database.ActionTaskStart:     true,
		database.ActionFileOperation: true,
		database.ActionGitCommit:     true,
		database.ActionStatusUpdate:  true,
	}

	if !validActionTypes[args.ActionType] {
		return LogActionResult{
			Success: false,
			Message: fmt.Sprintf("Invalid action type: %s", args.ActionType),
		}, nil
	}

	// Create action
	action := &database.Action{
		AgentID:     args.AgentID,
		ActionType:  args.ActionType,
		Description: args.Description,
		Details:     args.Details,
	}

	if err := t.db.CreateAction(action); err != nil {
		return LogActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to log action: %v", err),
		}, nil
	}

	return LogActionResult{
		Success: true,
		Message: "Action logged successfully",
	}, nil
}

// LogCommits implements the log_commits MCP tool
func (t *Tools) LogCommits(args LogCommitsArgs) (LogCommitsResult, error) {
	if len(args.CommitHashes) == 0 {
		return LogCommitsResult{
			Success: false,
			Message: "No commit hashes provided",
		}, nil
	}

	// Log commits to database
	if err := t.db.CreateCommits(args.AgentID, args.CommitHashes, args.CommitMessages); err != nil {
		return LogCommitsResult{
			Success: false,
			Message: fmt.Sprintf("Failed to log commits: %v", err),
		}, nil
	}

	// Log as action as well
	description := fmt.Sprintf("Logged %d commit(s): %v", len(args.CommitHashes), args.CommitHashes)
	details := map[string]interface{}{
		"commit_hashes":   args.CommitHashes,
		"commit_messages": args.CommitMessages,
	}
	detailsJSON, _ := json.Marshal(details)
	detailsStr := string(detailsJSON)

	action := &database.Action{
		AgentID:     args.AgentID,
		ActionType:  database.ActionGitCommit,
		Description: description,
		Details:     &detailsStr,
	}

	if err := t.db.CreateAction(action); err != nil {
		fmt.Printf("Warning: Failed to log commit action: %v\n", err)
	}

	return LogCommitsResult{
		Success: true,
		Message: fmt.Sprintf("Logged %d commit(s) successfully", len(args.CommitHashes)),
	}, nil
}

// EndSession implements the end_session MCP tool
func (t *Tools) EndSession(args EndSessionArgs) (EndSessionResult, error) {
	// Default final status if not provided
	finalStatus := database.StatusCompleted
	if args.FinalStatus != nil {
		finalStatus = *args.FinalStatus
	}

	// Validate final status
	validStatuses := map[string]bool{
		database.StatusCompleted: true,
		database.StatusFailed:    true,
	}

	if !validStatuses[finalStatus] {
		return EndSessionResult{
			Success: false,
			Message: fmt.Sprintf("Invalid final status: %s", finalStatus),
		}, nil
	}

	// End session
	summary := ""
	if args.Summary != nil {
		summary = *args.Summary
	}

	if err := t.db.EndAgentSession(args.AgentID, summary, finalStatus); err != nil {
		return EndSessionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to end session: %v", err),
		}, nil
	}

	message := fmt.Sprintf("Session ended with status: %s", finalStatus)
	if summary != "" {
		message += fmt.Sprintf(" (Summary: %s)", summary)
	}

	return EndSessionResult{
		Success: true,
		Message: message,
	}, nil
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}