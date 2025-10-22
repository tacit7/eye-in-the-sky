package app

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/api"
)

// renderEyeInTheSkyUsage renders the Eye-in-the-Sky session metrics table
func (m *Model) renderEyeInTheSkyUsage(summary UsageSummary) string {
	if !summary.HasData {
		return ""
	}

	// Build table with consistent column definitions
	table := NewTableBuilder().
		SetBorderStyle(BorderSimple).
		SetHeaderStyle(m.styles.Primary).
		AddColumn("Agent", 10, AlignLeft, true).
		AddColumn("Usage %", 8, AlignRight, false).
		AddColumn("Tokens", 12, AlignRight, false).
		AddColumn("Cost USD", 10, AlignRight, false).
		AddColumn("Model", 12, AlignLeft, true)

	// Build rows
	rows := make([][]string, 0, len(summary.Rows))
	for _, row := range summary.Rows {
		agentID := truncate(row.AgentID, 8)
		usage := fmt.Sprintf("%.1f%%", row.Usage)
		tokens := fmt.Sprintf("%d", row.TokensUsed)
		cost := fmt.Sprintf("$%.4f", row.Cost)
		model := truncate(row.Model, 12)

		rows = append(rows, []string{agentID, usage, tokens, cost, model})
	}

	rendered, _ := table.RenderTable(rows)

	// Add totals line
	totalsLine := m.styles.Subtle.Render(
		fmt.Sprintf("Total: $%.4f | %d tokens", summary.Totals.CostUSD, summary.Totals.Tokens))

	return rendered + "\n" + totalsLine
}

// renderClaudeDailyUsage renders the Claude Code daily usage table
func (m *Model) renderClaudeDailyUsage(summary UsageSummary) string {
	if !summary.HasData {
		return ""
	}

	// Build bordered table with consistent columns
	table := NewTableBuilder().
		SetBorderStyle(BorderBoxed).
		SetHeaderStyle(m.styles.Primary).
		SetAlternatingRowStyle(lipgloss.Color("233"), lipgloss.Color("")).
		AddColumn("Date", 12, AlignLeft, false).
		AddColumn("Input", 14, AlignRight, false).
		AddColumn("Output", 14, AlignRight, false).
		AddColumn("CacheWr", 14, AlignRight, false).
		AddColumn("CacheRd", 14, AlignRight, false).
		AddColumn("Total Tokens", 16, AlignRight, false).
		AddColumn("Cost USD", 12, AlignRight, false)

	// Build rows from summary data
	// Note: We stored the daily reports in summary.Rows, need to extract token details
	rows := make([][]string, 0, len(summary.Rows))

	// We need access to original daily reports here
	// For now, we'll pass them through the Model
	if len(m.ccusageDaily) > 0 {
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

		for _, row := range summary.Rows {
			if report, exists := dailyTotals[row.Timestamp]; exists {
				totalTokens := report.InputTokens + report.OutputTokens +
					report.CacheCreationTokens + report.CacheReadTokens
				rows = append(rows, []string{
					row.Timestamp,
					formatNumber(report.InputTokens),
					formatNumber(report.OutputTokens),
					formatNumber(report.CacheCreationTokens),
					formatNumber(report.CacheReadTokens),
					formatNumber(totalTokens),
					fmt.Sprintf("$%.4f", report.TotalCost),
				})
			}
		}
	}

	rendered, _ := table.RenderTable(rows)

	// Add border and totals
	result := table.RenderBorder() + "\n" + rendered + "\n" + table.RenderBorder()
	totalsLine := m.styles.Subtle.Render(fmt.Sprintf("Total Cost: $%.4f", summary.Totals.CostUSD))

	return result + "\n" + totalsLine
}

// renderClaudeMonthlyUsage renders the Claude Code monthly summary table
func (m *Model) renderClaudeMonthlyUsage(summary UsageSummary) string {
	if !summary.HasData {
		return ""
	}

	// Build bordered table
	table := NewTableBuilder().
		SetBorderStyle(BorderBoxed).
		SetHeaderStyle(m.styles.Primary).
		SetAlternatingRowStyle(lipgloss.Color("233"), lipgloss.Color("")).
		AddColumn("Month", 12, AlignLeft, false).
		AddColumn("Days", 12, AlignRight, false).
		AddColumn("Input Tokens", 16, AlignRight, false).
		AddColumn("Output Tokens", 16, AlignRight, false).
		AddColumn("Total Tokens", 16, AlignRight, false).
		AddColumn("Cost USD", 12, AlignRight, false)

	// Build rows
	rows := make([][]string, 0, len(summary.Rows))

	// We need access to original monthly reports
	if len(m.ccusageMonthly) > 0 {
		for i, monthly := range m.ccusageMonthly {
			if i < len(summary.Rows) {
				totalTokens := monthly.TotalInputTokens + monthly.TotalOutputTokens + monthly.TotalCacheTokens
				rows = append(rows, []string{
					monthly.Month,
					fmt.Sprintf("%d", monthly.Days),
					formatNumber(monthly.TotalInputTokens),
					formatNumber(monthly.TotalOutputTokens),
					formatNumber(totalTokens),
					fmt.Sprintf("$%.4f", monthly.TotalCost),
				})
			}
		}
	}

	rendered, _ := table.RenderTable(rows)

	// Add borders
	result := table.RenderBorder() + "\n" + rendered + "\n" + table.RenderBorder()

	return result
}

// renderBillingBlock renders the current billing block information
func (m *Model) renderBillingBlock(block *api.ActiveBlockReport) string {
	if block == nil {
		return ""
	}

	table := NewTableBuilder().
		SetBorderStyle(BorderSimple).
		SetHeaderStyle(m.styles.Primary).
		AddColumn("Period", 20, AlignLeft, false).
		AddColumn("Input", 15, AlignRight, false).
		AddColumn("Output", 15, AlignRight, false).
		AddColumn("Cost", 12, AlignRight, false).
		AddColumn("Time Left", 15, AlignLeft, false)

	rows := [][]string{{
		fmt.Sprintf("%s - %s", block.StartTime, block.EndTime),
		formatNumber(block.InputTokens),
		formatNumber(block.OutputTokens),
		fmt.Sprintf("$%.4f", block.TotalCost),
		block.TimeRemaining,
	}}

	rendered, _ := table.RenderTable(rows)
	return rendered
}