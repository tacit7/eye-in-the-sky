package dashboard

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jroimartin/gocui"
	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/window"
)

// continueSession opens a new shell and runs claude -c with the session ID
func (a *App) continueSession(g *gocui.Gui, v *gocui.View) error {
	if a.selectedIdx < 0 || a.selectedIdx >= len(a.agents) {
		return nil
	}

	agent := a.agents[a.selectedIdx]

	// Get session ID
	sessionID := ""
	if agent.CurrentSessionID != nil {
		sessionID = *agent.CurrentSessionID
	} else {
		return fmt.Errorf("no session ID for agent %s", agent.ID)
	}

	// Suspend GUI to run external command
	a.gui.Close()

	// Run claude -c command
	cmd := exec.Command("claude", "-c", sessionID)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("\n\033[1;36m→ Continuing session: %s\033[0m\n", sessionID)
	if err := cmd.Run(); err != nil {
		fmt.Printf("\033[1;31m✗ Failed to continue session: %v\033[0m\n", err)
	}

	fmt.Println("\n\033[1;32m✓ Session ended. Press Enter to return to dashboard...\033[0m")
	fmt.Scanln()

	// Reinitialize GUI
	newGui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	a.gui = newGui
	a.gui.SetManagerFunc(a.layout)
	a.gui.Cursor = true
	a.gui.Mouse = false

	// Reapply keybindings
	if err := a.setupKeybindings(); err != nil {
		return err
	}

	// Refresh data
	return a.refreshAgents()
}

// goToWindow brings the agent's window to front using AppleScript
func (a *App) goToWindow(g *gocui.Gui, v *gocui.View) error {
	if a.selectedIdx < 0 || a.selectedIdx >= len(a.agents) {
		return nil
	}

	agent := a.agents[a.selectedIdx]

	// Check if agent has window ID
	if agent.WindowID == nil || *agent.WindowID == "" {
		return a.showMessage("No window ID for this agent")
	}

	// Use window manager to bring window to front
	wm := window.NewManager()

	// Determine application based on source
	application := "Claude Desktop"
	if agent.Source == "worktree" {
		application = "Ghostty" // Default terminal for worktree agents
	}

	if err := wm.BringToFront(application, ""); err != nil {
		return a.showMessage(fmt.Sprintf("Failed to bring window to front: %v", err))
	}

	return a.showMessage("Window brought to front")
}

// showMessage displays a temporary message to the user
func (a *App) showMessage(msg string) error {
	// For now, just update the status bar
	// TODO: Implement popup message view
	v, err := a.gui.View(viewStatus)
	if err != nil {
		return err
	}

	v.Clear()
	fmt.Fprintf(v, " %s", msg)
	return nil
}

// showLogs shows the logs view (placeholder for now)
func (a *App) showLogs(g *gocui.Gui, v *gocui.View) error {
	if a.selectedIdx < 0 || a.selectedIdx >= len(a.agents) {
		return nil
	}

	agent := a.agents[a.selectedIdx]

	// Get actions/logs for this agent
	actions, err := a.db.ListActions(agent.ID)
	if err != nil {
		return a.showMessage(fmt.Sprintf("Failed to load logs: %v", err))
	}

	// Create or update logs view
	maxX, maxY := g.Size()
	if v, err := g.SetView(viewLogs, maxX/4, maxY/4, 3*maxX/4, 3*maxY/4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = fmt.Sprintf(" Logs - %s ", agent.ID[:8])
		v.Wrap = true
		v.Autoscroll = false

		// Write logs
		for _, action := range actions {
			fmt.Fprintf(v, "[%s] %s: %s\n",
				action.Timestamp.Format("15:04:05"),
				action.ActionType,
				action.Description)
		}

		// Set as current view
		if _, err := g.SetCurrentView(viewLogs); err != nil {
			return err
		}

		// Delete any existing keybinding and add close keybinding
		g.DeleteKeybinding(viewLogs, 'q', gocui.ModNone)
		if err := g.SetKeybinding(viewLogs, 'q', gocui.ModNone, a.closeLogs); err != nil {
			return err
		}
	}

	return nil
}

