package http

import (
	"fmt"
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
	EventName       string  `json:"event_name" validate:"required"`
	EventDate       *string `json:"event_date"`                     // YYYY-MM-DD
	EventTime       string  `json:"event_time" validate:"required"` // HH:MM or HH:MM:SS
	Comment         string  `json:"comment"`
	DurationMinutes *int    `json:"duration_minutes"`
	SortOrder       *int32  `json:"sort_order"`
}

type UpdateEventRequest struct {
	EventName       *string `json:"event_name"`
	EventDate       *string `json:"event_date"` // YYYY-MM-DD
	EventTime       *string `json:"event_time"` // HH:MM or HH:MM:SS
	Comment         *string `json:"comment"`
	DurationMinutes *int    `json:"duration_minutes"`
	SortOrder       *int32  `json:"sort_order"`
}

type EventResponse struct {
	EventUuid       string `json:"event_uuid"`
	EventListUuid   string `json:"event_list_uuid"`
	EventName       string `json:"event_name"`
	EventDate       string `json:"event_date,omitempty"` // YYYY-MM-DD
	EventTime       string `json:"event_time"`           // HH:MM:SS
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

// stringToTime converts string time (HH:MM or HH:MM:SS) to pgtype.Time
func stringToTime(timeStr string) (pgtype.Time, error) {
	if timeStr == "" {
		return pgtype.Time{Valid: false}, nil
	}

	layout := "15:04"
	if len(timeStr) > 5 {
		layout = "15:04:05"
	}

	t, err := time.Parse(layout, timeStr)
	if err != nil {
		return pgtype.Time{}, err
	}

	// Convert to microseconds since midnight
	micros := int64(t.Hour())*3600000000 + int64(t.Minute())*60000000 + int64(t.Second())*1000000
	return pgtype.Time{
		Microseconds: micros,
		Valid:        true,
	}, nil
}

// timeToString converts pgtype.Time to string (HH:MM:SS format)
func timeToString(t pgtype.Time) string {
	if !t.Valid {
		return ""
	}
	h := t.Microseconds / 3600000000
	m := (t.Microseconds % 3600000000) / 60000000
	s := (t.Microseconds % 60000000) / 1000000
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// eventToResponse converts sqlc.Event to EventResponse
func eventToResponse(event sqlc.Event) EventResponse {
	return EventResponse{
		EventUuid:       uuidToString(event.EventUuid),
		EventListUuid:   uuidToString(event.EventListUuid),
		EventName:       event.EventName,
		EventDate:       dateToString(event.EventDate),
		EventTime:       timeToString(event.EventTime),
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

	// Parse date
	var eventDate pgtype.Date
	if req.EventDate != nil {
		d, err := stringToDate(*req.EventDate)
		if err != nil {
			return ValidationError(c, "Invalid event_date format (expected: YYYY-MM-DD)")
		}
		eventDate = d
	}

	// Parse time
	eventTime, err := stringToTime(req.EventTime)
	if err != nil {
		return ValidationError(c, "Invalid event_time format (expected: HH:MM or HH:MM:SS)")
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
		EventDate:       eventDate,
		EventTime:       eventTime,
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

	eventDate := existingEvent.EventDate
	if req.EventDate != nil {
		d, err := stringToDate(*req.EventDate)
		if err != nil {
			return ValidationError(c, "Invalid event_date format (expected: YYYY-MM-DD)")
		}
		eventDate = d
	}

	eventTime := existingEvent.EventTime
	if req.EventTime != nil {
		t, err := stringToTime(*req.EventTime)
		if err != nil {
			return ValidationError(c, "Invalid event_time format (expected: HH:MM or HH:MM:SS)")
		}
		eventTime = t
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
		EventDate:       eventDate,
		EventTime:       eventTime,
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
