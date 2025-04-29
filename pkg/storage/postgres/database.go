package postgres

import (
	"database/sql"
	"fmt"

	"wscollector/config"

	_ "github.com/lib/pq"
)

// CreateDatabase connects to the postgres server and creates a new database if it doesn't exist.
func CreateDatabase(cfg config.PostgresConfig) error {
	// Connect to the default 'postgres' DB
	dsn := cfg.DSN("dev")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}
	defer db.Close()

	// Check if database exists
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1);`
	if err := db.QueryRow(query, cfg.DBName).Scan(&exists); err != nil {
		return fmt.Errorf("check db exists failed: %w", err)
	}

	if exists {
		return nil // DB already exists
	}

	// Create the database
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DBName))
	if err != nil {
		return fmt.Errorf("create db failed: %w", err)
	}

	return nil
}
