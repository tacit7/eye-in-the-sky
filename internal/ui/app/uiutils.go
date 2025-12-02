package app

import (
	"fmt"
)

// formatCostUSD formats a cost value as USD currency
func formatCostUSD(cost float64) string {
	return fmt.Sprintf("$%.4f", cost)
}

// formatTokenCount formats a token count with comma separators
func formatTokenCount(tokens int) string {
	return formatNumber(tokens)
}

// calculateUsagePercent calculates usage percentage from tokens used and budget
func calculateUsagePercent(tokensUsed, tokensBudget int) float64 {
	if tokensBudget == 0 {
		return 0.0
	}
	return float64(tokensUsed) / float64(tokensBudget) * 100
}

// formatUsagePercent formats a usage percentage
func formatUsagePercent(percent float64) string {
	return fmt.Sprintf("%.1f%%", percent)
}

// colorCost applies color to cost values based on thresholds
// Green: < $1.00, Yellow: $1.00-$5.00, Red: > $5.00
func colorCost(cost float64, styles *Styles) string {
	formatted := fmt.Sprintf("$%.4f", cost)
	switch {
	case cost > 5.0:
		return styles.Error.Render(formatted)
	case cost > 1.0:
		return styles.Warning.Render(formatted)
	default:
		return styles.Success.Render(formatted)
	}
}