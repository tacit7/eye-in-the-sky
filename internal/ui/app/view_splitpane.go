package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderSplitPaneView renders a view with left list and right detail panes
func (m *Model) renderSplitPaneView(
	title string,
	items []string,
	selectedIndex int,
	totalCount int,
	detailContent string,
	footer string,
) string {
	var b strings.Builder

	// Header with count
	header := m.renderHeader()
	viewTitle := fmt.Sprintf("%s (%d/%d)", title, selectedIndex+1, totalCount)
	if totalCount == 0 {
		viewTitle = fmt.Sprintf("%s (0)", title)
	}
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(m.styles.Title.Render(viewTitle))
	b.WriteString("\n\n")

	// Calculate pane widths (40/60 split)
	leftWidth := (m.width * 40) / 100
	rightWidth := m.width - leftWidth - 3 // -3 for border

	// Available height for content
	contentHeight := m.height - 8 // Reserve for header, title, footer

	// Left pane: item list
	var leftBuilder strings.Builder
	if len(items) == 0 {
		emptyMsg := fmt.Sprintf("No %s found", strings.ToLower(title))
		leftBuilder.WriteString(m.styles.Subtle.Render(emptyMsg))
	} else {
		for i, item := range items {
			if i == selectedIndex {
				leftBuilder.WriteString(m.styles.Primary.Reverse(true).Render(item))
			} else {
				leftBuilder.WriteString(m.styles.Text.Render(item))
			}
			leftBuilder.WriteString("\n")
		}
	}

	// Right pane: detail content with scrolling
	rightContent := detailContent
	if rightContent == "" && len(items) > 0 {
		rightContent = m.styles.Subtle.Render("Select an item to view details")
	}

	// Apply scrolling to right pane content
	rightLines := strings.Split(rightContent, "\n")
	visibleLines := contentHeight - 2 // Account for padding
	if visibleLines < 1 {
		visibleLines = 1
	}

	// Window the content based on scroll offset
	startLine := m.rightPaneOffset
	endLine := startLine + visibleLines
	if startLine >= len(rightLines) {
		startLine = 0
		m.rightPaneOffset = 0
	}
	if endLine > len(rightLines) {
		endLine = len(rightLines)
	}

	visibleContent := strings.Join(rightLines[startLine:endLine], "\n")

	// Style panes
	leftPane := lipgloss.NewStyle().
		Width(leftWidth).
		Height(contentHeight).
		BorderStyle(lipgloss.NormalBorder()).
		BorderRight(true).
		BorderForeground(lipgloss.Color(m.theme.Colors.Border)).
		Padding(0, 1).
		Render(leftBuilder.String())

	rightPane := lipgloss.NewStyle().
		Width(rightWidth).
		Height(contentHeight).
		Padding(0, 1).
		Render(visibleContent)

	// Join panes horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
	b.WriteString(content)

	// Footer
	b.WriteString("\n")
	b.WriteString(footer)

	return b.String()
}

// renderFooterWithKeys renders a footer with key bindings
func (m *Model) renderFooterWithKeys(keys string) string {
	footerStyle := m.styles.Subtle.
		Width(m.width).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(lipgloss.Color(m.theme.Colors.Border))

	return footerStyle.Render(keys)
}

// renderSplitPaneContent renders just the split-pane content for tabs (no header/footer)
func (m *Model) renderSplitPaneContent(
	items []string,
	selectedIndex int,
	detailContent string,
) string {
	if len(items) == 0 {
		return ""
	}

	// Calculate pane widths (40/60 split)
	leftWidth := (m.width * 40) / 100
	rightWidth := m.width - leftWidth - 7 // Account for borders and padding

	// Available height for content (tab view has less space)
	contentHeight := m.height - 12 // Reserve for header, tabs, footer, borders

	if contentHeight < 5 {
		contentHeight = 5
	}

	// Left pane: item list
	var leftBuilder strings.Builder
	for i, item := range items {
		if i == selectedIndex {
			leftBuilder.WriteString(m.styles.Primary.Reverse(true).Render(item))
		} else {
			leftBuilder.WriteString(m.styles.Text.Render(item))
		}
		leftBuilder.WriteString("\n")
	}

	// Right pane: detail content with scrolling
	rightContent := detailContent
	if rightContent == "" {
		rightContent = m.styles.Subtle.Render("Select an item to view details")
	}

	// Apply scrolling to right pane content
	rightLines := strings.Split(rightContent, "\n")
	visibleLines := contentHeight - 2
	if visibleLines < 1 {
		visibleLines = 1
	}

	// Window the content based on scroll offset
	startLine := m.rightPaneOffset
	endLine := startLine + visibleLines
	if startLine >= len(rightLines) {
		startLine = 0
		m.rightPaneOffset = 0
	}
	if endLine > len(rightLines) {
		endLine = len(rightLines)
	}

	visibleContent := strings.Join(rightLines[startLine:endLine], "\n")

	// Style panes
	leftPane := lipgloss.NewStyle().
		Width(leftWidth).
		Height(contentHeight).
		BorderStyle(lipgloss.NormalBorder()).
		BorderRight(true).
		BorderForeground(lipgloss.Color(m.theme.Colors.Border)).
		Padding(0, 1).
		Render(leftBuilder.String())

	rightPane := lipgloss.NewStyle().
		Width(rightWidth).
		Height(contentHeight).
		Padding(0, 1).
		Render(visibleContent)

	// Join panes horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
}
