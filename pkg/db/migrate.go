package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func RunMigrations() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// Create migrations table if it doesn't exist
	createMigrationsTable := `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL UNIQUE,
			executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`
	
	if _, err := DB.Exec(context.Background(), createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get migration files
	migrationDir := "migrations"
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Filter and sort SQL files
	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sort.Strings(sqlFiles)

	// Check which migrations have been executed
	executedMigrations := make(map[string]bool)
	rows, err := DB.Query(context.Background(), "SELECT filename FROM migrations")
	if err != nil {
		return fmt.Errorf("failed to query executed migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return fmt.Errorf("failed to scan migration filename: %w", err)
		}
		executedMigrations[filename] = true
	}

	// Execute pending migrations
	for _, filename := range sqlFiles {
		if executedMigrations[filename] {
			fmt.Printf("Migration %s already executed, skipping\n", filename)
			continue
		}

		fmt.Printf("Executing migration %s\n", filename)
		
		// Read migration file
		content, err := os.ReadFile(filepath.Join(migrationDir, filename))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Execute migration
		if _, err := DB.Exec(context.Background(), string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		// Record migration as executed
		if _, err := DB.Exec(context.Background(), 
			"INSERT INTO migrations (filename) VALUES ($1)", filename); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		fmt.Printf("Migration %s executed successfully\n", filename)
	}

	return nil
}