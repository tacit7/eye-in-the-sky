package dashboard

import (
	"github.com/jroimartin/gocui"
)

// setupKeybindings sets up all key bindings
func (a *App) setupKeybindings() error {
	// Quit
	for _, key := range a.keymap.Quit {
		if err := a.gui.SetKeybinding("", rune(key[0]), gocui.ModNone, a.quit); err != nil {
			return err
		}
	}

	// Refresh
	for _, key := range a.keymap.Refresh {
		if err := a.gui.SetKeybinding("", rune(key[0]), gocui.ModNone, a.refresh); err != nil {
			return err
		}
	}

	// Toggle all agents
	for _, key := range a.keymap.ToggleAllAgents {
		if err := a.gui.SetKeybinding("", rune(key[0]), gocui.ModNone, a.toggleFilter); err != nil {
			return err
		}
	}

	// Navigation - Up
	for _, key := range a.keymap.Up {
		if err := a.gui.SetKeybinding(viewMain, rune(key[0]), gocui.ModNone, a.cursorUp); err != nil {
			return err
		}
	}

	// Navigation - Down
	for _, key := range a.keymap.Down {
		if err := a.gui.SetKeybinding(viewMain, rune(key[0]), gocui.ModNone, a.cursorDown); err != nil {
			return err
		}
	}

	// Continue session
	for _, key := range a.keymap.ContinueSession {
		if err := a.gui.SetKeybinding("", rune(key[0]), gocui.ModNone, a.continueSession); err != nil {
			return err
		}
	}

	// Go to window
	for _, key := range a.keymap.GoToWindow {
		if err := a.gui.SetKeybinding("", rune(key[0]), gocui.ModNone, a.goToWindow); err != nil {
			return err
		}
	}

	// Logs
	for _, key := range a.keymap.Logs {
		if err := a.gui.SetKeybinding("", rune(key[0]), gocui.ModNone, a.showLogs); err != nil {
			return err
		}
	}

	// View details (Enter key)
	if err := a.gui.SetKeybinding(viewMain, gocui.KeyEnter, gocui.ModNone, a.handleViewDetails); err != nil {
		return err
	}

	// Arrow keys
	if err := a.gui.SetKeybinding(viewMain, gocui.KeyArrowUp, gocui.ModNone, a.cursorUp); err != nil {
		return err
	}
	if err := a.gui.SetKeybinding(viewMain, gocui.KeyArrowDown, gocui.ModNone, a.cursorDown); err != nil {
		return err
	}

	// Delete/Archive agent
	if err := a.gui.SetKeybinding("", 'D', gocui.ModNone, a.archiveAgent); err != nil {
		return err
	}

	return nil
}

// cursorUp moves the cursor up
func (a *App) cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		// Don't go above line 2 (header is lines 0-1)
		if cy > 2 {
			if err := v.SetCursor(cx, cy-1); err != nil {
				if oy > 0 {
					if err := v.SetOrigin(ox, oy-1); err != nil {
						return err
					}
				}
			}
			if a.selectedIdx > 0 {
				a.selectedIdx--
			}
		}
	}
	return nil
}

// cursorDown moves the cursor down
func (a *App) cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if a.selectedIdx < len(a.agents)-1 {
			if err := v.SetCursor(cx, cy+1); err != nil {
				if err := v.SetOrigin(ox, oy+1); err != nil {
					return err
				}
			}
			a.selectedIdx++
		}
	}
	return nil
}

// quit exits the application
func (a *App) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// refresh manually refreshes the agent list
func (a *App) refresh(g *gocui.Gui, v *gocui.View) error {
	return a.refreshAgents()
}

// toggleFilter toggles between active and all agents
func (a *App) toggleFilter(g *gocui.Gui, v *gocui.View) error {
	if a.filter == "active" {
		a.filter = ""
	} else {
		a.filter = "active"
	}
	a.selectedIdx = 0

	// Reset cursor to first agent (line 2)
	if mainView, err := g.View(viewMain); err == nil {
		mainView.SetCursor(0, 2)
		mainView.SetOrigin(0, 0)
	}

	return a.refreshAgents()
}

// handleViewDetails handles Enter key to view agent details
func (a *App) handleViewDetails(g *gocui.Gui, v *gocui.View) error {
	// Use the selected index to get the agent directly
	if a.selectedIdx >= 0 && a.selectedIdx < len(a.agents) {
		agent := a.agents[a.selectedIdx]
		return a.viewAgentDetails(agent.ID)
	}
	return nil
}
