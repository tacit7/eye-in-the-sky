package main

import (
	"flag"
	"log"

	"github.com/urielmaldonado/eye-in-the-sky/internal/dashboard"
)

func main() {
	// Command line flags
	port := flag.String("port", "8080", "Port to run the dashboard server on")
	flag.Parse()

	// Create dashboard server
	server := dashboard.NewServer(*port)

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