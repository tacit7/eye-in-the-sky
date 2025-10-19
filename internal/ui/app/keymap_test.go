package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/bubbles/key"
)

func TestDefaultKeyBindings(t *testing.T) {
	keys := DefaultKeyBindings()

	// Test that all bindings are initialized
	tests := []struct {
		name    string
		binding key.Binding
	}{
		{"Quit", keys.Quit},
		{"Refresh", keys.Refresh},
		{"HelpToggle", keys.HelpToggle},
		{"NavigateDown", keys.NavigateDown},
		{"NavigateUp", keys.NavigateUp},
		{"Select", keys.Select},
		{"ToggleFilter", keys.ToggleFilter},
		{"NewSession", keys.NewSession},
		{"ContinueSession", keys.ContinueSession},
		{"StartSession", keys.StartSession},
		{"GoToWindow", keys.GoToWindow},
		{"Archive", keys.Archive},
		{"ShowLogs", keys.ShowLogs},
		{"Back", keys.Back},
		{"ScrollUp", keys.ScrollUp},
		{"ScrollDown", keys.ScrollDown},
		{"PageUp", keys.PageUp},
		{"PageDown", keys.PageDown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.binding.Keys()) == 0 {
				t.Errorf("%s binding has no keys", tt.name)
			}
		})
	}
}

func TestLoadKeyBindingsCreatesDefaultFile(t *testing.T) {
	// Create temporary config directory
	tempDir := t.TempDir()
	oldConfigDir := ConfigDir
	defer func() { ConfigDir = oldConfigDir }()

	// Override ConfigDir to return temp directory
	ConfigDir = func() (string, error) {
		return tempDir, nil
	}

	keys, err := LoadKeyBindings()
	if err != nil {
		t.Fatalf("LoadKeyBindings failed: %v", err)
	}

	// Verify default keys were returned
	if len(keys.Quit.Keys()) == 0 {
		t.Error("Expected default Quit binding")
	}

	// Verify file was created
	keysPath := filepath.Join(tempDir, "keys.yaml")
	if _, err := os.Stat(keysPath); os.IsNotExist(err) {
		t.Error("Expected keys.yaml to be created")
	}
}

func TestLoadKeyBindingsWithValidYAML(t *testing.T) {
	// Create temporary config directory
	tempDir := t.TempDir()
	oldConfigDir := ConfigDir
	defer func() { ConfigDir = oldConfigDir }()

	ConfigDir = func() (string, error) {
		return tempDir, nil
	}

	// Write valid YAML
	yamlContent := `
global:
  quit: ["q", "ctrl+c"]
  refresh: ["r"]
  help_toggle: ["?"]

list_view:
  navigate_down: ["j"]
  navigate_up: ["k"]
  select: ["enter"]
  toggle_filter: ["a"]
  new_session: ["n"]
  continue_session: ["c"]
  start_session: ["s"]
  go_to_window: ["w"]
  archive: ["D"]
  show_logs: ["L"]

detail_view:
  back: ["esc"]
  scroll_up: ["k"]
  scroll_down: ["j"]
  page_up: ["ctrl+u"]
  page_down: ["ctrl+d"]
`

	keysPath := filepath.Join(tempDir, "keys.yaml")
	if err := os.WriteFile(keysPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test YAML: %v", err)
	}

	keys, err := LoadKeyBindings()
	if err != nil {
		t.Fatalf("LoadKeyBindings failed: %v", err)
	}

	// Verify keys were loaded correctly
	if keys.Quit.Keys()[0] != "q" {
		t.Errorf("Expected Quit key 'q', got '%s'", keys.Quit.Keys()[0])
	}
	if keys.NavigateDown.Keys()[0] != "j" {
		t.Errorf("Expected NavigateDown key 'j', got '%s'", keys.NavigateDown.Keys()[0])
	}
}

