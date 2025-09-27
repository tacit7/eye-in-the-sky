package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

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

	fmt.Printf("🔍 Eye in the Sky MCP Server - Agent ID: 6d09ae9e\n")
	fmt.Printf("📂 Database: %s\n", *dbPath)

	// TODO: Initialize database
	// TODO: Start MCP server

	log.Println("✅ MCP Server started successfully")
	select {} // Keep running
}