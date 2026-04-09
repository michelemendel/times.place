package http

import (
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	sqlc "github.com/michelemendel/times.place/db/sqlc"
	"github.com/michelemendel/times.place/internal/store"
)

// PublicHandler handles public endpoints
type PublicHandler struct {
	store *store.Store
}

// NewPublicHandler creates a new public handler
func NewPublicHandler(store *store.Store) *PublicHandler {
	return &PublicHandler{
		store: store,
	}
}

// Response types

type VenueWithEventListsResponse struct {
	Venue      VenueResponse       `json:"venue"`
	EventLists []EventListResponse `json:"event_lists"`
}

type EventListWithVenueAndEventsResponse struct {
	Venue     VenueResponse     `json:"venue"`
	EventList EventListResponse `json:"event_list"`
	Events    []EventResponse   `json:"events"`
}

// Helper functions to convert public query results

// venueRowToResponse converts ListPublicVenuesRow to VenueResponse
func venueRowToResponse(venue sqlc.ListPublicVenuesRow) VenueResponse {
	return VenueResponse{
		VenueUuid:        uuidToString(venue.VenueUuid),
		OwnerUuid:        "", // Public endpoints don't expose owner info
		Name:             venue.Name,
		BannerImage:      venue.BannerImage,
		Address:          venue.Address,
		Geolocation:      venue.Geolocation,
		DistanceKm:       nil,
		Comment:          textToString(venue.Comment),
		Timezone:         venue.Timezone,
		PrivateLinkToken: uuidToString(venue.PrivateLinkToken),
		OwnerName:        venue.OwnerName,
		OwnerEmail:       venue.OwnerEmail,
		CreatedAt:        timestamptzToString(venue.CreatedAt),
		ModifiedAt:       timestamptzToString(venue.ModifiedAt),
	}
}

// searchVenueRowToResponse converts SearchPublicVenuesRow to VenueResponse
func searchVenueRowToResponse(venue sqlc.SearchPublicVenuesRow) VenueResponse {
	return VenueResponse{
		VenueUuid:        uuidToString(venue.VenueUuid),
		OwnerUuid:        "", // Public endpoints don't expose owner info
		Name:             venue.Name,
		BannerImage:      venue.BannerImage,
		Address:          venue.Address,
		Geolocation:      venue.Geolocation,
		DistanceKm:       nil,
		Comment:          textToString(venue.Comment),
		Timezone:         venue.Timezone,
		PrivateLinkToken: uuidToString(venue.PrivateLinkToken),
		OwnerName:        venue.OwnerName,
		OwnerEmail:       venue.OwnerEmail,
		CreatedAt:        timestamptzToString(venue.CreatedAt),
		ModifiedAt:       timestamptzToString(venue.ModifiedAt),
	}
}

// venueTokenRowToResponse converts GetVenueByTokenRow to VenueResponse
func venueTokenRowToResponse(venue sqlc.GetVenueByTokenRow) VenueResponse {
	return VenueResponse{
		VenueUuid:        uuidToString(venue.VenueUuid),
		OwnerUuid:        "", // Public endpoints don't expose owner info
		Name:             venue.Name,
		BannerImage:      venue.BannerImage,
		Address:          venue.Address,
		Geolocation:      venue.Geolocation,
		DistanceKm:       nil,
		Comment:          textToString(venue.Comment),
		Timezone:         venue.Timezone,
		PrivateLinkToken: uuidToString(venue.PrivateLinkToken),
		OwnerName:        venue.OwnerName,
		OwnerEmail:       venue.OwnerEmail,
		CreatedAt:        timestamptzToString(venue.CreatedAt),
		ModifiedAt:       timestamptzToString(venue.ModifiedAt),
	}
}

// Handlers

type latLng struct {
	Lat float64
	Lng float64
}

func parseGeolocation(geolocation string) *latLng {
	parts := strings.Split(geolocation, ",")
	if len(parts) != 2 {
		return nil
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return nil
	}
	lng, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return nil
	}
	return &latLng{Lat: lat, Lng: lng}
}

func haversineKm(a, b latLng) float64 {
	const earthRadiusKm = 6371.0
	toRad := func(deg float64) float64 { return deg * (math.Pi / 180) }

	dLat := toRad(b.Lat - a.Lat)
	dLng := toRad(b.Lng - a.Lng)

	lat1 := toRad(a.Lat)
	lat2 := toRad(b.Lat)

	sinDLat := math.Sin(dLat / 2)
	sinDLng := math.Sin(dLng / 2)

	h := sinDLat*sinDLat + math.Cos(lat1)*math.Cos(lat2)*sinDLng*sinDLng
	return 2 * earthRadiusKm * math.Asin(math.Sqrt(h))
}

