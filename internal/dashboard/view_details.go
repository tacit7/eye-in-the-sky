package dashboard

import (
	"fmt"
	"sync"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/tacit7/eye-in-the-sky/internal/database"
)

// viewAgentDetails fetches and displays detailed information for an agent
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
		session, _ = a.db.GetSession(*agent.CurrentSessionID)
	}

	// Get logs for the session (limit to 20)
	var logs []*database.Log
	if session != nil {
		if fetchedLogs, err := a.db.GetLogs(session.ID); err == nil && fetchedLogs != nil {
			logs = fetchedLogs
			if len(logs) > 20 {
				logs = logs[:20]
			}
		}
	}

	// Get actions for the agent (limit to 20)
	var actions []*database.Action
	if fetchedActions, err := a.db.ListActions(agent.ID); err == nil && fetchedActions != nil {
		actions = fetchedActions
		if len(actions) > 20 {
			actions = actions[:20]
		}
	}

	// Get compactions for the agent
	compactions, _ := a.db.GetCompactionsForAgent(agent.ID)

	// Get initial context from persona if available
	var initialContext string
	if agent.PersonaID != nil && *agent.PersonaID != "" {
		if persona, err := a.db.GetPersona(*agent.PersonaID); err == nil && persona != nil {
			initialContext = persona.InitialContext
		}
	}

	// Render the detail view
	return a.renderDetail(agent, session, logs, actions, compactions, initialContext)
}

// renderDetail renders the agent detail view
func (a *App) renderDetail(agent *database.Agent, session *database.Session, logs []*database.Log, actions []*database.Action, compactions []*database.Compaction, initialContext string) error {
	v, err := a.ensureDetailView()
	if err != nil {
		return err
	}

	v.Clear()
	v.SetCursor(0, 0)
	v.SetOrigin(0, 0)

	lineNum := 0
	lineNum = a.renderAgentSection(v, agent, session, lineNum)
	lineNum = a.renderActionsSection(v, actions, lineNum)
	lineNum = a.renderLogsSection(v, logs, lineNum)
	lineNum = a.renderCompactionsSection(v, compactions, lineNum)
	lineNum = a.renderContextsSection(v, agent.ID, lineNum)
	lineNum = a.renderInitialContextSection(v, initialContext, lineNum)

	// Switch to details view with proper tag management
	if err := a.views.SwitchToView(
		[]string{"main", "navigation"},
		[]string{"details", "navigation"},
		ModeDetails,
		"details",
	); err != nil {
		return err
	}

	return a.updateStatusBar(" [j/k] Scroll • [x] Contexts • [l] Logs • [R] Refresh • [s] Start • [w] Window • [q] Back")
}

// ensureDetailView creates or retrieves the detail view
func (a *App) ensureDetailView() (*gocui.View, error) {
	maxX, maxY := a.gui.Size()

	// Initialize maps
	a.currentCompactions = nil
	a.compactionLineMap = make(map[int]int)
	a.currentContexts = nil
	a.contextLineMap = make(map[int]int)
	a.sectionLines = make(map[string]int)

	// Create detail view (full screen) with scrolling enabled
	v, err := a.gui.SetView("details", 0, 3, maxX-1, maxY-3)
	if err != nil && err != gocui.ErrUnknownView {
		return nil, err
	}

	v.Title = " Agent Details "
	v.Wrap = true
	v.Autoscroll = false
	v.Highlight = true
	v.SelBgColor = gocui.ColorGreen
	v.SelFgColor = gocui.ColorBlack

	return v, nil
}

