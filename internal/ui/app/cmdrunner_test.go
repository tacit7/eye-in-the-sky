package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestResolveClaudePath(t *testing.T) {
	tests := []struct {
		name        string
		cfgPath     string
		setup       func() (string, func())
		expectError bool
		errorMsg    string
	}{
		{
			name:    "Empty config uses PATH lookup",
			cfgPath: "",
			setup: func() (string, func()) {
				// This test relies on claude being in PATH or not
				// We just verify it attempts lookup without error handling
				return "", func() {}
			},
			expectError: false, // May or may not find claude in PATH
		},
		{
			name: "Valid configured path",
			cfgPath: "", // Will be set in setup
			setup: func() (string, func()) {
				// Create a temporary executable file
				tmpDir := t.TempDir()
				claudePath := filepath.Join(tmpDir, "claude")
				f, err := os.Create(claudePath)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				f.Close()
				os.Chmod(claudePath, 0755)

				return claudePath, func() {}
			},
			expectError: false,
		},
		{
			name:    "Invalid configured path",
			cfgPath: "/nonexistent/path/to/claude",
			setup: func() (string, func()) {
				return "", func() {}
			},
			expectError: true,
			errorMsg:    "claude binary not found at configured path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgPath := tt.cfgPath
			if tt.setup != nil {
				path, cleanup := tt.setup()
				defer cleanup()
				if path != "" {
					cfgPath = path
				}
			}

			result, err := ResolveClaudePath(cfgPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else if err != nil && cfgPath != "" {
				// Only fail if we provided a valid path
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && cfgPath != "" && result != cfgPath {
				t.Errorf("Expected path '%s', got '%s'", cfgPath, result)
			}
		})
	}
}

func TestResolveClaudePathFallbackToSystem(t *testing.T) {
	// Test that empty config falls back to system PATH lookup
	_, err := ResolveClaudePath("")

	// Check if claude is in PATH
	_, lookupErr := exec.LookPath("claude")

	if lookupErr == nil {
		// Claude is in PATH, should succeed
		if err != nil {
			t.Errorf("Expected success when claude is in PATH, got error: %v", err)
		}
	} else {
		// Claude not in PATH, should return the lookup error
		if err == nil {
			t.Error("Expected error when claude not in PATH, got nil")
		}
	}
}

func TestResolveClaudePathConfigTakesPrecedence(t *testing.T) {
	// Create a temporary executable
	tmpDir := t.TempDir()
	customPath := filepath.Join(tmpDir, "custom-claude")
	f, err := os.Create(customPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	f.Close()
	os.Chmod(customPath, 0755)

	result, err := ResolveClaudePath(customPath)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != customPath {
		t.Errorf("Expected configured path '%s', got '%s'", customPath, result)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
			 s[len(s)-len(substr):] == substr ||
			 containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
