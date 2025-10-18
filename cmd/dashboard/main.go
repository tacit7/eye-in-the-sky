package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tacit7/eye-in-the-sky/internal/dashboard"
	"github.com/tacit7/eye-in-the-sky/internal/database"
)

func main() {
	// Command line flags
	dbPath := flag.String("db", "./data/agents.db", "Path to the SQLite database")
	configPath := flag.String("config", "./cmd/dashboard/config/config.json", "Path to config file")
	keysPath := flag.String("keys", "./cmd/dashboard/config/keys.json", "Path to keys config file")
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
