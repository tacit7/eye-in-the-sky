package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TabsModel handles which tab is active, rendering, and keyboard navigation
type TabsModel struct {
	Titles      []string
	ActiveIndex int
	StyleActive lipgloss.Style
	StyleNormal lipgloss.Style
}

// NewTabsModel creates a new tabs model
func NewTabsModel(titles []string, activeColor, normalColor string) TabsModel {
	return TabsModel{
		Titles:      titles,
		ActiveIndex: 0,
		StyleActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color(activeColor)).
			Background(lipgloss.Color("#6")).
			Bold(true).
			Padding(0, 2),
		StyleNormal: lipgloss.NewStyle().
			Foreground(lipgloss.Color(normalColor)).
			Padding(0, 2),
	}
}

// View renders the tabs
func (m TabsModel) View() string {
	var out []string
	for i, t := range m.Titles {
		if i == m.ActiveIndex {
			out = append(out, m.StyleActive.Render(t))
		} else {
			out = append(out, m.StyleNormal.Render(t))
		}
	}
	return strings.Join(out, " ")
}

// Next moves to the next tab
func (m *TabsModel) Next() {
	m.ActiveIndex = (m.ActiveIndex + 1) % len(m.Titles)
}

// Prev moves to the previous tab
func (m *TabsModel) Prev() {
	m.ActiveIndex = (m.ActiveIndex - 1 + len(m.Titles)) % len(m.Titles)
}

// Set sets the active tab by index
func (m *TabsModel) Set(index int) {
	if index >= 0 && index < len(m.Titles) {
		m.ActiveIndex = index
	}
}

// Update handles mouse clicks on tabs
func (m *TabsModel) Update(msg tea.Msg) {
	if mouse, ok := msg.(tea.MouseMsg); ok && mouse.Type == tea.MouseLeft {
		// Approximate click region widths (you can store actual positions if needed)
		width := 10 // assume fixed width per tab for simplicity
		m.ActiveIndex = mouse.X / (width + 2)
		if m.ActiveIndex >= len(m.Titles) {
			m.ActiveIndex = len(m.Titles) - 1
		}
		if m.ActiveIndex < 0 {
			m.ActiveIndex = 0
		}
	}
}
