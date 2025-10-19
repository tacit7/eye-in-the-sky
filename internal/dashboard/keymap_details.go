package dashboard

import "github.com/jroimartin/gocui"

// RegisterDetailsKeys registers keybindings for the details view
// Tag: "details" - Agent detail view actions
func (a *App) RegisterDetailsKeys() {
	// Back to main - q
	a.keys.Register(KeySpec{
		View: "details",
		Key:  'q',
		Mod:  gocui.ModNone,
		Fn:   a.closeDetails,
		Tag:  "details",
	})

	// Refresh details - R
	a.keys.Register(KeySpec{
		View: "details",
		Key:  'R',
		Mod:  gocui.ModNone,
		Fn:   a.refreshDetails,
		Tag:  "details",
	})

	// View all contexts - c
	a.keys.Register(KeySpec{
		View: "details",
		Key:  'c',
		Mod:  gocui.ModNone,
		Fn:   a.viewAllContexts,
		Tag:  "details",
	})

	// View all contexts - x (alternative)
	a.keys.Register(KeySpec{
		View: "details",
		Key:  'x',
		Mod:  gocui.ModNone,
		Fn:   a.viewAllContexts,
		Tag:  "details",
	})

	// View all actions/logs - l
	a.keys.Register(KeySpec{
		View: "details",
		Key:  'l',
		Mod:  gocui.ModNone,
		Fn:   a.viewAllActions,
		Tag:  "details",
	})

	// Start session - s
	a.keys.Register(KeySpec{
		View: "details",
		Key:  's',
		Mod:  gocui.ModNone,
		Fn:   a.startSession,
		Tag:  "details",
	})

	// Go to window - w
	a.keys.Register(KeySpec{
		View: "details",
		Key:  'w',
		Mod:  gocui.ModNone,
		Fn:   a.goToWindow,
		Tag:  "details",
	})

	// Navigation keys will be added via "navigation" tag
}
