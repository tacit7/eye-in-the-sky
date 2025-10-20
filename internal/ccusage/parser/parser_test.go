package parser

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tacit7/eye-in-the-sky/internal/ccusage/models"
)

func TestCreateHash(t *testing.T) {
	input := "msg123req456"
	hash := createHash(input)

	// Verify it's a valid SHA256 hash (64 hex characters)
	if len(hash) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}

	// Verify it's deterministic
	hash2 := createHash(input)
	if hash != hash2 {
		t.Errorf("Hash not deterministic: %s != %s", hash, hash2)
	}

	// Verify different inputs produce different hashes
	hash3 := createHash("different")
	if hash == hash3 {
		t.Errorf("Different inputs produced same hash")
	}
}

func TestParserDeduplication(t *testing.T) {
	p := New(2)

	// First entry should not be duplicate
	hash1 := createHash("msg1req1")
	if p.isDuplicate(hash1) {
		t.Error("First entry marked as duplicate")
	}

	// Same hash should be duplicate
	if !p.isDuplicate(hash1) {
		t.Error("Second entry with same hash not marked as duplicate")
	}

	// Different hash should not be duplicate
	hash2 := createHash("msg2req2")
	if p.isDuplicate(hash2) {
		t.Error("Entry with different hash marked as duplicate")
	}
}

func TestConvertToDBRow(t *testing.T) {
	p := New(2)

	entry := models.UsageEntry{
		Cwd:       "/home/user/project",
		SessionID: "session123",
		Timestamp: time.Now(),
		Version:   "1.0",
		CostUSD:   0.05,
		RequestID: "req456",
		Message: models.Message{
			Model: "claude-3-sonnet-20240229",
			ID:    "msg123",
			Usage: models.UsageMetrics{
				InputTokens:              100,
				OutputTokens:             200,
				CacheCreationInputTokens: 10,
				CacheReadInputTokens:     5,
			},
		},
		IsApiErrorMessage: false,
	}

	row := p.convertToDBRow(entry, "testproject")

	if row == nil {
		t.Error("Expected non-nil row")
		return
	}

	if row.SessionID != "session123" {
		t.Errorf("Expected SessionID 'session123', got '%s'", row.SessionID)
	}

	if row.Model != "claude-3-sonnet-20240229" {
		t.Errorf("Expected Model 'claude-3-sonnet-20240229', got '%s'", row.Model)
	}

	if row.InputTokens != 100 {
		t.Errorf("Expected InputTokens 100, got %d", row.InputTokens)
	}

	if row.OutputTokens != 200 {
		t.Errorf("Expected OutputTokens 200, got %d", row.OutputTokens)
	}

	if row.CacheCreationTokens != 10 {
		t.Errorf("Expected CacheCreationTokens 10, got %d", row.CacheCreationTokens)
	}

	if row.Project != "testproject" {
		t.Errorf("Expected Project 'testproject', got '%s'", row.Project)
	}
}

func TestConvertToDBRowMissingFields(t *testing.T) {
	p := New(2)

	// Entry with missing SessionID
	entry := models.UsageEntry{
		Cwd:       "/home/user/project",
		SessionID: "", // Empty
		Timestamp: time.Now(),
		RequestID: "req456",
		Message: models.Message{
			ID: "msg123",
		},
	}

	row := p.convertToDBRow(entry, "project")
	if row != nil {
		t.Error("Expected nil row for entry with missing SessionID")
	}

	// Entry with missing MessageID
	entry.SessionID = "session123"
	entry.Message.ID = "" // Empty
	row = p.convertToDBRow(entry, "project")
	if row != nil {
		t.Error("Expected nil row for entry with missing MessageID")
	}

	// Entry with missing RequestID
	entry.Message.ID = "msg123"
	entry.RequestID = "" // Empty
	row = p.convertToDBRow(entry, "project")
	if row != nil {
		t.Error("Expected nil row for entry with missing RequestID")
	}
}

func TestDiscoveryGetsClaudePaths(t *testing.T) {
	// Save original env var
	origEnv := os.Getenv("CLAUDE_CONFIG_DIR")
	defer func() {
		if origEnv != "" {
			os.Setenv("CLAUDE_CONFIG_DIR", origEnv)
		} else {
			os.Unsetenv("CLAUDE_CONFIG_DIR")
		}
	}()

	// Test with CLAUDE_CONFIG_DIR env var
	testPath := "/custom/path1,/custom/path2"
	os.Setenv("CLAUDE_CONFIG_DIR", testPath)

	paths := getClaudePaths()
	if len(paths) != 2 {
		t.Errorf("Expected 2 paths with env var, got %d", len(paths))
	}
	if paths[0] != "/custom/path1" || paths[1] != "/custom/path2" {
		t.Errorf("Paths mismatch: %v", paths)
	}

	// Test with default paths
	os.Unsetenv("CLAUDE_CONFIG_DIR")
	paths = getClaudePaths()
	if len(paths) != 2 {
		t.Errorf("Expected 2 default paths, got %d", len(paths))
	}

	home, _ := os.UserHomeDir()
	expectedPath := filepath.Join(home, ".config", "claude", "projects")
	if paths[0] != expectedPath {
		t.Errorf("First path mismatch: expected %s, got %s", expectedPath, paths[0])
	}
}

func TestHashConsistency(t *testing.T) {
	// Ensure hash matches TypeScript implementation style
	input := "messageId123requestId456"
	hash := createHash(input)

	// Manually verify SHA256
	manualHash := fmt.Sprintf("%x", sha256.Sum256([]byte(input)))

	if hash != manualHash {
		t.Errorf("Hash mismatch: %s != %s", hash, manualHash)
	}
}
