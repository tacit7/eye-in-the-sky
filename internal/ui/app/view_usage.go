package app

import (
	"fmt"
	"strings"
	"github.com/charmbracelet/lipgloss"
)

// renderUsageTab renders session costs for all agents with monthly breakdown
func (m *Model) renderUsageTab() string {
	// Check for Claude Code empty state
	if m.ccusageDB != nil && m.ccusageEntryCount == 0 {
		return m.renderClaudeCodeEmptyState()
	}

	// Check if we have any data at all
	if len(m.allSessionMetrics) == 0 && len(m.monthlyCosts) == 0 && m.ccusageEntryCount == 0 {
		return m.renderUsageEmptyState()
	}

	// Prepare data summaries
	sessionSummary := BuildSessionMetricsSummary(m.allSessionMetrics)
	monthlySummary := BuildMonthlyCostsSummary(m.monthlyCosts)
	dailySummary := BuildClaudeDailySummary(m.ccusageDaily)
	monthlyReportSummary := BuildClaudeMonthlyReport(m.ccusageMonthly)

	// Render individual sections
	var sections []string

	// Eye-in-the-Sky Session Metrics
	if sessionSummary.HasData {
		content := m.renderEyeInTheSkyUsage(sessionSummary)
		sections = append(sections, m.renderSectionWithTitle("Eye-in-the-Sky Session Metrics", content))
	}

	// Monthly Cost Breakdown
	if monthlySummary.HasData {
		content := m.renderMonthlyCostsBreakdown(monthlySummary)
		sections = append(sections, m.renderSectionWithTitle("Monthly Cost Breakdown", content))
	}

	// Claude Code Daily Usage
	if dailySummary.HasData {
		content := m.renderClaudeDailyUsage(dailySummary)
		sections = append(sections, m.renderSectionWithTitle("Claude Code Daily Usage", content))
	}

	// Claude Code Monthly Summary
	if monthlyReportSummary.HasData {
		content := m.renderClaudeMonthlyUsage(monthlyReportSummary)
		sections = append(sections, m.renderSectionWithTitle("Claude Code Monthly Summary", content))
	}

	// Current Billing Block
	if m.ccusageBlock != nil {
		content := m.renderBillingBlock(m.ccusageBlock)
		sections = append(sections, m.renderSectionWithTitle("Current Billing Block", content))
	}

	// Total All-Time Cost
	if len(m.ccusageMonthly) > 0 {
		content := m.renderTotalAllTimeCost()
		sections = append(sections, m.renderSectionWithTitle("Total All-Time Cost", content))
	}

	// Build complete usage dashboard with enhancements
	header := m.renderGradientHeader()
	summaryBar := m.renderSummaryBar()
	separator := m.styles.Border.Render(strings.Repeat("─", 70))
	sectionsContent := m.renderUsageSections(sections...)
	footer := m.renderUsageFooter()

	// Compose final view
	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		summaryBar,
		separator,
		"",
		sectionsContent,
		footer,
	)

	// Set viewport content and return view
	m.usageViewport.SetContent(content)
	return m.usageViewport.View()
}

// renderMonthlyCostsBreakdown renders monthly cost breakdown using table builder
func (m *Model) renderMonthlyCostsBreakdown(summary UsageSummary) string {
	if !summary.HasData {
		return ""
	}

	// Build table with consistent column definitions
	table := NewTableBuilder().
		SetBorderStyle(BorderSimple).
		SetHeaderStyle(m.styles.Primary).
		AddColumn("Timestamp", 20, AlignLeft, false).
		AddColumn("Usage %", 8, AlignRight, false).
		AddColumn("Tokens Used", 12, AlignRight, false).
		AddColumn("Cost USD", 10, AlignRight, false).
		AddColumn("Model", 8, AlignLeft, true)

	// Build rows
	rows := make([][]string, 0, len(summary.Rows))
	for _, row := range summary.Rows {
		usage := formatUsagePercent(row.Usage)
		tokens := formatTokenCount(row.TokensUsed)
		cost := formatCostUSD(row.Cost)
		model := truncate(row.Model, 8)

		rows = append(rows, []string{row.Timestamp, usage, tokens, cost, model})
	}

	rendered, _ := table.RenderTable(rows)

	// Add totals line
	totalsLine := m.styles.Subtle.Render(
		fmt.Sprintf("Monthly Total: $%.4f | %d tokens across %d snapshots",
			summary.Totals.CostUSD, summary.Totals.Tokens, summary.Totals.Snapshots))

	return rendered + "\n" + totalsLine
}

// renderTotalAllTimeCost renders the total all-time cost summary
func (m *Model) renderTotalAllTimeCost() string {
	// Calculate totals across all months
	totalCost := 0.0
	for _, monthly := range m.ccusageMonthly {
		totalCost += monthly.TotalCost
	}

	return m.styles.Primary.Render(fmt.Sprintf("Claude Code Usage: $%.4f", totalCost))
}