// renderAgentSection renders agent basic information
func (a *App) renderAgentSection(v *gocui.View, agent *database.Agent, session *database.Session, lineNum int) int {
	fmt.Fprintf(v, "Agent: %s  (%s)\n", agent.ID[:8], agent.Status)
	lineNum++

	description := "N/A"
	if agent.Description != nil {
		description = *agent.Description
	}
	fmt.Fprintf(v, "Description: %s\n", description)
	lineNum++

	projectName := "N/A"
	if agent.ProjectName != nil {
		projectName = *agent.ProjectName
	}
	fmt.Fprintf(v, "Project: %s\n", projectName)
	lineNum++

	currentTask := "No task"
	if agent.CurrentTask != nil {
		currentTask = *agent.CurrentTask
	}
	fmt.Fprintf(v, "Current task: %s\n", currentTask)
	lineNum++

	// Window ID (for desktop agents)
	if agent.WindowID != nil && *agent.WindowID != "" {
		fmt.Fprintf(v, "Window ID: %s\n", *agent.WindowID)
		lineNum++
	}

	// Session information
	if session != nil {
		sessionName := "-"
		if session.Name != nil {
			sessionName = *session.Name
		}
		fmt.Fprintf(v, "Session: %s\n", sessionName)
		lineNum++
		fmt.Fprintf(v, "Started: %s\n", session.StartedAt.Format("2006-01-02 15:04"))
		lineNum++
	}

	if agent.LastActivityAt != nil {
		fmt.Fprintf(v, "Last activity: %s\n", agent.LastActivityAt.Format("2006-01-02 15:04:05"))
		lineNum++
	}

	return lineNum
}

// renderActionsSection renders the actions list
func (a *App) renderActionsSection(v *gocui.View, actions []*database.Action, lineNum int) int {
	fmt.Fprintln(v, "────────────────────────────────────────────────────────────────")
	lineNum++
	a.sectionLines["actions"] = lineNum
	fmt.Fprintln(v, "[ACTIONS - last 10 (press 'l' for all)]")
	lineNum++

	displayActions := actions
	if len(actions) > 10 {
		displayActions = actions[:10]
	}

	if len(displayActions) > 0 {
		for _, action := range displayActions {
			fmt.Fprintf(v, "%s  [%s] %s\n",
				action.Timestamp.Format("15:04"),
				action.ActionType,
				action.Description)
			lineNum++
		}
	} else {
		fmt.Fprintln(v, "No actions logged")
		lineNum++
	}

	return lineNum
}

// renderLogsSection renders the logs list
func (a *App) renderLogsSection(v *gocui.View, logs []*database.Log, lineNum int) int {
	fmt.Fprintln(v, "────────────────────────────────────────────────────────────────")
	lineNum++
	a.sectionLines["logs"] = lineNum
	fmt.Fprintln(v, "[LOGS - current session]")
	lineNum++

	if len(logs) > 0 {
		for _, log := range logs {
			fmt.Fprintf(v, "%s  %s\n",
				log.Timestamp.Format("15:04"),
				log.Message)
			lineNum++
		}
	} else {
		fmt.Fprintln(v, "No logs available")
		lineNum++
	}

	return lineNum
}

// renderCompactionsSection renders the compactions list
func (a *App) renderCompactionsSection(v *gocui.View, compactions []*database.Compaction, lineNum int) int {
	fmt.Fprintln(v, "────────────────────────────────────────────────────────────────")
	lineNum++
	a.sectionLines["compactions"] = lineNum
	fmt.Fprintln(v, "[COMPACTIONS - last 5]")
	lineNum++

	// Store for later access
	a.currentCompactions = compactions

	displayCompactions := compactions
	if len(compactions) > 5 {
		displayCompactions = compactions[:5]
	}

	if len(displayCompactions) > 0 {
		for i, comp := range displayCompactions {
			// Map this line to the compaction index
			a.compactionLineMap[lineNum] = i
			a.debugf("[renderDetail] Mapped line %d to compaction index %d (session: %s)", lineNum, i, comp.SessionID)

			fmt.Fprintf(v, "%s  Session: %s\n",
				comp.CompactedAt.Format("2006-01-02 15:04"),
				truncateString(comp.SessionID, 16))
			lineNum++

			if comp.Summary != nil && *comp.Summary != "" {
				fmt.Fprintf(v, "  Summary: %s\n", *comp.Summary)
				lineNum++
			}
			if comp.MessageCount != nil {
				fmt.Fprintf(v, "  Messages: %d\n", *comp.MessageCount)
				lineNum++
			}
		}
	} else {
		fmt.Fprintln(v, "No compactions")
		lineNum++
	}

	return lineNum
}

