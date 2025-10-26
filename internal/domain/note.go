package domain

import "time"

// NoteID represents a unique note identifier
type NoteID int

// Note represents a session note
type Note struct {
	ID        NoteID
	AgentID   AgentID
	Title     string
	Content   string
	CreatedAt time.Time
}