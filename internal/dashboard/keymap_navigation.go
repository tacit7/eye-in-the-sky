package dashboard

import "github.com/jroimartin/gocui"

// RegisterNavigationKeys registers shared navigation keybindings
// Tag: "navigation" - Reusable scroll/cursor bindings for list-like views
func (a *App) RegisterNavigationKeys() {
	// Cursor/Scroll down - j (Vim style: j = down)
	a.keys.Register(KeySpec{
		View: "", // Will be bound per-view
		Key:  'j',
		Mod:  gocui.ModNone,
		Fn:   a.navigationDown, // Generic handler
		Tag:  "navigation",
	})

	// Cursor/Scroll up - k (Vim style: k = up)
	a.keys.Register(KeySpec{
		View: "",
		Key:  'k',
		Mod:  gocui.ModNone,
		Fn:   a.navigationUp,
		Tag:  "navigation",
	})

	// Arrow up
	a.keys.Register(KeySpec{
		View: "",
		Key:  gocui.KeyArrowUp,
		Mod:  gocui.ModNone,
		Fn:   a.navigationUp,
		Tag:  "navigation",
	})

	// Arrow down
	a.keys.Register(KeySpec{
		View: "",
		Key:  gocui.KeyArrowDown,
		Mod:  gocui.ModNone,
		Fn:   a.navigationDown,
		Tag:  "navigation",
	})

	// Page down - Space
	a.keys.Register(KeySpec{
		View: "",
		Key:  ' ',
		Mod:  gocui.ModNone,
		Fn:   a.pageDown,
		Tag:  "navigation",
	})

	// Page up - b
	a.keys.Register(KeySpec{
		View: "",
		Key:  'b',
		Mod:  gocui.ModNone,
		Fn:   a.pageUp,
		Tag:  "navigation",
	})

	// Page down - PgDn
	a.keys.Register(KeySpec{
		View: "",
		Key:  gocui.KeyPgdn,
		Mod:  gocui.ModNone,
		Fn:   a.pageDown,
		Tag:  "navigation",
	})

	// Page up - PgUp
	a.keys.Register(KeySpec{
		View: "",
		Key:  gocui.KeyPgup,
		Mod:  gocui.ModNone,
		Fn:   a.pageUp,
		Tag:  "navigation",
	})

	// Top - g (Vim-style)
	a.keys.Register(KeySpec{
		View: "",
		Key:  'g',
		Mod:  gocui.ModNone,
		Fn:   a.jumpToTop,
		Tag:  "navigation",
	})

	// Bottom - G (Vim-style)
	a.keys.Register(KeySpec{
		View: "",
		Key:  'G',
		Mod:  gocui.ModNone,
		Fn:   a.jumpToBottom,
		Tag:  "navigation",
	})

	// Home
	a.keys.Register(KeySpec{
		View: "",
		Key:  gocui.KeyHome,
		Mod:  gocui.ModNone,
		Fn:   a.jumpToTop,
		Tag:  "navigation",
	})

	// End
	a.keys.Register(KeySpec{
		View: "",
		Key:  gocui.KeyEnd,
		Mod:  gocui.ModNone,
		Fn:   a.jumpToBottom,
		Tag:  "navigation",
	})
}

// navigationUp handles up movement - cursor in main, scroll in others
func (a *App) navigationUp(g *gocui.Gui, v *gocui.View) error {
	if v.Name() == viewMain {
		return a.cursorUp(g, v)
	}
	return a.scrollUp(g, v)
}

// navigationDown handles down movement - cursor in main, scroll in others
func (a *App) navigationDown(g *gocui.Gui, v *gocui.View) error {
	if v.Name() == viewMain {
		return a.cursorDown(g, v)
	}
	return a.scrollDown(g, v)
}
