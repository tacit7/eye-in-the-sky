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

// isRunningAsMCP detects if we're being called by Claude Desktop via stdin
func isRunningAsMCP() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func main() {
	var (
		dbPath = flag.String("db", "./data/agents.db", "SQLite database path")
		help   = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		fmt.Println("Eye in the Sky - Claude Code Multi-Agent Management System (MCP Server)")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Ensure data directory exists
	if err := os.MkdirAll(filepath.Dir(*dbPath), 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	fmt.Fprintf(os.Stderr, "🔍 Eye in the Sky MCP Server - Agent ID: 6d09ae9e\n")
	fmt.Fprintf(os.Stderr, "📂 Database: %s\n", *dbPath)

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

	fmt.Fprintln(os.Stderr, "✅ Database initialized successfully")

	// Initialize MCP server
	mcpServer := mcp.NewServer(db)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\n🛑 Shutdown signal received...")
		cancel()
	}()

	// Detect mode based on context
	runningAsMCP := isRunningAsMCP()

	if runningAsMCP {
		// MCP-only mode (called by Claude Desktop)
		fmt.Fprintln(os.Stderr, "🚀 Starting MCP Server (stdio mode)...")
		if err := mcpServer.Start(ctx); err != nil {
			log.Fatalf("MCP Server failed: %v", err)
		}
	} else {
		// Interactive mode - just MCP server
		fmt.Fprintln(os.Stderr, "🚀 Starting MCP Server...")
		if err := mcpServer.Start(ctx); err != nil {
			log.Fatalf("MCP Server failed: %v", err)
		}
	}

	fmt.Fprintln(os.Stderr, "✅ Server shutdown complete")
}
