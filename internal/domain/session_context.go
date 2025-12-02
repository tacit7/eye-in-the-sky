package domain

import (
	"time"
)

// SessionContext represents saved context for a session (simplified schema after migration 002)
type SessionContext struct {
	ID        int
	AgentID   AgentID
	SessionID string
	Context   string    // Markdown-formatted context
	CreatedAt time.Time
	UpdatedAt time.Time
}
