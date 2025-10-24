package keybindings

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadKeybindings loads and parses keybindings from YAML file
func LoadKeybindings() (*Resolver, error) {
	// Locate keybindings file
	path := GetKeybindingsPath()

	// Check if file exists; if not, create defaults
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("Keybindings file not found at %s, creating defaults\n", path)
		if err := createDefaultKeybindings(path); err != nil {
			return nil, fmt.Errorf("failed to create default keybindings: %w", err)
		}
	}

	// Read and unmarshal YAML
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read keybindings file: %w", err)
	}

	var config KeybindingsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal keybindings YAML: %w", err)
	}

	// Normalize and create resolver
	resolver := &Resolver{
		config:   config,
		resolved: normalizeKeybindings(&config),
	}

	return resolver, nil
}

// LoadKeybindingsDefaults loads hardcoded keybindings (no YAML parsing)
func LoadKeybindingsDefaults() (*Resolver, error) {
	// Build keybindings directly in code - NO YAML parsing
	resolved := make(ResolvedKeybindings)

	// Global keybindings
	resolved["global"] = map[string]map[string]bool{
		"quit":              {"ctrl+c": true, "shift+q": true},
		"help":              {"?": true, "shift+h": true},
		"toggle_filter":     {"a": true},
		"continue_session":  {"c": true},
		"start_session":     {"s": true},
		"go_to_window":      {"w": true},
		"archive":           {"d": true},
	}

	// List view keybindings
	resolved["list_overview"] = map[string]map[string]bool{
		"new_session": {"n": true},
		"select":      {"enter": true, "right": true},
		"down":        {"j": true, "down": true},
		"up":          {"k": true, "up": true},
	}

	resolved["list_project"] = map[string]map[string]bool{
		"new_ticket": {"n": true},
		"select":     {"enter": true},
		"down":       {"j": true, "down": true},
		"up":         {"k": true, "up": true},
	}

	resolved["list_claude"] = map[string]map[string]bool{
		"validate":   {"v": true},
		"edit":       {"e": true},
		"refresh":    {"r": true},
		"open":       {"enter": true},
		"down":       {"j": true, "down": true},
		"up":         {"k": true, "up": true},
		"page_down":  {"pgdown": true},
		"page_up":    {"pgup": true},
	}

	resolved["list_usage"] = map[string]map[string]bool{
		"down":      {"j": true, "down": true},
		"up":        {"k": true, "up": true},
		"page_down": {"pgdown": true},
		"page_up":   {"pgup": true},
	}

	resolved["list_config"] = map[string]map[string]bool{
		"down":      {"j": true, "down": true},
		"up":        {"k": true, "up": true},
		"page_down": {"pgdown": true},
		"page_up":   {"pgup": true},
	}

	// Detail view keybindings
	resolved["detail_overview"] = map[string]map[string]bool{
		"select": {"enter": true},
	}

	resolved["detail_commits"] = map[string]map[string]bool{
		"down":      {"j": true, "down": true},
		"up":        {"k": true, "up": true},
		"refresh":   {"r": true},
		"page_down": {"pgdown": true},
		"page_up":   {"pgup": true},
	}

	resolved["detail_logs"] = map[string]map[string]bool{
		"down":      {"j": true, "down": true},
		"up":        {"k": true, "up": true},
		"refresh":   {"r": true},
		"page_down": {"pgdown": true},
		"page_up":   {"pgup": true},
	}

	resolved["detail_notes"] = map[string]map[string]bool{
		"down":       {"j": true, "down": true},
		"up":         {"k": true, "up": true},
		"refresh":    {"r": true},
		"new_note":   {"n": true},
		"page_down":  {"pgdown": true},
		"page_up":    {"pgup": true},
	}

	resolved["detail_actions"] = map[string]map[string]bool{
		"down":      {"j": true, "down": true},
		"up":        {"k": true, "up": true},
		"refresh":   {"r": true},
		"page_down": {"pgdown": true},
		"page_up":   {"pgup": true},
	}

	resolved["detail_tasks"] = map[string]map[string]bool{
		"down":       {"j": true, "down": true},
		"up":         {"k": true, "up": true},
		"refresh":    {"r": true},
		"page_down":  {"pgdown": true},
		"page_up":    {"pgup": true},
	}

	resolver := &Resolver{
		config:   KeybindingsConfig{},
		resolved: resolved,
	}

	log.Printf("Loaded hardcoded keybindings\n")
	return resolver, nil
}

