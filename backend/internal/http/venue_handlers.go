package http

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	sqlc "github.com/michelemendel/times.place/db/sqlc"
	"github.com/michelemendel/times.place/internal/store"
)

// VenueHandler handles venue endpoints
type VenueHandler struct {
	store *store.Store
}

// NewVenueHandler creates a new venue handler
func NewVenueHandler(store *store.Store) *VenueHandler {
	return &VenueHandler{
		store: store,
	}
}

// Request/Response types

type CreateVenueRequest struct {
	Name             string `json:"name" validate:"required"`
	BannerImage      string `json:"banner_image"`
	Address          string `json:"address"`
	Geolocation      string `json:"geolocation"`
	Comment          string `json:"comment"`
	Timezone         string `json:"timezone"`
	PrivateLinkToken string `json:"private_link_token"` // Optional UUID string
}

type UpdateVenueRequest struct {
	Name             *string `json:"name"`
	BannerImage      *string `json:"banner_image"`
	Address          *string `json:"address"`
	Geolocation      *string `json:"geolocation"`
	Comment          *string `json:"comment"`
	Timezone         *string `json:"timezone"`
	PrivateLinkToken *string `json:"private_link_token"` // Optional UUID string
}

type VenueResponse struct {
	VenueUuid        string `json:"venue_uuid"`
	OwnerUuid        string `json:"owner_uuid"`
	Name             string `json:"name"`
	BannerImage      string `json:"banner_image"`
	Address          string `json:"address"`
	Geolocation      string `json:"geolocation"`
	Comment          string `json:"comment"`
	Timezone         string `json:"timezone"`
	PrivateLinkToken string `json:"private_link_token"`
	OwnerName        string `json:"owner_name,omitempty"`  // Set by public endpoints; empty for owner endpoints
	OwnerEmail       string `json:"owner_email,omitempty"` // Set by public endpoints; empty for owner endpoints
	CreatedAt        string `json:"created_at"`
	ModifiedAt       string `json:"modified_at"`
}

// OwnerVenueWithEventListsResponse is returned by List so the frontend gets event lists in one call.
type OwnerVenueWithEventListsResponse struct {
	VenueResponse
	EventLists []EventListResponse `json:"event_lists"`
}

// Helper functions

