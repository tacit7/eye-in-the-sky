package app

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/api"
)

// renderUsageTab renders session costs for all agents with monthly breakdown
func (m *Model) renderUsageTab() string {
	var b strings.Builder

	// Show Claude Code empty state if database is empty
	if m.ccusageDB != nil && m.ccusageEntryCount == 0 {
		b.WriteString(m.styles.Title.Render("Claude Code Usage"))
		b.WriteString("\n\n")

		if m.ccusageSyncing {
			b.WriteString(m.styles.Working.Render("  ⏳ " + m.ccusageSyncStatus))
		} else {
			b.WriteString(m.styles.Subtle.Render("  No Claude Code usage data found"))
			b.WriteString("\n\n")
			b.WriteString(m.styles.Primary.Render("  Press 'I' to initialize database"))
			b.WriteString("\n")
			b.WriteString(m.styles.Subtle.Render("  This will discover and parse JSONL files from:"))
			b.WriteString("\n")
			b.WriteString(m.styles.Subtle.Render("  ~/.config/claude/projects/"))
			b.WriteString("\n")
			b.WriteString(m.styles.Subtle.Render("  ~/.claude/projects/"))
		}

		b.WriteString("\n\n")
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", 70)))
		b.WriteString("\n\n")
	}

	if len(m.allSessionMetrics) == 0 && len(m.monthlyCosts) == 0 && m.ccusageEntryCount == 0 {
		return m.styles.Subtle.Render("  No usage data available")
	}

	// Current Session Summary
	if len(m.allSessionMetrics) > 0 {
		b.WriteString(m.styles.Title.Render("Eye-in-the-Sky Session Metrics"))
		b.WriteString("\n\n")

		// Column headers
		headerLine := fmt.Sprintf("%-10s %-8s %-12s %-10s %-12s",
			"Agent", "Usage %", "Tokens", "Cost USD", "Model")
		b.WriteString(m.styles.Primary.Render(headerLine))
		b.WriteString("\n")
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", 70)))
		b.WriteString("\n")

		// Data rows
		totalCost := 0.0
		totalTokens := 0

		for _, metric := range m.allSessionMetrics {
			usagePercent := float64(metric.TokensUsed) / float64(metric.TokensBudget) * 100

			agentID := truncate(string(metric.AgentID), 8)
			usage := fmt.Sprintf("%.1f%%", usagePercent)
			tokens := fmt.Sprintf("%d", metric.TokensUsed)
			cost := fmt.Sprintf("$%.4f", metric.EstimatedCostUSD)
			model := truncate(metric.ModelName, 12)

			line := fmt.Sprintf("%-10s %-8s %-12s %-10s %-12s",
				agentID, usage, tokens, cost, model)
			b.WriteString(m.styles.Text.Render(line))
			b.WriteString("\n")

			totalCost += metric.EstimatedCostUSD
			totalTokens += metric.TokensUsed
		}

		b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("Total: $%.4f | %d tokens\n", totalCost, totalTokens)))
	}

	// Monthly Cost Breakdown
	if len(m.monthlyCosts) > 0 {
		b.WriteString("\n\n")
		b.WriteString(m.styles.Title.Render("Monthly Cost Breakdown"))
		b.WriteString("\n\n")

		// Column headers
		headerLine := fmt.Sprintf("%-20s %-8s %-12s %-10s %-8s",
			"Timestamp", "Usage %", "Tokens Used", "Cost USD", "Model")
		b.WriteString(m.styles.Primary.Render(headerLine))
		b.WriteString("\n")
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", 70)))
		b.WriteString("\n")

		// Data rows
		totalMonthlyCost := 0.0
		totalMonthlyTokens := 0

		for _, metric := range m.monthlyCosts {
			usagePercent := float64(metric.TokensUsed) / float64(metric.TokensBudget) * 100

			timestamp := metric.Timestamp.Format("2006-01-02 15:04")
			usage := fmt.Sprintf("%.1f%%", usagePercent)
			tokens := fmt.Sprintf("%d", metric.TokensUsed)
			cost := fmt.Sprintf("$%.4f", metric.EstimatedCostUSD)
			model := truncate(metric.ModelName, 8)

			line := fmt.Sprintf("%-20s %-8s %-12s %-10s %-8s",
				timestamp, usage, tokens, cost, model)
			b.WriteString(m.styles.Text.Render(line))
			b.WriteString("\n")

			totalMonthlyCost += metric.EstimatedCostUSD
			totalMonthlyTokens += metric.TokensUsed
		}

		b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("Monthly Total: $%.4f | %d tokens across %d snapshots\n",
			totalMonthlyCost, totalMonthlyTokens, len(m.monthlyCosts))))
	}

	// Claude Code metrics (if available)
	if m.ccusageDB != nil && m.ccusageEntryCount > 0 {
		b.WriteString(m.renderClaudeCodeMetrics())
		b.WriteString("\n\n")
		b.WriteString(m.renderCombinedSummary())
	}

	return b.String()
}

