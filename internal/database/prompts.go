package database

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// SubagentPrompt represents a reusable prompt template
type SubagentPrompt struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	PromptText  string `json:"prompt_text"`
	ProjectID   string `json:"project_id,omitempty"`
	Active      bool   `json:"active"`
	Version     int    `json:"version"`
	Tags        string `json:"tags,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	CreatedBy   string `json:"created_by,omitempty"`
}

var slugRegex = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// ValidateSlug checks if slug is valid kebab-case
func ValidateSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("slug cannot be empty")
	}
	if !slugRegex.MatchString(slug) {
		return fmt.Errorf("slug must be kebab-case (lowercase, alphanumeric, hyphens only): %s", slug)
	}
	return nil
}

// CreateSubagentPrompt inserts a new prompt
func (db *DB) CreateSubagentPrompt(prompt *SubagentPrompt) error {
	// Validate slug
	if err := ValidateSlug(prompt.Slug); err != nil {
		return fmt.Errorf("invalid slug: %w", err)
	}

	// Validate project_id exists if provided
	if prompt.ProjectID != "" {
		var exists bool
		err := db.conn.QueryRow("SELECT EXISTS(SELECT 1 FROM projects WHERE id = ?)", prompt.ProjectID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check project existence: %w", err)
		}
		if !exists {
			return fmt.Errorf("project_id %s does not exist", prompt.ProjectID)
		}
	}

	// Auto-generate UUID if not provided
	if prompt.ID == "" {
		prompt.ID = uuid.New().String()
	}

	// Set defaults
	if prompt.Version == 0 {
		prompt.Version = 1
	}

	query := `
		INSERT INTO subagent_prompts (id, name, slug, description, prompt_text, project_id, active, version, tags, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	projectID := sql.NullString{String: prompt.ProjectID, Valid: prompt.ProjectID != ""}
	_, err := db.conn.Exec(query,
		prompt.ID,
		prompt.Name,
		prompt.Slug,
		prompt.Description,
		prompt.PromptText,
		projectID,
		prompt.Active,
		prompt.Version,
		prompt.Tags,
		prompt.CreatedBy,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if prompt.ProjectID == "" {
				return fmt.Errorf("global prompt with slug '%s' already exists", prompt.Slug)
			}
			return fmt.Errorf("prompt with slug '%s' already exists for project %s", prompt.Slug, prompt.ProjectID)
		}
		return fmt.Errorf("failed to create subagent prompt: %w", err)
	}

	return nil
}

// GetSubagentPromptByID retrieves a prompt by ID
func (db *DB) GetSubagentPromptByID(id string) (*SubagentPrompt, error) {
	query := `
		SELECT id, name, slug, description, prompt_text, COALESCE(project_id, ''), active, version, tags, created_at, updated_at, COALESCE(created_by, '')
		FROM subagent_prompts
		WHERE id = ?
	`
	var prompt SubagentPrompt
	err := db.conn.QueryRow(query, id).Scan(
		&prompt.ID,
		&prompt.Name,
		&prompt.Slug,
		&prompt.Description,
		&prompt.PromptText,
		&prompt.ProjectID,
		&prompt.Active,
		&prompt.Version,
		&prompt.Tags,
		&prompt.CreatedAt,
		&prompt.UpdatedAt,
		&prompt.CreatedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("prompt not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get subagent prompt: %w", err)
	}
	return &prompt, nil
}

// GetSubagentPromptBySlug retrieves a prompt by slug with project-aware fallback
// If projectID is provided, checks project-scoped first, then falls back to global
func (db *DB) GetSubagentPromptBySlug(slug string, projectID string) (*SubagentPrompt, error) {
	if err := ValidateSlug(slug); err != nil {
		return nil, fmt.Errorf("invalid slug: %w", err)
	}

	var prompt SubagentPrompt

	// If project_id provided, try project-scoped first
	if projectID != "" {
		query := `
			SELECT id, name, slug, description, prompt_text, COALESCE(project_id, ''), active, version, tags, created_at, updated_at, COALESCE(created_by, '')
			FROM subagent_prompts
			WHERE slug = ? AND project_id = ? AND active = 1
			LIMIT 1
		`
		err := db.conn.QueryRow(query, slug, projectID).Scan(
			&prompt.ID,
			&prompt.Name,
			&prompt.Slug,
			&prompt.Description,
			&prompt.PromptText,
			&prompt.ProjectID,
			&prompt.Active,
			&prompt.Version,
			&prompt.Tags,
			&prompt.CreatedAt,
			&prompt.UpdatedAt,
			&prompt.CreatedBy,
		)
		if err == nil {
			return &prompt, nil // Found project-scoped prompt
		}
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to query project-scoped prompt: %w", err)
		}
	}

	// Fallback to global prompt
	query := `
		SELECT id, name, slug, description, prompt_text, COALESCE(project_id, ''), active, version, tags, created_at, updated_at, COALESCE(created_by, '')
		FROM subagent_prompts
		WHERE slug = ? AND project_id IS NULL AND active = 1
		LIMIT 1
	`
	err := db.conn.QueryRow(query, slug).Scan(
		&prompt.ID,
		&prompt.Name,
		&prompt.Slug,
		&prompt.Description,
		&prompt.PromptText,
		&prompt.ProjectID,
		&prompt.Active,
		&prompt.Version,
		&prompt.Tags,
		&prompt.CreatedAt,
		&prompt.UpdatedAt,
		&prompt.CreatedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("prompt not found: %s", slug)
		}
		return nil, fmt.Errorf("failed to get subagent prompt: %w", err)
	}

	return &prompt, nil
}