// closeLogs closes the logs view
func (a *App) closeLogs(g *gocui.Gui, v *gocui.View) error {
	// Delete keybinding
	g.DeleteKeybinding(viewLogs, 'q', gocui.ModNone)

	if err := g.DeleteView(viewLogs); err != nil {
		return err
	}
	if _, err := g.SetCurrentView(viewMain); err != nil {
		return err
	}
	return nil
}

// viewAgentDetails displays detailed information about an agent
func (a *App) viewAgentDetails(agentID string) error {
	// Store the current agent ID for refresh
	a.currentAgentID = agentID

	// Get agent details
	agent, err := a.db.GetAgent(agentID)
	if err != nil {
		return a.showMessage(fmt.Sprintf("Failed to get agent: %v", err))
	}

	// Get session if available
	var session *database.Session
	if agent.CurrentSessionID != nil && *agent.CurrentSessionID != "" {
		session, err = a.db.GetSession(*agent.CurrentSessionID)
		if err != nil {
			// Session error is not critical, continue without it
			session = nil
		}
	}

	// Get logs for the session (limit to 20)
	var logs []*database.Log
	if session != nil {
		logs, err = a.db.GetLogs(session.ID)
		if err != nil {
			// Log error is not critical, continue without logs
			logs = nil
		}
		// Limit to 20 most recent
		if len(logs) > 20 {
			logs = logs[:20]
		}
	}

	// Get actions for the agent (limit to 20)
	actions, err := a.db.ListActions(agent.ID)
	if err != nil {
		// Action error is not critical, continue without actions
		actions = nil
	}
	// Limit to 20 most recent
	if len(actions) > 20 {
		actions = actions[:20]
	}

	// Get compactions for the agent
	compactions, err := a.db.GetCompactionsForAgent(agent.ID)
	if err != nil {
		// Compaction error is not critical, continue without compactions
		compactions = nil
	}

	// Get initial context from persona if available
	var initialContext string
	if agent.PersonaID != nil && *agent.PersonaID != "" {
		persona, err := a.db.GetPersona(*agent.PersonaID)
		if err == nil && persona != nil {
			initialContext = persona.InitialContext
		}
	}

	// Render the detail view
	return a.renderDetail(agent, session, logs, actions, compactions, initialContext)
}