// renderClaudeCodeMetrics renders Claude Code usage data
func (m *Model) renderClaudeCodeMetrics() string {
	var b strings.Builder

	// If no data, try to sync on-demand
	if m.ccusageEntryCount == 0 && m.ccusageDB != nil {
		b.WriteString(m.styles.Subtle.Render("No data available\n\n"))
		b.WriteString(m.styles.Text.Render("Press 'I' to sync and load Claude Code usage data from your local projects.\n"))
		b.WriteString(m.styles.Subtle.Render("This will scan ~/.config/claude/projects/ and ~/.claude/projects/ for usage logs.\n"))
		return b.String()
	}

	// Daily Usage Section
	if len(m.ccusageDaily) > 0 {
		b.WriteString(m.styles.Title.Render("Claude Code Daily Usage"))
		b.WriteString("\n\n")

		// Aggregate by date only
		dailyTotals := make(map[string]*api.DailyReport)
		for _, report := range m.ccusageDaily {
			if daily, exists := dailyTotals[report.Date]; exists {
				daily.InputTokens += report.InputTokens
				daily.OutputTokens += report.OutputTokens
				daily.CacheCreationTokens += report.CacheCreationTokens
				daily.CacheReadTokens += report.CacheReadTokens
				daily.TotalCost += report.TotalCost
			} else {
				newReport := report
				dailyTotals[report.Date] = &newReport
			}
		}

		// Sort dates in reverse order (most recent first)
		dates := make([]string, 0, len(dailyTotals))
		for date := range dailyTotals {
			dates = append(dates, date)
		}
		sort.Strings(dates)
		sort.Sort(sort.Reverse(sort.StringSlice(dates)))

		// Table header with borders
		tableWidth := 110
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", tableWidth)))
		b.WriteString("\n")

		headerLine := fmt.Sprintf("│ %-12s │ %-14s │ %-14s │ %-14s │ %-14s │ %-16s │ %-12s │",
			"Date", "Input", "Output", "CacheWr", "CacheRd", "Total Tokens", "Cost USD")
		b.WriteString(m.styles.Primary.Render(headerLine))
		b.WriteString("\n")
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", tableWidth)))
		b.WriteString("\n")

		totalCost := 0.0
		for i, date := range dates {
			report := dailyTotals[date]
			totalTokens := report.InputTokens + report.OutputTokens + report.CacheCreationTokens + report.CacheReadTokens
			line := fmt.Sprintf("│ %-12s │ %-14s │ %-14s │ %-14s │ %-14s │ %-16s │ %-12s │",
				date,
				formatNumber(report.InputTokens),
				formatNumber(report.OutputTokens),
				formatNumber(report.CacheCreationTokens),
				formatNumber(report.CacheReadTokens),
				formatNumber(totalTokens),
				fmt.Sprintf("$%.4f", report.TotalCost))

			// Alternate row colors for better readability (only within table width)
			if i%2 == 0 {
				b.WriteString(m.styles.Text.Render(line))
			} else {
				// Darker background for alternating rows - only applies to text width
				altStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("250")).
					Background(lipgloss.Color("233")).
					Padding(0, 0)
				b.WriteString(altStyle.Render(line))
			}
			b.WriteString("\n")
			totalCost += report.TotalCost
		}

		// Table bottom border
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", tableWidth)))
		b.WriteString("\n")
		b.WriteString(m.styles.Subtle.Render(fmt.Sprintf("Total Cost: $%.4f\n", totalCost)))
	}

	// Monthly Summary Section
	if len(m.ccusageMonthly) > 0 {
		b.WriteString("\n\n")
		b.WriteString(m.styles.Title.Render("Claude Code Monthly Summary"))
		b.WriteString("\n\n")

		// Monthly stats table
		monthlyTableWidth := 95
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", monthlyTableWidth)))
		b.WriteString("\n")

		monthHeader := fmt.Sprintf("│ %-12s │ %-12s │ %-16s │ %-16s │ %-16s │ %-12s │",
			"Month", "Days", "Input Tokens", "Output Tokens", "Total Tokens", "Cost USD")
		b.WriteString(m.styles.Primary.Render(monthHeader))
		b.WriteString("\n")
		b.WriteString(m.styles.Border.Render(strings.Repeat("─", monthlyTableWidth)))
		b.WriteString("\n")

		// Display each month's data
		for i, monthly := range m.ccusageMonthly {
			totalTokensMonth := monthly.TotalInputTokens + monthly.TotalOutputTokens + monthly.TotalCacheTokens

			monthLine := fmt.Sprintf("│ %-12s │ %-12d │ %-16s │ %-16s │ %-16s │ %-12s │",
				monthly.Month,
				monthly.Days,
				formatNumber(monthly.TotalInputTokens),
				formatNumber(monthly.TotalOutputTokens),
				formatNumber(totalTokensMonth),
				fmt.Sprintf("$%.4f", monthly.TotalCost))

			// Alternate row colors for better readability
			if i%2 == 0 {
				b.WriteString(m.styles.Text.Render(monthLine))
			} else {
				altStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("250")).
					Background(lipgloss.Color("233")).
					Padding(0, 0)
				b.WriteString(altStyle.Render(monthLine))
			}
			b.WriteString("\n")
		}

		b.WriteString(m.styles.Border.Render(strings.Repeat("─", monthlyTableWidth)))
		b.WriteString("\n")
	}

	// Active Block Section
	if m.ccusageBlock != nil {
		b.WriteString("\n\n")
		b.WriteString(m.styles.Title.Render("Current Billing Block"))
		b.WriteString("\n\n")

		blockInfo := fmt.Sprintf("Block: %s to %s | Time Remaining: %s",
			m.ccusageBlock.StartTime,
			m.ccusageBlock.EndTime,
			m.ccusageBlock.TimeRemaining)
		b.WriteString(m.styles.Text.Render(blockInfo))
		b.WriteString("\n")

		blockUsage := fmt.Sprintf("Usage: %s input | %s output | $%.4f total",
			formatNumber(m.ccusageBlock.InputTokens),
			formatNumber(m.ccusageBlock.OutputTokens),
			m.ccusageBlock.TotalCost)
		b.WriteString(m.styles.Text.Render(blockUsage))
		b.WriteString("\n")
	}

	return b.String()
}

// renderCombinedSummary renders Claude Code usage summary
func (m *Model) renderCombinedSummary() string {
	var b strings.Builder

	if len(m.ccusageMonthly) == 0 {
		return ""
	}

	b.WriteString(m.styles.Title.Render("Total All-Time Cost"))
	b.WriteString("\n\n")

	// Calculate totals across all months
	totalCost := 0.0
	for _, monthly := range m.ccusageMonthly {
		totalCost += monthly.TotalCost
	}

	line := fmt.Sprintf("Claude Code Usage: $%.4f", totalCost)
	b.WriteString(m.styles.Primary.Render(line))
	b.WriteString("\n")

	return b.String()
}