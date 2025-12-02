package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ui/theme"
)

// Dropdown represents a dropdown selection component
type Dropdown struct {
	Options   []string
	Selected  int
	Open      bool
	Width     int
	Style     lipgloss.Style
	Highlight lipgloss.Style
	Normal    lipgloss.Style
}

// NewDropdown creates a new dropdown component with theme styles
func NewDropdown(options []string) Dropdown {
	return Dropdown{
		Options:   options,
		Selected:  0,
		Open:      false,
		Width:     30,
		Style:     theme.Panel.Copy().Padding(0, 1),
		Highlight: theme.TextHighlight,
		Normal:    theme.TextMuted,
	}
}

// View renders the dropdown component
func (d Dropdown) View() string {
	current := d.Style.Render(d.Options[d.Selected])

	if !d.Open {
		return fmt.Sprintf("▼ %s", current)
	}

	var opts []string
	for i, opt := range d.Options {
		style := d.Normal
		if i == d.Selected {
			style = d.Highlight
			opt = "• " + opt
		} else {
			opt = "  " + opt
		}
		opts = append(opts, style.Render(opt))
	}

	return fmt.Sprintf("▲ %s\n%s", current, lipgloss.JoinVertical(lipgloss.Left, opts...))
}

// Update handles dropdown state updates
func (d Dropdown) Update(msg tea.Msg) (Dropdown, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			d.Open = !d.Open
		case "up", "k":
			if d.Open && d.Selected > 0 {
				d.Selected--
			}
		case "down", "j":
			if d.Open && d.Selected < len(d.Options)-1 {
				d.Selected++
			}
		case "g":
			if d.Open {
				d.Selected = 0 // Select Global
				d.Open = false
			}
		case "p":
			if d.Open && len(d.Options) > 1 {
				d.Selected = 1 // Select Project
				d.Open = false
			}
		case "a":
			if d.Open && len(d.Options) > 2 {
				d.Selected = 2 // Select Agent
				d.Open = false
			}
		case "esc":
			d.Open = false
		}
	}
	return d, nil
}

// SelectedValue returns the currently selected option
func (d Dropdown) SelectedValue() string {
	if d.Selected >= 0 && d.Selected < len(d.Options) {
		return d.Options[d.Selected]
	}
	return ""
}

// SetSelected sets the selected index
func (d *Dropdown) SetSelected(index int) {
	if index >= 0 && index < len(d.Options) {
		d.Selected = index
	}
}
