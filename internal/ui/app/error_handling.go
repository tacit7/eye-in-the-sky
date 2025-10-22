package app

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// renderErrorOverlay renders an error overlay on top of the current view
func (m *Model) renderErrorOverlay(baseView string) string {
	if m.err == nil {
		return baseView
	}

	errorBox := m.styles.ErrorBox.
		Width(StatusMessageWidth).
		Render(fmt.Sprintf("Error: %v", m.err))

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		errorBox,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(m.styles.Subtle.GetForeground()),
	)
}

// setError sets an error and optionally a status message
func (m *Model) setError(err error) {
	m.err = err
	if err != nil {
		m.statusMsg = fmt.Sprintf("Error: %v", err)
		m.statusTime = time.Now()
	}
}

// clearError clears the current error
func (m *Model) clearError() {
	m.err = nil
}

// hasError returns whether there is a current error
func (m *Model) hasError() bool {
	return m.err != nil
}

// getErrorStatus returns a formatted error status for the footer
func (m *Model) getErrorStatus() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf(" ✗ %v", m.err))
	}
	return ""
}