package http

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	sqlc "github.com/michelemendel/times.place/db/sqlc"
	"github.com/michelemendel/times.place/internal/store"
)

// EventListHandler handles event list endpoints
type EventListHandler struct {
	store *store.Store
}

// NewEventListHandler creates a new event list handler
func NewEventListHandler(store *store.Store) *EventListHandler {
	return &EventListHandler{
		store: store,
	}
}

// Request/Response types

type CreateEventListRequest struct {
	Name             string `json:"name"`
	Date             string `json:"date"` // RFC3339 date format: "2006-01-02"
	Comment          string `json:"comment"`
	Visibility       string `json:"visibility" validate:"required,oneof=public private"`
	PrivateLinkToken string `json:"private_link_token"` // Optional UUID string
	SortOrder        *int32 `json:"sort_order"`
}

type UpdateEventListRequest struct {
	Name             *string `json:"name"`
	Date             *string `json:"date"` // RFC3339 date format: "2006-01-02"
	Comment          *string `json:"comment"`
	Visibility       *string `json:"visibility" validate:"omitempty,oneof=public private"`
	PrivateLinkToken *string `json:"private_link_token"` // Optional UUID string
	SortOrder        *int32  `json:"sort_order"`
}

type EventListResponse struct {
	EventListUuid    string `json:"event_list_uuid"`
	VenueUuid        string `json:"venue_uuid"`
	Name             string `json:"name"`
	Date             string `json:"date"` // RFC3339 date format or empty
	Comment          string `json:"comment"`
	Visibility       string `json:"visibility"`
	PrivateLinkToken string `json:"private_link_token"`
	SortOrder        int32  `json:"sort_order"`
	CreatedAt        string `json:"created_at"`
	ModifiedAt       string `json:"modified_at"`
}

// Helper functions

// dateToString converts pgtype.Date to string (RFC3339 date format)
func dateToString(d pgtype.Date) string {
	if !d.Valid {
		return ""
	}
	// pgx/v5 pgtype.Date stores the value as time.Time
	return d.Time.Format("2006-01-02")
}

// stringToDate converts string date (RFC3339 date format) to pgtype.Date
func stringToDate(dateStr string) (pgtype.Date, error) {
	if dateStr == "" {
		return pgtype.Date{Valid: false}, nil
	}

	// Parse date string (format: "2006-01-02")
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return pgtype.Date{}, err
	}

	return pgtype.Date{
		Time:  t,
		Valid: true,
	}, nil
}

// eventListToResponse converts sqlc.EventList to EventListResponse (used internally).
func eventListToResponse(eventList sqlc.EventList) EventListResponse {
	return EventListToResponse(eventList)
}

// EventListToResponse converts sqlc.EventList to EventListResponse (exported for use by venue handler).
func EventListToResponse(eventList sqlc.EventList) EventListResponse {
	return EventListResponse{
		EventListUuid:    uuidToString(eventList.EventListUuid),
		VenueUuid:        uuidToString(eventList.VenueUuid),
		Name:             eventList.Name,
		Date:             dateToString(eventList.Date),
		Comment:          textToString(eventList.Comment),
		Visibility:       eventList.Visibility,
		PrivateLinkToken: uuidToString(eventList.PrivateLinkToken),
		SortOrder:        eventList.SortOrder,
		CreatedAt:        timestamptzToString(eventList.CreatedAt),
		ModifiedAt:       timestamptzToString(eventList.ModifiedAt),
	}
}

// Handlers

