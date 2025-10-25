package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the sql.DB connection
type DB struct {
	*sql.DB
}

// New creates a new database connection and initializes the database
func New(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Enable foreign key constraints
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}

	return &DB{db}, nil
}

// Migrate runs all migrations in the migrations directory
func (db *DB) Migrate(migrationsDir string) error {
	// Read migration files
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("error reading migrations directory: %w", err)
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Create migrations table if it doesn't exist
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("error creating migrations table: %w", err)
	}

	// Apply each migration
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		// Check if migration has already been applied
		var exists bool
		err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM migrations WHERE name = ?)", file.Name()).Scan(&exists)
		if err != nil {
			return fmt.Errorf("error checking migration status: %w", err)
		}
		if exists {
			continue
		}

		// Read and execute migration
		migration, err := os.ReadFile(filepath.Join(migrationsDir, file.Name()))
		if err != nil {
			return fmt.Errorf("error reading migration file %s: %w", file.Name(), err)
		}

		if _, err := tx.Exec(string(migration)); err != nil {
			return fmt.Errorf("error executing migration %s: %w", file.Name(), err)
		}

		// Record migration
		_, err = tx.Exec("INSERT INTO migrations (name) VALUES (?)", file.Name())
		if err != nil {
			return fmt.Errorf("error recording migration %s: %w", file.Name(), err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing migrations: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
