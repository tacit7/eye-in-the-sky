package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	// Refresh interval in seconds
	RefreshInterval int `yaml:"refresh_interval"`

	// Database path
	DatabasePath string `yaml:"database_path"`

	// Default view mode (list or detail)
	DefaultView string `yaml:"default_view"`

	// Show all agents or only active ones
	ShowAllAgents bool `yaml:"show_all_agents"`

	// Theme name
	Theme string `yaml:"theme"`

	// Claude binary path (empty to use $PATH)
	ClaudePath string `yaml:"claude_path"`
}

// Theme holds color and style configuration
type Theme struct {
	Name   string `yaml:"name"`
	Colors struct {
		Active    string `yaml:"active"`
		Working   string `yaml:"working"`
		Idle      string `yaml:"idle"`
		Stale     string `yaml:"stale"`
		Unknown   string `yaml:"unknown"`
		Completed string `yaml:"completed"`
		Failed    string `yaml:"failed"`
		Primary   string `yaml:"primary"`
		Secondary string `yaml:"secondary"`
		Border    string `yaml:"border"`
		Title     string `yaml:"title"`
		Text      string `yaml:"text"`
		Subtle    string `yaml:"subtle"`
	} `yaml:"colors"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	homeDir, _ := os.UserHomeDir()
	return Config{
		RefreshInterval: 5,
		DatabasePath:    filepath.Join(homeDir, "projects", "eye-in-the-sky", "data", "agents.db"),
		DefaultView:     "list",
		ShowAllAgents:   false,
		Theme:           "default",
		ClaudePath:      "", // Empty = use $PATH
	}
}

// DefaultTheme returns the default theme
func DefaultTheme() Theme {
	theme := Theme{
		Name: "default",
	}
	theme.Colors.Active = "#00FF00"    // Green
	theme.Colors.Working = "#0088FF"   // Blue
	theme.Colors.Idle = "#FFAA00"      // Yellow
	theme.Colors.Stale = "#888888"     // Gray
	theme.Colors.Unknown = "#888888"   // Gray
	theme.Colors.Completed = "#00FFFF" // Cyan
	theme.Colors.Failed = "#FF0000"    // Red
	theme.Colors.Primary = "#00AAFF"   // Light Blue
	theme.Colors.Secondary = "#8888FF" // Purple
	theme.Colors.Border = "#444444"    // Dark Gray
	theme.Colors.Title = "#FFFFFF"     // White
	theme.Colors.Text = "#CCCCCC"      // Light Gray
	theme.Colors.Subtle = "#666666"    // Medium Gray
	return theme
}

// ConfigDir returns the configuration directory path
func ConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", ".eyeinthesky"), nil
}

// LoadConfig loads configuration from file or creates default
func LoadConfig() (Config, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return Config{}, err
	}

	configPath := filepath.Join(configDir, "config.yaml")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return Config{}, fmt.Errorf("failed to create config directory: %w", err)
	}

	// If config file doesn't exist, create it with defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := DefaultConfig()
		if err := SaveConfig(config); err != nil {
			return config, fmt.Errorf("failed to save default config: %w", err)
		}
		return config, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config Config) error {
	configDir, err := ConfigDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.yaml")

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadTheme loads theme from file or creates default
func LoadTheme(themeName string) (Theme, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return Theme{}, err
	}

	themePath := filepath.Join(configDir, fmt.Sprintf("theme-%s.yaml", themeName))

	// If theme file doesn't exist, create it with defaults
	if _, err := os.Stat(themePath); os.IsNotExist(err) {
		theme := DefaultTheme()
		if err := SaveTheme(theme); err != nil {
			return theme, fmt.Errorf("failed to save default theme: %w", err)
		}
		return theme, nil
	}

	// Read theme file
	data, err := os.ReadFile(themePath)
	if err != nil {
		return Theme{}, fmt.Errorf("failed to read theme file: %w", err)
	}

	// Parse YAML
	var theme Theme
	if err := yaml.Unmarshal(data, &theme); err != nil {
		return Theme{}, fmt.Errorf("failed to parse theme file: %w", err)
	}

	return theme, nil
}

// SaveTheme saves theme to file
func SaveTheme(theme Theme) error {
	configDir, err := ConfigDir()
	if err != nil {
		return err
	}

	themePath := filepath.Join(configDir, fmt.Sprintf("theme-%s.yaml", theme.Name))

	data, err := yaml.Marshal(theme)
	if err != nil {
		return fmt.Errorf("failed to marshal theme: %w", err)
	}

	if err := os.WriteFile(themePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write theme file: %w", err)
	}

	return nil
}

// RefreshDuration returns the refresh interval as a time.Duration
func (c *Config) RefreshDuration() time.Duration {
	if c.RefreshInterval <= 0 {
		return 5 * time.Second
	}
	return time.Duration(c.RefreshInterval) * time.Second
}
