package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
)

// KeyBindings holds all key bindings for the application
type KeyBindings struct {
	// Global bindings
	Quit       key.Binding
	Refresh    key.Binding
	HelpToggle key.Binding

	// List view bindings
	NavigateDown    key.Binding
	NavigateUp      key.Binding
	Select          key.Binding
	ToggleFilter    key.Binding
	NewSession      key.Binding
	ContinueSession key.Binding
	StartSession    key.Binding
	GoToWindow      key.Binding
	Archive         key.Binding
	ShowLogs        key.Binding

	// Detail view bindings
	Back      key.Binding
	ScrollUp  key.Binding
	ScrollDown key.Binding
	PageUp    key.Binding
	PageDown  key.Binding
}

// keyBindingsYAML represents the YAML structure for key bindings
type keyBindingsYAML struct {
	Global struct {
		Quit       []string `yaml:"quit"`
		Refresh    []string `yaml:"refresh"`
		HelpToggle []string `yaml:"help_toggle"`
	} `yaml:"global"`
	ListView struct {
		NavigateDown    []string `yaml:"navigate_down"`
		NavigateUp      []string `yaml:"navigate_up"`
		Select          []string `yaml:"select"`
		ToggleFilter    []string `yaml:"toggle_filter"`
		NewSession      []string `yaml:"new_session"`
		ContinueSession []string `yaml:"continue_session"`
		StartSession    []string `yaml:"start_session"`
		GoToWindow      []string `yaml:"go_to_window"`
		Archive         []string `yaml:"archive"`
		ShowLogs        []string `yaml:"show_logs"`
	} `yaml:"list_view"`
	DetailView struct {
		Back       []string `yaml:"back"`
		ScrollUp   []string `yaml:"scroll_up"`
		ScrollDown []string `yaml:"scroll_down"`
		PageUp     []string `yaml:"page_up"`
		PageDown   []string `yaml:"page_down"`
	} `yaml:"detail_view"`
}

// DefaultKeyBindings returns the default key bindings
func DefaultKeyBindings() KeyBindings {
	return KeyBindings{
		// Global
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "refresh"),
		),
		HelpToggle: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),

		// List view
		NavigateDown: key.NewBinding(
			key.WithKeys("j"),
			key.WithHelp("j", "move down"),
		),
		NavigateUp: key.NewBinding(
			key.WithKeys("k"),
			key.WithHelp("k", "move up"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select/details"),
		),
		ToggleFilter: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "toggle filter"),
		),
		NewSession: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new session"),
		),
		ContinueSession: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "continue session"),
		),
		StartSession: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "start session"),
		),
		GoToWindow: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "focus window"),
		),
		Archive: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "archive agent"),
		),
		ShowLogs: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "show logs"),
		),

		// Detail view
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		ScrollUp: key.NewBinding(
			key.WithKeys("k"),
			key.WithHelp("k", "scroll up"),
		),
		ScrollDown: key.NewBinding(
			key.WithKeys("j"),
			key.WithHelp("j", "scroll down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "page down"),
		),
	}
}

// LoadKeyBindings loads key bindings from YAML file or returns defaults
func LoadKeyBindings() (KeyBindings, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return DefaultKeyBindings(), err
	}

	keysPath := filepath.Join(configDir, "keys.yaml")

	// If keys file doesn't exist, create it with defaults
	if _, err := os.Stat(keysPath); os.IsNotExist(err) {
		keys := DefaultKeyBindings()
		if err := SaveKeyBindingsYAML(keys); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save default keybindings: %v\n", err)
		}
		return keys, nil
	}

	// Read keys file
	data, err := os.ReadFile(keysPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to read keys file: %v, using defaults\n", err)
		return DefaultKeyBindings(), nil
	}

	// Parse YAML
	var yamlKeys keyBindingsYAML
	if err := yaml.Unmarshal(data, &yamlKeys); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to parse keys file: %v, using defaults\n", err)
		return DefaultKeyBindings(), nil
	}

	// Convert YAML to KeyBindings with conflict detection
	keys, conflicts := convertYAMLToBindings(yamlKeys)
	if len(conflicts) > 0 {
		fmt.Fprintf(os.Stderr, "warning: key conflicts detected, using defaults for conflicting keys:\n")
		for _, conflict := range conflicts {
			fmt.Fprintf(os.Stderr, "  - %s\n", conflict)
		}
	}

	return keys, nil
}

