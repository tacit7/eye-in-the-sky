package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ModalOptions configures the modal component
type ModalOptions struct {
	Title   string
	Width   int
	Height  int
	Content string
	OnClose func()
}

// Modal represents a simple display modal component
type Modal struct {
	Visible bool
	Title   string
	Content string
	Width   int
	Height  int
	OnClose func()
}

// NewModal creates a new modal component
func NewModal(opts ModalOptions) Modal {
	return Modal{
		Title:   opts.Title,
		Width:   opts.Width,
		Height:  opts.Height,
		Content: opts.Content,
		OnClose: opts.OnClose,
	}
}

// Init initializes the modal
func (m Modal) Init() tea.Cmd { return nil }

// Update handles modal updates
func (m Modal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.Visible {
			return m, nil
		}
		switch msg.String() {
		case "esc", "q":
			m.Visible = false
			if m.OnClose != nil {
				m.OnClose()
			}
		}
	}
	return m, nil
}

// View renders the modal
func (m Modal) View() string {
	if !m.Visible {
		return ""
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(m.Width).
		Height(m.Height).
		Padding(1, 2)

	title := lipgloss.NewStyle().Bold(true).Render(m.Title)
	body := lipgloss.NewStyle().MarginTop(1).Render(m.Content)

	return box.Render(title + "\n" + body)
}

// Show displays the modal with content
func (m *Modal) Show(content string) {
	m.Content = content
	m.Visible = true
}

// Hide hides the modal
func (m *Modal) Hide() {
	m.Visible = false
}
