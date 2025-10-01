package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/tacit7/eye-in-the-sky/internal/database"
	"github.com/tacit7/eye-in-the-sky/internal/dashboard"
	"github.com/tacit7/eye-in-the-sky/internal/mcp"
)

//go:embed templates/*.html
var templateFS embed.FS

// isRunningAsMCP detects if we're being called by Claude Desktop via stdin
func isRunningAsMCP() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func main() {
	var (
		dbPath        = flag.String("db", "./data/agents.db", "SQLite database path")
		help          = flag.Bool("help", false, "Show help")
		dashboardOnly = flag.Bool("dashboard", false, "Run dashboard only (no MCP server)")
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

	// Initialize dashboard server
	dashboardServer := dashboard.NewServer("8080", db)

	// Load templates
	if err := dashboardServer.LoadTemplates(templateFS, "templates/*.html"); err != nil {
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
		fmt.Fprintln(os.Stderr, "\n🛑 Shutdown signal received...")
		cancel()
	}()

	// Detect mode based on flags and context
	runningAsMCP := isRunningAsMCP()

	if *dashboardOnly {
		// Dashboard-only mode
		fmt.Fprintln(os.Stderr, "🚀 Starting Dashboard Server on :8080...")
		fmt.Fprintln(os.Stderr, "📊 Dashboard available at: http://localhost:8080")
		fmt.Fprintln(os.Stderr, "💡 Use Ctrl+C to stop the server")

		// Start dashboard and wait for shutdown signal
		go func() {
			if err := dashboardServer.Start(); err != nil {
				log.Fatalf("Dashboard server failed: %v", err)
			}
		}()

		// Wait for shutdown signal
		<-ctx.Done()
	} else if runningAsMCP {
		// MCP-only mode (called by Claude Desktop)
		fmt.Fprintln(os.Stderr, "🚀 Starting MCP Server (stdio mode)...")
		if err := mcpServer.Start(ctx); err != nil {
			log.Fatalf("MCP Server failed: %v", err)
		}
	} else {
		// Interactive mode - just MCP server (no dashboard)
		// Dashboard should be run separately with --dashboard flag
		fmt.Fprintln(os.Stderr, "🚀 Starting MCP Server...")
		fmt.Fprintln(os.Stderr, "💡 Use --dashboard flag to run dashboard server")
		if err := mcpServer.Start(ctx); err != nil {
			log.Fatalf("MCP Server failed: %v", err)
		}
	}

	fmt.Fprintln(os.Stderr, "✅ Server shutdown complete")
}
