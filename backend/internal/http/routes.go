package http

import (
	"github.com/labstack/echo/v4"
	"github.com/michelemendel/times.place/internal/mailer"
	"github.com/michelemendel/times.place/internal/service"
	"github.com/michelemendel/times.place/internal/store"
)

// RegisterRoutes registers all API routes and frontend static files (if built assets exist).
func RegisterRoutes(e *echo.Echo, store *store.Store, authService *service.AuthService, mailerSender mailer.Sender) {
	// Create handlers
	authHandler := NewAuthHandler(store, authService, mailerSender)
	healthcheckHandler := NewHealthcheckHandler(store)

	// Health check endpoint (public, no auth)
	e.GET("/health", healthcheckHandler.Health)

	// API routes group
	api := e.Group("/api")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)
	auth.POST("/logout", authHandler.Logout)
	auth.GET("/verify-email", authHandler.VerifyEmail)

	// Protected auth routes
	auth.GET("/me", authHandler.Me, JWTAuthMiddleware(authService))
	auth.DELETE("/me", authHandler.DeleteMe, JWTAuthMiddleware(authService))
	auth.POST("/resend-verification", authHandler.ResendVerification, JWTAuthMiddleware(authService))

	// Owner-scoped routes (protected)
	// Register more specific paths before :venue_uuid so /venues/:id/event-lists is matched correctly
	venues := api.Group("/venues", JWTAuthMiddleware(authService))
	venueHandler := NewVenueHandler(store)
	eventListHandler := NewEventListHandler(store)
	venues.GET("", venueHandler.List)
	venues.POST("", venueHandler.Create)
	venues.GET("/:venue_uuid/event-lists", eventListHandler.ListByVenue)
	venues.POST("/:venue_uuid/event-lists", eventListHandler.Create)
	venues.GET("/:venue_uuid", venueHandler.Get)
	venues.PATCH("/:venue_uuid", venueHandler.Update)
	venues.DELETE("/:venue_uuid", venueHandler.Delete)

	// Owner-scoped: list event lists by venue (unique path to avoid any router ambiguity)
	owner := api.Group("/owner", JWTAuthMiddleware(authService))
	owner.GET("/venues/:venue_uuid/event-lists", eventListHandler.ListByVenue)

	// Event lists: direct access by UUID
	eventLists := api.Group("/event-lists", JWTAuthMiddleware(authService))
	eventLists.GET("/:event_list_uuid", eventListHandler.Get)
	eventLists.PATCH("/:event_list_uuid", eventListHandler.Update)
	eventLists.DELETE("/:event_list_uuid", eventListHandler.Delete)

	// Events (nested under event lists + direct access)
	eventHandler := NewEventHandler(store)
	eventLists.GET("/:event_list_uuid/events", eventHandler.ListByEventList)
	eventLists.POST("/:event_list_uuid/events", eventHandler.Create)

	events := api.Group("/events", JWTAuthMiddleware(authService))
	events.GET("/:event_uuid", eventHandler.Get)
	events.PATCH("/:event_uuid", eventHandler.Update)
	events.DELETE("/:event_uuid", eventHandler.Delete)

	// Public routes (no auth)
	public := api.Group("/public")
	publicHandler := NewPublicHandler(store)
	public.GET("/venues", publicHandler.ListVenues)
	public.GET("/venues/:venue_uuid/event-lists", publicHandler.GetEventListsByVenue)
	public.GET("/venues/by-token/:token", publicHandler.GetVenueByToken)
	public.GET("/event-lists/by-token/:token", publicHandler.GetEventListByToken)
	public.GET("/event-lists/:event_list_uuid/events", publicHandler.GetEventsByEventList)

	// Admin routes (protected by JWT + Admin check)
	admin := api.Group("/admin", JWTAuthMiddleware(authService), AdminOnlyMiddleware(store))
	adminHandler := NewAdminHandler(store)
	admin.GET("/owners", adminHandler.ListOwners)
	admin.GET("/owners/:uuid", adminHandler.GetOwner)
	admin.GET("/venues", adminHandler.ListVenues)
	admin.DELETE("/owners/:uuid", adminHandler.DeleteOwner)

	// Serve frontend static files (if available).
	// This must be registered AFTER all API routes
	if err := setupFrontendRoutes(e); err != nil {
		// Log error but don't fail - frontend serving is optional.
		// The server will still work for API routes.
		e.Logger.Warnf("Failed to setup frontend routes: %v (set FRONTEND_BUILD_DIR to frontend/build path if needed)", err)
	}
}
