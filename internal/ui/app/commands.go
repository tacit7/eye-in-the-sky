package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/parser"
)

// archiveAgentCmd creates a command to archive an agent
func (m *Model) archiveAgentCmd(agentID string) tea.Cmd {
	return func() tea.Msg {
		// Update agent status to archived in database
		query := `UPDATE agents SET status = 'archived', updated_at = CURRENT_TIMESTAMP WHERE id = ?`
		if _, err := m.db.Exec(query, agentID); err != nil {
			return cmdResult{
				success: false,
				message: "Failed to archive agent",
				err:     err,
			}
		}

		return cmdResult{
			success: true,
			message: fmt.Sprintf("Archived agent %s", truncateID(agentID, 8)),
		}
	}
}

// initCCUsageCmd runs database initialization in background
func (m *Model) initCCUsageCmd() tea.Cmd {
	return func() tea.Msg {
		if m.ccusageDB == nil {
			return initCCUsageMsg{}
		}

		syncMgr := parser.NewSyncManager(m.ccusageDB)
		if err := syncMgr.Sync(); err != nil {
			m.ccusageSyncStatus = fmt.Sprintf("Error: %v", err)
			m.ccusageSyncing = false
			return initCCUsageMsg{}
		}

		// Reload data
		if err := m.loadCCUsageData(); err != nil {
			m.ccusageSyncStatus = fmt.Sprintf("Error loading data: %v", err)
		} else {
			m.ccusageSyncStatus = fmt.Sprintf("Initialized! Found %d entries", m.ccusageEntryCount)
		}

		m.ccusageSyncing = false
		return initCCUsageMsg{}
	}
}
