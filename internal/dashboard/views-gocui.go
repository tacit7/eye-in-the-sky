package dashboard

// ViewManager provides helpers for view lifecycle with scoped keybindings
type ViewManager struct {
	app *App
}

// NewViewManager creates a new view manager
func NewViewManager(app *App) *ViewManager {
	return &ViewManager{app: app}
}

// SwitchToView switches from current view to a new view with tag management
// This is the core view transition function
func (vm *ViewManager) SwitchToView(fromTags []string, toTags []string, toMode Mode, viewName string) error {
	// Unbind all current tags
	for _, tag := range fromTags {
		if err := vm.app.keys.UnbindTag(tag); err != nil {
			return err
		}
	}

	// Bind new tags
	for _, tag := range toTags {
		if err := vm.app.keys.BindTag(tag); err != nil {
			return err
		}
	}

	// Push new mode
	vm.app.modes.Push(toMode)

	// Set current view
	if _, err := vm.app.gui.SetCurrentView(viewName); err != nil {
		return err
	}

	return nil
}

// PopView returns from current view to previous view
// Used for modal closes and back navigation
func (vm *ViewManager) PopView(currentTags []string, previousTags []string, viewName string) error {
	// Unbind current tags
	for _, tag := range currentTags {
		if err := vm.app.keys.UnbindTag(tag); err != nil {
			return err
		}
	}

	// Pop mode (returns to previous)
	vm.app.modes.Pop()

	// Bind previous tags
	for _, tag := range previousTags {
		if err := vm.app.keys.BindTag(tag); err != nil {
			return err
		}
	}

	// Set current view
	if _, err := vm.app.gui.SetCurrentView(viewName); err != nil {
		return err
	}

	return nil
}

// CloseModalView is a centralized helper for closing modal views
// It handles: unbinding view tags, deleting the view, popping mode, and restoring previous tags
func (vm *ViewManager) CloseModalView(viewName string, modalTags []string, previousViewName string, previousTags []string) error {
	// Unbind modal tags
	for _, tag := range modalTags {
		vm.app.keys.UnbindTag(tag)
	}

	// Delete the view
	if err := vm.app.gui.DeleteView(viewName); err != nil {
		return err
	}

	// Pop mode and get the now-current mode
	vm.app.modes.Pop()

	// Rebind previous tags
	for _, tag := range previousTags {
		if err := vm.app.keys.BindTag(tag); err != nil {
			return err
		}
	}

	// Return to previous view
	if _, err := vm.app.gui.SetCurrentView(previousViewName); err != nil {
		return err
	}

	return nil
}

// CurrentTag returns the tag that matches the current mode
// This is a helper for mapping modes to their primary tags
func (vm *ViewManager) CurrentTag() string {
	mode := vm.app.modes.Current()
	return mode.String()
}
