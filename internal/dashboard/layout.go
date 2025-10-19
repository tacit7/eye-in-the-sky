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

	// Header view
	if v, err := g.SetView("header", 0, 0, maxX-1, 2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.FgColor = gocui.ColorCyan | gocui.AttrBold
		v.Clear()
		v.Write([]byte("╔══════════════════════════════════════════════════════════════════════════════╗\n"))
		v.Write([]byte("║                         EYE IN THE SKY - TUI                                 ║\n"))
		v.Write([]byte("╚══════════════════════════════════════════════════════════════════════════════╝"))
	}

	// Main agents list view
	if v, err := g.SetView(viewMain, 0, 3, maxX-1, maxY-3); err != nil {
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

	// Status bar view
	if v, err := g.SetView(viewStatus, 0, maxY-2, maxX-1, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.FgColor = gocui.ColorWhite
		v.Clear()
		v.Write([]byte(" [q] Quit | [r] Refresh | [a] Toggle All | [D] Archive | [c] Continue | [w] Window | [L] Logs"))
	}

	return nil
}
