package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/michelemendel/times.place/internal/service"
)

// Helper to create AuthService for testing
func createTestAuthServiceForMiddleware(t *testing.T) *service.AuthService {
	t.Helper()
	// Save original value
	oldValue := os.Getenv("JWT_SECRET")
	defer func() {
		if oldValue != "" {
			os.Setenv("JWT_SECRET", oldValue)
		} else {
			os.Unsetenv("JWT_SECRET")
		}
	}()
	
	os.Setenv("JWT_SECRET", "test-jwt-secret-for-middleware")
	authService, err := service.NewAuthService()
	if err != nil {
		t.Fatalf("Failed to create AuthService: %v", err)
	}
	return authService
}

// Test JWTAuthMiddleware

func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	authService := createTestAuthServiceForMiddleware(t)
	e := echo.New()
	
	ownerUUID := uuid.New().String()
	token, err := authService.GenerateAccessToken(ownerUUID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	
	// Create a handler that checks if owner UUID is in context
	handler := func(c echo.Context) error {
		uuidFromContext, err := GetOwnerUUIDFromContext(c)
		if err != nil {
			return err
		}
		if uuidFromContext != ownerUUID {
			return echo.NewHTTPError(http.StatusInternalServerError, "UUID mismatch")
		}
		return c.String(http.StatusOK, "success")
	}
	
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	middleware := JWTAuthMiddleware(authService)
	err = middleware(handler)(c)
	
	if err != nil {
		t.Fatalf("Middleware should allow valid token: %v", err)
	}
	
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestJWTAuthMiddleware_MissingHeader(t *testing.T) {
	authService := createTestAuthServiceForMiddleware(t)
	e := echo.New()
	
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}
	
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No Authorization header
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	middleware := JWTAuthMiddleware(authService)
	_ = middleware(handler)(c)
	
	// UnauthorizedError writes the response, so check the response code
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestJWTAuthMiddleware_InvalidFormat(t *testing.T) {
	authService := createTestAuthServiceForMiddleware(t)
	e := echo.New()
	
	testCases := []struct {
		name  string
		header string
	}{
		{"NoBearer", "token123"},
		{"WrongPrefix", "Basic token123"},
		{"MultipleSpaces", "Bearer  token123"},
		{"NoSpace", "Bearertoken123"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "success")
			}
			
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", tc.header)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			
			middleware := JWTAuthMiddleware(authService)
			_ = middleware(handler)(c)
			
			// UnauthorizedError writes the response, so check the response code
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("Expected status 401, got %d", rec.Code)
			}
		})
	}
}

func TestJWTAuthMiddleware_EmptyToken(t *testing.T) {
	authService := createTestAuthServiceForMiddleware(t)
	e := echo.New()
	
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}
	
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer ")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	middleware := JWTAuthMiddleware(authService)
	_ = middleware(handler)(c)
	
	// UnauthorizedError writes the response, so check the response code
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestJWTAuthMiddleware_InvalidToken(t *testing.T) {
	authService := createTestAuthServiceForMiddleware(t)
	e := echo.New()
	
	testCases := []struct {
		name  string
		token string
	}{
		{"InvalidJWT", "invalid.jwt.token"},
		{"Malformed", "not-a-jwt"},
		{"EmptyString", ""},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "success")
			}
			
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			
			middleware := JWTAuthMiddleware(authService)
			_ = middleware(handler)(c)
			
			// UnauthorizedError writes the response, so check the response code
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("Expected status 401, got %d", rec.Code)
			}
		})
	}
}

func TestJWTAuthMiddleware_WrongSecret(t *testing.T) {
	// Create service with one secret
	authService1 := createTestAuthServiceForMiddleware(t)
	
	// Create token with service1
	ownerUUID := uuid.New().String()
	token, err := authService1.GenerateAccessToken(ownerUUID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	
	// Create another service with different secret
	oldValue := os.Getenv("JWT_SECRET")
	defer func() {
		if oldValue != "" {
			os.Setenv("JWT_SECRET", oldValue)
		} else {
			os.Unsetenv("JWT_SECRET")
		}
	}()
	
	os.Setenv("JWT_SECRET", "different-secret")
	authService2, err := service.NewAuthService()
	if err != nil {
		t.Fatalf("Failed to create second AuthService: %v", err)
	}
	
	e := echo.New()
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}
	
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Use service2 (different secret) to validate token from service1
	middleware := JWTAuthMiddleware(authService2)
	_ = middleware(handler)(c)
	
	// UnauthorizedError writes the response, so check the response code
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestJWTAuthMiddleware_ContextPropagation(t *testing.T) {
	authService := createTestAuthServiceForMiddleware(t)
	e := echo.New()
	
	ownerUUID := uuid.New().String()
	token, err := authService.GenerateAccessToken(ownerUUID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	
	// Handler that extracts UUID from context
	var extractedUUID string
	handler := func(c echo.Context) error {
		var err error
		extractedUUID, err = GetOwnerUUIDFromContext(c)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, "success")
	}
	
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	middleware := JWTAuthMiddleware(authService)
	err = middleware(handler)(c)
	
	if err != nil {
		t.Fatalf("Middleware should allow valid token: %v", err)
	}
	
	if extractedUUID != ownerUUID {
		t.Errorf("Expected UUID %s in context, got %s", ownerUUID, extractedUUID)
	}
}

// Test GetOwnerUUIDFromContext

func TestGetOwnerUUIDFromContext_Valid(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	ownerUUID := uuid.New().String()
	c.Set(ContextKeyOwnerUUID, ownerUUID)
	
	extractedUUID, err := GetOwnerUUIDFromContext(c)
	if err != nil {
		t.Fatalf("GetOwnerUUIDFromContext failed: %v", err)
	}
	
	if extractedUUID != ownerUUID {
		t.Errorf("Expected UUID %s, got %s", ownerUUID, extractedUUID)
	}
}

func TestGetOwnerUUIDFromContext_Missing(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Don't set owner UUID in context
	
	_, err := GetOwnerUUIDFromContext(c)
	if err == nil {
		t.Error("GetOwnerUUIDFromContext should fail when UUID not in context")
	}
	
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("Expected HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", httpErr.Code)
	}
}

func TestGetOwnerUUIDFromContext_Empty(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Set empty string
	c.Set(ContextKeyOwnerUUID, "")
	
	_, err := GetOwnerUUIDFromContext(c)
	if err == nil {
		t.Error("GetOwnerUUIDFromContext should fail when UUID is empty string")
	}
	
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("Expected HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", httpErr.Code)
	}
}

func TestGetOwnerUUIDFromContext_WrongType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	// Set wrong type
	c.Set(ContextKeyOwnerUUID, 123)
	
	_, err := GetOwnerUUIDFromContext(c)
	if err == nil {
		t.Error("GetOwnerUUIDFromContext should fail when value is wrong type")
	}
	
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("Expected HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", httpErr.Code)
	}
}
