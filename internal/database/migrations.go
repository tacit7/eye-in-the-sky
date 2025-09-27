package database

import (
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migration represents a database migration
type Migration struct {
	Version string
	SQL     string
}

// RunMigrations executes all pending database migrations
func (db *DB) RunMigrations() error {
	// Ensure migration tracking table exists
	if err := db.createMigrationTable(); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	// Get all migration files
	migrations, err := db.loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get applied migrations
	applied, err := db.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, migration := range migrations {
		if !applied[migration.Version] {
			if err := db.runMigration(migration); err != nil {
				return fmt.Errorf("failed to run migration %s: %w", migration.Version, err)
			}
			fmt.Printf("✅ Applied migration: %s\n", migration.Version)
		}
	}

	return nil
}

// createMigrationTable creates the schema_migrations table if it doesn't exist
func (db *DB) createMigrationTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.conn.Exec(query)
	return err
}

// loadMigrations loads all migration files from the embedded filesystem
func (db *DB) loadMigrations() ([]Migration, error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			content, err := migrationsFS.ReadFile(filepath.Join("migrations", entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
			}

			// Extract version from filename (e.g., "0001_init.sql" -> "0001_init")
			version := strings.TrimSuffix(entry.Name(), ".sql")

			migrations = append(migrations, Migration{
				Version: version,
				SQL:     string(content),
			})
		}
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// getAppliedMigrations returns a map of applied migration versions
func (db *DB) getAppliedMigrations() (map[string]bool, error) {
	query := `SELECT version FROM schema_migrations`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, nil
}

// runMigration executes a single migration within a transaction
func (db *DB) runMigration(migration Migration) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration as applied (if not already in the SQL)
	if !strings.Contains(migration.SQL, "INSERT") || !strings.Contains(migration.SQL, "schema_migrations") {
		recordQuery := `INSERT INTO schema_migrations (version) VALUES (?)`
		if _, err := tx.Exec(recordQuery, migration.Version); err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}
	}

	return tx.Commit()
}

// GetMigrationStatus returns the current migration status
func (db *DB) GetMigrationStatus() ([]MigrationStatus, error) {
	migrations, err := db.loadMigrations()
	if err != nil {
		return nil, err
	}

	applied, err := db.getAppliedMigrations()
	if err != nil {
		return nil, err
	}

	var status []MigrationStatus
	for _, migration := range migrations {
		status = append(status, MigrationStatus{
			Version: migration.Version,
			Applied: applied[migration.Version],
		})
	}

	return status, nil
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Version string `json:"version"`
	Applied bool   `json:"applied"`
}
