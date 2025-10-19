package dashboard

import (
	"github.com/jroimartin/gocui"
)

const (
	viewMain    = "main"
	viewLogs    = "logs"
	viewCommand = "command"
	viewStatus  = "status"
)

// layout defines the UI layout
func (a *App) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Header view - only create once, don't recreate on every layout
	if v, err := g.SetView("header", 0, 0, maxX-1, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.FgColor = gocui.ColorCyan | gocui.AttrBold
		v.Clear()
		v.SetCursor(0, 0)

		// Simple header without ANSI escapes
		v.Write([]byte("================================================================================\n"))
		v.Write([]byte("                       EYE IN THE SKY - TUI                                 \n"))
		v.Write([]byte("================================================================================"))
	}
	// Don't update header on every layout pass - it steals focus

	// Notification area (space reserved at top, view created dynamically by showMessage)
	notificationY := 5 // Reserve 2 lines for notifications (3-5)

	// Main agents list view (starts after notification area)
	if v, err := g.SetView(viewMain, 0, notificationY, maxX-1, maxY-3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Active Sessions "
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		if _, err := g.SetCurrentView(viewMain); err != nil {
			return err
		}
		// Set cursor to first agent (skip header and separator line)
		v.SetCursor(0, 2)
		// Render agents on first load
		a.renderAgents()
	}

	// Status bar view - only create once
	if v, err := g.SetView(viewStatus, 0, maxY-2, maxX-1, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.FgColor = gocui.ColorWhite
		v.Clear()
		v.SetCursor(0, 0)
		v.Write([]byte(" [q] Quit | [R] Refresh | [a] Toggle All | [e] Edit | [d] Done | [D] Archive | [n] New | [r] Resume | [s] Start | [w] Window | [L] Logs"))
	}
	// Don't update status bar on every layout - updates happen via explicit calls

	return nil
}
