package http

import (
	"github.com/labstack/echo/v4"
	"github.com/michelemendel/times.place/internal/service"
	"github.com/michelemendel/times.place/internal/store"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(e *echo.Echo, store *store.Store, authService *service.AuthService) {
	// Create handlers
	authHandler := NewAuthHandler(store, authService)

	// API routes group
	api := e.Group("/api")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)
	auth.POST("/logout", authHandler.Logout)

	// Protected auth routes
	auth.GET("/me", authHandler.Me, JWTAuthMiddleware(authService))

	// TODO: Add other route groups here:
	// - venues (protected)
	// - event-lists (protected)
	// - events (protected)
	// - public (public)
}
