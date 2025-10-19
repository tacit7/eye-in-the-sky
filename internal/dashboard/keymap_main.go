package dashboard

import "github.com/jroimartin/gocui"

// RegisterMainKeys registers all keybindings for the main view
// Tag: "main" - Main agent list view
func (a *App) RegisterMainKeys() {
	// Quit - bound to main view only
	a.keys.Register(KeySpec{
		View: viewMain,
		Key:  'q',
		Mod:  gocui.ModNone,
		Fn:   a.quit,
		Tag:  "main",
	})

	// View details - Enter key
	a.keys.Register(KeySpec{
		View: viewMain,
		Key:  gocui.KeyEnter,
		Mod:  gocui.ModNone,
		Fn:   a.handleViewDetails,
		Tag:  "main",
	})

	// Edit description
	a.keys.Register(KeySpec{
		View: viewMain,
		Key:  'e',
		Mod:  gocui.ModNone,
		Fn:   a.editDescription,
		Tag:  "main",
	})

	// Refresh
	a.keys.Register(KeySpec{
		View: viewMain,
		Key:  'R',
		Mod:  gocui.ModNone,
		Fn:   a.refresh,
		Tag:  "main",
	})

	// Toggle filter
	a.keys.Register(KeySpec{
		View: viewMain,
		Key:  'a',
		Mod:  gocui.ModNone,
		Fn:   a.toggleFilter,
		Tag:  "main",
	})

	// Archive
	a.keys.Register(KeySpec{
		View: viewMain,
		Key:  'D',
		Mod:  gocui.ModNone,
		Fn:   a.archiveAgent,
		Tag:  "main",
	})

	// Mark done
	a.keys.Register(KeySpec{
		View: viewMain,
		Key:  'd',
		Mod:  gocui.ModNone,
		Fn:   a.markAgentDone,
		Tag:  "main",
	})

	// New session
	a.keys.Register(KeySpec{
		View: viewMain,
		Key:  'n',
		Mod:  gocui.ModNone,
		Fn:   a.createNewSession,
		Tag:  "main",
	})

	// Resume session
	a.keys.Register(KeySpec{
		View: viewMain,
		Key:  'r',
		Mod:  gocui.ModNone,
		Fn:   a.continueSession,
		Tag:  "main",
	})

	// Start session
	a.keys.Register(KeySpec{
		View: viewMain,
		Key:  's',
		Mod:  gocui.ModNone,
		Fn:   a.startSession,
		Tag:  "main",
	})

	// Go to window
	a.keys.Register(KeySpec{
		View: viewMain,
		Key:  'w',
		Mod:  gocui.ModNone,
		Fn:   a.goToWindow,
		Tag:  "main",
	})

	// Logs
	a.keys.Register(KeySpec{
		View: viewMain,
		Key:  'L',
		Mod:  gocui.ModNone,
		Fn:   a.showLogs,
		Tag:  "main",
	})

	// Navigation bindings will be added via "navigation" tag
}
