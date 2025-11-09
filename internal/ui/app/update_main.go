package app

import (
	"context"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tacit7/eye-in-the-sky/internal/ui/modal"
	"github.com/tacit7/eye-in-the-sky/internal/ui/util"
)

// handleListKeys handles keys in list view
func (m *Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Log all key presses for debugging
	debugf("Key pressed: %s", msg.String())

	// Set keybindings resolver context for list view tabs
	if m.keybindResolver != nil {
		var tabName string
		if m.listTabs.ActiveIndex < len(m.listTabs.Titles) {
			switch m.listTabs.ActiveIndex {
			case 0:
				tabName = "overview"
			case 1:
				tabName = "project"
			case 2:
				tabName = "claude"
			case 3:
				tabName = "usage"
			case 4:
				tabName = "config"
			}
		}
		if tabName != "" {
			debugf("Setting resolver context to list.%s", tabName)
			m.keybindResolver.SetContext("list", tabName)
		}
	} else {
		debugf("keybindResolver is NIL in handleListKeys!")
	}

	// Tab navigation (direct keys that don't go through resolver)
	switch msg.String() {
	case "tab", "right":
		m.listTabs.Next()
		return m, nil
	case "shift+tab", "left":
		m.listTabs.Prev()
		return m, nil
	case "i", "I":
		// Initialize CCUsage database (only in Usage tab when empty)
		if m.listTabs.ActiveIndex == 3 && m.ccusageDB != nil && m.ccusageEntryCount == 0 && !m.ccusageSyncing {
			m.ccusageSyncing = true
			m.ccusageSyncStatus = "Initializing database..."
			return m, tea.Batch(
				m.initCCUsageCmd(),
				m.tickCmd(),
			)
		}
		return m, nil
	}

	// Use resolver for tab-specific actions
	if m.keybindResolver != nil {
		action, found := m.keybindResolver.Resolve(msg, false)
		if found {
			switch action {
			case "new_session":
				// Open New Session form in Overview tab only
				if m.listTabs.ActiveIndex == 0 {
					formSpec := modal.NewSessionFormSpec()
					m.modalManager.OpenForm(formSpec)
					return m, nil
				}

			case "new_ticket":
				// Open New Ticket form in Project tab only
				if m.listTabs.ActiveIndex == 1 {
					var projectName string
					if m.projectInfo != nil {
						projectName = m.projectInfo.RepoName
					}
					formSpec := modal.NewTicketFormSpec(projectName)
					m.modalManager.OpenForm(formSpec)
					return m, nil
				}

			case "toggle_filter":
				m.showAll = !m.showAll
				return m, loadAgentsCmd(m.data.Agents, m.showAll)

			case "continue_session":
				agent := m.SelectedAgent()
				if agent == nil {
					m.statusMsg = "No agent selected"
					return m, nil
				}
				if m.claudePath == "" {
					m.statusMsg = "Claude binary not found"
					return m, nil
				}
				return m, ResumeSession(m.claudePath, string(agent.ID))

			case "start_session":
				agent := m.SelectedAgent()
				if agent == nil {
					m.statusMsg = "No agent selected"
					return m, nil
				}
				if m.claudePath == "" {
					m.statusMsg = "Claude binary not found"
					return m, nil
				}
				return m, StartSession(m.claudePath, string(agent.ID))

			case "go_to_window":
				agent := m.SelectedAgent()
				if agent == nil {
					m.statusMsg = "No agent selected"
					return m, nil
				}
				if agent.WindowID == "" {
					m.statusMsg = "No window ID for agent"
					return m, nil
				}
				return m, util.FocusWindowCmd(m.windowFocuser, agent.WindowID, agent.TerminalApplication)

			case "archive":
				agent := m.SelectedAgent()
				if agent == nil {
					m.statusMsg = "No agent selected"
					return m, nil
				}
				return m, m.archiveAgentCmd(string(agent.ID))
			}
		}
	}

	// Direct key handlers (not in resolver)
	switch msg.String() {
	case "d":
		// Mark session as done
		agent := m.SelectedAgent()
		if agent == nil {
			m.statusMsg = "No agent selected"
			return m, nil
		}
		ctx := context.Background()
		err := m.data.Agents.MarkComplete(ctx, agent.ID)
		if err != nil {
			m.statusMsg = "Failed to mark session complete: " + err.Error()
			return m, nil
		}
		m.statusMsg = "Session marked complete"
		// Reload agents to show updated status
		return m, loadAgentsCmd(m.data.Agents, m.showAll)
	}

	// Navigation for Overview tab (tab 0) using resolver
	if m.listTabs.ActiveIndex == 0 && m.keybindResolver != nil {
		action, found := m.keybindResolver.Resolve(msg, false)
		debugf("Overview: Key: %s, Action: %s, Found: %v", msg.String(), action, found)
		if found {
			switch action {
			case "down":
				debugf("[Overview] Moving down (index %d → %d of %d)", m.selectedIndex, m.selectedIndex+1, len(m.agents))
				if m.selectedIndex < len(m.agents)-1 {
					m.selectedIndex++
					m.adjustListScroll()
				}
				return m, nil

			case "up":
				debugf("[Overview] Moving up (index %d → %d)", m.selectedIndex, m.selectedIndex-1)
				if m.selectedIndex > 0 {
					m.selectedIndex--
					m.adjustListScroll()
				}
				return m, nil

			case "select":
				// Switch to detail view - load details with command
				if m.selectedIndex >= 0 && m.selectedIndex < len(m.agents) {
					m.selectedAgent = &m.agents[m.selectedIndex]
					m.currentView = ViewDetail
					m.detailOffset = 0
					// Set active tab to Overview (index 1, since 0 is back arrow)
					m.tabs.Set(1)
					return m, loadAgentDetailsCmd(m.data, m.selectedAgent.ID)
				}
				return m, nil
			}
		}
	}

	// Usage tab scrolling (only when Usage tab is active)
	if m.listTabs.ActiveIndex == 3 && m.keybindResolver != nil {
		action, found := m.keybindResolver.Resolve(msg, false)
		if found {
			switch action {
			case "down":
				m.usageViewport.LineDown(1)
				return m, nil
			case "up":
				m.usageViewport.LineUp(1)
				return m, nil
			case "page_down":
				m.usageViewport.HalfViewDown()
				return m, nil
			case "page_up":
				m.usageViewport.HalfViewUp()
				return m, nil
			}
		}
	}

	// Project tab navigation and scrolling (only when Project tab is active)
	if m.listTabs.ActiveIndex == 1 {
		keyStr := msg.String()

		// Section switching (direct keys)
		switch keyStr {
		case "1":
			m.projectSelectedSection = 0 // Tasks
			m.projectTasksIndex = 0
			return m, nil
		case "2":
			m.projectSelectedSection = 1 // CLAUDE.md
			return m, nil
		case "3":
			m.projectSelectedSection = 2 // Markdown files
			m.projectMDFilesIndex = 0
			return m, nil
		}

		// Use resolver for navigation actions
		if m.keybindResolver != nil {
			action, found := m.keybindResolver.Resolve(msg, false)
			if found {
				switch action {
				case "new_ticket":
					var projectName string
					if m.projectInfo != nil {
						projectName = m.projectInfo.RepoName
					}
					formSpec := modal.NewTicketFormSpec(projectName)
					m.modalManager.OpenForm(formSpec)
					return m, nil

				case "down":
					// Navigate within section
					switch m.projectSelectedSection {
					case 0: // Tasks
						if m.projectTasksIndex < len(m.projectTasks)-1 {
							m.projectTasksIndex++
						}
					case 2: // Markdown files
						if m.projectMDFilesIndex < len(m.projectMDFiles)-1 {
							m.projectMDFilesIndex++
						}
					}
					return m, nil

				case "up":
					// Navigate within section
					switch m.projectSelectedSection {
					case 0: // Tasks
						if m.projectTasksIndex > 0 {
							m.projectTasksIndex--
						}
					case 2: // Markdown files
						if m.projectMDFilesIndex > 0 {
							m.projectMDFilesIndex--
						}
					}
					return m, nil
				}
			}
		}
	}

	// Claude tab navigation and operations (only when Claude tab is active)
	if m.listTabs.ActiveIndex == 2 && m.keybindResolver != nil {
		action, found := m.keybindResolver.Resolve(msg, false)
		if found {
			switch action {
			case "down":
				// Navigate file list down
				if m.claudeSelectedIndex < len(m.claudeFiles)-1 {
					m.claudeSelectedIndex++
					m.claudeFilesViewport.LineDown(1)

					// Auto-load content for selected file or directory
					if m.claudeSelectedIndex < len(m.claudeFiles) {
						selectedFile := m.claudeFiles[m.claudeSelectedIndex]
						if !selectedFile.IsParent {
							if selectedFile.IsDir {
								return m, m.loadClaudeDirectoryContentCmd(selectedFile.Path)
							} else {
								return m, loadClaudeFileContentCmd(selectedFile.Path)
							}
						}
					}
				}
				return m, nil

			case "up":
				// Navigate file list up
				if m.claudeSelectedIndex > 0 {
					m.claudeSelectedIndex--
					m.claudeFilesViewport.LineUp(1)

					// Auto-load content for selected file or directory
					if m.claudeSelectedIndex < len(m.claudeFiles) {
						selectedFile := m.claudeFiles[m.claudeSelectedIndex]
						if !selectedFile.IsParent {
							if selectedFile.IsDir {
								return m, m.loadClaudeDirectoryContentCmd(selectedFile.Path)
							} else {
								return m, loadClaudeFileContentCmd(selectedFile.Path)
							}
						}
					}
				}
				return m, nil

			case "open":
				// Handle file/directory selection
				if m.claudeSelectedIndex >= 0 && m.claudeSelectedIndex < len(m.claudeFiles) {
					selectedFile := m.claudeFiles[m.claudeSelectedIndex]

					// Handle parent directory (..)
					if selectedFile.IsParent {
						// Navigate up one level
						if m.claudeCurrentPath != "" {
							// Remove last path component
							m.claudeCurrentPath = filepath.Dir(m.claudeCurrentPath)
							if m.claudeCurrentPath == "." {
								m.claudeCurrentPath = ""
							}
						}
						m.claudeShowingContent = false
						m.claudeContent = ""
						m.claudeSelectedIndex = 0
						m.statusMsg = "Navigated up"
						return m, m.loadClaudeFilesCmd()
					}

					// Handle directories
					if selectedFile.IsDir {
						// Navigate into directory
						if m.claudeCurrentPath != "" {
							m.claudeCurrentPath = filepath.Join(m.claudeCurrentPath, selectedFile.Name)
						} else {
							m.claudeCurrentPath = selectedFile.Name
						}
						m.claudeShowingContent = false
						m.claudeContent = ""
						m.claudeSelectedIndex = 0
						m.statusMsg = "Entered directory"
						return m, m.loadClaudeFilesCmd()
					}

					// Handle files
					m.claudeShowingContent = true
					return m, loadClaudeFileContentCmd(selectedFile.Path)
				}
				return m, nil

			case "validate":
				// Validate JSON
				if m.claudeShowingContent && m.claudeContent != "" {
					return m, validateClaudeFileCmd(m.claudeContent)
				}
				return m, nil

			case "edit":
				// Open in editor (only for files, not directories)
				if m.claudeSelectedIndex >= 0 && m.claudeSelectedIndex < len(m.claudeFiles) {
					selectedFile := m.claudeFiles[m.claudeSelectedIndex]
					if !selectedFile.IsDir {
						return m, openClaudeFileInEditor(selectedFile.Path)
					}
				}
				return m, nil

			case "refresh":
				// Refresh file list
				m.statusMsg = "Refreshing Claude config files..."
				return m, m.loadClaudeFilesCmd()

			case "page_down":
				if m.claudeShowingContent {
					var cmd tea.Cmd
					m.claudeViewport, cmd = m.claudeViewport.Update(msg)
					return m, cmd
				}
				return m, nil

			case "page_up":
				if m.claudeShowingContent {
					var cmd tea.Cmd
					m.claudeViewport, cmd = m.claudeViewport.Update(msg)
					return m, cmd
				}
				return m, nil
			}
		}

		// Handle esc/q to go back from content view (direct keys)
		if msg.String() == "esc" || msg.String() == "q" {
			if m.claudeShowingContent {
				m.claudeShowingContent = false
				m.claudeContent = ""
				return m, nil
			}
			return m, nil
		}
	}

	// Config tab navigation and operations (only when Config tab is active)
	if m.listTabs.ActiveIndex == 4 && m.keybindResolver != nil {
		keyStr := msg.String()

		// Direct key handling for edit mode toggle and special keys
		if keyStr == "e" || keyStr == "E" {
			// Toggle edit mode
			if !m.keybindingsEditing {
				m.keybindingsEditing = true
				m.keybindingsEditBuf = m.keybindingsYAML
				m.keybindingsModified = false
				m.statusMsg = "Edit mode on. Press Ctrl+S to save, Esc to cancel."
			} else {
				m.keybindingsEditing = false
				m.keybindingsEditBuf = ""
				m.keybindingsModified = false
				m.statusMsg = "Edit mode cancelled."
			}
			return m, nil
		}

		if keyStr == "ctrl+s" {
			// Save changes
			if m.keybindingsEditing && m.keybindingsModified {
				return m, m.saveKeybindingsCmd(m.keybindingsEditBuf)
			}
			m.statusMsg = "No changes to save."
			return m, nil
		}

		if keyStr == "esc" {
			// Cancel edit mode
			if m.keybindingsEditing {
				m.keybindingsEditing = false
				m.keybindingsEditBuf = ""
				m.keybindingsModified = false
				m.statusMsg = "Edit mode cancelled."
				return m, nil
			}
			return m, nil
		}

		// Use resolver for navigation actions when not in edit mode
		if !m.keybindingsEditing {
			action, found := m.keybindResolver.Resolve(msg, false)
			if found {
				switch action {
				case "refresh":
					// Reload keybindings from file
					return m, m.reloadKeybindingsCmd()

				case "down":
					// Scroll down
					var cmd tea.Cmd
					m.keybindingsViewport, cmd = m.keybindingsViewport.Update(msg)
					return m, cmd

				case "up":
					// Scroll up
					var cmd tea.Cmd
					m.keybindingsViewport, cmd = m.keybindingsViewport.Update(msg)
					return m, cmd

				case "page_down":
					// Page down
					var cmd tea.Cmd
					m.keybindingsViewport, cmd = m.keybindingsViewport.Update(msg)
					return m, cmd

				case "page_up":
					// Page up
					var cmd tea.Cmd
					m.keybindingsViewport, cmd = m.keybindingsViewport.Update(msg)
					return m, cmd
				}
			}
		}

		// In edit mode, handle text input
		if m.keybindingsEditing {
			// Handle character input
			if len(msg.String()) == 1 && msg.Runes[0] >= 32 && msg.Runes[0] <= 126 {
				// Printable character
				m.keybindingsEditBuf += msg.String()
				m.keybindingsModified = true
				return m, nil
			}

			// Handle special keys
			switch msg.String() {
			case "enter":
				m.keybindingsEditBuf += "\n"
				m.keybindingsModified = true
				return m, nil
			case "backspace":
				if len(m.keybindingsEditBuf) > 0 {
					m.keybindingsEditBuf = m.keybindingsEditBuf[:len(m.keybindingsEditBuf)-1]
					m.keybindingsModified = true
				}
				return m, nil
			case "tab":
				m.keybindingsEditBuf += "\t"
				m.keybindingsModified = true
				return m, nil
			}
		}
	}

	// Fallback keys for agent list navigation (not in resolver yet)
	if m.listTabs.ActiveIndex == 0 {
		switch msg.String() {
		case "g":
			// Go to top
			m.selectedIndex = 0
			m.listOffset = 0
			return m, nil
		case "G":
			// Go to bottom
			if len(m.agents) > 0 {
				m.selectedIndex = len(m.agents) - 1
				m.adjustListScroll()
			}
			return m, nil
		}
	}

	return m, nil
}
