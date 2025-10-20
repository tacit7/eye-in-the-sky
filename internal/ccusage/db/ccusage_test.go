package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Verify database file was created
	if _, err := os.Stat(dbPath); err != nil {
		t.Errorf("Database file not created: %v", err)
	}
}

func TestInsertUsageEntry(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	entry := UsageEntryRow{
		SessionID:           "session123",
		Timestamp:           time.Now().Format(time.RFC3339),
		Project:             "myproject",
		Model:               "claude-3-sonnet-20240229",
		InputTokens:         100,
		OutputTokens:        200,
		CacheCreationTokens: 10,
		CacheReadTokens:     5,
		TotalCost:           0.05,
		MessageID:           "msg123",
		RequestID:           "req456",
		UniqueHash:          "hash123",
	}

	err = db.InsertUsageEntry(entry)
	if err != nil {
		t.Fatalf("Failed to insert entry: %v", err)
	}
}

func TestBatchInsertUsageEntries(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	entries := []UsageEntryRow{
		{
			SessionID:           "session123",
			Timestamp:           time.Now().Format(time.RFC3339),
			Project:             "project1",
			Model:               "claude-3-sonnet-20240229",
			InputTokens:         100,
			OutputTokens:        200,
			CacheCreationTokens: 10,
			CacheReadTokens:     5,
			TotalCost:           0.05,
			MessageID:           "msg1",
			RequestID:           "req1",
			UniqueHash:          "hash1",
		},
		{
			SessionID:           "session124",
			Timestamp:           time.Now().Format(time.RFC3339),
			Project:             "project2",
			Model:               "claude-3-opus-20240229",
			InputTokens:         150,
			OutputTokens:        250,
			CacheCreationTokens: 15,
			CacheReadTokens:     8,
			TotalCost:           0.08,
			MessageID:           "msg2",
			RequestID:           "req2",
			UniqueHash:          "hash2",
		},
	}

	err = db.BatchInsertUsageEntries(entries)
	if err != nil {
		t.Fatalf("Failed to batch insert entries: %v", err)
	}
}

func TestUpdateFileMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	filePath := "/path/to/file.jsonl"
	mtime := int64(1234567890)

	err = db.UpdateFileMetadata(filePath, mtime)
	if err != nil {
		t.Fatalf("Failed to update file metadata: %v", err)
	}

	// Verify metadata was stored
	retrievedMtime, err := db.GetFileMetadata(filePath)
	if err != nil {
		t.Fatalf("Failed to get file metadata: %v", err)
	}

	if retrievedMtime != mtime {
		t.Errorf("Expected mtime %d, got %d", mtime, retrievedMtime)
	}
}

func TestGetFileMetadataNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Query for non-existent file
	mtime, err := db.GetFileMetadata("/nonexistent/file.jsonl")
	if err != nil {
		t.Fatalf("Failed to get file metadata: %v", err)
	}

	if mtime != 0 {
		t.Errorf("Expected mtime 0 for non-existent file, got %d", mtime)
	}
}

func TestDuplicateInsertRejected(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	entry := UsageEntryRow{
		SessionID:           "session123",
		Timestamp:           time.Now().Format(time.RFC3339),
		Project:             "project1",
		Model:               "claude-3-sonnet-20240229",
		InputTokens:         100,
		OutputTokens:        200,
		CacheCreationTokens: 10,
		CacheReadTokens:     5,
		TotalCost:           0.05,
		MessageID:           "msg1",
		RequestID:           "req1",
		UniqueHash:          "unique-hash-123",
	}

	// First insert should succeed
	err = db.InsertUsageEntry(entry)
	if err != nil {
		t.Fatalf("First insert failed: %v", err)
	}

	// Duplicate insert should fail due to UNIQUE constraint
	err = db.InsertUsageEntry(entry)
	if err == nil {
		t.Error("Expected error for duplicate unique_hash, but got none")
	}
}
