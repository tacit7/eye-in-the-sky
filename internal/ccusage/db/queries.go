package db

import (
	"fmt"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/ccusage/models"
)

// GetDailyUsage retrieves daily usage data within a time range
func (c *CCUsageDB) GetDailyUsage(since, until time.Time, project string) ([]models.DailyUsage, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	query := `
		SELECT DATE(timestamp) as date, project,
		       SUM(input_tokens) as input_tokens,
		       SUM(output_tokens) as output_tokens,
		       SUM(cache_creation_tokens) as cache_creation_tokens,
		       SUM(cache_read_tokens) as cache_read_tokens,
		       SUM(total_cost) as total_cost,
		       GROUP_CONCAT(DISTINCT model) as models_used
		FROM usage_entries
		WHERE timestamp >= ? AND timestamp < ?
	`

	args := []interface{}{since.Format(time.RFC3339), until.Format(time.RFC3339)}

	if project != "" {
		query += " AND project = ?"
		args = append(args, project)
	}

	query += ` GROUP BY DATE(timestamp), project ORDER BY date DESC`

	rows, err := c.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily usage: %w", err)
	}
	defer rows.Close()

	var results []models.DailyUsage
	for rows.Next() {
		var daily models.DailyUsage
		var modelsStr string

		err := rows.Scan(
			&daily.Date,
			&daily.Project,
			&daily.InputTokens,
			&daily.OutputTokens,
			&daily.CacheCreationTokens,
			&daily.CacheReadTokens,
			&daily.TotalCost,
			&modelsStr,
		)
		if err != nil {
			continue // Skip on error
		}

		results = append(results, daily)
	}

	return results, rows.Err()
}

