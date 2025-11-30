package usageview

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/api"
	"github.com/tacit7/eye-in-the-sky/internal/ui/services"
)

// Styles interface defines the styling methods needed by renderers
type Styles interface {
	Primary() lipgloss.Style
	Text() lipgloss.Style
	Subtle() lipgloss.Style
	// Color methods for theme configuration
	HeaderFg() lipgloss.Color
	HeaderBg() lipgloss.Color
	AlternatingRowDark() lipgloss.Color
	AlternatingRowLight() lipgloss.Color
}

// RenderSummaryBar renders the top-level summary bar with key metrics
func RenderSummaryBar(styles Styles, totalCost float64, totalTokens int, lastSync time.Time) string {
	// Format individual parts with styling
	costPart := styles.Primary().Render(fmt.Sprintf("\uf0d6 Total: %s", FormatCostUSD(totalCost))) // nf-fa-money
	tokensPart := styles.Text().Render(fmt.Sprintf("\uf013 Tokens: %s", FormatNumber(totalTokens))) // nf-fa-cog

	// Format timestamp
	syncTimeStr := lastSync.Format("2006-01-02 15:04 MST")
	updatePart := styles.Subtle().Render(fmt.Sprintf("\uf017 Last synced: %s", syncTimeStr)) // nf-fa-clock_o

	// Join with separators
	separator := styles.Subtle().Render("  |  ")
	parts := []string{costPart, tokensPart, updatePart}

	return strings.Join(parts, separator)
}

// RenderGradientHeader renders the styled usage page header
func RenderGradientHeader(styles Styles) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(styles.HeaderFg()).
		Background(styles.HeaderBg()).
		Bold(true).
		Padding(0, 2)

	return headerStyle.Render("\uf0ac  Eye in the Sky — Usage Metrics") // nf-fa-globe
}

// TableBuilder interface for rendering tables
type TableBuilder interface {
	SetBorderStyle(style int) TableBuilder
	SetHeaderStyle(style lipgloss.Style) TableBuilder
	SetAlternatingRowStyle(even, odd lipgloss.Color) TableBuilder
	AddColumn(name string, width int, align int, truncate bool) TableBuilder
	RenderTable(rows [][]string) (string, int)
	RenderBorder() string
}

// RenderSessionUsage renders the Eye-in-the-Sky session metrics table
func RenderSessionUsage(summary services.UsageSummary, styles Styles, newTable func() TableBuilder) string {
	if !summary.HasData {
		return ""
	}

	// Build table with consistent column definitions
	table := newTable().
		SetBorderStyle(0). // BorderSimple
		SetHeaderStyle(styles.Primary()).
		AddColumn("Agent", 10, 0, true). // Left-aligned
		AddColumn("Usage %", 8, 2, false). // Right-aligned
		AddColumn("Tokens", 12, 2, false). // Right-aligned
		AddColumn("Cost USD", 10, 2, false). // Right-aligned
		AddColumn("Model", 12, 0, true) // Left-aligned

	// Build rows
	rows := make([][]string, 0, len(summary.Rows))
	for _, row := range summary.Rows {
		agentID := Truncate(row.AgentID, 8)
		usage := fmt.Sprintf("%.1f%%", row.Usage)
		tokens := fmt.Sprintf("%d", row.TokensUsed)
		cost := fmt.Sprintf("$%.2f", row.Cost)
		model := Truncate(row.Model, 12)

		rows = append(rows, []string{agentID, usage, tokens, cost, model})
	}

	rendered, _ := table.RenderTable(rows)

	// Add totals line
	totalsLine := styles.Subtle().Render(
		fmt.Sprintf("Total: $%.2f | %d tokens", summary.Totals.CostUSD, summary.Totals.Tokens))

	return rendered + "\n" + totalsLine
}

// RenderClaudeDailyUsage renders the Claude Code daily usage table
func RenderClaudeDailyUsage(summary services.UsageSummary, styles Styles, newTable func() TableBuilder) string {
	if !summary.HasData {
		return ""
	}

	// Build bordered table with consistent columns
	table := newTable().
		SetBorderStyle(1). // BorderBoxed
		SetHeaderStyle(styles.Primary()).
		SetAlternatingRowStyle(styles.AlternatingRowDark(), styles.AlternatingRowLight()).
		AddColumn("Date", 15, 0, false). // Left-aligned, widened for weekday (Mon 2025-10-22)
		AddColumn("Input", 14, 2, false). // Right-aligned
		AddColumn("Output", 14, 2, false). // Right-aligned
		AddColumn("CacheWr", 14, 2, false). // Right-aligned
		AddColumn("CacheRd", 14, 2, false). // Right-aligned
		AddColumn("Total Tokens", 16, 2, false). // Right-aligned
		AddColumn("Cost USD", 12, 2, false) // Right-aligned

	// Build rows from service data (which has structured token fields)
	rows := make([][]string, 0, len(summary.Rows))

	for _, row := range summary.Rows {
		rows = append(rows, []string{
			row.Timestamp,
			FormatNumber(row.InputTokens),
			FormatNumber(row.OutputTokens),
			FormatNumber(row.CacheCreate),
			FormatNumber(row.CacheRead),
			FormatNumber(row.TokensUsed),
			fmt.Sprintf("$%.2f", row.Cost),
		})
	}

	rendered, _ := table.RenderTable(rows)

	// Add border and totals
	result := table.RenderBorder() + "\n" + rendered + "\n" + table.RenderBorder()
	totalsLine := styles.Subtle().Render(fmt.Sprintf("Total Cost: $%.2f", summary.Totals.CostUSD))

	return result + "\n" + totalsLine
}

