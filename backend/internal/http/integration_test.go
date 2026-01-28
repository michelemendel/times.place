package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	sqlc "github.com/michelemendel/times.place/db/sqlc"
	"github.com/michelemendel/times.place/internal/service"
	"github.com/michelemendel/times.place/internal/store"
	"github.com/michelemendel/times.place/internal/test"
)

type testServer struct {
	E       *echo.Echo
	Store   *store.Store
	Auth    *service.AuthService
	Cleanup func()
}

func setupIntegrationServer(t *testing.T) *testServer {
	t.Helper()
	test.RequireDatabase(t)

	// Ensure auth secrets exist for token generation/parsing
	if os.Getenv("JWT_SECRET") == "" {
		t.Setenv("JWT_SECRET", "test-jwt-secret")
	}
	if os.Getenv("REFRESH_TOKEN_SECRET") == "" {
		t.Setenv("REFRESH_TOKEN_SECRET", "test-refresh-secret")
	}

	// Point app DB at the test DB for the store.
	t.Setenv("DATABASE_URL", test.GetDatabaseURL())

	st, err := store.NewStore(os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Run everything inside a transaction so tests are isolated.
	tx, err := st.DB().Begin(t.Context())
	if err != nil {
		st.Close()
		t.Fatalf("failed to begin tx: %v", err)
	}
	st.Queries = sqlc.New(tx)

	auth, err := service.NewAuthService()
	if err != nil {
		_ = tx.Rollback(t.Context())
		st.Close()
		t.Fatalf("failed to create auth service: %v", err)
	}

	e := echo.New()
	e.HideBanner = true
	e.Validator = NewCustomValidator()
	RegisterRoutes(e, st, auth)

	cleanup := func() {
		_ = tx.Rollback(t.Context())
		st.Close()
	}

	return &testServer{
		E:       e,
		Store:   st,
		Auth:    auth,
		Cleanup: cleanup,
	}
}

func doJSONRequest(t *testing.T, e *echo.Echo, method, path string, body any, headers map[string]string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("failed to encode request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

type authResp struct {
	Owner struct {
		OwnerUUID string `json:"owner_uuid"`
		Email     string `json:"email"`
	} `json:"owner"`
	AccessToken string `json:"access_token"`
}

type venueResp struct {
	VenueUUID string `json:"venue_uuid"`
	Name      string `json:"name"`
}

func TestIntegration_AuthAndVenueCRUD_MinimalFlow(t *testing.T) {
	s := setupIntegrationServer(t)
	defer s.Cleanup()

	// Register user -> get access token
	reg := doJSONRequest(t, s.E, http.MethodPost, "/api/auth/register", map[string]any{
		"name":     "Owner One",
		"email":    "owner1@example.com",
		"mobile":   "555-0001",
		"password": "password123",
	}, nil, nil)
	if reg.Code != http.StatusCreated {
		t.Fatalf("expected register 201, got %d: %s", reg.Code, reg.Body.String())
	}

	var ar authResp
	if err := json.Unmarshal(reg.Body.Bytes(), &ar); err != nil {
		t.Fatalf("failed to unmarshal register response: %v", err)
	}
	if ar.AccessToken == "" {
		t.Fatalf("expected non-empty access token")
	}

	authz := map[string]string{"Authorization": "Bearer " + ar.AccessToken}

	// Create venue
	createVenue := doJSONRequest(t, s.E, http.MethodPost, "/api/venues", map[string]any{
		"name":              "Test Venue",
		"banner_image":      "",
		"address":           "",
		"geolocation":       "",
		"comment":           "",
		"timezone":          "UTC",
		"visibility":        "public",
		"private_link_token": "",
	}, authz, nil)
	if createVenue.Code != http.StatusCreated {
		t.Fatalf("expected create venue 201, got %d: %s", createVenue.Code, createVenue.Body.String())
	}

	var vr venueResp
	if err := json.Unmarshal(createVenue.Body.Bytes(), &vr); err != nil {
		t.Fatalf("failed to unmarshal create venue response: %v", err)
	}
	if vr.VenueUUID == "" {
		t.Fatalf("expected venue_uuid")
	}

	// Get venue
	getVenue := doJSONRequest(t, s.E, http.MethodGet, "/api/venues/"+vr.VenueUUID, nil, authz, nil)
	if getVenue.Code != http.StatusOK {
		t.Fatalf("expected get venue 200, got %d: %s", getVenue.Code, getVenue.Body.String())
	}

	// Update venue (name)
	patchVenue := doJSONRequest(t, s.E, http.MethodPatch, "/api/venues/"+vr.VenueUUID, map[string]any{
		"name": "Updated Venue",
	}, authz, nil)
	if patchVenue.Code != http.StatusOK {
		t.Fatalf("expected patch venue 200, got %d: %s", patchVenue.Code, patchVenue.Body.String())
	}

	// Delete venue
	delVenue := doJSONRequest(t, s.E, http.MethodDelete, "/api/venues/"+vr.VenueUUID, nil, authz, nil)
	if delVenue.Code != http.StatusNoContent {
		t.Fatalf("expected delete venue 204, got %d: %s", delVenue.Code, delVenue.Body.String())
	}

	// Get after delete -> 404
	getAfterDelete := doJSONRequest(t, s.E, http.MethodGet, "/api/venues/"+vr.VenueUUID, nil, authz, nil)
	if getAfterDelete.Code != http.StatusNotFound {
		t.Fatalf("expected get-after-delete 404, got %d: %s", getAfterDelete.Code, getAfterDelete.Body.String())
	}
}

func TestIntegration_TokenBasedAccessControl(t *testing.T) {
	s := setupIntegrationServer(t)
	defer s.Cleanup()

	// No token -> protected route returns 401
	noAuth := doJSONRequest(t, s.E, http.MethodGet, "/api/venues", nil, nil, nil)
	if noAuth.Code != http.StatusUnauthorized {
		t.Fatalf("expected no-auth 401, got %d: %s", noAuth.Code, noAuth.Body.String())
	}

	// Register two owners
	register := func(email string) string {
		reg := doJSONRequest(t, s.E, http.MethodPost, "/api/auth/register", map[string]any{
			"name":     "Owner " + email,
			"email":    email,
			"mobile":   "555-9999",
			"password": "password123",
		}, nil, nil)
		if reg.Code != http.StatusCreated {
			t.Fatalf("expected register 201 for %s, got %d: %s", email, reg.Code, reg.Body.String())
		}
		var ar authResp
		if err := json.Unmarshal(reg.Body.Bytes(), &ar); err != nil {
			t.Fatalf("failed to unmarshal register response: %v", err)
		}
		if ar.AccessToken == "" {
			t.Fatalf("expected access token for %s", email)
		}
		return ar.AccessToken
	}

	tokenA := register("ownerA@example.com")
	tokenB := register("ownerB@example.com")

	authzA := map[string]string{"Authorization": "Bearer " + tokenA}
	authzB := map[string]string{"Authorization": "Bearer " + tokenB}

	// Owner A creates a venue
	createVenue := doJSONRequest(t, s.E, http.MethodPost, "/api/venues", map[string]any{
		"name":               "A Venue",
		"banner_image":       "",
		"address":            "",
		"geolocation":        "",
		"comment":            "",
		"timezone":           "UTC",
		"visibility":         "public",
		"private_link_token": "",
	}, authzA, nil)
	if createVenue.Code != http.StatusCreated {
		t.Fatalf("expected create venue 201, got %d: %s", createVenue.Code, createVenue.Body.String())
	}
	var vr venueResp
	if err := json.Unmarshal(createVenue.Body.Bytes(), &vr); err != nil {
		t.Fatalf("failed to unmarshal create venue response: %v", err)
	}

	// Owner B cannot read Owner A's venue -> 404 (do not leak existence)
	getAsB := doJSONRequest(t, s.E, http.MethodGet, "/api/venues/"+vr.VenueUUID, nil, authzB, nil)
	if getAsB.Code != http.StatusNotFound {
		t.Fatalf("expected cross-owner get 404, got %d: %s", getAsB.Code, getAsB.Body.String())
	}

	// Owner B cannot delete Owner A's venue -> 404
	delAsB := doJSONRequest(t, s.E, http.MethodDelete, "/api/venues/"+vr.VenueUUID, nil, authzB, nil)
	if delAsB.Code != http.StatusNotFound {
		t.Fatalf("expected cross-owner delete 404, got %d: %s", delAsB.Code, delAsB.Body.String())
	}

	// Owner A can still read it
	getAsA := doJSONRequest(t, s.E, http.MethodGet, "/api/venues/"+vr.VenueUUID, nil, authzA, nil)
	if getAsA.Code != http.StatusOK {
		t.Fatalf("expected owner get 200, got %d: %s", getAsA.Code, getAsA.Body.String())
	}
}

