package keybindings

import (
	"testing"
)

// TestResolverPrecedence tests that resolver respects scope precedence
func TestResolverPrecedence(t *testing.T) {
	// Create a test resolver with overlapping keybindings
	resolver := &Resolver{
		config: KeybindingsConfig{
			Global: map[string][]string{
				"quit": {"ctrl+c", "shift+q"},
				"help": {"?"},
			},
			List: map[string]map[string][]string{
				"overview": {
					"new_session": {"n"},
					"quit":        {"q"}, // Override global quit
				},
			},
		},
		resolved: make(ResolvedKeybindings),
	}

	// Build resolved map manually for testing
	resolver.resolved = normalizeKeybindings(&resolver.config)

	tests := []struct {
		name     string
		scope    string
		action   string
		key      string
		expected bool
	}{
		{
			name:     "Global quit with ctrl+c",
			scope:    "global",
			action:   "quit",
			key:      "ctrl+c",
			expected: true,
		},
		{
			name:     "Global quit with shift+q",
			scope:    "global",
			action:   "quit",
			key:      "shift+q",
			expected: true,
		},
		{
			name:     "Global help with ?",
			scope:    "global",
			action:   "help",
			key:      "?",
			expected: true,
		},
		{
			name:     "List overview new_session",
			scope:    "list_overview",
			action:   "new_session",
			key:      "n",
			expected: true,
		},
		{
			name:     "List overview quit override",
			scope:    "list_overview",
			action:   "quit",
			key:      "q",
			expected: true,
		},
		{
			name:     "Invalid action",
			scope:    "global",
			action:   "nonexistent",
			key:      "x",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if scopeBindings, ok := resolver.resolved[tt.scope]; ok {
				if actionBindings, ok := scopeBindings[tt.action]; ok {
					result := actionBindings[normalizeKey(tt.key)]
					if result != tt.expected {
						t.Errorf("got %v, want %v", result, tt.expected)
					}
					return
				}
			}
			if tt.expected {
				t.Errorf("expected to find binding for %s:%s:%s", tt.scope, tt.action, tt.key)
			}
		})
	}
}

// TestKeyNormalization tests that keys are normalized consistently
func TestKeyNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Ctrl+C", "ctrl+c"},
		{"SHIFT+Q", "shift+q"},
		{"  ?  ", "?"},
		{"Enter", "enter"},
		{"n", "n"},
		{"N", "n"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeKey(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestGetScopeKeybindings tests getting keybindings for a scope
func TestGetScopeKeybindings(t *testing.T) {
	resolver := &Resolver{
		config: KeybindingsConfig{
			Global: map[string][]string{
				"quit": {"ctrl+c"},
				"help": {"?"},
			},
			List: map[string]map[string][]string{
				"overview": {
					"new_session": {"n"},
				},
			},
		},
		resolved: make(ResolvedKeybindings),
	}

	resolver.resolved = normalizeKeybindings(&resolver.config)

	// Get global scope keybindings
	globalBindings := resolver.GetScopeKeybindings("global")
	if len(globalBindings) != 2 {
		t.Errorf("expected 2 global bindings, got %d", len(globalBindings))
	}

	if _, ok := globalBindings["quit"]; !ok {
		t.Error("expected 'quit' action in global bindings")
	}

	if _, ok := globalBindings["help"]; !ok {
		t.Error("expected 'help' action in global bindings")
	}

	// Get list_overview scope keybindings
	listBindings := resolver.GetScopeKeybindings("list_overview")
	if len(listBindings) != 1 {
		t.Errorf("expected 1 overview binding, got %d", len(listBindings))
	}

	if _, ok := listBindings["new_session"]; !ok {
		t.Error("expected 'new_session' action in overview bindings")
	}
}

// TestGetKeysForAction tests getting keys for a specific action
func TestGetKeysForAction(t *testing.T) {
	resolver := &Resolver{
		config: KeybindingsConfig{
			Global: map[string][]string{
				"quit": {"ctrl+c", "shift+q"},
				"help": {"?"},
			},
		},
		resolved: make(ResolvedKeybindings),
	}

	resolver.resolved = normalizeKeybindings(&resolver.config)

	tests := []struct {
		scope   string
		action  string
		minKeys int
	}{
		{"global", "quit", 2},
		{"global", "help", 1},
		{"global", "nonexistent", 0},
	}

	for _, tt := range tests {
		t.Run(tt.scope+":"+tt.action, func(t *testing.T) {
			keys := resolver.GetKeysForAction(tt.scope, tt.action)
			if len(keys) < tt.minKeys {
				t.Errorf("GetKeysForAction(%q, %q) returned %d keys, want at least %d", tt.scope, tt.action, len(keys), tt.minKeys)
			}
		})
	}
}
