package dashboard

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
)

// KeySpec defines a single keybinding specification
type KeySpec struct {
	View string                                      // gocui view name ("" for global)
	Key  interface{}                                 // rune or gocui.Key
	Mod  gocui.Modifier                              // Key modifier (gocui.ModNone, etc.)
	Fn   func(*gocui.Gui, *gocui.View) error         // Handler function
	Tag  string                                      // Tag group for bind/unbind
}

// KeyRegistry manages scoped keybindings by tag
type KeyRegistry struct {
	gui    *gocui.Gui
	specs  []KeySpec           // All registered specs
	active map[string]bool     // Currently active tags
}

// NewKeyRegistry creates a new key registry
func NewKeyRegistry(gui *gocui.Gui) *KeyRegistry {
	return &KeyRegistry{
		gui:    gui,
		specs:  []KeySpec{},
		active: make(map[string]bool),
	}
}

// Register adds a keybinding spec to the registry
func (r *KeyRegistry) Register(spec KeySpec) {
	r.specs = append(r.specs, spec)
}

// BindTag activates all keybindings with the given tag
func (r *KeyRegistry) BindTag(tag string) error {
	if r.active[tag] {
		return nil // Already active
	}

	for _, spec := range r.specs {
		if spec.Tag == tag {
			if err := r.gui.SetKeybinding(spec.View, spec.Key, spec.Mod, spec.Fn); err != nil {
				return fmt.Errorf("failed to bind key in tag %s: %w", tag, err)
			}
		}
	}

	r.active[tag] = true
	return nil
}

// UnbindTag deactivates all keybindings with the given tag
func (r *KeyRegistry) UnbindTag(tag string) error {
	if !r.active[tag] {
		return nil // Not active
	}

	for _, spec := range r.specs {
		if spec.Tag == tag {
			if err := r.gui.DeleteKeybinding(spec.View, spec.Key, spec.Mod); err != nil {
				// Ignore errors - binding might not exist
				continue
			}
		}
	}

	delete(r.active, tag)
	return nil
}

// RebindTag unbinds then binds a tag (useful for refresh)
func (r *KeyRegistry) RebindTag(tag string) error {
	if err := r.UnbindTag(tag); err != nil {
		return err
	}
	return r.BindTag(tag)
}

// UnbindAll deactivates all currently active tags
func (r *KeyRegistry) UnbindAll() error {
	for tag := range r.active {
		if err := r.UnbindTag(tag); err != nil {
			return err
		}
	}
	return nil
}

// IsActive returns whether a tag is currently active
func (r *KeyRegistry) IsActive(tag string) bool {
	return r.active[tag]
}

// ActiveTags returns a list of currently active tags
func (r *KeyRegistry) ActiveTags() []string {
	tags := make([]string, 0, len(r.active))
	for tag := range r.active {
		tags = append(tags, tag)
	}
	return tags
}

// ParseKey converts a string key representation to gocui.Key or rune
// Supports: literals (a, R), specials (<Up>, <Enter>), modifiers (Ctrl+X, Alt+X), Vim aliases (<CR>, <C-c>)
func ParseKey(keyStr string) (interface{}, gocui.Modifier, error) {
	mod := gocui.ModNone

	// Handle modifiers
	if strings.HasPrefix(keyStr, "Ctrl+") || strings.HasPrefix(keyStr, "<C-") {
		mod = gocui.ModNone // gocui uses special keys for Ctrl combinations
		keyStr = strings.TrimPrefix(keyStr, "Ctrl+")
		keyStr = strings.TrimPrefix(keyStr, "<C-")
		keyStr = strings.TrimSuffix(keyStr, ">")

		// Map Ctrl+letter to gocui.KeyCtrl[Letter]
		if len(keyStr) == 1 {
			letter := strings.ToUpper(keyStr)[0]
			switch letter {
			case 'A':
				return gocui.KeyCtrlA, mod, nil
			case 'C':
				return gocui.KeyCtrlC, mod, nil
			case 'E':
				return gocui.KeyCtrlE, mod, nil
			case 'U':
				return gocui.KeyCtrlU, mod, nil
			case 'W':
				return gocui.KeyCtrlW, mod, nil
			default:
				return nil, mod, fmt.Errorf("unsupported Ctrl+%c", letter)
			}
		}
	}

	if strings.HasPrefix(keyStr, "Alt+") || strings.HasPrefix(keyStr, "<M-") {
		mod = gocui.ModAlt
		keyStr = strings.TrimPrefix(keyStr, "Alt+")
		keyStr = strings.TrimPrefix(keyStr, "<M-")
		keyStr = strings.TrimSuffix(keyStr, ">")
	}

	// Handle special keys
	keyStr = strings.TrimPrefix(keyStr, "<")
	keyStr = strings.TrimSuffix(keyStr, ">")

	// Vim aliases
	switch keyStr {
	case "CR":
		keyStr = "Enter"
	}

	// Special keys
	switch keyStr {
	case "Up":
		return gocui.KeyArrowUp, mod, nil
	case "Down":
		return gocui.KeyArrowDown, mod, nil
	case "Left":
		return gocui.KeyArrowLeft, mod, nil
	case "Right":
		return gocui.KeyArrowRight, mod, nil
	case "Enter":
		return gocui.KeyEnter, mod, nil
	case "Esc":
		return gocui.KeyEsc, mod, nil
	case "PageUp", "PgUp":
		return gocui.KeyPgup, mod, nil
	case "PageDown", "PgDn":
		return gocui.KeyPgdn, mod, nil
	case "Tab":
		return gocui.KeyTab, mod, nil
	case "Backspace":
		return gocui.KeyBackspace, mod, nil
	case "Delete":
		return gocui.KeyDelete, mod, nil
	case "Home":
		return gocui.KeyHome, mod, nil
	case "End":
		return gocui.KeyEnd, mod, nil
	case "Space":
		return ' ', mod, nil
	case "F1":
		return gocui.KeyF1, mod, nil
	case "F2":
		return gocui.KeyF2, mod, nil
	case "F3":
		return gocui.KeyF3, mod, nil
	case "F4":
		return gocui.KeyF4, mod, nil
	case "F5":
		return gocui.KeyF5, mod, nil
	case "F6":
		return gocui.KeyF6, mod, nil
	case "F7":
		return gocui.KeyF7, mod, nil
	case "F8":
		return gocui.KeyF8, mod, nil
	case "F9":
		return gocui.KeyF9, mod, nil
	case "F10":
		return gocui.KeyF10, mod, nil
	case "F11":
		return gocui.KeyF11, mod, nil
	case "F12":
		return gocui.KeyF12, mod, nil
	}

	// Literal rune
	if len(keyStr) == 1 {
		return rune(keyStr[0]), mod, nil
	}

	return nil, mod, fmt.Errorf("unknown key: %s", keyStr)
}