func roundTo4Decimals(v float64) float64 {
	return math.Round(v*1e4) / 1e4
}

// ListVenues handles GET /api/public/venues
func (h *PublicHandler) ListVenues(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	ctx := c.Request().Context()

	// Check for search query parameter
	query := c.QueryParam("query")

	latStr := c.QueryParam("lat")
	lngStr := c.QueryParam("lng")
	radiusStr := c.QueryParam("radius_km")

	hasUserLocation := latStr != "" && lngStr != ""
	var userLat, userLng float64
	var radiusKm pgtype.Float8

	if hasUserLocation {
		var err error
		userLat, err = strconv.ParseFloat(latStr, 64)
		if err != nil {
			return ValidationError(c, "Invalid lat")
		}
		userLng, err = strconv.ParseFloat(lngStr, 64)
		if err != nil {
			return ValidationError(c, "Invalid lng")
		}
		if radiusStr != "" {
			r, err := strconv.ParseFloat(radiusStr, 64)
			if err != nil {
				return ValidationError(c, "Invalid radius_km")
			}
			radiusKm = pgtype.Float8{Float64: r, Valid: true}
		} else {
			radiusKm = pgtype.Float8{Valid: false}
		}
	}

	var response []VenueResponse

	if query != "" {
		if hasUserLocation {
			searchVenues, err := h.store.Queries.SearchPublicVenues(ctx, pgtype.Text{
				String: query,
				Valid:  true,
			})
			if err != nil {
				return InternalError(c, "Failed to search venues")
			}

			user := latLng{Lat: userLat, Lng: userLng}
			tmp := make([]VenueResponse, 0, len(searchVenues))
			for _, venue := range searchVenues {
				r := searchVenueRowToResponse(venue)
				if coords := parseGeolocation(venue.Geolocation); coords != nil {
					d := roundTo4Decimals(haversineKm(user, *coords))
					r.DistanceKm = &d
				}
				if radiusKm.Valid && (r.DistanceKm == nil || *r.DistanceKm > radiusKm.Float64) {
					continue
				}
				tmp = append(tmp, r)
			}
			sort.Slice(tmp, func(i, j int) bool {
				ai, aj := tmp[i].DistanceKm, tmp[j].DistanceKm
				if ai == nil && aj == nil {
					return tmp[i].CreatedAt > tmp[j].CreatedAt
				}
				if ai == nil {
					return false
				}
				if aj == nil {
					return true
				}
				if *ai == *aj {
					return tmp[i].CreatedAt > tmp[j].CreatedAt
				}
				return *ai < *aj
			})
			response = tmp
		} else {
			// Use search query
			searchVenues, err := h.store.Queries.SearchPublicVenues(ctx, pgtype.Text{
				String: query,
				Valid:  true,
			})
			if err != nil {
				return InternalError(c, "Failed to search venues")
			}
			// Convert to response format
			response = make([]VenueResponse, len(searchVenues))
			for i, venue := range searchVenues {
				response[i] = searchVenueRowToResponse(venue)
			}
		}
	} else {
		if hasUserLocation {
			venues, err := h.store.Queries.ListPublicVenues(ctx)
			if err != nil {
				return InternalError(c, "Failed to list venues")
			}

			user := latLng{Lat: userLat, Lng: userLng}
			tmp := make([]VenueResponse, 0, len(venues))
			for _, venue := range venues {
				r := venueRowToResponse(venue)
				if coords := parseGeolocation(venue.Geolocation); coords != nil {
					d := roundTo4Decimals(haversineKm(user, *coords))
					r.DistanceKm = &d
				}
				if radiusKm.Valid && (r.DistanceKm == nil || *r.DistanceKm > radiusKm.Float64) {
					continue
				}
				tmp = append(tmp, r)
			}
			sort.Slice(tmp, func(i, j int) bool {
				ai, aj := tmp[i].DistanceKm, tmp[j].DistanceKm
				if ai == nil && aj == nil {
					return tmp[i].CreatedAt > tmp[j].CreatedAt
				}
				if ai == nil {
					return false
				}
				if aj == nil {
					return true
				}
				if *ai == *aj {
					return tmp[i].CreatedAt > tmp[j].CreatedAt
				}
				return *ai < *aj
			})
			response = tmp
		} else {
			// Use list query
			venues, err := h.store.Queries.ListPublicVenues(ctx)
			if err != nil {
				return InternalError(c, "Failed to list venues")
			}
			// Convert to response format
			response = make([]VenueResponse, len(venues))
			for i, venue := range venues {
				response[i] = venueRowToResponse(venue)
			}
		}
	}

	return c.JSON(http.StatusOK, response)
}