// convertYAMLToBindings converts YAML structure to KeyBindings
func convertYAMLToBindings(yamlKeys keyBindingsYAML) (KeyBindings, []string) {
	defaults := DefaultKeyBindings()
	keys := KeyBindings{}
	var conflicts []string

	// Track used keys for conflict detection
	globalKeys := make(map[string]string)
	listKeys := make(map[string]string)
	detailKeys := make(map[string]string)

	// Helper to create binding with conflict detection
	createBinding := func(yamlKeys []string, defaultBinding key.Binding, name string, context string, keyMap map[string]string) key.Binding {
		if len(yamlKeys) == 0 {
			return defaultBinding
		}

		// Check for conflicts
		for _, k := range yamlKeys {
			if existing, exists := keyMap[k]; exists {
				conflicts = append(conflicts, fmt.Sprintf("key '%s' used for both %s and %s", k, existing, name))
				return defaultBinding
			}
			keyMap[k] = name
		}

		// Get help from default binding
		help := defaultBinding.Help()
		return key.NewBinding(
			key.WithKeys(yamlKeys...),
			key.WithHelp(help.Key, help.Desc),
		)
	}

	// Global bindings
	keys.Quit = createBinding(yamlKeys.Global.Quit, defaults.Quit, "quit", "global", globalKeys)
	keys.Refresh = createBinding(yamlKeys.Global.Refresh, defaults.Refresh, "refresh", "global", globalKeys)
	keys.HelpToggle = createBinding(yamlKeys.Global.HelpToggle, defaults.HelpToggle, "help_toggle", "global", globalKeys)

	// List view bindings
	keys.NavigateDown = createBinding(yamlKeys.ListView.NavigateDown, defaults.NavigateDown, "navigate_down", "list_view", listKeys)
	keys.NavigateUp = createBinding(yamlKeys.ListView.NavigateUp, defaults.NavigateUp, "navigate_up", "list_view", listKeys)
	keys.Select = createBinding(yamlKeys.ListView.Select, defaults.Select, "select", "list_view", listKeys)
	keys.ToggleFilter = createBinding(yamlKeys.ListView.ToggleFilter, defaults.ToggleFilter, "toggle_filter", "list_view", listKeys)
	keys.NewSession = createBinding(yamlKeys.ListView.NewSession, defaults.NewSession, "new_session", "list_view", listKeys)
	keys.ContinueSession = createBinding(yamlKeys.ListView.ContinueSession, defaults.ContinueSession, "continue_session", "list_view", listKeys)
	keys.StartSession = createBinding(yamlKeys.ListView.StartSession, defaults.StartSession, "start_session", "list_view", listKeys)
	keys.GoToWindow = createBinding(yamlKeys.ListView.GoToWindow, defaults.GoToWindow, "go_to_window", "list_view", listKeys)
	keys.Archive = createBinding(yamlKeys.ListView.Archive, defaults.Archive, "archive", "list_view", listKeys)
	keys.ShowLogs = createBinding(yamlKeys.ListView.ShowLogs, defaults.ShowLogs, "show_logs", "list_view", listKeys)

	// Detail view bindings
	keys.Back = createBinding(yamlKeys.DetailView.Back, defaults.Back, "back", "detail_view", detailKeys)
	keys.ScrollUp = createBinding(yamlKeys.DetailView.ScrollUp, defaults.ScrollUp, "scroll_up", "detail_view", detailKeys)
	keys.ScrollDown = createBinding(yamlKeys.DetailView.ScrollDown, defaults.ScrollDown, "scroll_down", "detail_view", detailKeys)
	keys.PageUp = createBinding(yamlKeys.DetailView.PageUp, defaults.PageUp, "page_up", "detail_view", detailKeys)
	keys.PageDown = createBinding(yamlKeys.DetailView.PageDown, defaults.PageDown, "page_down", "detail_view", detailKeys)

	return keys, conflicts
}

// SaveKeyBindingsYAML saves key bindings to YAML file
func SaveKeyBindingsYAML(keys KeyBindings) error {
	configDir, err := ConfigDir()
	if err != nil {
		return err
	}

	keysPath := filepath.Join(configDir, "keys.yaml")

	// Convert KeyBindings to YAML structure
	yamlKeys := keyBindingsYAML{}
	yamlKeys.Global.Quit = keys.Quit.Keys()
	yamlKeys.Global.Refresh = keys.Refresh.Keys()
	yamlKeys.Global.HelpToggle = keys.HelpToggle.Keys()

	yamlKeys.ListView.NavigateDown = keys.NavigateDown.Keys()
	yamlKeys.ListView.NavigateUp = keys.NavigateUp.Keys()
	yamlKeys.ListView.Select = keys.Select.Keys()
	yamlKeys.ListView.ToggleFilter = keys.ToggleFilter.Keys()
	yamlKeys.ListView.NewSession = keys.NewSession.Keys()
	yamlKeys.ListView.ContinueSession = keys.ContinueSession.Keys()
	yamlKeys.ListView.StartSession = keys.StartSession.Keys()
	yamlKeys.ListView.GoToWindow = keys.GoToWindow.Keys()
	yamlKeys.ListView.Archive = keys.Archive.Keys()
	yamlKeys.ListView.ShowLogs = keys.ShowLogs.Keys()

	yamlKeys.DetailView.Back = keys.Back.Keys()
	yamlKeys.DetailView.ScrollUp = keys.ScrollUp.Keys()
	yamlKeys.DetailView.ScrollDown = keys.ScrollDown.Keys()
	yamlKeys.DetailView.PageUp = keys.PageUp.Keys()
	yamlKeys.DetailView.PageDown = keys.PageDown.Keys()

	data, err := yaml.Marshal(yamlKeys)
	if err != nil {
		return fmt.Errorf("failed to marshal keys: %w", err)
	}

	if err := os.WriteFile(keysPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write keys file: %w", err)
	}

	return nil
}

// ShortHelp returns short help for the key bindings
func (k KeyBindings) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit, k.Refresh, k.HelpToggle, k.Select, k.ToggleFilter}
}

// FullHelp returns full help for the key bindings grouped by context
func (k KeyBindings) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// Global
		{k.Quit, k.Refresh, k.HelpToggle},
		// List view
		{k.NavigateDown, k.NavigateUp, k.Select, k.ToggleFilter},
		{k.NewSession, k.ContinueSession, k.StartSession},
		{k.GoToWindow, k.Archive, k.ShowLogs},
		// Detail view
		{k.Back, k.ScrollUp, k.ScrollDown, k.PageUp, k.PageDown},
	}
}

// NewHelp creates a new help model
func NewHelp() help.Model {
	h := help.New()
	h.ShowAll = false
	return h
}

// Matches checks if a key message matches a key binding
func Matches(msg tea.KeyMsg, binding key.Binding) bool {
	return key.Matches(msg, binding)
}
