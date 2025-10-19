package dashboard

import "github.com/jroimartin/gocui"

// RegisterEditKeys registers keybindings for edit mode
// Tag: "edit" - Text input mode with readline support
func (a *App) RegisterEditKeys() {
	// Submit - Enter
	a.keys.Register(KeySpec{
		View: "edit-input",
		Key:  gocui.KeyEnter,
		Mod:  gocui.ModNone,
		Fn:   a.saveDescription,
		Tag:  "edit",
	})

	// Cancel - Esc
	a.keys.Register(KeySpec{
		View: "edit-input",
		Key:  gocui.KeyEsc,
		Mod:  gocui.ModNone,
		Fn:   a.cancelEdit,
		Tag:  "edit",
	})

	// Readline: Clear line - Ctrl+U
	a.keys.Register(KeySpec{
		View: "edit-input",
		Key:  gocui.KeyCtrlU,
		Mod:  gocui.ModNone,
		Fn:   a.readlineClearLine,
		Tag:  "edit",
	})

	// Readline: Delete word - Ctrl+W
	a.keys.Register(KeySpec{
		View: "edit-input",
		Key:  gocui.KeyCtrlW,
		Mod:  gocui.ModNone,
		Fn:   a.readlineDeleteWord,
		Tag:  "edit",
	})

	// Readline: Jump to start - Ctrl+A
	a.keys.Register(KeySpec{
		View: "edit-input",
		Key:  gocui.KeyCtrlA,
		Mod:  gocui.ModNone,
		Fn:   a.readlineJumpStart,
		Tag:  "edit",
	})

	// Readline: Jump to end - Ctrl+E
	a.keys.Register(KeySpec{
		View: "edit-input",
		Key:  gocui.KeyCtrlE,
		Mod:  gocui.ModNone,
		Fn:   a.readlineJumpEnd,
		Tag:  "edit",
	})

	// Allow arrow keys, backspace, delete - handled by gocui.DefaultEditor
	// No need to bind these; DefaultEditor handles them when view is editable
}

// Readline helper functions

func (a *App) readlineClearLine(g *gocui.Gui, v *gocui.View) error {
	v.Clear()
	v.SetCursor(0, 0)
	return nil
}

func (a *App) readlineDeleteWord(g *gocui.Gui, v *gocui.View) error {
	cx, cy := v.Cursor()
	if cx == 0 {
		return nil
	}

	line := v.Buffer()
	if len(line) == 0 {
		return nil
	}

	// Find start of current word (scan backward to space or start)
	start := cx - 1
	for start > 0 && line[start] != ' ' {
		start--
	}
	if line[start] == ' ' {
		start++
	}

	// Delete from start to cursor
	newLine := line[:start] + line[cx:]
	v.Clear()
	v.Write([]byte(newLine))
	v.SetCursor(start, cy)

	return nil
}

func (a *App) readlineJumpStart(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	v.SetCursor(0, cy)
	return nil
}

func (a *App) readlineJumpEnd(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	line := v.Buffer()
	v.SetCursor(len(line), cy)
	return nil
}
