package parser

import (
	"fmt"

	"github.com/tacit7/eye-in-the-sky/internal/ccusage/db"
)

// SyncManager handles syncing JSONL files to the database
type SyncManager struct {
	database *db.CCUsageDB
	parser   *Parser
}

// NewSyncManager creates a new SyncManager
func NewSyncManager(database *db.CCUsageDB) *SyncManager {
	return &SyncManager{
		database: database,
		parser:   New(4),
	}
}

// Sync discovers JSONL files, checks for modifications, and syncs to database
func (sm *SyncManager) Sync() error {
	// Discover all files
	files, err := DiscoverFiles()
	if err != nil {
		return fmt.Errorf("failed to discover files: %w", err)
	}

	// Filter files that need parsing (changed or new)
	filesToParse := sm.filterModifiedFiles(files)
	if len(filesToParse) == 0 {
		// No new or modified files
		return nil
	}

	// Parse the modified files
	entries, err := sm.parser.ParseFiles(filesToParse)
	if err != nil {
		return fmt.Errorf("failed to parse files: %w", err)
	}

	// Insert entries into database
	if len(entries) > 0 {
		if err := sm.database.BatchInsertUsageEntries(entries); err != nil {
			return fmt.Errorf("failed to insert entries: %w", err)
		}
	}

	// Update file metadata for all files
	for _, file := range filesToParse {
		if err := sm.database.UpdateFileMetadata(file.Path, file.MTime); err != nil {
			// Log error but continue
			continue
		}
	}

	return nil
}

// filterModifiedFiles filters files that have been modified since last parse
func (sm *SyncManager) filterModifiedFiles(files []FileInfo) []FileInfo {
	var toProcess []FileInfo

	for _, file := range files {
		lastMTime, err := sm.database.GetFileMetadata(file.Path)
		if err != nil {
			// On error, include the file to be safe
			toProcess = append(toProcess, file)
			continue
		}

		// If file mtime has changed, or this is the first time seeing it
		if lastMTime == 0 || file.MTime > lastMTime {
			toProcess = append(toProcess, file)
		}
	}

	return toProcess
}

// FullSync performs a full re-parse of all files
func (sm *SyncManager) FullSync() error {
	// Discover all files
	files, err := DiscoverFiles()
	if err != nil {
		return fmt.Errorf("failed to discover files: %w", err)
	}

	// Parse all files
	entries, err := sm.parser.ParseFiles(files)
	if err != nil {
		return fmt.Errorf("failed to parse files: %w", err)
	}

	// Insert entries into database
	if len(entries) > 0 {
		if err := sm.database.BatchInsertUsageEntries(entries); err != nil {
			return fmt.Errorf("failed to insert entries: %w", err)
		}
	}

	// Update file metadata for all files
	for _, file := range files {
		if err := sm.database.UpdateFileMetadata(file.Path, file.MTime); err != nil {
			// Log error but continue
			continue
		}
	}

	return nil
}
