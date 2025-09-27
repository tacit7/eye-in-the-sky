package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/tacit7/eye-in-the-sky/internal/dashboard"
	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/mcp"
)

func main() {
	// Command line flags
	port := flag.String("port", "8080", "Port to run the dashboard server on")
	dbPath := flag.String("db", "./data/agents.db", "Path to the SQLite database")
	help := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *help {
		fmt.Println("Eye in the Sky - Claude Code Multi-Agent Management System")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Ensure data directory exists
	if err := os.MkdirAll(filepath.Dir(*dbPath), 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	fmt.Printf("🔍 Eye in the Sky - Integrated Dashboard & MCP Server\n")
	fmt.Printf("📂 Database: %s\n", *dbPath)
	fmt.Printf("🌐 Dashboard: http://localhost:%s\n", *port)

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

	fmt.Println("✅ Database initialized successfully")

	// Create servers
	dashboardServer := dashboard.NewServer(*port, db)
	mcpServer := mcp.NewServer(db)

	// Load templates
	if err := dashboardServer.LoadTemplates("web/templates"); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutdown signal received...")
		cancel()
	}()

	// Start both servers concurrently
	var wg sync.WaitGroup
	wg.Add(2)

	// Start Dashboard Server
	go func() {
		defer wg.Done()
		fmt.Println("🚀 Starting Dashboard Server...")
		if err := dashboardServer.Start(); err != nil {
			log.Printf("Dashboard Server failed: %v", err)
			cancel()
		}
	}()

	// Start MCP Server
	go func() {
		defer wg.Done()
		fmt.Println("🚀 Starting MCP Server...")
		if err := mcpServer.Start(ctx); err != nil && ctx.Err() == nil {
			log.Printf("MCP Server failed: %v", err)
			cancel()
		}
	}()

	fmt.Println("✅ Both servers started successfully")
	fmt.Printf("📊 Access dashboard at: http://localhost:%s\n", *port)

	// Wait for shutdown signal or server failure
	<-ctx.Done()

	fmt.Println("✅ Server shutdown complete")
}
