package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/michelemendel/times.place/internal/service"
	"github.com/michelemendel/times.place/internal/store"
)

// Server wraps the Echo server and dependencies
type Server struct {
	echo      *echo.Echo
	store     *store.Store
	authService *service.AuthService
}

// NewServer creates and configures a new server
func NewServer() (*Server, error) {
	// Load environment variables from .env file
	// This will not error if .env doesn't exist (useful for production)
	_ = godotenv.Load("backend/.env")

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true

	// Validator for request validation
	e.Validator = NewCustomValidator()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	// CORS middleware (for development)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // In production, restrict this
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	// JSON middleware
	e.Use(middleware.BodyLimit("1M"))

	// Connect to database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	store, err := store.NewStore(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize auth service
	authService, err := service.NewAuthService()
	if err != nil {
		store.Close()
		return nil, fmt.Errorf("failed to initialize auth service: %w", err)
	}

	// Register routes
	RegisterRoutes(e, store, authService)

	return &Server{
		echo:        e,
		store:       store,
		authService: authService,
	}, nil
}

// Start starts the server
func (s *Server) Start() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	address := fmt.Sprintf(":%s", port)
	log.Printf("Starting server on %s", address)

	// Start server in a goroutine
	go func() {
		if err := s.echo.Start(address); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.echo.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	// Close database connection
	s.store.Close()

	log.Println("Server stopped")
	return nil
}

// Echo returns the underlying Echo instance (for testing)
func (s *Server) Echo() *echo.Echo {
	return s.echo
}
