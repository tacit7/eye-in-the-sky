package models

import (
	"time"
)

// UsageEntry represents a single Claude Code usage entry from JSONL
type UsageEntry struct {
	Cwd                string    `json:"cwd"`
	SessionID          string    `json:"sessionId"`
	Timestamp          time.Time `json:"timestamp"`
	Version            string    `json:"version"`
	Message            Message   `json:"message"`
	CostUSD            float64   `json:"costUSD"`
	RequestID          string    `json:"requestId"`
	IsApiErrorMessage  bool      `json:"isApiErrorMessage"`
}

// Message contains the usage data and model information
type Message struct {
	Usage  UsageMetrics `json:"usage"`
	Model  string       `json:"model"`
	ID     string       `json:"id"`
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

// UsageMetrics contains token counts and cache information
type UsageMetrics struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

// DailyUsage aggregates usage data for a single day
type DailyUsage struct {
	Date                   string
	Project                string
	InputTokens            int
	OutputTokens           int
	CacheCreationTokens    int
	CacheReadTokens        int
	TotalCost              float64
	ModelsUsed             []string
	ModelBreakdowns        map[string]ModelBreakdown
}

// ModelBreakdown contains token and cost breakdown for a specific model
type ModelBreakdown struct {
	InputTokens             int
	OutputTokens            int
	CacheCreationTokens     int
	CacheReadTokens         int
	TotalCost               float64
}

// SessionData aggregates usage for a specific session
type SessionData struct {
	SessionID      string
	Project        string
	StartTime      time.Time
	EndTime        time.Time
	InputTokens    int
	OutputTokens   int
	CacheTokens    int
	TotalCost      float64
	ModelsUsed     map[string]int // model -> count
}

// BlockUsage represents a 5-hour billing block
type BlockUsage struct {
	StartTime       time.Time
	EndTime         time.Time
	IsActive        bool
	InputTokens     int
	OutputTokens    int
	CacheTokens     int
	TotalCost       float64
	ProjectBreakdown map[string]BlockProjectData
}

// BlockProjectData contains breakdown by project within a block
type BlockProjectData struct {
	InputTokens  int
	OutputTokens int
	CacheTokens  int
	TotalCost    float64
}
