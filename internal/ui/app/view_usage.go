package app

import (
	"strings"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ui/services"
	"github.com/tacit7/eye-in-the-sky/internal/ui/view/usage"
)

// renderUsageTab renders session costs for all agents with monthly breakdown
func (m *Model) renderUsageTab() string {
	// Show loading state
	if m.ccusageLoading {
		return m.styles.Working.Render("\uf021 " + m.ccusageSyncStatus)
	}

	// Check for empty states
	if emptyState := m.checkEmptyState(); emptyState != "" {
		return emptyState
	}

	// Get viewport view
	view := m.usageViewport.View()

	// Add scroll indicators
	var b strings.Builder

	// Top indicator
	if m.usageViewport.AtTop() {
		b.WriteString("\n")
	} else {
		b.WriteString(m.styles.Subtle.Render("▲") + "\n")
	}

	b.WriteString(view)

	// Bottom indicator
	if !m.usageViewport.AtBottom() {
		b.WriteString("\n" + m.styles.Subtle.Render("▼"))
	}

	return b.String()
}

// checkEmptyState checks if we should display an empty state
func (m *Model) checkEmptyState() string {
	if m.ccusageDB != nil && m.ccusageEntryCount == 0 {
		return m.renderClaudeCodeEmptyState()
	}

	if len(m.allSessionMetrics) == 0 && len(m.monthlyCosts) == 0 && m.ccusageEntryCount == 0 {
		return m.renderUsageEmptyState()
	}

	return ""
}

// checkCachedUsageView returns cached view if valid
func (m *Model) checkCachedUsageView() string {
	widthUnchanged := m.width == m.lastRenderWidth
	if !m.usageDirty && widthUnchanged && m.cachedUsageRender != "" {
		m.usageViewport.SetContent(m.cachedUsageRender)
		return m.usageViewport.View()
	}
	return ""
}

// buildUsageSummary builds the usage summary from service
func (m *Model) buildUsageSummary() services.Summary {
	inputs := services.Inputs{
		Sessions:  m.allSessionMetrics,
		Monthly:   convertToValueSlice(m.monthlyCosts),
		DailyCC:   m.ccusageDaily,
		MonthlyCC: m.ccusageMonthly,
		LastSync:  m.lastCCUsageSync,
	}

	debugf("Usage data: %d sessions, %d monthly, %d daily, %d monthlyCC",
		len(m.allSessionMetrics),
		len(m.monthlyCosts),
		len(m.ccusageDaily),
		len(m.ccusageMonthly))

	summary, err := m.usageSvc.Build(inputs)
	if err != nil {
		debugf("Usage service error: %v", err)
	}

	return summary
}

// renderUsageContent renders all usage sections and composes the view
func (m *Model) renderUsageContent(summary services.Summary) string {
	usageStyles := NewUsageStyles(&m.styles)
	sections := m.renderUsageSections(summary, usageStyles)

	// Use adaptive width for separator
	separatorWidth := m.usageViewport.Width
	if separatorWidth == 0 {
		separatorWidth = m.width - 4
	}

	header := usageview.RenderGradientHeader(usageStyles)
	summaryBar := usageview.RenderSummaryBar(usageStyles, summary.TotalsUSD, summary.TokensTotal, summary.LastSync)
	separator := m.styles.Border.Render(strings.Repeat("─", separatorWidth))
	sectionsContent := usageview.RenderUsageSections(usageStyles, sections...)
	footer := usageview.RenderUsageFooter(usageStyles, m.ccusageSyncStatus, separatorWidth)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		summaryBar,
		separator,
		"",
		sectionsContent,
		footer,
	)
}

// renderUsageSections renders all individual sections
func (m *Model) renderUsageSections(summary services.Summary, styles *UsageStyles) []string {
	var sections []string

	if summary.Sessions.HasData {
		content := usageview.RenderSessionUsage(summary.Sessions, styles, NewUsageTableBuilder)
		sections = append(sections, usageview.RenderSectionWithTitle("Eye-in-the-Sky Session Metrics", content, styles))
	}

	if summary.Monthly.HasData {
		content := usageview.RenderMonthlyUsage(summary.Monthly, styles, NewUsageTableBuilder)
		sections = append(sections, usageview.RenderSectionWithTitle("Monthly Cost Breakdown", content, styles))
	}

	if summary.DailyCC.HasData {
		content := usageview.RenderClaudeDailyUsage(summary.DailyCC, styles, NewUsageTableBuilder)
		sections = append(sections, usageview.RenderSectionWithTitle("Claude Code Daily Usage", content, styles))
	}

	if summary.MonthlyCC.HasData {
		content := usageview.RenderClaudeMonthlyUsage(summary.MonthlyCC, styles, NewUsageTableBuilder)
		sections = append(sections, usageview.RenderSectionWithTitle("Claude Code Monthly Summary", content, styles))
	}

	if m.ccusageBlock != nil {
		content := usageview.RenderBillingBlock(m.ccusageBlock, styles)
		sections = append(sections, usageview.RenderSectionWithTitle("Current Billing Block", content, styles))
	}

	if len(m.ccusageMonthly) > 0 {
		// Use DailyCC total (MonthlyCC is the same data, just aggregated differently)
		content := usageview.RenderTotalCost(summary.DailyCC.Totals.CostUSD, styles)
		sections = append(sections, usageview.RenderSectionWithTitle("Claude Code Total Cost", content, styles))
	}

	return sections
}

// cacheAndReturnUsageView caches the content and returns the viewport view
func (m *Model) cacheAndReturnUsageView(content string) string {
	m.cachedUsageRender = content
	m.usageDirty = false
	m.lastRenderWidth = m.width

	m.usageViewport.SetContent(content)
	return m.usageViewport.View()
}

// convertToValueSlice converts pointer slice to value slice
func convertToValueSlice(metrics []*SessionMetric) []SessionMetric {
	result := make([]SessionMetric, 0, len(metrics))
	for _, m := range metrics {
		if m != nil {
			result = append(result, *m)
		}
	}
	return result
}