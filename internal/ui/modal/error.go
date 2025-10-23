package modal

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ErrorContent represents formatted error message
type ErrorContent struct {
	Title   string
	Message string
	Details string
}

// NewError creates a new error modal content
func NewError(title string, message string) ErrorContent {
	return ErrorContent{
		Title:   title,
		Message: message,
		Details: "",
	}
}

// NewErrorWithDetails creates an error with additional details
func NewErrorWithDetails(title string, message string, details string) ErrorContent {
	return ErrorContent{
		Title:   title,
		Message: message,
		Details: details,
	}
}

// FormatError formats error content into a readable string
func (e ErrorContent) Format() string {
	lines := []string{
		"",
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("ERROR"),
		"",
		lipgloss.NewStyle().Bold(true).Render(e.Title),
		"",
		e.Message,
	}

	if e.Details != "" {
		lines = append(lines, "")
		lines = append(lines, "Details:")
		lines = append(lines, "────────────────────────────")
		lines = append(lines, e.Details)
	}

	lines = append(lines, "")
	lines = append(lines, "[Esc] Close")

	return strings.Join(lines, "\n")
}

// OpenError opens an error modal with the given message
func (m *Modal) OpenErrorMsg(title string, message string) {
	err := NewError(title, message)
	m.OpenError(err.Title, err.Format())
}

// OpenErrorWithDetails opens an error modal with detailed information
func (m *Modal) OpenErrorWithDetails(title string, message string, details string) {
	err := NewErrorWithDetails(title, message, details)
	m.OpenError(err.Title, err.Format())
}

// IsErrorModal returns true if the current modal is an error
func (m *Modal) IsErrorModal() bool {
	return m.Type == ModalError
}

// DisplayCommandError opens an error modal for a command execution error
func (m *Modal) DisplayCommandError(command string, err error) {
	title := fmt.Sprintf("Command Failed: %s", command)
	message := "The command execution failed. See details below."
	details := err.Error()
	m.OpenErrorWithDetails(title, message, details)
}
