package dashboard

import "github.com/jroimartin/gocui"

// RegisterGlobalKeys registers global keybindings (active in all views)
// Tag: "global" - Only Ctrl+C and ?
func (a *App) RegisterGlobalKeys() {
	// Quit - Ctrl+C (global emergency exit)
	a.keys.Register(KeySpec{
		View: "", // Global
		Key:  gocui.KeyCtrlC,
		Mod:  gocui.ModNone,
		Fn:   a.quit,
		Tag:  "global",
	})

	// Help - ? (show keybinding help modal)
	a.keys.Register(KeySpec{
		View: "", // Global
		Key:  '?',
		Mod:  gocui.ModNone,
		Fn:   a.showHelp,
		Tag:  "global",
	})
}

// showHelp displays a help modal with current keybindings
func (a *App) showHelp(g *gocui.Gui, v *gocui.View) error {
	// TODO: Implement help modal
	// For now, just show a message
	return a.showMessage("Press Ctrl+C to quit, q to go back")
}
