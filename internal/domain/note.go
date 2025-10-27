package domain

import "time"

// NoteID represents a unique note identifier
type NoteID string

// Note represents a polymorphic note
type Note struct {
	ID         NoteID
	AgentID    AgentID // For compatibility with existing code
	ParentID   string
	ParentType string // "global", "projects", "agents", "sessions"
	Body       string
	CreatedAt  time.Time

	// Legacy fields - kept for backward compatibility during migration
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}