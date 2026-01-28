package http

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/michelemendel/times.place/internal/store"
)

// HealthcheckHandler handles health check endpoints
type HealthcheckHandler struct {
	store *store.Store
}

// NewHealthcheckHandler creates a new healthcheck handler
func NewHealthcheckHandler(store *store.Store) *HealthcheckHandler {
	return &HealthcheckHandler{
		store: store,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Database  string `json:"database,omitempty"`
}

// Health handles GET /health
// Returns a simple health check response
func (h *HealthcheckHandler) Health(c echo.Context) error {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Optionally check database connectivity
	if h.store != nil {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()

		err := h.store.DB().Ping(ctx)
		if err != nil {
			response.Status = "degraded"
			response.Database = "unavailable"
			return c.JSON(http.StatusServiceUnavailable, response)
		}
		response.Database = "connected"
	}

	return c.JSON(http.StatusOK, response)
}
