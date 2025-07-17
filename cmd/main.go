package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/phuocnguyen/user-api/internal/user"
	"github.com/phuocnguyen/user-api/pkg/db"
	httpxecho "github.com/phuocnguyen/user-api/pkg/httpx-echo"
)

func main() {
	// Initialize database
	if err := db.InitPostgres(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
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
