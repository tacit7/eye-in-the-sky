package domain

import "time"

// CommitHash represents a git commit hash
type CommitHash string

// Commit represents a git commit
type Commit struct {
	ID        int
	AgentID   AgentID
	Hash      CommitHash
	Message   string
	Timestamp time.Time
}