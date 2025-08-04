package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/phuocnguyen/user-api/pkg/db"
)

func main() {
	var (
		up    = flag.Bool("up", false, "Run all pending migrations")
		down  = flag.Bool("down", false, "Rollback the last migration")
		force = flag.Bool("force", false, "Force run migrations (ignores migration history)")
		help  = flag.Bool("help", false, "Show this help message")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Initialize database connection
	if err := db.InitPostgres(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	switch {
	case *up:
		runMigrationsUp(*force)
	case *down:
		runMigrationsDown()
	default:
		fmt.Println("No action specified. Use -help for usage information.")
		showHelp()
	}
}

func runMigrationsUp(force bool) {
	fmt.Println("Running migrations...")
	
	if force {
		fmt.Println("⚠️  FORCE MODE: This will run all migrations regardless of history!")
		fmt.Println("Are you sure? This could cause data loss. (y/N): ")
		
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Migration cancelled.")
			return
		}
		
		// For force mode, you might want to implement a ForceRunMigrations function
		// For now, we'll use the regular function
	}
	
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	
	fmt.Println("✅ Migrations completed successfully!")
}

func runMigrationsDown() {
	fmt.Println("⚠️  Rollback functionality not implemented yet.")
	fmt.Println("This would rollback the last migration.")
	fmt.Println("For now, you can manually rollback by:")
	fmt.Println("1. Connecting to your database")
	fmt.Println("2. Running the reverse SQL commands")
	fmt.Println("3. Removing the entry from the migrations table")
}

func showHelp() {
	fmt.Println("Database Migration Tool")
	fmt.Println("Usage: go run cmd/migrate/main.go [OPTIONS]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -up     Run all pending migrations")
	fmt.Println("  -down   Rollback the last migration (not implemented yet)")
	fmt.Println("  -force  Force run all migrations (dangerous!)")
	fmt.Println("  -help   Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/migrate/main.go -up")
	fmt.Println("  go run cmd/migrate/main.go -down")
	fmt.Println("  go run cmd/migrate/main.go -help")
	fmt.Println("")
	fmt.Println("Environment Variables:")
	fmt.Println("  DATABASE_URL - PostgreSQL connection string")
}