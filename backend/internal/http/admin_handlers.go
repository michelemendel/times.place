package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/michelemendel/times.place/internal/store"
	"github.com/michelemendel/times.place/utils"
)

// AdminHandler handles admin-only endpoints
type AdminHandler struct {
	store *store.Store
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(store *store.Store) *AdminHandler {
	return &AdminHandler{
		store: store,
	}
}

// ListOwners handles GET /api/admin/owners
func (h *AdminHandler) ListOwners(c echo.Context) error {
	ctx := c.Request().Context()

	owners, err := h.store.Queries.ListDetailsAllOwners(ctx)
	if err != nil {
		return InternalError(c, "Failed to list owners")
	}

	// Transform to response
	response := make([]map[string]any, len(owners))
	for i, o := range owners {
		response[i] = map[string]any{
			"owner_uuid":  utils.UUIDToString(o.OwnerUuid),
			"name":        o.Name,
			"email":       o.Email,
			"is_admin":    o.IsAdmin,
			"is_demo":     o.IsDemo,
			"created_at":  timestamptzToString(o.CreatedAt),
			"venue_count": o.VenueCount,
		}
	}

	return c.JSON(http.StatusOK, response)
}

// GetOwner handles GET /api/admin/owners/:uuid
func (h *AdminHandler) GetOwner(c echo.Context) error {
	idStr := c.Param("uuid")
	uuid, err := utils.StringToUUID(idStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	ctx := c.Request().Context()
	owner, err := h.store.Queries.GetOwnerDetails(ctx, uuid)
	if err != nil {
		return NotFoundError(c, "Owner not found")
	}

	response := map[string]any{
		"owner_uuid":  utils.UUIDToString(owner.OwnerUuid),
		"name":        owner.Name,
		"email":       owner.Email,
		"mobile":      owner.Mobile,
		"is_admin":    owner.IsAdmin,
		"is_demo":     owner.IsDemo,
		"created_at":  timestamptzToString(owner.CreatedAt),
		"modified_at": timestamptzToString(owner.ModifiedAt),
	}

	return c.JSON(http.StatusOK, response)
}

// ListVenues handles GET /api/admin/venues
func (h *AdminHandler) ListVenues(c echo.Context) error {
	ctx := c.Request().Context()

	venues, err := h.store.Queries.ListAllVenues(ctx)
	if err != nil {
		return InternalError(c, "Failed to list venues")
	}

	// Transform to response
	response := make([]map[string]any, len(venues))
	for i, v := range venues {
		response[i] = map[string]any{
			"venue_uuid":  utils.UUIDToString(v.VenueUuid),
			"name":        v.Name,
			"address":     v.Address,
			"owner_uuid":  utils.UUIDToString(v.OwnerUuid),
			"owner_name":  v.OwnerName,
			"owner_email": v.OwnerEmail,
			// No visibility field
		}
	}

	return c.JSON(http.StatusOK, response)
}

// DeleteOwner handles DELETE /api/admin/owners/:uuid
func (h *AdminHandler) DeleteOwner(c echo.Context) error {
	idStr := c.Param("uuid")
	uuid, err := utils.StringToUUID(idStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	ctx := c.Request().Context()

	// Check if target is a demo account or self (optional, but good practice)
	// We can check is_demo first
	owner, err := h.store.Queries.GetOwnerDetails(ctx, uuid)
	if err == nil {
		if owner.IsDemo {
			return ForbiddenError(c, "Cannot delete demo accounts")
		}
		// Prevent self-deletion if needed (optional)
		// requesterUUID, _ := GetOwnerUUIDFromContext(c)
		// if requesterUUID == idStr { return ForbiddenError(c, "Cannot delete yourself") }
	}

	err = h.store.Queries.AdminDeleteOwner(ctx, uuid)
	if err != nil {
		return InternalError(c, "Failed to delete owner")
	}

	// Also revoke refresh tokens? Handled by DB CASCADE DELETE
	// Also delete venues? Handled by DB CASCADE DELETE

	return c.NoContent(http.StatusNoContent)
}

// Helper to reuse timestamptzToString from auth_handlers if possible, or duplicate/move it
// Since it's unexported in auth_handlers, I'll duplicate a version here that uses utils or copy logic.
// However, auth_handlers uses unexported timestamptzToString.
// I can add `TimestamptzToString` to utils/utils.go or just write it inline/helper here.
// I'll make it a helper here for now.

// Note: service.DateFormat might not be defined or correct context.
// Let's use time.RFC3339 directly or check utils.go which has GetTimestampAsString.
