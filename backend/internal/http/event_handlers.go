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

// EventHandler handles event endpoints
type EventHandler struct {
	store *store.Store
}

// NewEventHandler creates a new event handler
func NewEventHandler(store *store.Store) *EventHandler {
	return &EventHandler{
		store: store,
	}
}

// Request/Response types

type CreateEventRequest struct {
	EventName       string `json:"event_name" validate:"required"`
	Datetime        string `json:"datetime" validate:"required"` // RFC3339 timestamp
	Comment         string `json:"comment"`
	DurationMinutes *int   `json:"duration_minutes"`
	SortOrder       *int32 `json:"sort_order"`
}

type UpdateEventRequest struct {
	EventName       *string `json:"event_name"`
	Datetime        *string `json:"datetime"` // RFC3339 timestamp
	Comment         *string `json:"comment"`
	DurationMinutes *int    `json:"duration_minutes"`
	SortOrder       *int32  `json:"sort_order"`
}

type EventResponse struct {
	EventUuid       string `json:"event_uuid"`
	EventListUuid   string `json:"event_list_uuid"`
	EventName       string `json:"event_name"`
	Datetime        string `json:"datetime"` // RFC3339 timestamp
	Comment         string `json:"comment"`
	DurationMinutes *int   `json:"duration_minutes"`
	SortOrder       int32  `json:"sort_order"`
	CreatedAt       string `json:"created_at"`
	ModifiedAt      string `json:"modified_at"`
}

// Helper functions

// int4ToInt converts pgtype.Int4 to *int (handle NULL)
func int4ToInt(i pgtype.Int4) *int {
	if !i.Valid {
		return nil
	}
	val := int(i.Int32)
	return &val
}

// intToInt4 converts *int to pgtype.Int4
func intToInt4(i *int) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{
		Int32: int32(*i),
		Valid: true,
	}
}

// eventToResponse converts sqlc.Event to EventResponse
func eventToResponse(event sqlc.Event) EventResponse {
	return EventResponse{
		EventUuid:       uuidToString(event.EventUuid),
		EventListUuid:   uuidToString(event.EventListUuid),
		EventName:       event.EventName,
		Datetime:        timestamptzToString(event.Datetime),
		Comment:         textToString(event.Comment),
		DurationMinutes: int4ToInt(event.DurationMinutes),
		SortOrder:       event.SortOrder,
		CreatedAt:       timestamptzToString(event.CreatedAt),
		ModifiedAt:      timestamptzToString(event.ModifiedAt),
	}
}

// Handlers

// ListByEventList handles GET /api/event-lists/:event_list_uuid/events
func (h *EventHandler) ListByEventList(c echo.Context) error {
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

	// List events
	events, err := h.store.Queries.ListEventsByEventListAndOwner(ctx, sqlc.ListEventsByEventListAndOwnerParams{
		EventListUuid: eventListUUID,
		OwnerUuid:     ownerUUID,
	})
	if err != nil {
		return InternalError(c, "Failed to list events")
	}

	// Convert to response format
	response := make([]EventResponse, len(events))
	for i, event := range events {
		response[i] = eventToResponse(event)
	}

	return c.JSON(http.StatusOK, response)
}

