package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NavBarZone defines the clickable region for a nav bar item
type NavBarZone struct {
	StartX int
	EndX   int
	Title  string
}

// NavBar handles which item is active, rendering, and keyboard navigation
type NavBar struct {
	Titles      []string
	ActiveIndex int
	Zones       []NavBarZone
	StyleActive lipgloss.Style
	StyleNormal lipgloss.Style
}

// NewNavBar creates a new navigation bar
func NewNavBar(titles []string, activeColor, normalColor string) NavBar {
	return NavBar{
		Titles:      titles,
		ActiveIndex: 0,
		StyleActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color(activeColor)).
			Bold(true).
			Underline(true).
			Padding(0, 2),
		StyleNormal: lipgloss.NewStyle().
			Foreground(lipgloss.Color(normalColor)).
			Padding(0, 2),
	}
}

// View renders the nav bar and calculates click zones
func (m *NavBar) View() string {
	var out []string
	m.Zones = make([]NavBarZone, 0)
	cursor := 0

	for i, t := range m.Titles {
		var rendered string
		if i == m.ActiveIndex {
			rendered = m.StyleActive.Render(t)
		} else {
			rendered = m.StyleNormal.Render(t)
		}
		width := lipgloss.Width(rendered)
		m.Zones = append(m.Zones, NavBarZone{
			StartX: cursor,
			EndX:   cursor + width,
			Title:  t,
		})
		cursor += width + 1 // +1 for space between items
		out = append(out, rendered)
	}

	return strings.Join(out, " ")
}

// Next moves to the next item
func (m *NavBar) Next() {
	m.ActiveIndex = (m.ActiveIndex + 1) % len(m.Titles)
}

// Prev moves to the previous item
func (m *NavBar) Prev() {
	m.ActiveIndex = (m.ActiveIndex - 1 + len(m.Titles)) % len(m.Titles)
}

// Set sets the active item by index
func (m *NavBar) Set(index int) {
	if index >= 0 && index < len(m.Titles) {
		m.ActiveIndex = index
	}
}

// Update handles mouse clicks on nav bar items using zone detection
func (m *NavBar) Update(msg tea.Msg) {
	if mouse, ok := msg.(tea.MouseMsg); ok && mouse.Type == tea.MouseLeft {
		for i, z := range m.Zones {
			if mouse.X >= z.StartX && mouse.X < z.EndX {
				m.ActiveIndex = i
				break
			}
		}
	}
}

// Render is an alias for View() for backwards compatibility
func (m *NavBar) Render() string {
	return m.View()
}
