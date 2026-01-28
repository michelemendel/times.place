package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/michelemendel/times.place/internal/service"
)

// Helper to create AuthHandler for testing (with nil store since we're only testing helpers)
func createTestAuthHandler(t *testing.T) *AuthHandler {
	t.Helper()
	// Save original values
	oldJWTSecret := os.Getenv("JWT_SECRET")
	defer func() {
		if oldJWTSecret != "" {
			os.Setenv("JWT_SECRET", oldJWTSecret)
		} else {
			os.Unsetenv("JWT_SECRET")
		}
	}()
	
	os.Setenv("JWT_SECRET", "test-secret")
	authService, err := service.NewAuthService()
	if err != nil {
		t.Fatalf("Failed to create AuthService: %v", err)
	}
	
	// Create handler with nil store (not used in helper tests)
	return NewAuthHandler(nil, authService)
}

// Helper to restore environment variables after test
func restoreEnvVars(t *testing.T, vars map[string]string) {
	t.Helper()
	t.Cleanup(func() {
		for key, value := range vars {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	})
}

// Test Cookie Setting

func TestSetRefreshTokenCookie_DefaultSettings(t *testing.T) {
	handler := createTestAuthHandler(t)
	e := echo.New()
	
	// Clear cookie env vars
	oldCookieDomain := os.Getenv("COOKIE_DOMAIN")
	oldCookieSecure := os.Getenv("COOKIE_SECURE")
	oldCookieSameSite := os.Getenv("COOKIE_SAME_SITE")
	defer func() {
		if oldCookieDomain != "" {
			os.Setenv("COOKIE_DOMAIN", oldCookieDomain)
		} else {
			os.Unsetenv("COOKIE_DOMAIN")
		}
		if oldCookieSecure != "" {
			os.Setenv("COOKIE_SECURE", oldCookieSecure)
		} else {
			os.Unsetenv("COOKIE_SECURE")
		}
		if oldCookieSameSite != "" {
			os.Setenv("COOKIE_SAME_SITE", oldCookieSameSite)
		} else {
			os.Unsetenv("COOKIE_SAME_SITE")
		}
	}()
	
	os.Unsetenv("COOKIE_DOMAIN")
	os.Unsetenv("COOKIE_SECURE")
	os.Unsetenv("COOKIE_SAME_SITE")
	
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	token := "test-refresh-token"
	handler.setRefreshTokenCookie(c, token)
	
	// Check cookie was set
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("No cookies set")
	}
	
	cookie := cookies[0]
	if cookie.Name != "refresh_token" {
		t.Errorf("Expected cookie name 'refresh_token', got '%s'", cookie.Name)
	}
	if cookie.Value != token {
		t.Errorf("Expected cookie value '%s', got '%s'", token, cookie.Value)
	}
	if cookie.Path != "/" {
		t.Errorf("Expected cookie path '/', got '%s'", cookie.Path)
	}
	if !cookie.HttpOnly {
		t.Error("Expected cookie to be HttpOnly")
	}
	if cookie.Secure {
		t.Error("Expected cookie Secure to be false by default")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("Expected cookie SameSite to be Lax, got %v", cookie.SameSite)
	}
	
	// Check MaxAge matches RefreshTokenLifetime (30 days)
	expectedMaxAge := int(service.RefreshTokenLifetime.Seconds())
	if cookie.MaxAge != expectedMaxAge {
		t.Errorf("Expected cookie MaxAge %d, got %d", expectedMaxAge, cookie.MaxAge)
	}
}

func TestSetRefreshTokenCookie_CookieSecure(t *testing.T) {
	handler := createTestAuthHandler(t)
	e := echo.New()
	
	oldCookieSecure := os.Getenv("COOKIE_SECURE")
	defer func() {
		if oldCookieSecure != "" {
			os.Setenv("COOKIE_SECURE", oldCookieSecure)
		} else {
			os.Unsetenv("COOKIE_SECURE")
		}
	}()
	
	os.Setenv("COOKIE_SECURE", "true")
	
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	token := "test-refresh-token"
	handler.setRefreshTokenCookie(c, token)
	
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("No cookies set")
	}
	
	cookie := cookies[0]
	if !cookie.Secure {
		t.Error("Expected cookie Secure to be true when COOKIE_SECURE=true")
	}
}

