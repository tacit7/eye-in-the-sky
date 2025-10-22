package domain

import "time"

// LogID represents a unique log identifier
type LogID int

// Log represents a session log entry
type Log struct {
	ID        LogID
	SessionID string
	Type      string
	Message   string
	Timestamp time.Time
}