func TestLoadKeyBindingsWithInvalidYAML(t *testing.T) {
	// Create temporary config directory
	tempDir := t.TempDir()
	oldConfigDir := ConfigDir
	defer func() { ConfigDir = oldConfigDir }()

	ConfigDir = func() (string, error) {
		return tempDir, nil
	}

	// Write invalid YAML
	yamlContent := `invalid: yaml: content: [[[`
	keysPath := filepath.Join(tempDir, "keys.yaml")
	if err := os.WriteFile(keysPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test YAML: %v", err)
	}

	// Should return defaults on parse error (not fail)
	keys, err := LoadKeyBindings()
	if err != nil {
		t.Fatalf("LoadKeyBindings should not fail on invalid YAML: %v", err)
	}

	// Should have defaults
	if len(keys.Quit.Keys()) == 0 {
		t.Error("Expected default keys on invalid YAML")
	}
}

func TestConvertYAMLToBindingsWithConflicts(t *testing.T) {
	yaml := keyBindingsYAML{}
	yaml.Global.Quit = []string{"q"}
	yaml.Global.Refresh = []string{"q"} // Conflict!
	yaml.ListView.NavigateDown = []string{"j"}
	yaml.ListView.NavigateUp = []string{"j"} // Conflict!

	keys, conflicts := convertYAMLToBindings(yaml)

	// Should have detected conflicts
	if len(conflicts) != 2 {
		t.Errorf("Expected 2 conflicts, got %d", len(conflicts))
	}

	// Conflicting keys should use defaults
	defaults := DefaultKeyBindings()
	if keys.Refresh.Keys()[0] != defaults.Refresh.Keys()[0] {
		t.Error("Expected conflicting Refresh to use default")
	}
	if keys.NavigateUp.Keys()[0] != defaults.NavigateUp.Keys()[0] {
		t.Error("Expected conflicting NavigateUp to use default")
	}
}

func TestConvertYAMLToBindingsNoConflictsAcrossViews(t *testing.T) {
	// Same key can exist in different views (list vs detail)
	yaml := keyBindingsYAML{}
	yaml.ListView.NavigateDown = []string{"j"}
	yaml.DetailView.ScrollDown = []string{"j"} // Not a conflict - different view

	keys, conflicts := convertYAMLToBindings(yaml)

	// Should NOT have conflicts
	if len(conflicts) != 0 {
		t.Errorf("Expected no conflicts across views, got %d", len(conflicts))
	}

	// Both should have 'j'
	if keys.NavigateDown.Keys()[0] != "j" {
		t.Error("Expected NavigateDown to have 'j'")
	}
	if keys.ScrollDown.Keys()[0] != "j" {
		t.Error("Expected ScrollDown to have 'j'")
	}
}

func TestShortHelp(t *testing.T) {
	keys := DefaultKeyBindings()
	shortHelp := keys.ShortHelp()

	if len(shortHelp) == 0 {
		t.Error("Expected short help to have bindings")
	}

	// Should include key bindings
	if len(shortHelp) < 3 {
		t.Errorf("Expected at least 3 short help bindings, got %d", len(shortHelp))
	}
}

func TestFullHelp(t *testing.T) {
	keys := DefaultKeyBindings()
	fullHelp := keys.FullHelp()

	if len(fullHelp) == 0 {
		t.Error("Expected full help to have binding groups")
	}

	// Should have multiple groups
	if len(fullHelp) < 3 {
		t.Errorf("Expected at least 3 help groups, got %d", len(fullHelp))
	}
}

func TestSaveKeyBindingsYAML(t *testing.T) {
	// Create temporary config directory
	tempDir := t.TempDir()
	oldConfigDir := ConfigDir
	defer func() { ConfigDir = oldConfigDir }()

	ConfigDir = func() (string, error) {
		return tempDir, nil
	}

	keys := DefaultKeyBindings()
	if err := SaveKeyBindingsYAML(keys); err != nil {
		t.Fatalf("SaveKeyBindingsYAML failed: %v", err)
	}

	// Verify file was created
	keysPath := filepath.Join(tempDir, "keys.yaml")
	if _, err := os.Stat(keysPath); os.IsNotExist(err) {
		t.Error("Expected keys.yaml to be created")
	}

	// Verify can be loaded back
	loadedKeys, err := LoadKeyBindings()
	if err != nil {
		t.Fatalf("LoadKeyBindings failed after save: %v", err)
	}

	if loadedKeys.Quit.Keys()[0] != keys.Quit.Keys()[0] {
		t.Error("Saved and loaded keys don't match")
	}
}