func TestSetRefreshTokenCookie_CookieDomain(t *testing.T) {
	handler := createTestAuthHandler(t)
	e := echo.New()
	
	oldCookieDomain := os.Getenv("COOKIE_DOMAIN")
	defer func() {
		if oldCookieDomain != "" {
			os.Setenv("COOKIE_DOMAIN", oldCookieDomain)
		} else {
			os.Unsetenv("COOKIE_DOMAIN")
		}
	}()
	
	testDomain := ".example.com"
	os.Setenv("COOKIE_DOMAIN", testDomain)
	
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	token := "test-refresh-token"
	handler.setRefreshTokenCookie(c, token)
	
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("No cookies set")
	}
	
	cookie := cookies[0]
	// Cookie domain should be set (Echo may normalize it, so check it's not empty)
	if cookie.Domain == "" {
		t.Error("Expected cookie domain to be set, got empty string")
	}
	// The domain should match what we set (Echo preserves the value as-is)
	if cookie.Domain != testDomain {
		// Some HTTP libraries normalize domain (remove leading dot), so accept either
		expectedAlt := "example.com"
		if cookie.Domain != expectedAlt {
			t.Errorf("Expected cookie domain '%s' or '%s', got '%s'", testDomain, expectedAlt, cookie.Domain)
		}
	}
}

func TestSetRefreshTokenCookie_SameSiteLax(t *testing.T) {
	handler := createTestAuthHandler(t)
	e := echo.New()
	
	oldCookieSameSite := os.Getenv("COOKIE_SAME_SITE")
	defer func() {
		if oldCookieSameSite != "" {
			os.Setenv("COOKIE_SAME_SITE", oldCookieSameSite)
		} else {
			os.Unsetenv("COOKIE_SAME_SITE")
		}
	}()
	
	// Test default (no env var)
	os.Unsetenv("COOKIE_SAME_SITE")
	
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	token := "test-refresh-token"
	handler.setRefreshTokenCookie(c, token)
	
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("No cookies set")
	}
	
	cookie := cookies[0]
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("Expected cookie SameSite to be Lax (default), got %v", cookie.SameSite)
	}
}

func TestSetRefreshTokenCookie_SameSiteStrict(t *testing.T) {
	handler := createTestAuthHandler(t)
	e := echo.New()
	
	oldCookieSameSite := os.Getenv("COOKIE_SAME_SITE")
	defer func() {
		if oldCookieSameSite != "" {
			os.Setenv("COOKIE_SAME_SITE", oldCookieSameSite)
		} else {
			os.Unsetenv("COOKIE_SAME_SITE")
		}
	}()
	
	os.Setenv("COOKIE_SAME_SITE", "strict")
	
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	token := "test-refresh-token"
	handler.setRefreshTokenCookie(c, token)
	
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("No cookies set")
	}
	
	cookie := cookies[0]
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("Expected cookie SameSite to be Strict, got %v", cookie.SameSite)
	}
}

func TestSetRefreshTokenCookie_SameSiteNone(t *testing.T) {
	handler := createTestAuthHandler(t)
	e := echo.New()
	
	oldCookieSameSite := os.Getenv("COOKIE_SAME_SITE")
	defer func() {
		if oldCookieSameSite != "" {
			os.Setenv("COOKIE_SAME_SITE", oldCookieSameSite)
		} else {
			os.Unsetenv("COOKIE_SAME_SITE")
		}
	}()
	
	os.Setenv("COOKIE_SAME_SITE", "none")
	
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	token := "test-refresh-token"
	handler.setRefreshTokenCookie(c, token)
	
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("No cookies set")
	}
	
	cookie := cookies[0]
	if cookie.SameSite != http.SameSiteNoneMode {
		t.Errorf("Expected cookie SameSite to be None, got %v", cookie.SameSite)
	}
}

func TestSetRefreshTokenCookie_SameSiteCaseInsensitive(t *testing.T) {
	handler := createTestAuthHandler(t)
	e := echo.New()
	
	oldCookieSameSite := os.Getenv("COOKIE_SAME_SITE")
	defer func() {
		if oldCookieSameSite != "" {
			os.Setenv("COOKIE_SAME_SITE", oldCookieSameSite)
		} else {
			os.Unsetenv("COOKIE_SAME_SITE")
		}
	}()
	
	// Test uppercase
	os.Setenv("COOKIE_SAME_SITE", "STRICT")
	
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	token := "test-refresh-token"
	handler.setRefreshTokenCookie(c, token)
	
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("No cookies set")
	}
	
	cookie := cookies[0]
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("Expected cookie SameSite to be Strict (case insensitive), got %v", cookie.SameSite)
	}
}

