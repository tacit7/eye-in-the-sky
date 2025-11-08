package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	_ "modernc.org/sqlite"

	ccdb "github.com/tacit7/eye-in-the-sky/internal/ccusage/db"
	"github.com/tacit7/eye-in-the-sky/internal/ui/app"
)

// SyncWriter wraps a file and ensures all writes are flushed immediately
type SyncWriter struct {
	file *os.File
}

func (sw *SyncWriter) Write(p []byte) (n int, err error) {
	n, err = sw.file.Write(p)
	if err == nil {
		sw.file.Sync()
	}
	return
}

func main() {
	// Setup logging to file - ALL logs go to tui.log
	home, _ := os.UserHomeDir()
	logDir := filepath.Join(home, ".config", "eye-in-the-sky")
	os.MkdirAll(logDir, 0755)

	logFile := filepath.Join(logDir, "tui.log")
	// Truncate the file on startup to clear old logs
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not open log file: %v\n", err)
	} else {
		defer f.Sync()
		defer f.Close()
		// Use SyncWriter to ensure logs are flushed immediately
		syncWriter := &SyncWriter{file: f}
		log.SetOutput(syncWriter)
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	log.Println("=== TUI Dashboard Started ===")

	// Load configuration
	config, err := app.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Open database
	db, err := sql.Open("sqlite", "file:"+config.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize CCUsage database
	ccusageDBPath := filepath.Join(logDir, "ccusage.sqlite")

	var ccusageDB *ccdb.CCUsageDB
	ccusageDB, err = ccdb.New(ccusageDBPath)
	if err != nil {
		log.Printf("Warning: CCUsage database unavailable: %v", err)
		ccusageDB = nil
	} else {
		defer ccusageDB.Close()
		// Note: CCUsage sync now happens on-demand when viewing Usage tab
	}

	// Create model
	model, err := app.NewModel(db, ccusageDB)
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// Create Bubble Tea program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)

	// Run program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
