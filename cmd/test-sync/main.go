package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tacit7/eye-in-the-sky/internal/ccusage/db"
	"github.com/tacit7/eye-in-the-sky/internal/ccusage/parser"
	_ "modernc.org/sqlite"
)

func main() {
	// Get database path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	dbPath := filepath.Join(homeDir, ".config", "eye-in-the-sky", "eits.db")
	fmt.Printf("Opening database: %s\n", dbPath)

	// Open database
	database, err := sql.Open("sqlite", "file:"+dbPath+"?cache=shared&mode=rwc&_journal_mode=WAL")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Create CCUsage DB instance
	ccdb, err := db.NewWithConnection(database)
	if err != nil {
		log.Fatalf("Failed to create CCUsage DB: %v", err)
	}

	// Check current entry count
	beforeCount, err := ccdb.GetEntryCount()
	if err != nil {
		log.Fatalf("Failed to get entry count: %v", err)
	}
	fmt.Printf("Entries before sync: %d\n", beforeCount)

	// Create sync manager
	fmt.Println("\nCreating sync manager...")
	syncMgr := parser.NewSyncManager(ccdb)

	// Run sync
	fmt.Println("\n=== Starting Sync ===")
	if err := syncMgr.Sync(); err != nil {
		log.Fatalf("Sync failed: %v", err)
	}

	// Check entry count after sync
	afterCount, err := ccdb.GetEntryCount()
	if err != nil {
		log.Fatalf("Failed to get entry count after sync: %v", err)
	}

	fmt.Printf("\n=== Sync Complete ===\n")
	fmt.Printf("Entries before: %d\n", beforeCount)
	fmt.Printf("Entries after:  %d\n", afterCount)
	fmt.Printf("New entries:    %d\n", afterCount-beforeCount)
}
