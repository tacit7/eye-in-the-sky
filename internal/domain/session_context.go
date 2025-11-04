package domain

import (
	"time"
)

// SessionContext represents saved context for a session
type SessionContext struct {
	ID              int
	AgentID         AgentID
	SessionID       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CurrentPhase    string
	OverallProgress float32
	PendingTasks    string // JSON string
	CompletedTasks  string // JSON string
	NextActions     string // JSON string
	Dependencies    string // JSON string
	ImportantFiles  string // JSON string
	Milestones      string // JSON string
	CurrentGoals    string // JSON string
	Blockers        string // JSON string
	KeyDecisions    string // JSON string
	Environment     string // JSON string
	Metrics         string // JSON string
	AutoSave        bool
	LearnedContext  string
}
