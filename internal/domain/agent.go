package domain

import "time"

// AgentID is a unique identifier for an agent (8-character hash)
type AgentID string

// Agent represents an Eye in the Sky agent
type Agent struct {
	ID                  AgentID
	Status              string    // active, working, idle, completed, failed
	Source              string    // worktree or desktop
	CreatedAt           time.Time
	UpdatedAt           time.Time
	GitWorktreePath     string
	FeatureDesc         string
	CurrentTask         string
	LastActivityAt      time.Time
	WindowID            string
	TerminalApplication string
	AgentDescription    string
	ProjectName         string
	SessionID           string
	ParentSessionID     string
	ParentAgentID       string
	CompletedAt         *time.Time
	TaskCount           int    // Number of tasks from TaskWarrior
	Bookmarked          bool   // Pinned/bookmarked agents appear at top
}