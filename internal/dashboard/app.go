package dashboard

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/tacit7/eye-in-the-sky/internal/database"
)

// App represents the TUI dashboard application
type App struct {
	gui            *gocui.Gui
	db             *database.DB
	config         *Config
	keymap         *Keymap
	agents         []*database.Agent
	selectedIdx    int
	filter         string
	ticker         *time.Ticker
	quitChan       chan bool
	currentAgentID string // Agent ID currently being viewed in detail
}

// Config represents the dashboard configuration
type Config struct {
	PollInterval   int               `json:"poll_interval"`
	DefaultFilter  string            `json:"default_filter"`
	Colors         map[string]string `json:"colors"`
	ScrollMode     string            `json:"scroll_mode"`
	PaginationMode string            `json:"pagination_mode"`
}

// Keymap represents the key bindings configuration
type Keymap struct {
	Quit             []string `json:"quit"`
	Refresh          []string `json:"refresh"`
	ContinueSession  []string `json:"continue_session"`
	GoToWindow       []string `json:"go_to_window"`
	ToggleAllAgents  []string `json:"toggle_all_agents"`
	Up               []string `json:"up"`
	Down             []string `json:"down"`
	Logs             []string `json:"logs"`
	CommandMode      []string `json:"command_mode"`
	Archive          []string `json:"archive"`
	ViewDetails      []string `json:"view_details"`
}

// NewApp creates a new dashboard application
func NewApp(db *database.DB, configPath, keysPath string) (*App, error) {
	// Load configuration
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Load keymap
	keymap, err := loadKeymap(keysPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load keymap: %w", err)
	}

	// Initialize gocui
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return nil, fmt.Errorf("failed to create GUI: %w", err)
	}

	app := &App{
		gui:        g,
		db:         db,
		config:     config,
		keymap:     keymap,
		filter:     config.DefaultFilter,
		quitChan:   make(chan bool),
	}

	return app, nil
}

// Run starts the dashboard application
func (a *App) Run() error {
	defer a.gui.Close()

	// Set up GUI
	a.gui.SetManagerFunc(a.layout)
	a.gui.Cursor = true
	a.gui.Mouse = false

	// Set up keybindings
	if err := a.setupKeybindings(); err != nil {
		return err
	}

	// Load initial data (before starting main loop)
	agents, err := a.db.ListAgents(a.filter)
	if err != nil {
		return err
	}

	// Filter archived agents
	filtered := make([]*database.Agent, 0)
	for _, agent := range agents {
		if agent.Status != database.StatusArchived {
			filtered = append(filtered, agent)
		}
	}
	a.agents = filtered

	// Start polling ticker
	a.ticker = time.NewTicker(time.Duration(a.config.PollInterval) * time.Second)
	go a.pollLoop()

	// Start main loop
	if err := a.gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}

	// Stop ticker
	a.ticker.Stop()
	a.quitChan <- true

	return nil
}

// pollLoop polls the database for updates
func (a *App) pollLoop() {
	for {
		select {
		case <-a.ticker.C:
			a.gui.Update(func(g *gocui.Gui) error {
				return a.refreshAgents()
			})
		case <-a.quitChan:
			return
		}
	}
}

// refreshAgents reloads agents from the database
func (a *App) refreshAgents() error {
	agents, err := a.db.ListAgents(a.filter)
	if err != nil {
		return err
	}

	// Filter archived agents
	filtered := make([]*database.Agent, 0)
	for _, agent := range agents {
		if agent.Status != database.StatusArchived {
			filtered = append(filtered, agent)
		}
	}

	a.agents = filtered
	return a.renderAgents()
}

// loadConfig loads configuration from file
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// loadKeymap loads keymap from file
func loadKeymap(path string) (*Keymap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var keymap Keymap
	if err := json.Unmarshal(data, &keymap); err != nil {
		return nil, err
	}

	return &keymap, nil
}
