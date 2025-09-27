package database

import (
	"errors"
	"fmt"
)

// Custom error types for better error handling
var (
	ErrAgentNotFound      = errors.New("agent not found")
	ErrAgentAlreadyExists = errors.New("agent already exists")
	ErrInvalidAgentID     = errors.New("invalid agent ID")
	ErrInvalidStatus      = errors.New("invalid status")
	ErrInvalidActionType  = errors.New("invalid action type")
	ErrDatabaseConnection = errors.New("database connection failed")
	ErrMigrationFailed    = errors.New("migration failed")
)

// AgentError wraps agent-related errors with context
type AgentError struct {
	AgentID string
	Op      string
	Err     error
}

func (e *AgentError) Error() string {
	return fmt.Sprintf("agent %s: %s: %v", e.AgentID, e.Op, e.Err)
}

func (e *AgentError) Unwrap() error {
	return e.Err
}

// ValidationError represents input validation errors
type ValidationError struct {
	Field string
	Value interface{}
	Err   error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s' with value '%v': %v", e.Field, e.Value, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// IsNotFound checks if an error is a "not found" error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrAgentNotFound)
}

// IsAlreadyExists checks if an error is an "already exists" error
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAgentAlreadyExists)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

// NewAgentError creates a new agent error with context
func NewAgentError(agentID, op string, err error) *AgentError {
	return &AgentError{
		AgentID: agentID,
		Op:      op,
		Err:     err,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(field string, value interface{}, err error) *ValidationError {
	return &ValidationError{
		Field: field,
		Value: value,
		Err:   err,
	}
}
