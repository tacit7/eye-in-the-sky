package modal

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// HelpContent represents formatted help text
type HelpContent struct {
	Title   string
	Actions map[string][]string // action -> keys
	Scope   string
}

// NewHelpFromKeybindings creates help content from a keybindings map
func NewHelpFromKeybindings(scope string, bindings map[string][]string) HelpContent {
	return HelpContent{
		Title:   formatScopeTitle(scope),
		Actions: bindings,
		Scope:   scope,
	}
}

// formatScopeTitle converts scope name to title case
func formatScopeTitle(scope string) string {
	switch scope {
	case "global":
		return "Global Keybindings"
	case "modal":
		return "Modal Keybindings"
	case "list_overview":
		return "Overview Tab Keybindings"
	case "list_project":
		return "Project Tab Keybindings"
	case "list_claude":
		return "Claude Config Tab Keybindings"
	case "list_usage":
		return "Token Usage Tab Keybindings"
	case "detail_overview":
		return "Agent Overview Tab Keybindings"
	case "detail_commits":
		return "Commits Tab Keybindings"
	case "detail_logs":
		return "Logs Tab Keybindings"
	case "detail_notes":
		return "Notes Tab Keybindings"
	case "detail_actions":
		return "Actions Tab Keybindings"
	case "detail_tasks":
		return "Tasks Tab Keybindings"
	default:
		return "Keybindings: " + scope
	}
}

// FormatHelp formats help content into a readable two-column string
func (h HelpContent) Format() string {
	if len(h.Actions) == 0 {
		return "No keybindings for this scope.\n\n[Esc] Close Help"
	}

	lines := []string{
		"",
		lipgloss.NewStyle().Bold(true).Render(h.Title),
		strings.Repeat("─", 60),
		"",
	}

	// Sort actions for consistent display
	actions := make([]string, 0, len(h.Actions))
	for action := range h.Actions {
		actions = append(actions, action)
	}
	sort.Strings(actions)

	// Find max widths for alignment
	maxActionLen := 0
	maxKeyLen := 0
	for _, action := range actions {
		if len(action) > maxActionLen {
			maxActionLen = len(action)
		}
		keyStr := strings.Join(h.Actions[action], ", ")
		if len(keyStr) > maxKeyLen {
			maxKeyLen = len(keyStr)
		}
	}

	// Split into two columns if many actions
	columnWidth := 70 // Total width for two columns
	col1Width := (columnWidth / 2) - 2

	var col1, col2 []string

	for i, action := range actions {
		keyStr := strings.Join(h.Actions[action], ", ")
		line := fmt.Sprintf("  %-*s  %s", maxActionLen, action, keyStr)

		if i%2 == 0 {
			col1 = append(col1, line)
		} else {
			col2 = append(col2, line)
		}
	}

	// Combine columns
	maxRows := len(col1)
	if len(col2) > maxRows {
		maxRows = len(col2)
	}

	for i := 0; i < maxRows; i++ {
		if i < len(col1) && i < len(col2) {
			// Both columns have content at this row
			line := fmt.Sprintf("%-*s%s", col1Width, col1[i], col2[i])
			lines = append(lines, line)
		} else if i < len(col1) {
			// Only col1 has content
			lines = append(lines, col1[i])
		} else if i < len(col2) {
			// Only col2 has content
			lines = append(lines, col2[i])
		}
	}

	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Italic(true).Render("[Esc] Close  [j/k or ↓/↑] Scroll"))

	return strings.Join(lines, "\n")
}

// HelpModal extends Modal with help-specific viewport for scrolling
type HelpModal struct {
	Modal
	viewport viewport.Model
	content  string
}

// OpenHelpForScope opens a help modal for a specific scope
func (m *Modal) OpenHelpForScope(scope string, bindings map[string][]string) {
	help := NewHelpFromKeybindings(scope, bindings)
	formatted := help.Format()

	m.Active = true
	m.Type = ModalHelp
	m.Title = help.Title
	m.Content = formatted
}

// UpdateHelpViewport updates the help modal viewport with current dimensions
func (m *Modal) UpdateHelpViewport(width int, height int) {
	// Reserve space for borders, title, and footer
	availWidth := width - 4  // Account for left/right borders
	availHeight := height - 6 // Account for title + footer + padding
	if availHeight < 5 {
		availHeight = 5
	}
	if availWidth < 40 {
		availWidth = 40
	}

	m.helpViewport.Width = availWidth
	m.helpViewport.Height = availHeight
	m.helpViewport.SetContent(m.Content)
}

// HandleHelpScroll handles scrolling within help modal
func (m *Modal) HandleHelpScroll(direction string) {
	switch direction {
	case "down", "j":
		m.helpViewport.LineDown(1)
	case "up", "k":
		m.helpViewport.LineUp(1)
	case "page_down":
		m.helpViewport.HalfPageDown()
	case "page_up":
		m.helpViewport.HalfPageUp()
	}
}
