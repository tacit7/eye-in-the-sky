package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ButtonMsg is the message emitted when a button is pressed.
type ButtonMsg struct {
	ID string
}

// Button represents a simple clickable/focusable button component.
type Button struct {
	ID       string
	Label    string
	Focused  bool
	Disabled bool

	// Visual styles
	styleNormal   lipgloss.Style
	styleFocused  lipgloss.Style
	styleDisabled lipgloss.Style
}

// ButtonOptions lets you customize colors and styles when constructing a button.
type ButtonOptions struct {
	ID       string
	Label    string
	Normal   lipgloss.Style
	Focused  lipgloss.Style
	Disabled lipgloss.Style
}

// NewButton creates a new button with default or custom styles.
func NewButton(opts ButtonOptions) Button {
	// Fallback defaults
	defaultNormal := lipgloss.NewStyle().
		Padding(0, 2).
		Background(lipgloss.Color("#333333")).
		Foreground(lipgloss.Color("#CCCCCC")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444444"))

	defaultFocused := defaultNormal.Copy().
		Background(lipgloss.Color("#00ADD8")).
		Foreground(lipgloss.Color("#FFFFFF")).
		BorderForeground(lipgloss.Color("#00ADD8"))

	defaultDisabled := defaultNormal.Copy().
		Foreground(lipgloss.Color("#777777")).
		Background(lipgloss.Color("#222222"))

	b := Button{
		ID:            opts.ID,
		Label:         opts.Label,
		styleNormal:   mergeStyle(opts.Normal, defaultNormal),
		styleFocused:  mergeStyle(opts.Focused, defaultFocused),
		styleDisabled: mergeStyle(opts.Disabled, defaultDisabled),
	}
	return b
}

// mergeStyle returns opts if set, otherwise def.
func mergeStyle(opts, def lipgloss.Style) lipgloss.Style {
	if opts.String() != "" {
		return opts
	}
	return def
}

// Init initializes the button
func (b Button) Init() tea.Cmd { return nil }

// Update handles button updates
func (b Button) Update(msg tea.Msg) (Button, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if b.Disabled {
			return b, nil
		}
		if b.Focused && (msg.String() == "enter" || msg.String() == " ") {
			return b, func() tea.Msg { return ButtonMsg{ID: b.ID} }
		}
	}
	return b, nil
}

// View renders the button
func (b Button) View() string {
	if b.Disabled {
		return b.styleDisabled.Render(b.Label)
	}
	if b.Focused {
		return b.styleFocused.Render(b.Label)
	}
	return b.styleNormal.Render(b.Label)
}
