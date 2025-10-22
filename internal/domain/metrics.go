package domain

import "time"

// SessionMetric represents usage metrics for an agent session
type SessionMetric struct {
	ID               int
	AgentID          AgentID
	SessionID        string
	Timestamp        time.Time
	CreatedAt        time.Time
	Duration         int // seconds
	TasksCompleted   int
	FilesModified    int
	LinesAdded       int
	LinesRemoved     int
	CommitCount      int
	ErrorCount       int
	EstimatedCost    float64
	EstimatedCostUSD float64
	InputTokens      int
	OutputTokens     int
	TotalTokens      int
	TokensUsed       int
	TokensBudget     int
	TokensRemaining  int
	CacheWriteTokens int
	CacheReadTokens  int
	ModelName        string
	Notes            string
}