package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	sqlc "github.com/michelemendel/times.place/db/sqlc"
	"github.com/michelemendel/times.place/internal/service"
	"github.com/michelemendel/times.place/internal/store"
	"github.com/michelemendel/times.place/utils"
)

// AdminHandler handles admin-only endpoints
type AdminHandler struct {
	store       *store.Store
	authService *service.AuthService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(store *store.Store, authService *service.AuthService) *AdminHandler {
	return &AdminHandler{
		store:       store,
		authService: authService,
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
			"venue_limit": o.VenueLimit,
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
		"venue_limit": owner.VenueLimit,
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
			"venue_uuid":           utils.UUIDToString(v.VenueUuid),
			"name":                 v.Name,
			"address":              v.Address,
			"owner_uuid":           utils.UUIDToString(v.OwnerUuid),
			"owner_name":           v.OwnerName,
			"owner_email":          v.OwnerEmail,
			"public_events_count":  v.PublicEventsCount,
			"private_events_count": v.PrivateEventsCount,
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

type UpdateOwnerVenueLimitRequest struct {
	VenueLimit int32 `json:"venue_limit" validate:"required,min=1"`
}

// UpdateOwnerVenueLimit handles PATCH /api/admin/owners/:uuid/venue-limit
func (h *AdminHandler) UpdateOwnerVenueLimit(c echo.Context) error {
	idStr := c.Param("uuid")
	uuid, err := utils.StringToUUID(idStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	var req UpdateOwnerVenueLimitRequest
	if err := c.Bind(&req); err != nil {
		return ValidationError(c, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return ValidationError(c, err.Error())
	}

	ctx := c.Request().Context()
	err = h.store.Queries.UpdateOwnerVenueLimit(ctx, sqlc.UpdateOwnerVenueLimitParams{
		OwnerUuid:  uuid,
		VenueLimit: req.VenueLimit,
	})
	if err != nil {
		return InternalError(c, "Failed to update venue limit")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Venue limit updated"})
}

type AdminResetPasswordRequest struct {
	Password string `json:"password" validate:"required,min=6"`
}

// AdminResetPassword handles POST /api/admin/owners/:uuid/reset-password
func (h *AdminHandler) AdminResetPassword(c echo.Context) error {
	idStr := c.Param("uuid")
	uuid, err := utils.StringToUUID(idStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	var req AdminResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return ValidationError(c, "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return ValidationError(c, err.Error())
	}

	// Hash password
	passwordHash, err := h.authService.HashPassword(req.Password)
	if err != nil {
		return InternalError(c, "Failed to process password")
	}

	ctx := c.Request().Context()
	err = h.store.Queries.UpdateOwnerPassword(ctx, sqlc.UpdateOwnerPasswordParams{
		PasswordHash: passwordHash,
		OwnerUuid:    uuid,
	})
	if err != nil {
		return InternalError(c, "Failed to update password")
	}

	// Revoke all refresh tokens for this user
	_ = h.store.Queries.RevokeAllTokensForOwner(ctx, uuid)

	return c.JSON(http.StatusOK, map[string]string{"message": "Password updated successfully"})
}

// Helper to reuse timestamptzToString from auth_handlers if possible, or duplicate/move it
// Since it's unexported in auth_handlers, I'll duplicate a version here that uses utils or copy logic.
// However, auth_handlers uses unexported timestamptzToString.
// I can add `TimestamptzToString` to utils/utils.go or just write it inline/helper here.
// I'll make it a helper here for now.

// Note: service.DateFormat might not be defined or correct context.
// Let's use time.RFC3339 directly or check utils.go which has GetTimestampAsString.
