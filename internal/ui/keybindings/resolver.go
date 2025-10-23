package keybindings

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// SetContext sets the current view and tab context for resolution
func (r *Resolver) SetContext(view string, tab string) {
	r.currentView = strings.ToLower(view)
	r.currentTab = strings.ToLower(tab)
}

// Resolve determines the action for a key press based on scope precedence
// Order: Modal → Tab → View → Global
func (r *Resolver) Resolve(msg tea.KeyMsg, modalActive bool) (action string, found bool) {
	key := normalizeKey(msg.String())

	// 1. Check modal scope if active
	if modalActive {
		if act, ok := r.findAction("modal", key); ok {
			return act, true
		}
	}

	// 2. Check tab scope
	if r.currentTab != "" {
		tabKey := "detail_" + r.currentTab
		if act, ok := r.findAction(tabKey, key); ok {
			return act, true
		}
	}

	// 3. Check view scope
	viewKey := "list_" + r.currentView
	if act, ok := r.findAction(viewKey, key); ok {
		return act, true
	}

	// 4. Check global scope
	if act, ok := r.findAction("global", key); ok {
		return act, true
	}

	return "", false
}

// findAction searches a scope for the given key and returns the action
func (r *Resolver) findAction(scope string, key string) (string, bool) {
	actions, ok := r.resolved[scope]
	if !ok {
		return "", false
	}

	for action, keys := range actions {
		if keys[key] {
			return action, true
		}
	}

	return "", false
}

// GetKeysForAction returns all keys bound to an action in a given scope
func (r *Resolver) GetKeysForAction(scope string, action string) []string {
	actions, ok := r.resolved[scope]
	if !ok {
		return []string{}
	}

	keys, ok := actions[action]
	if !ok {
		return []string{}
	}

	result := make([]string, 0, len(keys))
	for key := range keys {
		result = append(result, key)
	}

	return result
}

// Helper functions for common actions

// IsQuit checks if the key is a quit action
func (r *Resolver) IsQuit(msg tea.KeyMsg) bool {
	action, found := r.Resolve(msg, false)
	return found && action == "quit"
}

// IsHelp checks if the key is a help action
func (r *Resolver) IsHelp(msg tea.KeyMsg) bool {
	action, found := r.Resolve(msg, false)
	return found && action == "help"
}

// GetScopeKeybindings returns all keybindings for a given scope
func (r *Resolver) GetScopeKeybindings(scope string) map[string][]string {
	result := make(map[string][]string)
	actions, ok := r.resolved[scope]
	if !ok {
		return result
	}

	for action, keys := range actions {
		keyList := make([]string, 0, len(keys))
		for key := range keys {
			keyList = append(keyList, key)
		}
		result[action] = keyList
	}

	return result
}

// GetAllScopes returns a list of all available scopes
func (r *Resolver) GetAllScopes() []string {
	scopes := make([]string, 0, len(r.resolved))
	for scope := range r.resolved {
		scopes = append(scopes, scope)
	}
	return scopes
}
