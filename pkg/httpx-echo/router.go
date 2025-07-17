package httpxecho

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/phuocnguyen/user-api/internal/user"
)

func RegisterRoutes(e *echo.Echo, userHandler *user.Handler) {
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// API routes
	api := e.Group("/api/v1")

	// User routes
	users := api.Group("/users")
	users.POST("", userHandler.CreateUser)
	users.GET("", userHandler.GetUsers)
	users.GET("/:id", userHandler.GetUser)
	users.PUT("/:id", userHandler.UpdateUser)
	users.DELETE("/:id", userHandler.DeleteUser)
}