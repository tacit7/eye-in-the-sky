package utils

import (
	"crypto/sha1"
	"fmt"
	"time"
)

// GenerateGitStyleAgentID generates a git-style 8-character hash for agent IDs
// Uses SHA1 hash of current timestamp and random component, truncated to 8 chars
func GenerateGitStyleAgentID() string {
	// Create input string with timestamp and nanoseconds for uniqueness
	input := fmt.Sprintf("agent-%d-%d", time.Now().Unix(), time.Now().Nanosecond())

	// Generate SHA1 hash
	hash := sha1.Sum([]byte(input))

	// Convert to hex string and take first 8 characters (git-style)
	return fmt.Sprintf("%x", hash)[:8]
}

// ValidateAgentID checks if an agent ID is a valid 8-character hex string
func ValidateAgentID(id string) bool {
	if len(id) != 8 {
		return false
	}

	// Check if all characters are valid hex
	for _, char := range id {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}

	return true
}