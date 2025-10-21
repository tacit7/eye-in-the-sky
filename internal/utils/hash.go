package utils

import (
	"github.com/google/uuid"
)

// GenerateGitStyleAgentID generates a UUID for agent IDs
// Returns a standard UUID string (36 characters)
func GenerateGitStyleAgentID() string {
	return uuid.New().String()
}

// ValidateAgentID checks if an agent ID is a valid UUID string
func ValidateAgentID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}