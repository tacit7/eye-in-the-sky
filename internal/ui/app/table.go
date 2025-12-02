package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Alignment represents text alignment within a table cell
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// BorderStyle represents the style of table borders
type BorderStyle int

const (
	BorderNone BorderStyle = iota
	BorderSimple
	BorderRounded
	BorderDouble
	BorderHeavy
	BorderBoxed // New: Renders table with │ column separators
)

// TableColumn defines a column in a table
type TableColumn struct {
	Title     string
	Width     int
	Alignment Alignment
	Truncate  bool
}

// TableBuilder builds and renders tables with configurable columns and styles
type TableBuilder struct {
	Columns      []TableColumn
	BorderStyle  BorderStyle
	RowStyleFunc func(rowIndex int) lipgloss.Style
	HeaderStyle  lipgloss.Style
	BorderColor  lipgloss.Color
}

// NewTableBuilder creates a new table builder with default settings
func NewTableBuilder() *TableBuilder {
	return &TableBuilder{
		BorderStyle: BorderSimple,
		HeaderStyle: lipgloss.NewStyle().Bold(true),
		BorderColor: lipgloss.Color("240"),
	}
}

// RenderTable renders the table and returns both the rendered string and computed height
func (tb *TableBuilder) RenderTable(rows [][]string) (string, int) {
	if len(tb.Columns) == 0 || len(rows) == 0 {
		return "", 0
	}

	var result strings.Builder
	height := 0

	// Render header
	header := tb.renderHeader()
	if header != "" {
		result.WriteString(header)
		result.WriteString("\n")
		height++
	}

	// Render separator
	if tb.BorderStyle != BorderNone {
		separator := tb.renderSeparator()
		result.WriteString(separator)
		result.WriteString("\n")
		height++
	}

	// Render rows
	for i, row := range rows {
		renderedRow := tb.renderRow(row, i)
		result.WriteString(renderedRow)
		if i < len(rows)-1 {
			result.WriteString("\n")
		}
		height++
	}

	return result.String(), height
}

// RenderBorder renders a horizontal border for the table
func (tb *TableBuilder) RenderBorder() string {
	totalWidth := 0
	for _, col := range tb.Columns {
		totalWidth += col.Width
	}
	// Account for column separators if using bordered style
	if tb.BorderStyle == BorderBoxed {
		totalWidth += (len(tb.Columns) - 1) * 3 // " │ " between columns
		totalWidth += 4 // "│ " at start and " │" at end
	} else if tb.BorderStyle == BorderSimple || tb.BorderStyle == BorderDouble {
		totalWidth += (len(tb.Columns) - 1) * 3 // " │ " between columns
	} else {
		totalWidth += (len(tb.Columns) - 1) * 2 // "  " between columns
	}

	char := "─"
	switch tb.BorderStyle {
	case BorderDouble:
		char = "═"
	case BorderHeavy:
		char = "━"
	}

	return lipgloss.NewStyle().Foreground(tb.BorderColor).Render(strings.Repeat(char, totalWidth))
}

// renderHeader renders the table header
func (tb *TableBuilder) renderHeader() string {
	if tb.BorderStyle == BorderNone {
		return ""
	}

	var cells []string
	for i, col := range tb.Columns {
		cell := tb.formatCell(col.Title, i, -1) // -1 indicates header
		cells = append(cells, cell)
	}

	var header string
	if tb.BorderStyle == BorderBoxed {
		header = "│ " + strings.Join(cells, " │ ") + " │"
	} else {
		header = strings.Join(cells, "  ")
	}
	return tb.HeaderStyle.Render(header)
}

// renderSeparator renders a separator line
func (tb *TableBuilder) renderSeparator() string {
	totalWidth := 0
	for _, col := range tb.Columns {
		totalWidth += col.Width
	}
	totalWidth += (len(tb.Columns) - 1) * 2 // Account for spacing between columns

	char := "─"
	switch tb.BorderStyle {
	case BorderDouble:
		char = "═"
	case BorderHeavy:
		char = "━"
	}

	separator := strings.Repeat(char, totalWidth)
	return lipgloss.NewStyle().Foreground(tb.BorderColor).Render(separator)
}

