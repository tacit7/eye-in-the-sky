package domain

import "time"

// TaskID is a unique identifier for a task (UUID from TaskWarrior)
type TaskID string

// Task represents a TaskWarrior task
type Task struct {
	ID              TaskID
	UUID            string // TaskWarrior UUID
	Description     string
	Status          string // pending, completed, deleted
	Priority        string // H, M, L
	Project         string
	Tags            []string
	Due             time.Time
	Entry           time.Time
	Annotations     []TaskAnnotation
	WorkflowStatus  string // ready, working, testing, review, blocked, etc.
}

// TaskAnnotation represents a task annotation with timestamp
type TaskAnnotation struct {
	Entry       time.Time
	Description string
}