// renderDetail renders the agent detail view
func (a *App) renderDetail(agent *database.Agent, session *database.Session, logs []*database.Log, actions []*database.Action, compactions []*database.Compaction, initialContext string) error {
	maxX, maxY := a.gui.Size()

	// Create detail view (full screen) with scrolling enabled
	v, err := a.gui.SetView("details", 0, 3, maxX-1, maxY-3)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	v.Title = " Agent Details "
	v.Wrap = true
	v.Autoscroll = false
	v.Clear()

	// Agent information
	fmt.Fprintf(v, "Agent: %s  (%s)\n", agent.ID[:8], agent.Status)

	description := "N/A"
	if agent.Description != nil {
		description = *agent.Description
	}
	fmt.Fprintf(v, "Description: %s\n", description)

	projectName := "N/A"
	if agent.ProjectName != nil {
		projectName = *agent.ProjectName
	}
	fmt.Fprintf(v, "Project: %s\n", projectName)

	currentTask := "No task"
	if agent.CurrentTask != nil {
		currentTask = *agent.CurrentTask
	}
	fmt.Fprintf(v, "Current task: %s\n", currentTask)

	// Window ID (for desktop agents)
	if agent.WindowID != nil && *agent.WindowID != "" {
		fmt.Fprintf(v, "Window ID: %s\n", *agent.WindowID)
	}

	// Session information
	if session != nil {
		sessionName := "-"
		if session.Name != nil {
			sessionName = *session.Name
		}
		fmt.Fprintf(v, "Session: %s\n", sessionName)
		fmt.Fprintf(v, "Started: %s\n", session.StartedAt.Format("2006-01-02 15:04"))
	}

	if agent.LastActivityAt != nil {
		fmt.Fprintf(v, "Last activity: %s\n", agent.LastActivityAt.Format("2006-01-02 15:04:05"))
	}

	// Actions section
	fmt.Fprintln(v, "---")
	fmt.Fprintln(v, "[ACTIONS - agent activity]")

	if len(actions) > 0 {
		for _, action := range actions {
			fmt.Fprintf(v, "%s  [%s] %s\n",
				action.Timestamp.Format("15:04"),
				action.ActionType,
				action.Description)
		}
	} else {
		fmt.Fprintln(v, "No actions logged")
	}

	// Logs section
	fmt.Fprintln(v, "---")
	fmt.Fprintln(v, "[LOGS - current session]")

	if len(logs) > 0 {
		for _, log := range logs {
			fmt.Fprintf(v, "%s  %s\n",
				log.Timestamp.Format("15:04"),
				log.Message)
		}
	} else {
		fmt.Fprintln(v, "No logs available")
	}

	// Compactions section
	fmt.Fprintln(v, "---")
	fmt.Fprintln(v, "[COMPACTIONS - conversation compactions]")

	if len(compactions) > 0 {
		for _, comp := range compactions {
			fmt.Fprintf(v, "%s  Session: %s → %s\n",
				comp.CompactedAt.Format("2006-01-02 15:04"),
				truncateString(comp.NewSessionID, 12),
				truncateString(ptrToString(comp.OldSessionID), 12))
			if comp.Summary != nil && *comp.Summary != "" {
				fmt.Fprintf(v, "  Summary: %s\n", *comp.Summary)
			}
			if comp.MessageCount != nil {
				fmt.Fprintf(v, "  Messages: %d\n", *comp.MessageCount)
			}
		}
	} else {
		fmt.Fprintln(v, "No compactions")
	}

	// Initial context section (if available)
	if initialContext != "" {
		fmt.Fprintln(v, "---")
		fmt.Fprintln(v, "[INITIAL CONTEXT - from persona]")
		// Show first 500 characters of initial context
		if len(initialContext) > 500 {
			fmt.Fprintf(v, "%s...\n", initialContext[:500])
		} else {
			fmt.Fprintln(v, initialContext)
		}
	}

	// Set as current view
	if _, err := a.gui.SetCurrentView("details"); err != nil {
		return err
	}

	// Delete any existing keybindings for this view
	a.gui.DeleteKeybinding("details", 'q', gocui.ModNone)
	a.gui.DeleteKeybinding("details", 'r', gocui.ModNone)
	a.gui.DeleteKeybinding("details", 'w', gocui.ModNone)
	a.gui.DeleteKeybinding("details", gocui.KeyArrowUp, gocui.ModNone)
	a.gui.DeleteKeybinding("details", gocui.KeyArrowDown, gocui.ModNone)
	// Delete configured scroll keys
	for _, key := range a.keymap.Up {
		a.gui.DeleteKeybinding("details", rune(key[0]), gocui.ModNone)
	}
	for _, key := range a.keymap.Down {
		a.gui.DeleteKeybinding("details", rune(key[0]), gocui.ModNone)
	}

	// Add close keybinding (from config)
	for _, key := range a.keymap.Quit {
		if err := a.gui.SetKeybinding("details", rune(key[0]), gocui.ModNone, a.closeDetails); err != nil {
			return err
		}
	}

	// Add refresh keybinding (from config)
	for _, key := range a.keymap.Refresh {
		if err := a.gui.SetKeybinding("details", rune(key[0]), gocui.ModNone, a.refreshDetails); err != nil {
			return err
		}
	}

	// Add window keybinding (from config)
	for _, key := range a.keymap.GoToWindow {
		if err := a.gui.SetKeybinding("details", rune(key[0]), gocui.ModNone, a.goToWindow); err != nil {
			return err
		}
	}

	// Add scroll keybindings (arrow keys for compatibility)
	if err := a.gui.SetKeybinding("details", gocui.KeyArrowUp, gocui.ModNone, a.scrollUp); err != nil {
		return err
	}
	if err := a.gui.SetKeybinding("details", gocui.KeyArrowDown, gocui.ModNone, a.scrollDown); err != nil {
		return err
	}

	// Add configured scroll keybindings (j/k from config)
	for _, key := range a.keymap.Up {
		if err := a.gui.SetKeybinding("details", rune(key[0]), gocui.ModNone, a.scrollUp); err != nil {
			return err
		}
	}
	for _, key := range a.keymap.Down {
		if err := a.gui.SetKeybinding("details", rune(key[0]), gocui.ModNone, a.scrollDown); err != nil {
			return err
		}
	}

	// Update status bar
	statusView, err := a.gui.View(viewStatus)
	if err != nil {
		return err
	}
	statusView.Clear()
	fmt.Fprint(statusView, " [j/k] Scroll • [r] Refresh • [w] Go to window • [L] View all logs • [q] Return to agents list")

	return nil
}