// renderRow renders a single data row
func (tb *TableBuilder) renderRow(row []string, rowIndex int) string {
	var cells []string
	for i := range tb.Columns {
		cell := ""
		if i < len(row) {
			cell = tb.formatCell(row[i], i, rowIndex)
		} else {
			cell = tb.formatCell("", i, rowIndex)
		}
		cells = append(cells, cell)
	}

	var rowStr string
	if tb.BorderStyle == BorderBoxed {
		rowStr = "│ " + strings.Join(cells, " │ ") + " │"
	} else {
		rowStr = strings.Join(cells, "  ")
	}

	// Apply row styling if provided
	if tb.RowStyleFunc != nil {
		style := tb.RowStyleFunc(rowIndex)
		return style.Render(rowStr)
	}

	return rowStr
}

// formatCell formats a cell according to column settings
func (tb *TableBuilder) formatCell(content string, colIndex int, rowIndex int) string {
	if colIndex >= len(tb.Columns) {
		return content
	}

	col := tb.Columns[colIndex]

	// Truncate if needed
	if col.Truncate && lipgloss.Width(content) > col.Width {
		if col.Width > 3 {
			content = content[:col.Width-3] + "..."
		} else {
			content = "..."
		}
	}

	// Apply alignment
	switch col.Alignment {
	case AlignCenter:
		return centerInWidth(content, col.Width)
	case AlignRight:
		return padLeft(content, col.Width)
	default: // AlignLeft
		return padRight(content, col.Width)
	}
}

// centerInWidth centers text within the specified width
func centerInWidth(text string, width int) string {
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}

	leftPad := (width - textWidth) / 2
	rightPad := width - textWidth - leftPad
	return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
}

// padLeft pads text to the left to reach the specified width
func padLeft(text string, width int) string {
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}
	return strings.Repeat(" ", width-textWidth) + text
}

// AddColumn adds a column to the table builder
func (tb *TableBuilder) AddColumn(title string, width int, alignment Alignment, truncate bool) *TableBuilder {
	tb.Columns = append(tb.Columns, TableColumn{
		Title:     title,
		Width:     width,
		Alignment: alignment,
		Truncate:  truncate,
	})
	return tb
}

// SetBorderStyle sets the border style for the table
func (tb *TableBuilder) SetBorderStyle(style BorderStyle) *TableBuilder {
	tb.BorderStyle = style
	return tb
}

// SetHeaderStyle sets the style for the header row
func (tb *TableBuilder) SetHeaderStyle(style lipgloss.Style) *TableBuilder {
	tb.HeaderStyle = style
	return tb
}

// SetRowStyleFunc sets a function to style individual rows
func (tb *TableBuilder) SetRowStyleFunc(fn func(rowIndex int) lipgloss.Style) *TableBuilder {
	tb.RowStyleFunc = fn
	return tb
}

// SetAlternatingRowStyle sets up alternating row colors for better readability
func (tb *TableBuilder) SetAlternatingRowStyle(evenColor, oddColor lipgloss.Color) *TableBuilder {
	tb.RowStyleFunc = func(rowIndex int) lipgloss.Style {
		if rowIndex%2 == 0 {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("250")).
				Background(evenColor).
				Padding(0, 0)
		}
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Background(oddColor).
			Padding(0, 0)
	}
	return tb
}

// RenderSimpleTable is a convenience function for quick table rendering
func RenderSimpleTable(headers []string, rows [][]string, widths []int) string {
	tb := NewTableBuilder()

	for i, header := range headers {
		width := 20 // default width
		if i < len(widths) {
			width = widths[i]
		}
		tb.AddColumn(header, width, AlignLeft, true)
	}

	result, _ := tb.RenderTable(rows)
	return result
}