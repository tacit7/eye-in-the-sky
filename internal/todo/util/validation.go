package util

import (
	"fmt"
	"strings"
	"time"
)

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (ve ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", ve.Field, ve.Message)
}

// ValidateDescription validates a task description.
func ValidateDescription(desc string) error {
	if strings.TrimSpace(desc) == "" {
		return ValidationError{
			Field:   "description",
			Message: "cannot be empty",
		}
	}
	return nil
}

// ValidateTagName validates a tag name.
func ValidateTagName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return ValidationError{
			Field:   "tag_name",
			Message: "cannot be empty",
		}
	}

	if len(name) > 64 {
		return ValidationError{
			Field:   "tag_name",
			Message: "must be 64 characters or less",
		}
	}

	return nil
}

// ValidatePriority validates a priority value.
func ValidatePriority(priority *int) error {
	if priority == nil {
		return nil // Priority is optional
	}

	if *priority < 1 || *priority > 5 {
		return ValidationError{
			Field:   "priority",
			Message: "must be between 1 and 5",
		}
	}

	return nil
}

// ValidateDueDate validates a due date.
func ValidateDueDate(dueDate *time.Time) error {
	if dueDate == nil {
		return nil // Due date is optional
	}

	// Optionally, you could add a check for past dates:
	// if dueDate.Before(time.Now()) {
	//     return ValidationError{
	//         Field:   "due_date",
	//         Message: "cannot be in the past",
	//     }
	// }

	return nil
}

// ValidateParentID validates a parent task ID.
func ValidateParentID(childID int, parentID *int) error {
	if parentID == nil {
		return nil // Parent is optional
	}

	if *parentID == childID {
		return ValidationError{
			Field:   "parent_id",
			Message: "cannot be the same as the task itself",
		}
	}

	return nil
}

// ValidateProjectName validates a project name.
func ValidateProjectName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ValidationError{
			Field:   "name",
			Message: "cannot be empty",
		}
	}

	return nil
}

// ValidateWorkflowStateCode validates a workflow state code.
func ValidateWorkflowStateCode(code string) error {
	code = strings.TrimSpace(code)

	if code == "" {
		return ValidationError{
			Field:   "code",
			Message: "cannot be empty",
		}
	}

	if len(code) > 50 {
		return ValidationError{
			Field:   "code",
			Message: "must be 50 characters or less",
		}
	}

	return nil
}

// ValidateWorkflowStateDisplayName validates a workflow state display name.
func ValidateWorkflowStateDisplayName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return ValidationError{
			Field:   "display_name",
			Message: "cannot be empty",
		}
	}

	if len(name) > 100 {
		return ValidationError{
			Field:   "display_name",
			Message: "must be 100 characters or less",
		}
	}

	return nil
}

// ValidateNoteContent validates note markdown content.
func ValidateNoteContent(content string) error {
	if strings.TrimSpace(content) == "" {
		return ValidationError{
			Field:   "body_markdown",
			Message: "cannot be empty",
		}
	}

	return nil
}

// ValidateWeight validates a task weight.
func ValidateWeight(weight *int) error {
	if weight == nil {
		return nil // Weight is optional
	}

	if *weight < 0 {
		return ValidationError{
			Field:   "weight",
			Message: "must be non-negative",
		}
	}

	if *weight > 1000 {
		return ValidationError{
			Field:   "weight",
			Message: "must be 1000 or less",
		}
	}

	return nil
}

// ValidatePaginationParams validates limit and offset for pagination.
func ValidatePaginationParams(limit, offset int) error {
	if limit < 1 || limit > 1000 {
		return ValidationError{
			Field:   "limit",
			Message: "must be between 1 and 1000",
		}
	}

	if offset < 0 {
		return ValidationError{
			Field:   "offset",
			Message: "must be non-negative",
		}
	}

	return nil
}

// ValidateSearchQuery validates a full-text search query.
func ValidateSearchQuery(query string) error {
	query = strings.TrimSpace(query)

	if query == "" {
		return ValidationError{
			Field:   "query",
			Message: "cannot be empty",
		}
	}

	if len(query) > 500 {
		return ValidationError{
			Field:   "query",
			Message: "must be 500 characters or less",
		}
	}

	return nil
}