// refreshDetails refreshes the detail view with latest data
func (a *App) refreshDetails(g *gocui.Gui, v *gocui.View) error {
	if a.currentAgentID == "" {
		return nil
	}
	return a.viewAgentDetails(a.currentAgentID)
}

// closeDetails closes the details view and returns to main list
func (a *App) closeDetails(g *gocui.Gui, v *gocui.View) error {
	// Clear current agent ID
	a.currentAgentID = ""

	// Delete keybindings for details view
	for _, key := range a.keymap.Quit {
		g.DeleteKeybinding("details", rune(key[0]), gocui.ModNone)
	}
	for _, key := range a.keymap.Refresh {
		g.DeleteKeybinding("details", rune(key[0]), gocui.ModNone)
	}
	for _, key := range a.keymap.GoToWindow {
		g.DeleteKeybinding("details", rune(key[0]), gocui.ModNone)
	}
	for _, key := range a.keymap.Up {
		g.DeleteKeybinding("details", rune(key[0]), gocui.ModNone)
	}
	for _, key := range a.keymap.Down {
		g.DeleteKeybinding("details", rune(key[0]), gocui.ModNone)
	}
	g.DeleteKeybinding("details", gocui.KeyArrowUp, gocui.ModNone)
	g.DeleteKeybinding("details", gocui.KeyArrowDown, gocui.ModNone)

	// Delete details view
	if err := g.DeleteView("details"); err != nil && err != gocui.ErrUnknownView {
		return err
	}

	// Return to main view
	if _, err := g.SetCurrentView(viewMain); err != nil {
		return err
	}

	// Restore status bar
	statusView, err := g.View(viewStatus)
	if err != nil {
		return err
	}
	statusView.Clear()
	fmt.Fprint(statusView, " [q] Quit | [r] Refresh | [a] Toggle All | [D] Archive | [c] Continue | [w] Window | [L] Logs")

	// Re-render agents
	return a.renderAgents()
}

// scrollUp scrolls the detail view up
func (a *App) scrollUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		if oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

// scrollDown scrolls the detail view down
func (a *App) scrollDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		// Ignore error if we're at the bottom
		v.SetOrigin(ox, oy+1)
	}
	return nil
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// ptrToString converts a string pointer to string
func ptrToString(s *string) string {
	if s == nil {
		return "N/A"
	}
	return *s
}

// archiveAgent archives (soft deletes) the selected agent
func (a *App) archiveAgent(g *gocui.Gui, v *gocui.View) error {
	if a.selectedIdx < 0 || a.selectedIdx >= len(a.agents) {
		return nil
	}

	agent := a.agents[a.selectedIdx]

	// Update agent status to archived
	if err := a.db.UpdateAgentStatus(agent.ID, database.StatusArchived, nil); err != nil {
		return a.showMessage(fmt.Sprintf("Failed to archive agent: %v", err))
	}

	// Refresh the agent list
	if err := a.refreshAgents(); err != nil {
		return err
	}

	// Adjust selected index if needed
	if a.selectedIdx >= len(a.agents) && a.selectedIdx > 0 {
		a.selectedIdx = len(a.agents) - 1
	}

	// Update cursor position
	if mainView, err := g.View(viewMain); err == nil {
		mainView.SetCursor(0, a.selectedIdx+2) // +2 for header
	}

	return a.showMessage(fmt.Sprintf("Agent %s archived", agent.ID[:8]))
}