// GetSessionData retrieves usage data grouped by session
func (c *CCUsageDB) GetSessionData(project string, limit int) ([]models.SessionData, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	query := `
		SELECT session_id, project,
		       MIN(timestamp) as start_time,
		       MAX(timestamp) as end_time,
		       SUM(input_tokens) as input_tokens,
		       SUM(output_tokens) as output_tokens,
		       SUM(cache_creation_tokens + cache_read_tokens) as cache_tokens,
		       SUM(total_cost) as total_cost
		FROM usage_entries
	`

	args := []interface{}{}

	if project != "" {
		query += " WHERE project = ?"
		args = append(args, project)
	}

	query += ` GROUP BY session_id ORDER BY start_time DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := c.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query session data: %w", err)
	}
	defer rows.Close()

	var results []models.SessionData
	for rows.Next() {
		var session models.SessionData
		var startStr, endStr string

		err := rows.Scan(
			&session.SessionID,
			&session.Project,
			&startStr,
			&endStr,
			&session.InputTokens,
			&session.OutputTokens,
			&session.CacheTokens,
			&session.TotalCost,
		)
		if err != nil {
			continue
		}

		session.StartTime, _ = time.Parse(time.RFC3339, startStr)
		session.EndTime, _ = time.Parse(time.RFC3339, endStr)
		session.ModelsUsed = make(map[string]int)

		results = append(results, session)
	}

	return results, rows.Err()
}

// GetTotalCost retrieves total cost within a time range
func (c *CCUsageDB) GetTotalCost(since, until time.Time, project string) (float64, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	query := `SELECT SUM(total_cost) FROM usage_entries WHERE timestamp >= ? AND timestamp < ?`
	args := []interface{}{since.Format(time.RFC3339), until.Format(time.RFC3339)}

	if project != "" {
		query += " AND project = ?"
		args = append(args, project)
	}

	var cost float64
	err := c.db.QueryRow(query, args...).Scan(&cost)
	if err != nil {
		return 0, fmt.Errorf("failed to query total cost: %w", err)
	}

	return cost, nil
}

// GetActiveBlock retrieves the current active billing block
func (c *CCUsageDB) GetActiveBlock() (*models.BlockUsage, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Calculate the current active 5-hour block
	now := time.Now()
	blockStart := getBlockStart(now)
	blockEnd := blockStart.Add(5 * time.Hour)

	query := `
		SELECT MIN(timestamp) as start_time,
		       MAX(timestamp) as end_time,
		       SUM(input_tokens) as input_tokens,
		       SUM(output_tokens) as output_tokens,
		       SUM(cache_creation_tokens + cache_read_tokens) as cache_tokens,
		       SUM(total_cost) as total_cost,
		       project
		FROM usage_entries
		WHERE timestamp >= ? AND timestamp < ?
		GROUP BY project
	`

	rows, err := c.db.Query(query,
		blockStart.Format(time.RFC3339),
		blockEnd.Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query active block: %w", err)
	}
	defer rows.Close()

	block := &models.BlockUsage{
		StartTime:        blockStart,
		EndTime:          blockEnd,
		IsActive:         true,
		ProjectBreakdown: make(map[string]models.BlockProjectData),
	}

	for rows.Next() {
		var startStr, endStr, project string
		var inputTokens, outputTokens, cacheTokens int
		var totalCost float64

		err := rows.Scan(&startStr, &endStr, &inputTokens, &outputTokens, &cacheTokens, &totalCost, &project)
		if err != nil {
			continue
		}

		block.InputTokens += inputTokens
		block.OutputTokens += outputTokens
		block.CacheTokens += cacheTokens
		block.TotalCost += totalCost

		block.ProjectBreakdown[project] = models.BlockProjectData{
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
			CacheTokens:  cacheTokens,
			TotalCost:    totalCost,
		}
	}

	return block, rows.Err()
}

// getBlockStart calculates the start time of the current 5-hour block
func getBlockStart(t time.Time) time.Time {
	// Blocks are 5 hours starting from 00:00, 05:00, 10:00, etc.
	hour := t.Hour()
	blockHour := (hour / 5) * 5
	return time.Date(t.Year(), t.Month(), t.Day(), blockHour, 0, 0, 0, t.Location())
}

// GetMonthlyUsage retrieves monthly usage data
func (c *CCUsageDB) GetMonthlyUsage(year, month int, project string) (models.DailyUsage, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	query := `
		SELECT SUM(input_tokens) as input_tokens,
		       SUM(output_tokens) as output_tokens,
		       SUM(cache_creation_tokens) as cache_creation_tokens,
		       SUM(cache_read_tokens) as cache_read_tokens,
		       SUM(total_cost) as total_cost
		FROM usage_entries
		WHERE timestamp >= ? AND timestamp < ?
	`

	args := []interface{}{startDate.Format(time.RFC3339), endDate.Format(time.RFC3339)}

	if project != "" {
		query += " AND project = ?"
		args = append(args, project)
	}

	var usage models.DailyUsage
	usage.Date = startDate.Format("2006-01")
	usage.Project = project

	err := c.db.QueryRow(query, args...).Scan(
		&usage.InputTokens,
		&usage.OutputTokens,
		&usage.CacheCreationTokens,
		&usage.CacheReadTokens,
		&usage.TotalCost,
	)

	if err != nil {
		return models.DailyUsage{}, fmt.Errorf("failed to query monthly usage: %w", err)
	}

	return usage, nil
}

// GetEntryCount returns the number of entries in the database
func (c *CCUsageDB) GetEntryCount() (int, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var count int
	err := c.db.QueryRow("SELECT COUNT(*) FROM usage_entries").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count entries: %w", err)
	}

	return count, nil
}

// GetDistinctMonths retrieves all distinct year-month combinations with data, ordered reverse chronologically
func (c *CCUsageDB) GetDistinctMonths() ([]string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	query := `
		SELECT DISTINCT strftime('%Y-%m', timestamp) as month
		FROM usage_entries
		ORDER BY month DESC
	`

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query distinct months: %w", err)
	}
	defer rows.Close()

	var months []string
	for rows.Next() {
		var month string
		err := rows.Scan(&month)
		if err != nil {
			continue
		}
		months = append(months, month)
	}

	return months, rows.Err()
}
