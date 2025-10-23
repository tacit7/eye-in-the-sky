package modal

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpContent represents formatted help text
type HelpContent struct {
	Title   string
	Actions map[string][]string // action -> keys
}

// NewHelpFromKeybindings creates help content from a keybindings map
func NewHelpFromKeybindings(scope string, bindings map[string][]string) HelpContent {
	return HelpContent{
		Title:   fmt.Sprintf("%s Keybindings", strings.Title(scope)),
		Actions: bindings,
	}
}

// FormatHelp formats help content into a readable string
func (h HelpContent) Format() string {
	lines := []string{
		"",
		lipgloss.NewStyle().Bold(true).Render(h.Title),
		strings.Repeat("─", 40),
		"",
	}

	// Create two columns for keys
	maxActionLen := 0
	for action := range h.Actions {
		if len(action) > maxActionLen {
			maxActionLen = len(action)
		}
	}

	// Format each action and its keys
	for action, keys := range h.Actions {
		keyStr := strings.Join(keys, ", ")
		line := fmt.Sprintf("  %-*s  %s", maxActionLen, action, keyStr)
		lines = append(lines, line)
	}

	lines = append(lines, "")
	lines = append(lines, "[Esc] Close Help")

	return strings.Join(lines, "\n")
}

// OpenHelpForScope opens a help modal for a specific scope
func (m *Modal) OpenHelpForScope(scope string, bindings map[string][]string) {
	help := NewHelpFromKeybindings(scope, bindings)
	m.OpenHelp(help.Title, help.Format())
}
