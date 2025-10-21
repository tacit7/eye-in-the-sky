package app

import (
	"log"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const dbClickWindow = 250 * time.Millisecond

// handleListClick handles mouse clicks on the list view
func (m *Model) handleListClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Check if click is in tabs area (only vertical bounds - tabs fill entire width)
	tabsTop := m.listLayout.HeaderH
	tabsBottom := tabsTop + m.listLayout.TabsH
	if msg.Y >= tabsTop && msg.Y < tabsBottom {
		// Click is in tabs area - check horizontal bounds
		// Tabs are rendered at X=0 and span the width
		if msg.X >= 0 && msg.X < m.width {
			m.listTabs.Update(msg)
		}
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
	// Account for header row (2 lines: header + separator)
	rowInView := msg.Y - (contentTop + 1) // +1 for top border
	headerRows := 2 // header + separator line
	if rowInView < headerRows {
		return m, nil // Click on header
	}

	dataRowIndex := rowInView - headerRows
	if dataRowIndex < 0 || dataRowIndex >= m.listLayout.Content.H-2-headerRows { // -2 for borders
		return m, nil
	}

	// Map to actual row index
	rowIndex := m.listOffset + dataRowIndex
	if rowIndex >= len(m.agents) {
		return m, nil
	}

	// Check if click is on session ID column (right side)
	// Session ID is positioned on the far right, roughly the last 15 chars
	sessionIDColStart := contentRight - 15
	if msg.X >= sessionIDColStart && msg.X < contentRight {
		agent := m.agents[rowIndex]
		if agent.CurrentSessionID != "" {
			copyToClipboard(agent.CurrentSessionID)
			m.statusMsg = "Session ID copied to clipboard"
		}
		return m, nil
	}

	// Handle click based on active tab
	now := time.Now()

	// On Overview tab (index 0), single click opens detail view
	if m.listTabs.ActiveIndex == 0 {
		m.selectedIndex = rowIndex
		if err := m.loadAgentDetails(); err != nil {
			m.err = err
			return m, nil
		}
		m.currentView = ViewDetail
		m.detailOffset = 0
		m.tabs.Set(1)
		return m, nil
	}

	// On other tabs, use double-click to open detail view
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

// copyToClipboard copies text to the system clipboard
func copyToClipboard(text string) {
	cmd := exec.Command("pbcopy")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("Error copying to clipboard: %v", err)
		return
	}

	go func() {
		defer stdin.Close()
		stdin.Write([]byte(text))
	}()

	if err := cmd.Run(); err != nil {
		log.Printf("Error running pbcopy: %v", err)
	}
}