func TestSetRefreshTokenCookie_MaxAge(t *testing.T) {
	handler := createTestAuthHandler(t)
	e := echo.New()
	
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	token := "test-refresh-token"
	handler.setRefreshTokenCookie(c, token)
	
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("No cookies set")
	}
	
	cookie := cookies[0]
	expectedMaxAge := int(service.RefreshTokenLifetime.Seconds())
	if cookie.MaxAge != expectedMaxAge {
		t.Errorf("Expected cookie MaxAge %d (30 days), got %d", expectedMaxAge, cookie.MaxAge)
	}
}

// Test Cookie Clearing

func TestClearRefreshTokenCookie(t *testing.T) {
	handler := createTestAuthHandler(t)
	e := echo.New()
	
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	handler.clearRefreshTokenCookie(c)
	
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("No cookies set")
	}
	
	cookie := cookies[0]
	if cookie.Name != "refresh_token" {
		t.Errorf("Expected cookie name 'refresh_token', got '%s'", cookie.Name)
	}
	if cookie.Value != "" {
		t.Errorf("Expected empty cookie value, got '%s'", cookie.Value)
	}
	if cookie.MaxAge != -1 {
		t.Errorf("Expected cookie MaxAge -1, got %d", cookie.MaxAge)
	}
	if !cookie.HttpOnly {
		t.Error("Expected cookie to be HttpOnly")
	}
}

// Test Refresh Token Extraction

func TestGetRefreshTokenFromRequest_CookieFirst(t *testing.T) {
	handler := createTestAuthHandler(t)
	e := echo.New()
	
	token := "cookie-token-value"
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: token,
	})
	
	// Also set body with different token (should be ignored)
	body := RefreshRequest{RefreshToken: "body-token-value"}
	bodyBytes, _ := json.Marshal(body)
	req.Body = &testBody{bytes.NewReader(bodyBytes)}
	req.Header.Set("Content-Type", "application/json")
	
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	extracted := handler.getRefreshTokenFromRequest(c)
	if extracted != token {
		t.Errorf("Expected token from cookie '%s', got '%s'", token, extracted)
	}
}

func TestGetRefreshTokenFromRequest_BodyFallback(t *testing.T) {
	handler := createTestAuthHandler(t)
	e := echo.New()
	
	token := "body-token-value"
	body := RefreshRequest{RefreshToken: token}
	bodyBytes, _ := json.Marshal(body)
	
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	// No cookie
	
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	extracted := handler.getRefreshTokenFromRequest(c)
	if extracted != token {
		t.Errorf("Expected token from body '%s', got '%s'", token, extracted)
	}
}

func TestGetRefreshTokenFromRequest_EmptyCookie(t *testing.T) {
	handler := createTestAuthHandler(t)
	e := echo.New()
	
	token := "body-token-value"
	body := RefreshRequest{RefreshToken: token}
	bodyBytes, _ := json.Marshal(body)
	
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	// Cookie exists but is empty
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "",
	})
	
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	extracted := handler.getRefreshTokenFromRequest(c)
	if extracted != token {
		t.Errorf("Expected token from body when cookie is empty '%s', got '%s'", token, extracted)
	}
}

func TestGetRefreshTokenFromRequest_Neither(t *testing.T) {
	handler := createTestAuthHandler(t)
	e := echo.New()
	
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	// No cookie, no body
	
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	extracted := handler.getRefreshTokenFromRequest(c)
	if extracted != "" {
		t.Errorf("Expected empty string when neither cookie nor body present, got '%s'", extracted)
	}
}

func TestGetRefreshTokenFromRequest_InvalidBody(t *testing.T) {
	handler := createTestAuthHandler(t)
	e := echo.New()
	
	// Invalid JSON body
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	// No cookie
	
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	extracted := handler.getRefreshTokenFromRequest(c)
	if extracted != "" {
		t.Errorf("Expected empty string when body is invalid JSON, got '%s'", extracted)
	}
}

func TestGetRefreshTokenFromRequest_EmptyBodyToken(t *testing.T) {
	handler := createTestAuthHandler(t)
	e := echo.New()
	
	// Valid JSON but empty refresh_token field
	body := RefreshRequest{RefreshToken: ""}
	bodyBytes, _ := json.Marshal(body)
	
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	// No cookie
	
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	extracted := handler.getRefreshTokenFromRequest(c)
	if extracted != "" {
		t.Errorf("Expected empty string when body token is empty, got '%s'", extracted)
	}
}

// Helper type for request body
type testBody struct {
	*bytes.Reader
}

func (tb *testBody) Close() error {
	return nil
}
