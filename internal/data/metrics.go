package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/domain"
)

// metricsStore implements SQL operations for metrics
type metricsStore struct {
	db *sql.DB
}

// NewMetricsStore creates a new metrics store
func NewMetricsStore(db *sql.DB) *metricsStore {
	return &metricsStore{db: db}
}

// LoadAll loads all session metrics
func (s *metricsStore) LoadAll(ctx context.Context) ([]domain.SessionMetric, error) {
	query := `
		SELECT id, agent_id, session_id, timestamp, duration,
		       tasks_completed, files_modified, lines_added, lines_removed,
		       commit_count, error_count, estimated_cost,
		       input_tokens, output_tokens, total_tokens,
		       cache_write_tokens, cache_read_tokens
		FROM session_metrics
		ORDER BY timestamp DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query metrics: %w", err)
	}
	defer rows.Close()

	var metrics []domain.SessionMetric
	for rows.Next() {
		var m domain.SessionMetric
		err := rows.Scan(
			&m.ID, &m.AgentID, &m.SessionID, &m.Timestamp, &m.Duration,
			&m.TasksCompleted, &m.FilesModified, &m.LinesAdded, &m.LinesRemoved,
			&m.CommitCount, &m.ErrorCount, &m.EstimatedCost,
			&m.InputTokens, &m.OutputTokens, &m.TotalTokens,
			&m.CacheWriteTokens, &m.CacheReadTokens,
		)
		if err != nil {
			return nil, fmt.Errorf("scan metric: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}

// LoadByAgent loads metrics for a specific agent
func (s *metricsStore) LoadByAgent(ctx context.Context, agentID domain.AgentID, limit int) ([]domain.SessionMetric, error) {
	query := `
		SELECT id, agent_id, session_id, timestamp, duration,
		       tasks_completed, files_modified, lines_added, lines_removed,
		       commit_count, error_count, estimated_cost,
		       input_tokens, output_tokens, total_tokens,
		       cache_write_tokens, cache_read_tokens
		FROM session_metrics
		WHERE agent_id = ?
		ORDER BY timestamp DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.QueryContext(ctx, query, string(agentID))
	if err != nil {
		return nil, fmt.Errorf("query metrics: %w", err)
	}
	defer rows.Close()

	var metrics []domain.SessionMetric
	for rows.Next() {
		var m domain.SessionMetric
		err := rows.Scan(
			&m.ID, &m.AgentID, &m.SessionID, &m.Timestamp, &m.Duration,
			&m.TasksCompleted, &m.FilesModified, &m.LinesAdded, &m.LinesRemoved,
			&m.CommitCount, &m.ErrorCount, &m.EstimatedCost,
			&m.InputTokens, &m.OutputTokens, &m.TotalTokens,
			&m.CacheWriteTokens, &m.CacheReadTokens,
		)
		if err != nil {
			return nil, fmt.Errorf("scan metric: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}

// LoadByTimeRange loads metrics within a time range
func (s *metricsStore) LoadByTimeRange(ctx context.Context, start, end time.Time) ([]domain.SessionMetric, error) {
	query := `
		SELECT id, agent_id, session_id, timestamp, duration,
		       tasks_completed, files_modified, lines_added, lines_removed,
		       commit_count, error_count, estimated_cost,
		       input_tokens, output_tokens, total_tokens,
		       cache_write_tokens, cache_read_tokens
		FROM session_metrics
		WHERE timestamp BETWEEN ? AND ?
		ORDER BY timestamp DESC
	`

	rows, err := s.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("query metrics: %w", err)
	}
	defer rows.Close()

	var metrics []domain.SessionMetric
	for rows.Next() {
		var m domain.SessionMetric
		err := rows.Scan(
			&m.ID, &m.AgentID, &m.SessionID, &m.Timestamp, &m.Duration,
			&m.TasksCompleted, &m.FilesModified, &m.LinesAdded, &m.LinesRemoved,
			&m.CommitCount, &m.ErrorCount, &m.EstimatedCost,
			&m.InputTokens, &m.OutputTokens, &m.TotalTokens,
			&m.CacheWriteTokens, &m.CacheReadTokens,
		)
		if err != nil {
			return nil, fmt.Errorf("scan metric: %w", err)
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}

// LoadMonthly loads metrics for a specific month
func (s *metricsStore) LoadMonthly(ctx context.Context, year int, month time.Month) ([]domain.SessionMetric, error) {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

	return s.LoadByTimeRange(ctx, start, end)
}