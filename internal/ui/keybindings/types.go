package keybindings

// KeyScope represents the current context for key binding resolution
type KeyScope int

const (
	ScopeGlobal KeyScope = iota
	ScopeModal
	ScopeView
	ScopeTab
)

// KeybindingsConfig holds the structure of keybindings from YAML
type KeybindingsConfig struct {
	Global map[string][]string            `yaml:"global"`
	List   map[string]map[string][]string `yaml:"list"`   // overview, project, claude, usage
	Detail map[string]map[string][]string `yaml:"detail"` // overview, commits, logs, notes, actions, tasks, projects
}

// ResolvedKeybindings normalizes keybindings into a searchable format
// Structure: map[scope][action][key] = true
type ResolvedKeybindings map[string]map[string]map[string]bool

// Resolver provides key binding resolution with scope precedence
type Resolver struct {
	config    KeybindingsConfig
	resolved  ResolvedKeybindings
	currentView string
	currentTab  string
}