// textToString converts pgtype.Text to string
func textToString(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

// venueToResponse converts sqlc.Venue to VenueResponse
func venueToResponse(venue sqlc.Venue) VenueResponse {
	return VenueResponse{
		VenueUuid:        uuidToString(venue.VenueUuid),
		OwnerUuid:        uuidToString(venue.OwnerUuid),
		Name:             venue.Name,
		BannerImage:      venue.BannerImage,
		Address:          venue.Address,
		Geolocation:      venue.Geolocation,
		Comment:          textToString(venue.Comment),
		Timezone:         venue.Timezone,
		PrivateLinkToken: uuidToString(venue.PrivateLinkToken),
		CreatedAt:        timestamptzToString(venue.CreatedAt),
		ModifiedAt:       timestamptzToString(venue.ModifiedAt),
	}
}

// parsePrivateLinkToken parses optional private link token string to pgtype.UUID
// If empty, generates a new UUID
func parsePrivateLinkToken(tokenStr string) (pgtype.UUID, error) {
	if tokenStr == "" {
		// Generate new UUID
		newUUID := uuid.New()
		return stringToUUID(newUUID.String())
	}
	return stringToUUID(tokenStr)
}

// Handlers

// List handles GET /api/venues
func (h *VenueHandler) List(c echo.Context) error {
	// Get owner UUID from context (set by JWT middleware)
	ownerUUIDStr, err := GetOwnerUUIDFromContext(c)
	if err != nil {
		return UnauthorizedError(c, "Unauthorized")
	}

	ctx := c.Request().Context()

	// Convert string UUID to pgtype.UUID
	ownerUUID, err := stringToUUID(ownerUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	// List venues
	venues, err := h.store.Queries.ListVenuesByOwner(ctx, ownerUUID)
	if err != nil {
		return InternalError(c, "Failed to list venues")
	}

	// Build response with event lists for each venue (one round-trip, no separate per-venue calls)
	response := make([]OwnerVenueWithEventListsResponse, len(venues))
	for i, venue := range venues {
		response[i] = OwnerVenueWithEventListsResponse{
			VenueResponse: venueToResponse(venue),
			EventLists:    []EventListResponse{},
		}
		eventLists, err := h.store.Queries.ListEventListsByVenueAndOwner(ctx, sqlc.ListEventListsByVenueAndOwnerParams{
			VenueUuid: venue.VenueUuid,
			OwnerUuid: ownerUUID,
		})
		if err != nil {
			return InternalError(c, "Failed to list event lists for venue")
		}
		for _, el := range eventLists {
			response[i].EventLists = append(response[i].EventLists, EventListToResponse(el))
		}
	}

	return c.JSON(http.StatusOK, response)
}

// Create handles POST /api/venues
func (h *VenueHandler) Create(c echo.Context) error {
	var req CreateVenueRequest
	if err := c.Bind(&req); err != nil {
		return ValidationError(c, "Invalid request body")
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return ValidationError(c, err.Error())
	}

	// Get owner UUID from context
	ownerUUIDStr, err := GetOwnerUUIDFromContext(c)
	if err != nil {
		return UnauthorizedError(c, "Unauthorized")
	}

	ctx := c.Request().Context()

	verified, err := IsEmailVerified(ctx, h.store.Queries, ownerUUIDStr)
	if err != nil {
		return InternalError(c, "Failed to check owner")
	}
	if !verified {
		return EmailNotVerifiedError(c, "Please verify your email address to make changes.")
	}

	// Convert owner UUID
	ownerUUID, err := stringToUUID(ownerUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	// Fetch owner to get their specific venue limit
	owner, err := h.store.Queries.GetOwnerByID(ctx, ownerUUID)
	if err != nil {
		return InternalError(c, "Failed to fetch owner details")
	}

	// Enforce per-user venue limit
	limit := int64(owner.VenueLimit)
	count, err := h.store.Queries.CountVenuesByOwner(ctx, ownerUUID)
	if err != nil {
		return InternalError(c, "Failed to check venue count")
	}
	if count >= limit {
		return ForbiddenError(c, "You have reached your limit of "+strconv.FormatInt(limit, 10)+" venues. Upgrade to add more.")
	}

	// Parse private link token (generate if not provided)
	privateLinkToken, err := parsePrivateLinkToken(req.PrivateLinkToken)
	if err != nil {
		return ValidationError(c, "Invalid private_link_token format")
	}

	// Prepare comment (convert string to pgtype.Text)
	var comment pgtype.Text
	if req.Comment != "" {
		comment = pgtype.Text{
			String: req.Comment,
			Valid:  true,
		}
	}

	// Create venue
	venue, err := h.store.Queries.CreateVenue(ctx, sqlc.CreateVenueParams{
		OwnerUuid:        ownerUUID,
		Name:             req.Name,
		BannerImage:      req.BannerImage,
		Address:          req.Address,
		Geolocation:      req.Geolocation,
		Comment:          comment,
		Timezone:         req.Timezone,
		PrivateLinkToken: privateLinkToken,
	})
	if err != nil {
		return InternalError(c, "Failed to create venue")
	}

	return c.JSON(http.StatusCreated, venueToResponse(venue))
}

// Get handles GET /api/venues/:venue_uuid
func (h *VenueHandler) Get(c echo.Context) error {
	venueUUIDStr := c.Param("venue_uuid")
	if venueUUIDStr == "" {
		return ValidationError(c, "venue_uuid is required")
	}

	// Get owner UUID from context
	ownerUUIDStr, err := GetOwnerUUIDFromContext(c)
	if err != nil {
		return UnauthorizedError(c, "Unauthorized")
	}

	ctx := c.Request().Context()

	// Convert UUIDs
	venueUUID, err := stringToUUID(venueUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid venue_uuid format")
	}

	ownerUUID, err := stringToUUID(ownerUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	// Get venue
	venue, err := h.store.Queries.GetVenueByIDAndOwner(ctx, sqlc.GetVenueByIDAndOwnerParams{
		VenueUuid: venueUUID,
		OwnerUuid: ownerUUID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return NotFoundError(c, "Venue not found")
		}
		return InternalError(c, "Failed to get venue")
	}

	return c.JSON(http.StatusOK, venueToResponse(venue))
}

// Update handles PATCH /api/venues/:venue_uuid
func (h *VenueHandler) Update(c echo.Context) error {
	venueUUIDStr := c.Param("venue_uuid")
	if venueUUIDStr == "" {
		return ValidationError(c, "venue_uuid is required")
	}

	var req UpdateVenueRequest
	if err := c.Bind(&req); err != nil {
		return ValidationError(c, "Invalid request body")
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return ValidationError(c, err.Error())
	}

	// Get owner UUID from context
	ownerUUIDStr, err := GetOwnerUUIDFromContext(c)
	if err != nil {
		return UnauthorizedError(c, "Unauthorized")
	}

	ctx := c.Request().Context()

	demo, err := IsDemoOwner(ctx, h.store.Queries, ownerUUIDStr)
	if err != nil {
		return InternalError(c, "Failed to check owner")
	}
	if demo {
		return ForbiddenError(c, "Demo data cannot be modified")
	}
	verified, err := IsEmailVerified(ctx, h.store.Queries, ownerUUIDStr)
	if err != nil {
		return InternalError(c, "Failed to check owner")
	}
	if !verified {
		return EmailNotVerifiedError(c, "Please verify your email address to make changes.")
	}

	// Convert UUIDs
	venueUUID, err := stringToUUID(venueUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid venue_uuid format")
	}

	ownerUUID, err := stringToUUID(ownerUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	// Get existing venue to use current values for COALESCE
	existingVenue, err := h.store.Queries.GetVenueByIDAndOwner(ctx, sqlc.GetVenueByIDAndOwnerParams{
		VenueUuid: venueUUID,
		OwnerUuid: ownerUUID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return NotFoundError(c, "Venue not found")
		}
		return InternalError(c, "Failed to get venue")
	}

	// Prepare update params (use existing values for COALESCE)
	name := existingVenue.Name
	if req.Name != nil {
		name = *req.Name
	}

	bannerImage := existingVenue.BannerImage
	if req.BannerImage != nil {
		bannerImage = *req.BannerImage
	}

	address := existingVenue.Address
	if req.Address != nil {
		address = *req.Address
	}

	geolocation := existingVenue.Geolocation
	if req.Geolocation != nil {
		geolocation = *req.Geolocation
	}

	var comment pgtype.Text
	if req.Comment != nil {
		if *req.Comment != "" {
			comment = pgtype.Text{
				String: *req.Comment,
				Valid:  true,
			}
		} else {
			comment = pgtype.Text{Valid: false}
		}
	} else {
		comment = existingVenue.Comment
	}

	timezone := existingVenue.Timezone
	if req.Timezone != nil {
		timezone = *req.Timezone
	}

	privateLinkToken := existingVenue.PrivateLinkToken
	if req.PrivateLinkToken != nil {
		if *req.PrivateLinkToken != "" {
			parsed, err := stringToUUID(*req.PrivateLinkToken)
			if err != nil {
				return ValidationError(c, "Invalid private_link_token format")
			}
			privateLinkToken = parsed
		} else {
			// Empty string means clear the token
			privateLinkToken = pgtype.UUID{Valid: false}
		}
	}

	// Update venue
	venue, err := h.store.Queries.UpdateVenue(ctx, sqlc.UpdateVenueParams{
		VenueUuid:        venueUUID,
		OwnerUuid:        ownerUUID,
		Name:             name,
		BannerImage:      bannerImage,
		Address:          address,
		Geolocation:      geolocation,
		Comment:          comment,
		Timezone:         timezone,
		PrivateLinkToken: privateLinkToken,
	})
	if err != nil {
		return InternalError(c, "Failed to update venue")
	}

	return c.JSON(http.StatusOK, venueToResponse(venue))
}

// Delete handles DELETE /api/venues/:venue_uuid
func (h *VenueHandler) Delete(c echo.Context) error {
	venueUUIDStr := c.Param("venue_uuid")
	if venueUUIDStr == "" {
		return ValidationError(c, "venue_uuid is required")
	}

	// Get owner UUID from context
	ownerUUIDStr, err := GetOwnerUUIDFromContext(c)
	if err != nil {
		return UnauthorizedError(c, "Unauthorized")
	}

	ctx := c.Request().Context()

	demo, err := IsDemoOwner(ctx, h.store.Queries, ownerUUIDStr)
	if err != nil {
		return InternalError(c, "Failed to check owner")
	}
	if demo {
		return ForbiddenError(c, "Demo data cannot be modified")
	}
	verified, err := IsEmailVerified(ctx, h.store.Queries, ownerUUIDStr)
	if err != nil {
		return InternalError(c, "Failed to check owner")
	}
	if !verified {
		return EmailNotVerifiedError(c, "Please verify your email address to make changes.")
	}

	// Convert UUIDs
	venueUUID, err := stringToUUID(venueUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid venue_uuid format")
	}

	ownerUUID, err := stringToUUID(ownerUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	// Ensure the venue exists and is owned by the caller.
	// (sqlc :exec delete queries don't report affected rows, so we pre-check
	// to guarantee cross-owner deletes return 404.)
	_, err = h.store.Queries.GetVenueByIDAndOwner(ctx, sqlc.GetVenueByIDAndOwnerParams{
		VenueUuid: venueUUID,
		OwnerUuid: ownerUUID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return NotFoundError(c, "Venue not found")
		}
		return InternalError(c, "Failed to get venue")
	}

	// Delete venue (cascade handled by DB via ON DELETE CASCADE)
	err = h.store.Queries.DeleteVenue(ctx, sqlc.DeleteVenueParams{
		VenueUuid: venueUUID,
		OwnerUuid: ownerUUID,
	})
	if err != nil {
		return InternalError(c, "Failed to delete venue")
	}

	return c.NoContent(http.StatusNoContent)
}
