package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func SeedDatabase() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// Get seed files
	seedDir := "seeds"
	files, err := os.ReadDir(seedDir)
	if err != nil {
		return fmt.Errorf("failed to read seeds directory: %w", err)
	}

	// Filter and sort SQL files
	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sort.Strings(sqlFiles)

	// Execute seed files
	for _, filename := range sqlFiles {
		fmt.Printf("Executing seed file %s\n", filename)
		
		// Read seed file
		content, err := os.ReadFile(filepath.Join(seedDir, filename))
		if err != nil {
			return fmt.Errorf("failed to read seed file %s: %w", filename, err)
		}

		// Execute seed data
		if _, err := DB.Exec(context.Background(), string(content)); err != nil {
			return fmt.Errorf("failed to execute seed file %s: %w", filename, err)
		}

		fmt.Printf("Seed file %s executed successfully\n", filename)
	}

	return nil
}