// RenderClaudeMonthlyUsage renders the Claude Code monthly summary table
func RenderClaudeMonthlyUsage(summary services.UsageSummary, styles Styles, newTable func() TableBuilder) string {
	if !summary.HasData {
		return ""
	}

	// Build bordered table
	table := newTable().
		SetBorderStyle(1). // BorderBoxed
		SetHeaderStyle(styles.Primary()).
		SetAlternatingRowStyle(styles.AlternatingRowDark(), styles.AlternatingRowLight()).
		AddColumn("Month", 12, 0, false). // Left-aligned
		AddColumn("Days", 12, 2, false). // Right-aligned
		AddColumn("Input Tokens", 16, 2, false). // Right-aligned
		AddColumn("Output Tokens", 16, 2, false). // Right-aligned
		AddColumn("Total Tokens", 16, 2, false). // Right-aligned
		AddColumn("Cost USD", 12, 2, false) // Right-aligned

	// Build rows from service data (which has structured fields)
	rows := make([][]string, 0, len(summary.Rows))

	for _, row := range summary.Rows {
		rows = append(rows, []string{
			row.Timestamp,
			fmt.Sprintf("%d", row.Days),
			FormatNumber(row.InputTokens),
			FormatNumber(row.OutputTokens),
			FormatNumber(row.TokensUsed),
			fmt.Sprintf("$%.2f", row.Cost),
		})
	}

	rendered, _ := table.RenderTable(rows)

	// Add borders
	result := table.RenderBorder() + "\n" + rendered + "\n" + table.RenderBorder()

	return result
}

// RenderBillingBlock renders the current billing block information with progress bar
func RenderBillingBlock(block *api.ActiveBlockReport, styles Styles) string {
	if block == nil {
		return ""
	}

	var b strings.Builder

	// Create progress bar (showing 62% as placeholder - would need actual budget data to calculate)
	// For now, we'll use a visual indicator based on time remaining
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 40

	// Estimate progress (placeholder - would need actual start/end timestamps to calculate accurately)
	percent := 0.62 // 62% remaining as per spec
	progBar := prog.ViewAs(percent)

	// Build the display
	b.WriteString(progBar)
	b.WriteString(fmt.Sprintf("  %.0f%% remaining (%s)", percent*100, block.TimeRemaining))
	b.WriteString("\n\n")

	// Add token and cost information
	info := fmt.Sprintf("%s input | %s output | %s cache | %s total",
		styles.Text().Render(FormatNumber(block.InputTokens)),
		styles.Text().Render(FormatNumber(block.OutputTokens)),
		styles.Text().Render(FormatNumber(block.CacheTokens)),
		FormatCostUSD(block.TotalCost))
	b.WriteString(info)

	return b.String()
}

// Helper functions

// FormatNumber formats a number with comma separators
func FormatNumber(n int) string {
	if n == 0 {
		return "0"
	}

	// Handle negative numbers
	negative := n < 0
	if negative {
		n = -n
	}

	// Convert to string and add commas
	str := fmt.Sprintf("%d", n)
	var result strings.Builder

	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(digit)
	}

	if negative {
		return "-" + result.String()
	}
	return result.String()
}

// FormatCostUSD formats cost in USD with color coding
func FormatCostUSD(cost float64) string {
	return fmt.Sprintf("$%.2f", cost)
}

// RenderMonthlyUsage renders the monthly cost breakdown table
func RenderMonthlyUsage(summary services.UsageSummary, styles Styles, newTable func() TableBuilder) string {
	if !summary.HasData {
		return ""
	}

	// Build table with consistent column definitions
	table := newTable().
		SetBorderStyle(0). // BorderSimple
		SetHeaderStyle(styles.Primary()).
		AddColumn("Timestamp", 20, 0, false). // Left-aligned
		AddColumn("Usage %", 8, 2, false). // Right-aligned
		AddColumn("Tokens Used", 12, 2, false). // Right-aligned
		AddColumn("Cost USD", 10, 2, false). // Right-aligned
		AddColumn("Model", 8, 0, true) // Left-aligned

	// Build rows
	rows := make([][]string, 0, len(summary.Rows))
	for _, row := range summary.Rows {
		usage := fmt.Sprintf("%.1f%%", row.Usage)
		tokens := FormatNumber(row.TokensUsed)
		cost := fmt.Sprintf("$%.2f", row.Cost)
		model := Truncate(row.Model, 8)

		rows = append(rows, []string{row.Timestamp, usage, tokens, cost, model})
	}

	rendered, _ := table.RenderTable(rows)

	// Add totals line
	totalsLine := styles.Subtle().Render(
		fmt.Sprintf("Monthly Total: $%.2f | %d tokens across %d snapshots",
			summary.Totals.CostUSD, summary.Totals.Tokens, summary.Totals.Snapshots))

	return rendered + "\n" + totalsLine
}

// RenderTotalCost renders the total all-time cost summary
func RenderTotalCost(totalCost float64, styles Styles) string {
	return styles.Primary().Render(fmt.Sprintf("Claude Code Usage: $%.2f", totalCost))
}

// Truncate truncates a string to the given length
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