// ListByVenueQuery handles GET /api/event-lists?venue_uuid=xxx
// Used by the frontend to list event lists for a venue (avoids path-matching issues with nested routes).
func (h *EventListHandler) ListByVenueQuery(c echo.Context) error {
	venueUUIDStr := c.QueryParam("venue_uuid")
	if venueUUIDStr == "" {
		return ValidationError(c, "venue_uuid query parameter is required")
	}

	ownerUUIDStr, err := GetOwnerUUIDFromContext(c)
	if err != nil {
		return UnauthorizedError(c, "Unauthorized")
	}

	ctx := c.Request().Context()

	venueUUID, err := stringToUUID(venueUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid venue_uuid format")
	}

	ownerUUID, err := stringToUUID(ownerUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	eventLists, err := h.store.Queries.ListEventListsByVenueAndOwner(ctx, sqlc.ListEventListsByVenueAndOwnerParams{
		VenueUuid: venueUUID,
		OwnerUuid: ownerUUID,
	})
	if err != nil {
		return InternalError(c, "Failed to list event lists")
	}

	response := make([]EventListResponse, len(eventLists))
	for i, eventList := range eventLists {
		response[i] = eventListToResponse(eventList)
	}

	return c.JSON(http.StatusOK, response)
}

// ListByVenue handles GET /api/venues/:venue_uuid/event-lists
func (h *EventListHandler) ListByVenue(c echo.Context) error {
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

	// List event lists
	eventLists, err := h.store.Queries.ListEventListsByVenueAndOwner(ctx, sqlc.ListEventListsByVenueAndOwnerParams{
		VenueUuid: venueUUID,
		OwnerUuid: ownerUUID,
	})
	if err != nil {
		return InternalError(c, "Failed to list event lists")
	}

	// Convert to response format
	response := make([]EventListResponse, len(eventLists))
	for i, eventList := range eventLists {
		response[i] = eventListToResponse(eventList)
	}

	return c.JSON(http.StatusOK, response)
}

// Create handles POST /api/venues/:venue_uuid/event-lists
func (h *EventListHandler) Create(c echo.Context) error {
	venueUUIDStr := c.Param("venue_uuid")
	if venueUUIDStr == "" {
		return ValidationError(c, "venue_uuid is required")
	}

	var req CreateEventListRequest
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

	// Convert UUIDs
	venueUUID, err := stringToUUID(venueUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid venue_uuid format")
	}

	ownerUUID, err := stringToUUID(ownerUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	// Validate venue ownership first
	_, err = h.store.Queries.GetVenueByIDAndOwner(ctx, sqlc.GetVenueByIDAndOwnerParams{
		VenueUuid: venueUUID,
		OwnerUuid: ownerUUID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return NotFoundError(c, "Venue not found")
		}
		return InternalError(c, "Failed to validate venue ownership")
	}

	// Parse date
	date, err := stringToDate(req.Date)
	if err != nil {
		return ValidationError(c, "Invalid date format (expected: YYYY-MM-DD)")
	}

	// Parse private link token
	privateLinkToken, err := parsePrivateLinkToken(req.PrivateLinkToken)
	if err != nil {
		return ValidationError(c, "Invalid private_link_token format")
	}

	// Prepare comment
	var comment pgtype.Text
	if req.Comment != "" {
		comment = pgtype.Text{
			String: req.Comment,
			Valid:  true,
		}
	}

	// Sort order (default to 0)
	sortOrder := int32(0)
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	// Create event list
	eventList, err := h.store.Queries.CreateEventList(ctx, sqlc.CreateEventListParams{
		VenueUuid:        venueUUID,
		Name:             req.Name,
		Date:             date,
		Comment:          comment,
		Visibility:       req.Visibility,
		PrivateLinkToken: privateLinkToken,
		SortOrder:        sortOrder,
	})
	if err != nil {
		return InternalError(c, "Failed to create event list")
	}

	return c.JSON(http.StatusCreated, eventListToResponse(eventList))
}

// Get handles GET /api/event-lists/:event_list_uuid
func (h *EventListHandler) Get(c echo.Context) error {
	eventListUUIDStr := c.Param("event_list_uuid")
	if eventListUUIDStr == "" {
		return ValidationError(c, "event_list_uuid is required")
	}

	// Get owner UUID from context
	ownerUUIDStr, err := GetOwnerUUIDFromContext(c)
	if err != nil {
		return UnauthorizedError(c, "Unauthorized")
	}

	ctx := c.Request().Context()

	// Convert UUIDs
	eventListUUID, err := stringToUUID(eventListUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid event_list_uuid format")
	}

	ownerUUID, err := stringToUUID(ownerUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	// Get event list
	eventList, err := h.store.Queries.GetEventListByIDAndOwner(ctx, sqlc.GetEventListByIDAndOwnerParams{
		EventListUuid: eventListUUID,
		OwnerUuid:     ownerUUID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return NotFoundError(c, "Event list not found")
		}
		return InternalError(c, "Failed to get event list")
	}

	return c.JSON(http.StatusOK, eventListToResponse(eventList))
}

// Update handles PATCH /api/event-lists/:event_list_uuid
func (h *EventListHandler) Update(c echo.Context) error {
	eventListUUIDStr := c.Param("event_list_uuid")
	if eventListUUIDStr == "" {
		return ValidationError(c, "event_list_uuid is required")
	}

	var req UpdateEventListRequest
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

	// Convert UUIDs
	eventListUUID, err := stringToUUID(eventListUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid event_list_uuid format")
	}

	ownerUUID, err := stringToUUID(ownerUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	// Get existing event list
	existingEventList, err := h.store.Queries.GetEventListByIDAndOwner(ctx, sqlc.GetEventListByIDAndOwnerParams{
		EventListUuid: eventListUUID,
		OwnerUuid:     ownerUUID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return NotFoundError(c, "Event list not found")
		}
		return InternalError(c, "Failed to get event list")
	}

	// Prepare update params
	name := existingEventList.Name
	if req.Name != nil {
		name = *req.Name
	}

	date := existingEventList.Date
	if req.Date != nil {
		parsedDate, err := stringToDate(*req.Date)
		if err != nil {
			return ValidationError(c, "Invalid date format (expected: YYYY-MM-DD)")
		}
		date = parsedDate
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
		comment = existingEventList.Comment
	}

	visibility := existingEventList.Visibility
	if req.Visibility != nil {
		visibility = *req.Visibility
	}

	privateLinkToken := existingEventList.PrivateLinkToken
	if req.PrivateLinkToken != nil {
		if *req.PrivateLinkToken != "" {
			parsed, err := stringToUUID(*req.PrivateLinkToken)
			if err != nil {
				return ValidationError(c, "Invalid private_link_token format")
			}
			privateLinkToken = parsed
		} else {
			privateLinkToken = pgtype.UUID{Valid: false}
		}
	}

	sortOrder := existingEventList.SortOrder
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	// Update event list
	eventList, err := h.store.Queries.UpdateEventList(ctx, sqlc.UpdateEventListParams{
		EventListUuid:    eventListUUID,
		OwnerUuid:        ownerUUID,
		Name:             name,
		Date:             date,
		Comment:          comment,
		Visibility:       visibility,
		PrivateLinkToken: privateLinkToken,
		SortOrder:        sortOrder,
	})
	if err != nil {
		return InternalError(c, "Failed to update event list")
	}

	return c.JSON(http.StatusOK, eventListToResponse(eventList))
}

// Delete handles DELETE /api/event-lists/:event_list_uuid
func (h *EventListHandler) Delete(c echo.Context) error {
	eventListUUIDStr := c.Param("event_list_uuid")
	if eventListUUIDStr == "" {
		return ValidationError(c, "event_list_uuid is required")
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

	// Convert UUIDs
	eventListUUID, err := stringToUUID(eventListUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid event_list_uuid format")
	}

	ownerUUID, err := stringToUUID(ownerUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	// Delete event list (cascade handled by DB)
	err = h.store.Queries.DeleteEventList(ctx, sqlc.DeleteEventListParams{
		EventListUuid: eventListUUID,
		OwnerUuid:     ownerUUID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return NotFoundError(c, "Event list not found")
		}
		return InternalError(c, "Failed to delete event list")
	}

	return c.NoContent(http.StatusNoContent)
}