// GetKeybindingsPath returns the path to the keybindings YAML file
func GetKeybindingsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "~"
	}
	return filepath.Join(home, ".eye-in-the-sky", "keybindings.yaml")
}

// normalizeKeybindings converts the YAML config into a normalized resolver map
func normalizeKeybindings(config *KeybindingsConfig) ResolvedKeybindings {
	resolved := make(ResolvedKeybindings)

	// Initialize scopes
	resolved["global"] = make(map[string]map[string]bool)
	resolved["list"] = make(map[string]map[string]bool)
	resolved["detail"] = make(map[string]map[string]bool)

	// Process global keybindings
	if config.Global != nil {
		for action, keys := range config.Global {
			resolved["global"][action] = make(map[string]bool)
			for _, key := range keys {
				resolved["global"][action][normalizeKey(key)] = true
			}
		}
	}

	// Process list view keybindings (overview, project, claude, usage)
	if config.List != nil {
		for view, actions := range config.List {
			viewKey := "list_" + view
			resolved[viewKey] = make(map[string]map[string]bool)
			for action, keys := range actions {
				resolved[viewKey][action] = make(map[string]bool)
				for _, key := range keys {
					resolved[viewKey][action][normalizeKey(key)] = true
				}
			}
		}
	}

	// Process detail view keybindings (tabs)
	if config.Detail != nil {
		for tab, actions := range config.Detail {
			tabKey := "detail_" + tab
			resolved[tabKey] = make(map[string]map[string]bool)
			for action, keys := range actions {
				resolved[tabKey][action] = make(map[string]bool)
				for _, key := range keys {
					resolved[tabKey][action][normalizeKey(key)] = true
				}
			}
		}
	}

	return resolved
}

// normalizeKey converts key strings to lowercase for consistent matching
func normalizeKey(key string) string {
	return strings.ToLower(strings.TrimSpace(key))
}

// createDefaultKeybindings creates a default keybindings YAML file
func createDefaultKeybindings(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	defaultConfig := `keybindings:
  global:
    quit: ["ctrl+c", "shift+q"]
    help: ["ctrl+h"]
    toggle_filter: ["a"]
    continue_session: ["c"]
    start_session: ["s"]
    go_to_window: ["w"]
    archive: ["D"]

  list:
    overview:
      new_session: ["n"]
      select: ["enter", "right"]
      down: ["j", "down"]
      up: ["k", "up"]

    project:
      new_ticket: ["n"]
      select: ["enter"]
      down: ["j", "down"]
      up: ["k", "up"]

    claude:
      validate: ["v"]
      edit: ["e"]
      refresh: ["r"]
      open: ["enter"]
      down: ["j", "down"]
      up: ["k", "up"]
      page_down: ["pgdown"]
      page_up: ["pgup"]

    usage:
      down: ["j", "down"]
      up: ["k", "up"]
      page_down: ["pgdown"]
      page_up: ["pgup"]

  detail:
    overview:
      select: ["enter"]
    commits:
      refresh: ["r"]
    logs:
      refresh: ["r"]
    notes:
      new_note: ["n"]
    actions:
      rerun: ["r"]
    tasks:
      new_task: ["n"]
`

	if err := os.WriteFile(path, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write keybindings file: %w", err)
	}

	log.Printf("Created default keybindings at %s\n", path)
	return nil
}

// LoadKeybindingsYAML reads and returns the keybindings YAML file content as a string
func LoadKeybindingsYAML() (string, error) {
	path := GetKeybindingsPath()

	// Check if file exists; if not, create defaults
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := createDefaultKeybindings(path); err != nil {
			return "", fmt.Errorf("failed to create default keybindings: %w", err)
		}
	}

	// Read file content
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read keybindings file: %w", err)
	}

	return string(data), nil
}

// SaveKeybindingsYAML atomically writes keybindings YAML content to file
func SaveKeybindingsYAML(content string) error {
	path := GetKeybindingsPath()
	dir := filepath.Dir(path)

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write to temporary file first (atomic save)
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Atomically rename temp file to actual file
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath) // Clean up temp file on error
		return fmt.Errorf("failed to save keybindings: %w", err)
	}

	log.Printf("Saved keybindings to %s\n", path)
	return nil
}