// Create handles POST /api/event-lists/:event_list_uuid/events
func (h *EventHandler) Create(c echo.Context) error {
	eventListUUIDStr := c.Param("event_list_uuid")
	if eventListUUIDStr == "" {
		return ValidationError(c, "event_list_uuid is required")
	}

	var req CreateEventRequest
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
	eventListUUID, err := stringToUUID(eventListUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid event_list_uuid format")
	}

	ownerUUID, err := stringToUUID(ownerUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	// Validate event list ownership first
	_, err = h.store.Queries.GetEventListByIDAndOwner(ctx, sqlc.GetEventListByIDAndOwnerParams{
		EventListUuid: eventListUUID,
		OwnerUuid:     ownerUUID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return NotFoundError(c, "Event list not found")
		}
		return InternalError(c, "Failed to validate event list ownership")
	}

	// Parse datetime (RFC3339)
	datetime, err := time.Parse(time.RFC3339, req.Datetime)
	if err != nil {
		return ValidationError(c, "Invalid datetime format (expected: RFC3339)")
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

	// Create event
	event, err := h.store.Queries.CreateEvent(ctx, sqlc.CreateEventParams{
		EventListUuid:   eventListUUID,
		EventName:       req.EventName,
		Datetime:        pgtype.Timestamptz{Time: datetime, Valid: true},
		Comment:         comment,
		DurationMinutes: intToInt4(req.DurationMinutes),
		SortOrder:       sortOrder,
	})
	if err != nil {
		return InternalError(c, "Failed to create event")
	}

	return c.JSON(http.StatusCreated, eventToResponse(event))
}

// Get handles GET /api/events/:event_uuid
func (h *EventHandler) Get(c echo.Context) error {
	eventUUIDStr := c.Param("event_uuid")
	if eventUUIDStr == "" {
		return ValidationError(c, "event_uuid is required")
	}

	// Get owner UUID from context
	ownerUUIDStr, err := GetOwnerUUIDFromContext(c)
	if err != nil {
		return UnauthorizedError(c, "Unauthorized")
	}

	ctx := c.Request().Context()

	// Convert UUIDs
	eventUUID, err := stringToUUID(eventUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid event_uuid format")
	}

	ownerUUID, err := stringToUUID(ownerUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	// Get event
	event, err := h.store.Queries.GetEventByIDAndOwner(ctx, sqlc.GetEventByIDAndOwnerParams{
		EventUuid: eventUUID,
		OwnerUuid: ownerUUID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return NotFoundError(c, "Event not found")
		}
		return InternalError(c, "Failed to get event")
	}

	return c.JSON(http.StatusOK, eventToResponse(event))
}

// Update handles PATCH /api/events/:event_uuid
func (h *EventHandler) Update(c echo.Context) error {
	eventUUIDStr := c.Param("event_uuid")
	if eventUUIDStr == "" {
		return ValidationError(c, "event_uuid is required")
	}

	var req UpdateEventRequest
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
	eventUUID, err := stringToUUID(eventUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid event_uuid format")
	}

	ownerUUID, err := stringToUUID(ownerUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	// Get existing event
	existingEvent, err := h.store.Queries.GetEventByIDAndOwner(ctx, sqlc.GetEventByIDAndOwnerParams{
		EventUuid: eventUUID,
		OwnerUuid: ownerUUID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return NotFoundError(c, "Event not found")
		}
		return InternalError(c, "Failed to get event")
	}

	// Prepare update params
	eventName := existingEvent.EventName
	if req.EventName != nil {
		eventName = *req.EventName
	}

	datetime := existingEvent.Datetime
	if req.Datetime != nil {
		parsedDatetime, err := time.Parse(time.RFC3339, *req.Datetime)
		if err != nil {
			return ValidationError(c, "Invalid datetime format (expected: RFC3339)")
		}
		datetime = pgtype.Timestamptz{Time: parsedDatetime, Valid: true}
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
		comment = existingEvent.Comment
	}

	durationMinutes := existingEvent.DurationMinutes
	if req.DurationMinutes != nil {
		durationMinutes = intToInt4(req.DurationMinutes)
	}

	sortOrder := existingEvent.SortOrder
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	// Update event
	event, err := h.store.Queries.UpdateEvent(ctx, sqlc.UpdateEventParams{
		EventUuid:       eventUUID,
		OwnerUuid:       ownerUUID,
		EventName:       eventName,
		Datetime:        datetime,
		Comment:         comment,
		DurationMinutes: durationMinutes,
		SortOrder:       sortOrder,
	})
	if err != nil {
		return InternalError(c, "Failed to update event")
	}

	return c.JSON(http.StatusOK, eventToResponse(event))
}

// Delete handles DELETE /api/events/:event_uuid
func (h *EventHandler) Delete(c echo.Context) error {
	eventUUIDStr := c.Param("event_uuid")
	if eventUUIDStr == "" {
		return ValidationError(c, "event_uuid is required")
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
	eventUUID, err := stringToUUID(eventUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid event_uuid format")
	}

	ownerUUID, err := stringToUUID(ownerUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	// Delete event
	err = h.store.Queries.DeleteEvent(ctx, sqlc.DeleteEventParams{
		EventUuid: eventUUID,
		OwnerUuid: ownerUUID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return NotFoundError(c, "Event not found")
		}
		return InternalError(c, "Failed to delete event")
	}

	return c.NoContent(http.StatusNoContent)
}
