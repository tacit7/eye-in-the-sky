package app

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current view
func (m *Model) View() string {
	start := time.Now()
	defer func(startTime time.Time) {
		elapsed := time.Since(startTime)
		endTime := time.Now().Format("15:04:05.000")
		if elapsed > 50*time.Millisecond {
			log.Printf("[PERF][%s] View() complete in %v (currentView=%v)", endTime, elapsed, m.currentView)
		}
	}(start)
	if m.width == 0 {
		return "Loading..."
	}

	// Render note modal if visible (highest priority)
	if m.noteModal.Visible {
		return m.noteModal.View()
	}

	// Render modal if active
	if m.modalManager.IsActive() {
		return m.renderModalOverlay()
	}

	// Render help overlay if shown
	if m.showHelp {
		return m.renderHelp()
	}

	// Use renderer map for view dispatching
	if fn, ok := m.renderers[m.currentView]; ok {
		return fn(m)
	}

	return "Unknown view"
}

// renderHeader renders the top header bar
func (m *Model) renderHeader() string {
	// Build title with version
	title := m.styles.Bold.Render(AppTitle)
	version := m.styles.Subtle.Render(" " + AppVersion)

	// Use Lipgloss to center the content
	content := lipgloss.JoinHorizontal(lipgloss.Top, title, version)

	// Calculate header width
	headerWidth := m.width - HeaderPadding
	if headerWidth < MinHeaderWidth {
		headerWidth = MinHeaderWidth
	}

	// Center the content horizontally
	centeredContent := lipgloss.PlaceHorizontal(
		headerWidth,
		lipgloss.Center,
		content,
	)

	// Apply header styling
	return m.styles.Header.
		Width(headerWidth).
		Render(centeredContent)
}

// renderFooter renders the bottom status bar
func (m *Model) renderFooter() string {
	footerWidth := m.width - FooterPadding

	// Build the three footer sections
	left := m.styles.KeyHelp.Render(m.getKeyHelp())
	middle := m.getStatusMessage()
	right := m.getRightStatus()

	// Use Lipgloss to place content horizontally
	footer := lipgloss.PlaceHorizontal(
		footerWidth,
		lipgloss.Center,
		middle,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(m.styles.Subtle.GetForeground()),
	)

	// Place left and right content
	if lipgloss.Width(left)+lipgloss.Width(right) < footerWidth {
		footer = lipgloss.JoinHorizontal(
			lipgloss.Top,
			left,
			lipgloss.PlaceHorizontal(
				footerWidth-lipgloss.Width(left)-lipgloss.Width(right),
				lipgloss.Center,
				middle,
			),
			right,
		)
	} else {
		// Not enough space, prioritize key help
		footer = left
	}

	// Apply footer styling
	return m.styles.Footer.
		Width(footerWidth).
		Render(footer)
}

// getStatusMessage returns the current status message if within timeout
func (m *Model) getStatusMessage() string {
	if m.statusMsg != "" && time.Since(m.statusTime) < StatusTimeout {
		return m.styles.StatusBar.Render(m.statusMsg)
	}
	return ""
}

// getKeyHelp returns context-sensitive key hints
func (m *Model) getKeyHelp() string {
	var keyHelp []string

	// Get context-specific keys
	switch m.currentView {
	case ViewList:
		keyHelp = m.getListKeyHelp()
	case ViewDetail:
		keyHelp = m.getDetailKeyHelp()
	default:
		keyHelp = m.getDefaultKeyHelp()
	}

	// Add common keys
	keyHelp = append(keyHelp, m.getCommonKeyHelp()...)
	return strings.Join(keyHelp, "  ")
}

// getListKeyHelp returns key help for list view
func (m *Model) getListKeyHelp() []string {
	switch m.listTabs.ActiveIndex {
	case 1: // Project tab
		return m.getProjectTabKeyHelp()
	default: // Overview and others
		keyHelp := []string{"[j/k] navigate", "[tab] switch tab", "[enter] details"}
		if m.selectedIndex >= 0 && m.selectedIndex < len(m.agents) {
			keyHelp = append(keyHelp, "[c] continue")
		}
		return keyHelp
	}
}

// getProjectTabKeyHelp returns key help for project tab navigation
func (m *Model) getProjectTabKeyHelp() []string {
	switch m.projectSection {
	case "tasks":
		return []string{"[j/k] navigate", "[enter] open task", "[d] mark done", "[←] back"}
	case "claude":
		return []string{"[j/k] scroll", "[←] back"}
	case "markdown":
		return []string{"[j/k] navigate files", "[←] back"}
	default:
		return []string{"[t] tasks", "[c] claude.md", "[m] markdown files", "[tab] back to overview"}
	}
}

// getDetailKeyHelp returns key help for detail view
func (m *Model) getDetailKeyHelp() []string {
	keyHelp := []string{"[o/c/l/n/a/t] tabs", "[j/k] navigate", "[h/l] scroll"}
	if m.tabs.ActiveIndex == 0 {
		// Back arrow selected
		keyHelp = append(keyHelp, "[enter/esc] back")
	}
	return keyHelp
}

// getDefaultKeyHelp returns default key help
func (m *Model) getDefaultKeyHelp() []string {
	return []string{"[j/k] navigate", "[enter] select"}
}

// getCommonKeyHelp returns keys that are always available
func (m *Model) getCommonKeyHelp() []string {
	return []string{"[r] refresh", "[?] help", "[q] quit"}
}

// getRightStatus returns the right-aligned status information
func (m *Model) getRightStatus() string {
	if m.isLoading {
		return m.styles.Warning.Render(" ⟳ Refreshing...")
	}

	// Use unified error status
	if errorStatus := m.getErrorStatus(); errorStatus != "" {
		return errorStatus
	}

	// Show last update time
	if !m.lastUpdate.IsZero() {
		updateStr := m.formatElapsedTime(time.Since(m.lastUpdate))
		return m.styles.Subtle.Render(fmt.Sprintf(" Updated %s", updateStr))
	}

	return ""
}

// formatElapsedTime formats an elapsed duration for display
func (m *Model) formatElapsedTime(elapsed time.Duration) string {
	switch {
	case elapsed < JustNowThreshold:
		return "just now"
	case elapsed < SecondsThreshold:
		return fmt.Sprintf("%ds ago", int(elapsed.Seconds()))
	case elapsed < MinutesThreshold:
		return fmt.Sprintf("%dm ago", int(elapsed.Minutes()))
	default:
		return fmt.Sprintf("%dh ago", int(elapsed.Hours()))
	}
}

// renderModalOverlay renders the modal on top of the current view
func (m *Model) renderModalOverlay() string {
	// Create modal overlay with semi-transparent background
	modalWidth := m.width - 4
	if modalWidth < 60 {
		modalWidth = 60
	}
	modalHeight := m.height - 4
	if modalHeight < 10 {
		modalHeight = 10
	}

	// Get modal content
	modalContent := m.modalManager.View()

	// Create a centered box with the modal content
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.styles.Primary.GetForeground()).
		Padding(1, 2).
		Width(modalWidth).
		MaxHeight(modalHeight)

	centeredModal := lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box.Render(modalContent),
	)

	return centeredModal
}