// GetEventListsByVenue handles GET /api/public/venues/:venue_uuid/event-lists
func (h *PublicHandler) GetEventListsByVenue(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	venueUUIDStr := c.Param("venue_uuid")
	if venueUUIDStr == "" {
		return ValidationError(c, "venue_uuid is required")
	}

	ctx := c.Request().Context()

	// Convert UUID
	venueUUID, err := stringToUUID(venueUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid venue_uuid format")
	}

	// Get public event lists
	eventLists, err := h.store.Queries.GetPublicEventListsByVenue(ctx, venueUUID)
	if err != nil {
		return InternalError(c, "Failed to get event lists")
	}

	// Convert to response format
	response := make([]EventListResponse, len(eventLists))
	for i, eventList := range eventLists {
		response[i] = eventListToResponse(eventList)
	}

	return c.JSON(http.StatusOK, response)
}

// GetVenueByToken handles GET /api/public/venues/by-token/:token
func (h *PublicHandler) GetVenueByToken(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	tokenStr := c.Param("token")
	if tokenStr == "" {
		return ValidationError(c, "token is required")
	}

	ctx := c.Request().Context()

	// Convert token to UUID
	tokenUUID, err := stringToUUID(tokenStr)
	if err != nil {
		return ValidationError(c, "Invalid token format")
	}

	// Get venue with event lists
	rows, err := h.store.Queries.GetVenueWithEventListsByToken(ctx, tokenUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return NotFoundError(c, "Venue not found")
		}
		return InternalError(c, "Failed to get venue")
	}

	if len(rows) == 0 {
		return NotFoundError(c, "Venue not found")
	}

	// First row contains venue info (all rows have same venue data)
	firstRow := rows[0]
	venue := venueTokenRowToResponse(sqlc.GetVenueByTokenRow{
		VenueUuid:        firstRow.VenueUuid,
		Name:             firstRow.Name,
		BannerImage:      firstRow.BannerImage,
		Address:          firstRow.Address,
		Geolocation:      firstRow.Geolocation,
		Comment:          firstRow.Comment,
		Timezone:         firstRow.Timezone,
		PrivateLinkToken: firstRow.PrivateLinkToken,
		CreatedAt:        firstRow.CreatedAt,
		ModifiedAt:       firstRow.ModifiedAt,
		OwnerName:        firstRow.OwnerName,
		OwnerEmail:       firstRow.OwnerEmail,
	})

	// Convert event lists from rows
	eventLists := make([]EventListResponse, 0)
	for _, row := range rows {
		// Skip if event list UUID is null (venue might have no event lists)
		if !row.EventListUuid.Valid {
			continue
		}

		eventList := EventListResponse{
			EventListUuid:    uuidToString(row.EventListUuid),
			VenueUuid:        uuidToString(row.VenueUuid),
			Name:             textToString(row.EventListName),
			Date:             dateToString(row.EventListDate),
			Comment:          textToString(row.EventListComment),
			Visibility:       textToString(row.EventListVisibility),
			PrivateLinkToken: uuidToString(row.EventListPrivateLinkToken),
			SortOrder:        int32(row.EventListSortOrder.Int32),
			CreatedAt:        timestamptzToString(row.EventListCreatedAt),
			ModifiedAt:       timestamptzToString(row.EventListModifiedAt),
		}
		eventLists = append(eventLists, eventList)
	}

	return c.JSON(http.StatusOK, VenueWithEventListsResponse{
		Venue:      venue,
		EventLists: eventLists,
	})
}

