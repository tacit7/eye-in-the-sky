package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tacit7/eye-in-the-sky/internal/ui/config"
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

// tabBorderWithBottom creates custom tab borders that connect to content
func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	if !config.UseNerdFonts {
		border = lipgloss.Border{
			Top:          "-",
			Bottom:       middle,
			Left:         "|",
			Right:        "|",
			TopLeft:      "+",
			TopRight:     "+",
			BottomLeft:   left,
			BottomRight:  right,
		}
	} else {
		border.BottomLeft = left
		border.Bottom = middle
		border.BottomRight = right
	}
	return border
}

// NewNavBar creates a new navigation bar
func NewNavBar(titles []string, activeColor, normalColor string) NavBar {
	highlightColor := lipgloss.Color("#00ADD8")
	inactiveTabBorder := tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder := tabBorderWithBottom("┘", " ", "└")

	return NavBar{
		Titles:      titles,
		ActiveIndex: 0,
		StyleActive: lipgloss.NewStyle().
			Border(activeTabBorder, true).
			BorderForeground(highlightColor).
			Padding(0, 1),
		StyleNormal: lipgloss.NewStyle().
			Border(inactiveTabBorder, true).
			BorderForeground(highlightColor).
			Padding(0, 1),
	}
}

// View renders the nav bar and calculates click zones
func (m *NavBar) View() string {
	var renderedTabs []string
	m.Zones = make([]NavBarZone, 0)
	cursor := 0

	for i, t := range m.Titles {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.Titles)-1, i == m.ActiveIndex

		if isActive {
			style = m.StyleActive
		} else {
			style = m.StyleNormal
		}

		// Get border and adjust corners for seamless connection
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)

		rendered := style.Render(t)
		width := lipgloss.Width(rendered)
		m.Zones = append(m.Zones, NavBarZone{
			StartX: cursor,
			EndX:   cursor + width,
			Title:  t,
		})
		cursor += width
		renderedTabs = append(renderedTabs, rendered)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
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
