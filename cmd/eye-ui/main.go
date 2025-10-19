package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/mattn/go-sqlite3"

	"github.com/tacit7/eye-in-the-sky/internal/ui/app"
)

func main() {
	// Load configuration
	config, err := app.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Open database
	db, err := sql.Open("sqlite3", config.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create model
	model, err := app.NewModel(db)
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// Create Bubble Tea program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