// GetEventListByToken handles GET /api/public/event-lists/by-token/:token
func (h *PublicHandler) GetEventListByToken(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	tokenStr := c.Param("token")
	if tokenStr == "" {
		return ValidationError(c, "token is required")
	}

	ctx := c.Request().Context()

	// Convert token to UUID
	tokenUUID, err := stringToUUID(tokenStr)
	if err != nil {
		return ValidationError(c, "Invalid token format")
	}

	// Get event list by token
	eventList, err := h.store.Queries.GetEventListByToken(ctx, tokenUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return NotFoundError(c, "Event list not found")
		}
		return InternalError(c, "Failed to get event list")
	}

	// Get parent venue (public fields only)
	// We need to query venue by venue_uuid without owner check
	// Since this is a public endpoint, we can query directly
	venueUUID := eventList.VenueUuid
	var venue VenueResponse

	// Query venue with owner name and email (no owner check needed for public endpoint)
	row := h.store.DB().QueryRow(ctx,
		"SELECT v.venue_uuid, v.name, v.banner_image, v.address, v.geolocation, v.comment, v.timezone, v.private_link_token, v.created_at, v.modified_at, o.name AS owner_name, o.email AS owner_email FROM venues v INNER JOIN venue_owners o ON o.owner_uuid = v.owner_uuid WHERE v.venue_uuid = $1",
		venueUUID)

	var v sqlc.GetVenueByTokenRow
	err = row.Scan(
		&v.VenueUuid,
		&v.Name,
		&v.BannerImage,
		&v.Address,
		&v.Geolocation,
		&v.Comment,
		&v.Timezone,
		&v.PrivateLinkToken,
		&v.CreatedAt,
		&v.ModifiedAt,
		&v.OwnerName,
		&v.OwnerEmail,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return NotFoundError(c, "Venue not found")
		}
		return InternalError(c, "Failed to get venue")
	}
	venue = venueTokenRowToResponse(v)

	// Get events for this event list (public access - no owner check)
	// Query events directly by event_list_uuid
	eventRows, err := h.store.DB().Query(ctx,
		"SELECT event_uuid, event_list_uuid, event_name, event_date, event_time, comment, duration_minutes, sort_order, created_at, modified_at FROM events WHERE event_list_uuid = $1 ORDER BY sort_order ASC, event_date ASC, event_time ASC",
		eventList.EventListUuid)
	if err != nil {
		return InternalError(c, "Failed to get events")
	}
	defer eventRows.Close()

	events := make([]EventResponse, 0)
	for eventRows.Next() {
		var e sqlc.Event
		err := eventRows.Scan(
			&e.EventUuid,
			&e.EventListUuid,
			&e.EventName,
			&e.EventDate,
			&e.EventTime,
			&e.Comment,
			&e.DurationMinutes,
			&e.SortOrder,
			&e.CreatedAt,
			&e.ModifiedAt,
		)
		if err != nil {
			return InternalError(c, "Failed to scan event")
		}
		events = append(events, eventToResponse(e))
	}
	if err := eventRows.Err(); err != nil {
		return InternalError(c, "Failed to read events")
	}

	return c.JSON(http.StatusOK, EventListWithVenueAndEventsResponse{
		Venue:     venue,
		EventList: eventListToResponse(eventList),
		Events:    events,
	})
}

// GetEventsByEventList handles GET /api/public/event-lists/:event_list_uuid/events
// Only returns events if the event list is public
func (h *PublicHandler) GetEventsByEventList(c echo.Context) error {
	c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	eventListUUIDStr := c.Param("event_list_uuid")
	if eventListUUIDStr == "" {
		return ValidationError(c, "event_list_uuid is required")
	}

	ctx := c.Request().Context()

	// Convert UUID
	eventListUUID, err := stringToUUID(eventListUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid event_list_uuid format")
	}

	// Query event list directly to check if it exists and is public
	row := h.store.DB().QueryRow(ctx,
		"SELECT event_list_uuid, venue_uuid, name, date, comment, visibility, private_link_token, sort_order, created_at, modified_at FROM event_lists WHERE event_list_uuid = $1",
		eventListUUID)

	var el sqlc.EventList
	err = row.Scan(
		&el.EventListUuid,
		&el.VenueUuid,
		&el.Name,
		&el.Date,
		&el.Comment,
		&el.Visibility,
		&el.PrivateLinkToken,
		&el.SortOrder,
		&el.CreatedAt,
		&el.ModifiedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return NotFoundError(c, "Event list not found")
		}
		return InternalError(c, "Failed to get event list")
	}

	// Only allow access to public event lists
	if el.Visibility != "public" {
		return NotFoundError(c, "Event list not found")
	}

	// Get events for this event list (public access - no owner check)
	eventRows, err := h.store.DB().Query(ctx,
		"SELECT event_uuid, event_list_uuid, event_name, event_date, event_time, comment, duration_minutes, sort_order, created_at, modified_at FROM events WHERE event_list_uuid = $1 ORDER BY sort_order ASC, event_date ASC, event_time ASC",
		eventListUUID)
	if err != nil {
		return InternalError(c, "Failed to get events")
	}
	defer eventRows.Close()

	events := make([]EventResponse, 0)
	for eventRows.Next() {
		var e sqlc.Event
		err := eventRows.Scan(
			&e.EventUuid,
			&e.EventListUuid,
			&e.EventName,
			&e.EventDate,
			&e.EventTime,
			&e.Comment,
			&e.DurationMinutes,
			&e.SortOrder,
			&e.CreatedAt,
			&e.ModifiedAt,
		)
		if err != nil {
			return InternalError(c, "Failed to scan event")
		}
		events = append(events, eventToResponse(e))
	}
	if err := eventRows.Err(); err != nil {
		return InternalError(c, "Failed to read events")
	}

	return c.JSON(http.StatusOK, events)
}
