package dashboard

import (
	"github.com/jroimartin/gocui"
)

// quit exits the application
func (a *App) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// refresh refreshes the agent list
func (a *App) refresh(g *gocui.Gui, v *gocui.View) error {
	return a.refreshAgents()
}

// handleViewDetails shows details for the selected agent
func (a *App) handleViewDetails(g *gocui.Gui, v *gocui.View) error {
	if a.selectedIdx < 0 || a.selectedIdx >= len(a.agents) {
		return nil
	}
	agent := a.agents[a.selectedIdx]
	return a.viewAgentDetails(agent.ID)
}

// toggleFilter toggles between showing all agents and active only
func (a *App) toggleFilter(g *gocui.Gui, v *gocui.View) error {
	if a.filter == "all" {
		a.filter = "active"
	} else {
		a.filter = "all"
	}
	return a.refreshAgents()
}

// editDescription enters edit mode for the selected agent's description
func (a *App) editDescription(g *gocui.Gui, v *gocui.View) error {
	// TODO: Implement edit mode (will be added later)
	return a.showMessage("Edit mode not yet implemented")
}

// saveDescription saves the edited description
func (a *App) saveDescription(g *gocui.Gui, v *gocui.View) error {
	// TODO: Implement save (will be added later)
	return nil
}

// cancelEdit cancels the edit mode
func (a *App) cancelEdit(g *gocui.Gui, v *gocui.View) error {
	// TODO: Implement cancel (will be added later)
	return nil
}

// cursorUp moves cursor up in the main list
func (a *App) cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		return nil
	}

	cx, cy := v.Cursor()
	// Don't move above the header (line 2)
	if cy > 2 {
		if err := v.SetCursor(cx, cy-1); err != nil {
			ox, oy := v.Origin()
			if oy > 0 {
				v.SetOrigin(ox, oy-1)
			}
		}
		if a.selectedIdx > 0 {
			a.selectedIdx--
		}
	}
	return nil
}

// cursorDown moves cursor down in the main list
func (a *App) cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		return nil
	}

	cx, cy := v.Cursor()
	lines := len(v.BufferLines())

	// Don't move past the last agent
	if cy < lines-1 && a.selectedIdx < len(a.agents)-1 {
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			v.SetOrigin(ox, oy+1)
		}
		a.selectedIdx++
	}
	return nil
}

// jumpToTop jumps to the top of the current view
func (a *App) jumpToTop(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		v.SetOrigin(0, 0)
		if v.Name() == viewMain {
			a.selectedIdx = 0
			v.SetCursor(0, 2) // Start at line 2 (after header)
		} else {
			v.SetCursor(0, 0)
		}
	}
	return nil
}

// jumpToBottom jumps to the bottom of the current view
func (a *App) jumpToBottom(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		// Get buffer line count
		lines := v.BufferLines()
		lineCount := len(lines)

		if v.Name() == viewMain {
			a.selectedIdx = len(a.agents) - 1
			v.SetCursor(0, a.selectedIdx)
		} else {
			// For detail views, scroll to bottom
			_, height := v.Size()
			if lineCount > height {
				v.SetOrigin(0, lineCount-height)
			}
		}
	}
	return nil
}
