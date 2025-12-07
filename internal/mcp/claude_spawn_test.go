package mcp

import (
	"os"
	"testing"
)

func TestFindClaudeBinary(t *testing.T) {
	path, err := FindClaudeBinary()
	if err != nil {
		t.Fatalf("FindClaudeBinary failed: %v", err)
	}

	if path == "" {
		t.Fatalf("FindClaudeBinary returned empty path")
	}

	// Verify the file exists
	_, err = os.Stat(path)
	if err != nil {
		t.Fatalf("Claude binary path does not exist: %s", path)
	}

	t.Logf("Found Claude binary at: %s", path)
}

func TestCreateSystemCommand(t *testing.T) {
	claudePath, err := FindClaudeBinary()
	if err != nil {
		t.Skipf("Claude binary not found: %v", err)
	}

	args := []string{"-p", "test prompt", "--model", "haiku"}
	projectPath := "/tmp"

	cmd := CreateSystemCommand(claudePath, args, projectPath)

	if cmd == nil {
		t.Fatalf("CreateSystemCommand returned nil")
	}

	if cmd.Path != claudePath {
		t.Fatalf("Expected path %s, got %s", claudePath, cmd.Path)
	}

	if cmd.Dir != projectPath {
		t.Fatalf("Expected dir %s, got %s", projectPath, cmd.Dir)
	}

	if len(cmd.Args) < 2 {
		t.Fatalf("Expected args to be populated")
	}

	t.Logf("Command created successfully with %d args", len(cmd.Args))
}
