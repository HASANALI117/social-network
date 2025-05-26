package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

const (
	dbPath     = "/app/data/social_network.db"
	migrations = "file://pkg/db/migrations/sqlite"
)

// DB represents a connection to the database
type DB struct {
	*sql.DB
}

var GlobalDB *sql.DB

// NewDB creates a new database connection
func InitDB() (*sql.DB, error) {
	// Ensure directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create DB directory: %w", err)
	}

	// Connect to SQLite database
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	GlobalDB = db

	log.Println("Database initialized successfully")
	return GlobalDB, nil
}

// runMigrations runs database migrations
func runMigrations(db *sql.DB) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrations, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	// Add this to check current version
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Printf("Migration version error: %v", err)
	} else {
		log.Printf("Current migration version: %d, Dirty: %v", version, dirty)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("No new migrations to apply")
		} else {
			log.Printf("Migration error: %v", err)
			return fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}
