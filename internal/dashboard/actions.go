package dashboard

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jroimartin/gocui"
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

		// Add close keybinding
		if err := g.SetKeybinding(viewLogs, 'q', gocui.ModNone, a.closeLogs); err != nil {
			return err
		}
	}

	return nil
}

// closeLogs closes the logs view
func (a *App) closeLogs(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView(viewLogs); err != nil {
		return err
	}
	if _, err := g.SetCurrentView(viewMain); err != nil {
		return err
	}
	return nil
}
