package overview

import tea "github.com/charmbracelet/bubbletea"

// ViewportUpdateMsg is sent when a viewport needs to process a Bubble Tea message
// Root Model will forward this to the appropriate viewport
type ViewportUpdateMsg struct {
	Target string      // "usage", "claude", "keybindings", or "project"
	Msg    tea.Msg     // The message to forward (typically tea.KeyMsg or tea.WindowSizeMsg)
}

// TabChangedMsg is sent when the active tab changes
// Root Model can use this to trigger viewport refreshes
type TabChangedMsg struct {
	NewTabIndex int // The new active tab index
}