// renderContextsSection renders the saved contexts list
func (a *App) renderContextsSection(v *gocui.View, agentID string, lineNum int) int {
	fmt.Fprintln(v, "────────────────────────────────────────────────────────────────")
	lineNum++
	a.sectionLines["contexts"] = lineNum
	fmt.Fprintln(v, "[SAVED CONTEXTS - last 5 (press 'c' for all)]")
	lineNum++

	// Get saved contexts from session_context table
	savedContexts, err := a.db.GetSessionContextsForAgent(agentID)
	a.debugf("[renderDetail] Getting contexts for agent %s, found %d, err: %v", agentID, len(savedContexts), err)

	// Guard against nil
	if err != nil || savedContexts == nil {
		fmt.Fprintln(v, "No saved contexts")
		lineNum++
		return lineNum
	}

	displayContexts := savedContexts
	if len(savedContexts) > 5 {
		displayContexts = savedContexts[:5]
	}

	if len(displayContexts) > 0 {
		// Store ALL contexts for later access (when pressing 'c')
		a.currentContexts = savedContexts

		// Already sorted by created_at DESC, show latest 5
		for i, ctx := range displayContexts {
			phase := "unknown"
			if ctx.CurrentPhase != nil {
				phase = *ctx.CurrentPhase
			}

			// Map this line to the context index
			a.contextLineMap[lineNum] = i
			a.debugf("[renderDetail] Mapped line %d to context index %d (phase: %s)", lineNum, i, phase)

			fmt.Fprintf(v, "%s  Phase: %s\n",
				ctx.CreatedAt.Format("2006-01-02 15:04"),
				phase)
			lineNum++
		}
	} else {
		fmt.Fprintln(v, "No saved contexts")
		lineNum++
	}

	return lineNum
}

// renderInitialContextSection renders the initial context from persona
func (a *App) renderInitialContextSection(v *gocui.View, initialContext string, lineNum int) int {
	if initialContext == "" {
		return lineNum
	}

	fmt.Fprintln(v, "────────────────────────────────────────────────────────────────")
	lineNum++
	fmt.Fprintln(v, "[INITIAL CONTEXT - from persona]")
	lineNum++

	// Show first 500 characters of initial context
	if len(initialContext) > 500 {
		fmt.Fprintf(v, "%s...\n", initialContext[:500])
	} else {
		fmt.Fprintln(v, initialContext)
	}
	lineNum++

	return lineNum
}

// refreshDetails refreshes the detail view with latest data
var refreshMutex sync.Mutex

func (a *App) refreshDetails(g *gocui.Gui, v *gocui.View) error {
	// Serialize refresh requests
	if !refreshMutex.TryLock() {
		return nil // Already refreshing
	}
	defer refreshMutex.Unlock()

	if a.currentAgentID == "" {
		return nil
	}

	// Show green dot
	a.showRefreshDot = true
	g.Update(func(g *gocui.Gui) error {
		return nil
	})

	// Refresh data
	err := a.viewAgentDetails(a.currentAgentID)

	// Hide green dot after 1 second
	go func() {
		time.Sleep(1 * time.Second)
		a.showRefreshDot = false
		g.Update(func(g *gocui.Gui) error {
			return nil
		})
	}()

	return err
}

// closeDetails closes the details view and returns to main list
func (a *App) closeDetails(g *gocui.Gui, v *gocui.View) error {
	// Clear current agent ID
	a.currentAgentID = ""

	// Delete the details view
	if err := g.DeleteView("details"); err != nil && err != gocui.ErrUnknownView {
		return err
	}

	// Use view manager to restore main view
	if err := a.views.PopView(
		[]string{"details", "navigation"},
		[]string{"main", "navigation"},
		viewMain,
	); err != nil {
		return err
	}

	// Restore status bar
	if err := a.updateStatusBar(" [q] Quit | [R] Refresh | [a] Toggle All | [D] Archive | [r] Resume | [w] Window | [L] Logs"); err != nil {
		return err
	}

	return a.refreshAgents()
}

// Helper functions

// truncateString truncates a string to maxLen, adding "..." if needed
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// updateStatusBar updates the status bar with new text
func (a *App) updateStatusBar(text string) error {
	statusView, err := a.gui.View(viewStatus)
	if err != nil {
		return err
	}
	statusView.Clear()
	fmt.Fprint(statusView, text)
	return nil
}

// debugf logs debug messages (can be controlled by debug flag)
func (a *App) debugf(format string, args ...interface{}) {
	// TODO: Add debug flag check
	// if a.config.Debug {
	// 	fmt.Fprintf(os.Stderr, format+"\n", args...)
	// }
	// For now, no-op to avoid stderr spam
}
