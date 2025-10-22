package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TabZone defines the clickable region for a tab
type TabZone struct {
	StartX int
	EndX   int
	Title  string
}

// TabsModel handles which tab is active, rendering, and keyboard navigation
type TabsModel struct {
	Titles      []string
	ActiveIndex int
	Zones       []TabZone
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
			Bold(true).
			Underline(true).
			Padding(0, 2),
		StyleNormal: lipgloss.NewStyle().
			Foreground(lipgloss.Color(normalColor)).
			Padding(0, 2),
	}
}

// View renders the tabs and calculates click zones
func (m *TabsModel) View() string {
	var out []string
	m.Zones = make([]TabZone, 0)
	cursor := 0

	for i, t := range m.Titles {
		var rendered string
		if i == m.ActiveIndex {
			rendered = m.StyleActive.Render(t)
		} else {
			rendered = m.StyleNormal.Render(t)
		}
		width := lipgloss.Width(rendered)
		m.Zones = append(m.Zones, TabZone{
			StartX: cursor,
			EndX:   cursor + width,
			Title:  t,
		})
		cursor += width + 1 // +1 for space between tabs
		out = append(out, rendered)
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

// Update handles mouse clicks on tabs using zone detection
func (m *TabsModel) Update(msg tea.Msg) {
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
func (m *TabsModel) Render() string {
	return m.View()
}