// ListSubagentPromptsOptions contains filtering options
type ListSubagentPromptsOptions struct {
	ProjectID   string
	Active      *bool
	Tags        []string
	Resolve     bool // Deduplicate by slug (project overrides global)
	IncludeText bool // Include prompt_text in results
	Limit       int
	Offset      int
}

// ListSubagentPrompts retrieves prompts with filtering
func (db *DB) ListSubagentPrompts(opts ListSubagentPromptsOptions) ([]SubagentPrompt, error) {
	query := "SELECT id, name, slug, description, %s, COALESCE(project_id, ''), active, version, tags, created_at, updated_at, COALESCE(created_by, '') FROM subagent_prompts WHERE 1=1"
	args := []interface{}{}

	// Include or exclude prompt_text
	promptTextField := "''"
	if opts.IncludeText {
		promptTextField = "prompt_text"
	}
	query = fmt.Sprintf(query, promptTextField)

	// Filter by project_id
	if opts.ProjectID != "" {
		if opts.Resolve {
			// Return project-scoped + global (with deduplication handled by application)
			query += " AND (project_id = ? OR project_id IS NULL)"
			args = append(args, opts.ProjectID)
		} else {
			query += " AND project_id = ?"
			args = append(args, opts.ProjectID)
		}
	}

	// Filter by active status
	if opts.Active != nil {
		query += " AND active = ?"
		args = append(args, *opts.Active)
	}

	// Filter by tags (simple contains check)
	if len(opts.Tags) > 0 {
		tagConditions := []string{}
		for _, tag := range opts.Tags {
			tagConditions = append(tagConditions, "tags LIKE ?")
			args = append(args, "%"+tag+"%")
		}
		query += " AND (" + strings.Join(tagConditions, " OR ") + ")"
	}

	// Order by project_id DESC (project-scoped first), then slug
	query += " ORDER BY project_id DESC, slug ASC"

	// Pagination
	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
		if opts.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, opts.Offset)
		}
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list subagent prompts: %w", err)
	}
	defer rows.Close()

	prompts := []SubagentPrompt{}
	seenSlugs := make(map[string]bool) // For resolve deduplication

	for rows.Next() {
		var prompt SubagentPrompt
		err := rows.Scan(
			&prompt.ID,
			&prompt.Name,
			&prompt.Slug,
			&prompt.Description,
			&prompt.PromptText,
			&prompt.ProjectID,
			&prompt.Active,
			&prompt.Version,
			&prompt.Tags,
			&prompt.CreatedAt,
			&prompt.UpdatedAt,
			&prompt.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan prompt: %w", err)
		}

		// Resolve: Skip global if we've already seen this slug (from project-scoped)
		if opts.Resolve {
			if seenSlugs[prompt.Slug] {
				continue
			}
			seenSlugs[prompt.Slug] = true
		}

		prompts = append(prompts, prompt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating prompts: %w", err)
	}

	return prompts, nil
}

// UpdateSubagentPrompt updates a prompt with optimistic locking
func (db *DB) UpdateSubagentPrompt(id string, updates map[string]interface{}, expectedVersion int) error {
	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Validate slug if being updated
	if slug, ok := updates["slug"].(string); ok {
		if err := ValidateSlug(slug); err != nil {
			return fmt.Errorf("invalid slug: %w", err)
		}
	}

	// Validate project_id if being updated
	if projectID, ok := updates["project_id"].(string); ok && projectID != "" {
		var exists bool
		err := db.conn.QueryRow("SELECT EXISTS(SELECT 1 FROM projects WHERE id = ?)", projectID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check project existence: %w", err)
		}
		if !exists {
			return fmt.Errorf("project_id %s does not exist", projectID)
		}
	}

	// Check current version (optimistic locking)
	var currentVersion int
	err := db.conn.QueryRow("SELECT version FROM subagent_prompts WHERE id = ?", id).Scan(&currentVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("prompt not found: %s", id)
		}
		return fmt.Errorf("failed to check version: %w", err)
	}

	if currentVersion != expectedVersion {
		return fmt.Errorf("version mismatch: expected %d, got %d (prompt was modified by another process)", expectedVersion, currentVersion)
	}

	// Build UPDATE query
	setClauses := []string{}
	args := []interface{}{}

	for key, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", key))
		args = append(args, value)
	}

	// Auto-update updated_at
	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")

	query := fmt.Sprintf("UPDATE subagent_prompts SET %s WHERE id = ? AND version = ?", strings.Join(setClauses, ", "))
	args = append(args, id, expectedVersion)

	result, err := db.conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update subagent prompt: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("prompt not found or version mismatch")
	}

	return nil
}

// DeactivateSubagentPrompt soft deletes a prompt
func (db *DB) DeactivateSubagentPrompt(id string) error {
	query := "UPDATE subagent_prompts SET active = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	result, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate prompt: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("prompt not found: %s", id)
	}

	return nil
}

// DeleteSubagentPrompt permanently deletes a prompt (use with caution)
func (db *DB) DeleteSubagentPrompt(id string) error {
	query := "DELETE FROM subagent_prompts WHERE id = ?"
	result, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete prompt: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("prompt not found: %s", id)
	}

	return nil
}
