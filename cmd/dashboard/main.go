package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tacit7/eye-in-the-sky/internal/dashboard"
	"github.com/tacit7/eye-in-the-sky/internal/database"
)

func main() {
	// Get executable directory for config paths
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	execDir := filepath.Dir(execPath)
	projectRoot := filepath.Dir(execDir) // bin is one level down from project root

	// Command line flags with paths relative to project root
	defaultDBPath := filepath.Join(projectRoot, "data", "agents.db")
	defaultConfigPath := filepath.Join(projectRoot, "cmd", "dashboard", "config", "config.json")
	defaultKeysPath := filepath.Join(projectRoot, "cmd", "dashboard", "config", "keys.json")

	dbPath := flag.String("db", defaultDBPath, "Path to the SQLite database")
	configPath := flag.String("config", defaultConfigPath, "Path to config file")
	keysPath := flag.String("keys", defaultKeysPath, "Path to keys config file")
	help := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *help {
		fmt.Println("Eye in the Sky - TUI Dashboard")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Initialize database
	db, err := database.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Health(); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}

	// Create and run dashboard
	app, err := dashboard.NewApp(db, *configPath, *keysPath)
	if err != nil {
		log.Fatalf("Failed to create dashboard: %v", err)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("Dashboard error: %v", err)
	}
}
