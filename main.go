package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/mcp"
)

func main() {
	// Command line flags
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

	fmt.Printf("🔍 Eye in the Sky - MCP Server\n")
	fmt.Printf("📂 Database: %s\n", *dbPath)

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

	// Create MCP server
	mcpServer := mcp.NewServer(db)

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

	// Start MCP Server
	fmt.Println("🚀 Starting MCP Server...")
	if err := mcpServer.Start(ctx); err != nil && ctx.Err() == nil {
		log.Printf("MCP Server failed: %v", err)
	}

	fmt.Println("✅ Server shutdown complete")
}
