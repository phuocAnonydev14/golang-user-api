package main

import (
	"log"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/phuocnguyen/user-api/internal/user"
	"github.com/phuocnguyen/user-api/pkg/db"
	"github.com/phuocnguyen/user-api/pkg/env"
	httpxecho "github.com/phuocnguyen/user-api/pkg/httpx-echo"
)

func main() {
	// Load .env file
	if err := env.LoadEnv(".env"); err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
	}

	// Initialize database
	if err := db.InitPostgres(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations (only if AUTO_MIGRATE is enabled)
	autoMigrate := strings.ToLower(os.Getenv("AUTO_MIGRATE"))
	if autoMigrate == "true" || autoMigrate == "1" || autoMigrate == "yes" {
		log.Println("Auto-migration enabled, running migrations...")
		if err := db.RunMigrations(); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
	} else {
		log.Println("Auto-migration disabled. Use 'go run cmd/migrate/main.go -up' to run migrations manually.")
	}

	// Seed database (optional, can be skipped if data already exists)
	if err := db.SeedDatabase(); err != nil {
		log.Printf("Warning: Failed to seed database: %v", err)
	}

	// Initialize repository and handler
	userRepo := user.NewRepository(db.DB)
	userHandler := user.NewHandler(userRepo)

	e := echo.New()
	httpxecho.RegisterRoutes(e, userHandler)

	if err := e.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
