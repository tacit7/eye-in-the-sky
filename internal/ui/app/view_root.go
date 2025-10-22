package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current view
func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Render help overlay if shown
	if m.showHelp {
		return m.renderHelp()
	}

	switch m.currentView {
	case ViewList:
		return m.renderListView()
	case ViewDetail:
		return m.renderDetailView()
	default:
		return "Unknown view"
	}
}

// renderHeader renders the top header bar
func (m *Model) renderHeader() string {
	// Calculate widths
	headerWidth := m.width - 2 // Account for padding
	if headerWidth < 40 {
		headerWidth = 40
	}

	// Title section
	title := " 👁️ Eye in the Sky - Agent Management "
	appVersion := "v0.2.0" // TODO: Make this configurable
	titleAndVersion := title + appVersion

	// Generate padding and center title
	padding := (headerWidth - len(titleAndVersion))
	leftPad := padding / 2
	rightPad := padding - leftPad

	// Build header with centered title
	var headerContent strings.Builder
	headerContent.WriteString(strings.Repeat(" ", leftPad))
	headerContent.WriteString(m.styles.Bold.Render(title))
	headerContent.WriteString(m.styles.Subtle.Render(appVersion))
	headerContent.WriteString(strings.Repeat(" ", rightPad))

	// Style the header with background
	styledHeader := m.styles.Header.
		Width(headerWidth).
		Render(headerContent.String())

	return styledHeader
}

// renderFooter renders the bottom status bar
func (m *Model) renderFooter() string {
	var footer string
	footerWidth := m.width - 2 // Account for padding

	// Build the key help section based on current view
	keyHelp := m.getKeyHelp()
	left := m.styles.KeyHelp.Render(keyHelp)

	// Status section (only show status messages temporarily)
	middle := ""
	if m.statusMsg != "" && time.Since(m.statusTime) < 3*time.Second {
		middle = m.styles.StatusBar.Render(m.statusMsg)
	}

	// Sync status - right aligned
	right := m.getRightStatus()

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	middleWidth := lipgloss.Width(middle)
	rightWidth := lipgloss.Width(right)
	spacing := footerWidth - leftWidth - middleWidth - rightWidth

	if spacing > 0 {
		spacer := strings.Repeat(" ", spacing)
		footer = left + spacer + middle + right
	} else {
		// If not enough space, just show key help
		footer = left
	}

	// Apply footer styling
	return m.styles.Footer.
		Width(footerWidth).
		Render(footer)
}

// getKeyHelp returns context-sensitive key hints
func (m *Model) getKeyHelp() string {
	// Build the key help section based on current view
	var keyHelp []string

	switch m.currentView {
	case ViewList:
		switch m.listTabs.ActiveIndex {
		case 1: // Project tab
			// Handle the sub-navigation within project tab
			switch m.projectSection {
			case "tasks":
				keyHelp = []string{"[j/k] navigate", "[enter] open task", "[d] mark done", "[←] back"}
			case "claude":
				keyHelp = []string{"[j/k] scroll", "[←] back"}
			case "markdown":
				keyHelp = []string{"[j/k] navigate files", "[←] back"}
			default:
				keyHelp = []string{"[t] tasks", "[c] claude.md", "[m] markdown files", "[tab] back to overview"}
			}
		default: // Overview and others
			keyHelp = []string{"[j/k] navigate", "[tab] switch tab", "[enter] details"}
			if m.selectedIndex >= 0 && m.selectedIndex < len(m.agents) {
				keyHelp = append(keyHelp, "[c] continue")
			}
		}
	case ViewDetail:
		keyHelp = []string{"[o/c/l/n/a/t] tabs", "[j/k] navigate", "[h/l] scroll"}
		if m.tabs.ActiveIndex == 0 {
			// Back arrow selected
			keyHelp = append(keyHelp, "[enter/esc] back")
		}
	default:
		keyHelp = []string{"[j/k] navigate", "[enter] select"}
	}

	keyHelp = append(keyHelp, "[r] refresh", "[?] help", "[q] quit")
	return strings.Join(keyHelp, "  ")
}

// getRightStatus returns the right-aligned status information
func (m *Model) getRightStatus() string {
	// Sync status - right aligned
	right := ""
	if m.isLoading {
		right = m.styles.Warning.Render(" ⟳ Refreshing...")
	} else if m.err != nil {
		right = m.styles.Error.Render(" ✗ Error")
	} else {
		// Show last update time
		if !m.lastUpdate.IsZero() {
			elapsed := time.Since(m.lastUpdate)
			var updateStr string
			if elapsed < 1*time.Second {
				updateStr = "just now"
			} else if elapsed < 60*time.Second {
				updateStr = fmt.Sprintf("%ds ago", int(elapsed.Seconds()))
			} else if elapsed < 60*time.Minute {
				updateStr = fmt.Sprintf("%dm ago", int(elapsed.Minutes()))
			} else {
				updateStr = fmt.Sprintf("%dh ago", int(elapsed.Hours()))
			}
			right = m.styles.Subtle.Render(fmt.Sprintf(" Updated %s", updateStr))
		}
	}
	return right
}