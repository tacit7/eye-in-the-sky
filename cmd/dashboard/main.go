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
	help := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *help {
		fmt.Println("Eye in the Sky - TUI Dashboard")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Get database path from standard config location
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}
	dbPath := filepath.Join(homeDir, ".config", "eye-in-the-sky", "agents.db")

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}

	fmt.Fprintf(os.Stderr, "📂 Database: %s\n", dbPath)

	// Initialize database
	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Health(); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}

	// Create and run dashboard (config paths from app package)
	app, err := dashboard.NewApp(db, "", "")
	if err != nil {
		log.Fatalf("Failed to create dashboard: %v", err)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("Dashboard error: %v", err)
	}
}
