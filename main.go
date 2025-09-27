package main

import (
	"flag"
	"log"

	"github.com/urielmaldonado/eye-in-the-sky/internal/dashboard"
	"github.com/urielmaldonado/eye-in-the-sky/internal/database"
)

func main() {
	// Command line flags
	port := flag.String("port", "8080", "Port to run the dashboard server on")
	dbPath := flag.String("db", "../data/agents.db", "Path to the SQLite database")
	flag.Parse()

	// Initialize database
	db, err := database.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Printf("Connected to database at %s", *dbPath)

	// Create dashboard server with database
	server := dashboard.NewServer(*port, db)

	// Load templates
	if err := server.LoadTemplates("web/templates"); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	// Start the server
	log.Printf("Starting Eye in the Sky Dashboard...")
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}