package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const dbClickWindow = 250 * time.Millisecond

// handleListClick handles mouse clicks on the list view
func (m *Model) handleListClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Check if click is in tabs area
	tabsTop := m.listLayout.HeaderH
	tabsBottom := tabsTop + m.listLayout.TabsH
	if msg.Y >= tabsTop && msg.Y < tabsBottom {
		m.listTabs.Update(msg)
		return m, nil
	}

	// Check if click is in content area
	contentTop := m.listLayout.Content.Y
	contentBottom := contentTop + m.listLayout.Content.H
	if msg.Y < contentTop || msg.Y >= contentBottom {
		return m, nil // Outside content box
	}

	// Check horizontal bounds
	contentLeft := m.listLayout.Content.X
	contentRight := contentLeft + m.listLayout.Content.W
	if msg.X < contentLeft || msg.X >= contentRight {
		return m, nil // Outside content box
	}

	// Calculate row index within visible window
	rowInView := msg.Y - (contentTop + 1) // +1 for top border
	if rowInView < 0 || rowInView >= m.listLayout.Content.H-2 { // -2 for top and bottom borders
		return m, nil
	}

	// Map to actual row index
	rowIndex := m.listOffset + rowInView
	if rowIndex >= len(m.agents) {
		return m, nil
	}

	// Handle single vs double click
	now := time.Now()
	if rowIndex == m.lastClickRow && now.Sub(m.lastClickAt) <= dbClickWindow {
		// Double click - open detail view
		m.selectedIndex = rowIndex
		if err := m.loadAgentDetails(); err != nil {
			m.err = err
			return m, nil
		}
		m.currentView = ViewDetail
		m.detailOffset = 0
		m.tabs.Set(1)
		m.lastClickRow = -1 // Reset
		return m, nil
	}

	// Single click - select row
	m.selectedIndex = rowIndex
	m.lastClickAt = now
	m.lastClickRow = rowIndex
	return m, nil
}

// handleDetailClick handles mouse clicks on the detail view
func (m *Model) handleDetailClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Pass to tabs component which handles its own click detection
	m.tabs.Update(msg)

	// If user clicked on back arrow (index 0), go back to list
	if m.tabs.ActiveIndex == 0 {
		m.currentView = ViewList
		m.detailOffset = 0
		return m, nil
	}

	// Load data for the new tab
	if err := m.loadTabData(); err != nil {
		m.err = err
	}

	return m, nil
}
