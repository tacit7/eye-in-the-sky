package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// renderHelp renders the help overlay
func (m *Model) renderHelp() string {
	configDir, _ := ConfigDir()

	title := m.styles.Title.Render("Eye in the Sky - Help")

	// Get help content from help model
	helpView := m.help.View(m.keys)

	// Footer with config info
	configPath := fmt.Sprintf("Config: %s", configDir)
	version := "Version: 0.1.0"
	footer := m.styles.Subtle.Render(fmt.Sprintf("%s | %s", configPath, version))

	// Create bordered box
	helpContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		helpView,
		"",
		footer,
	)

	// Style the help box
	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.theme.Colors.Border)).
		Padding(1, 2).
		Width(m.width - 4).
		Render(helpContent)

	// Center the help box
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		helpBox,
	)
}