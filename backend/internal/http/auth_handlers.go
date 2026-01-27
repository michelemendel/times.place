package http

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	sqlc "github.com/michelemendel/times.place/db/sqlc"
	"github.com/michelemendel/times.place/internal/service"
	"github.com/michelemendel/times.place/internal/store"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	store      *store.Store
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(store *store.Store, authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		store:       store,
		authService: authService,
	}
}

// Request/Response types

type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Mobile   string `json:"mobile" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"` // Optional fallback if not in cookie
}

type AuthResponse struct {
	Owner       OwnerResponse `json:"owner"`
	AccessToken string       `json:"access_token"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

type OwnerResponse struct {
	OwnerUUID  string `json:"owner_uuid"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Mobile     string `json:"mobile"`
	CreatedAt  string `json:"created_at"`
	ModifiedAt string `json:"modified_at"`
}

// Helper functions

// uuidToString converts pgtype.UUID to string
func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	// Convert [16]byte to uuid.UUID, then to string
	uuidVal, err := uuid.FromBytes(u.Bytes[:])
	if err != nil {
		return ""
	}
	return uuidVal.String()
}

// stringToUUID converts string UUID to pgtype.UUID
func stringToUUID(s string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	var result pgtype.UUID
	// Convert uuid.UUID to [16]byte
	var bytes [16]byte
	copy(bytes[:], parsed[:])
	result.Bytes = bytes
	result.Valid = true
	return result, nil
}

// timestamptzToString converts pgtype.Timestamptz to RFC3339 string
func timestamptzToString(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format(time.RFC3339)
}

// ownerToResponse converts sqlc VenueOwner to OwnerResponse
func ownerToResponse(owner sqlc.VenueOwner) OwnerResponse {
	return OwnerResponse{
		OwnerUUID:  uuidToString(owner.OwnerUuid),
		Name:       owner.Name,
		Email:      owner.Email,
		Mobile:     owner.Mobile,
		CreatedAt:  timestamptzToString(owner.CreatedAt),
		ModifiedAt: timestamptzToString(owner.ModifiedAt),
	}
}

// setRefreshTokenCookie sets the refresh token as an HttpOnly cookie
func (h *AuthHandler) setRefreshTokenCookie(c echo.Context, token string) {
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	cookieSecure := os.Getenv("COOKIE_SECURE") == "true"
	cookieSameSite := http.SameSiteLaxMode
	
	sameSiteStr := os.Getenv("COOKIE_SAME_SITE")
	switch strings.ToLower(sameSiteStr) {
	case "strict":
		cookieSameSite = http.SameSiteStrictMode
	case "none":
		cookieSameSite = http.SameSiteNoneMode
	default:
		cookieSameSite = http.SameSiteLaxMode
	}

	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		MaxAge:   int(service.RefreshTokenLifetime.Seconds()),
		HttpOnly: true,
		Secure:   cookieSecure,
		SameSite: cookieSameSite,
	}
	
	if cookieDomain != "" {
		cookie.Domain = cookieDomain
	}
	
	c.SetCookie(cookie)
}

// clearRefreshTokenCookie clears the refresh token cookie
func (h *AuthHandler) clearRefreshTokenCookie(c echo.Context) {
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	c.SetCookie(cookie)
}

// getRefreshTokenFromRequest extracts refresh token from cookie or request body
func (h *AuthHandler) getRefreshTokenFromRequest(c echo.Context) string {
	// Try cookie first
	cookie, err := c.Cookie("refresh_token")
	if err == nil && cookie != nil && cookie.Value != "" {
		return cookie.Value
	}
	
	// Fallback to request body
	var req RefreshRequest
	if err := c.Bind(&req); err == nil && req.RefreshToken != "" {
		return req.RefreshToken
	}
	
	return ""
}

// Handlers

// Register handles POST /api/auth/register
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return ValidationError(c, "Invalid request body")
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return ValidationError(c, err.Error())
	}

	ctx := c.Request().Context()

	// Check if email already exists
	_, err := h.store.Queries.GetOwnerByEmail(ctx, req.Email)
	if err == nil {
		// Owner exists
		return ConflictError(c, "Email already registered")
	}

	// Hash password
	passwordHash, err := h.authService.HashPassword(req.Password)
	if err != nil {
		return InternalError(c, "Failed to process password")
	}

	// Create owner
	owner, err := h.store.Queries.CreateOwner(ctx, sqlc.CreateOwnerParams{
		Name:         req.Name,
		Email:        req.Email,
		Mobile:       req.Mobile,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return InternalError(c, "Failed to create account")
	}

	// Generate tokens
	accessToken, err := h.authService.GenerateAccessToken(uuidToString(owner.OwnerUuid))
	if err != nil {
		return InternalError(c, "Failed to generate access token")
	}

	refreshToken, err := h.authService.GenerateRefreshToken()
	if err != nil {
		return InternalError(c, "Failed to generate refresh token")
	}

	// Store refresh token
	tokenHash := h.authService.HashRefreshToken(refreshToken)
	expiresAt := time.Now().Add(service.RefreshTokenLifetime)
	
	ownerUUID, _ := stringToUUID(uuidToString(owner.OwnerUuid))
	_, err = h.store.Queries.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		OwnerUuid: ownerUUID,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{
			Time:  expiresAt,
			Valid: true,
		},
		UserAgent: pgtype.Text{
			String: c.Request().UserAgent(),
			Valid:  true,
		},
		IpAddress: pgtype.Text{
			String: c.RealIP(),
			Valid:  true,
		},
	})
	if err != nil {
		return InternalError(c, "Failed to store refresh token")
	}

	// Set refresh token cookie
	h.setRefreshTokenCookie(c, refreshToken)

	// Return response
	return c.JSON(http.StatusCreated, AuthResponse{
		Owner:       ownerToResponse(owner),
		AccessToken: accessToken,
	})
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return ValidationError(c, "Invalid request body")
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return ValidationError(c, err.Error())
	}

	ctx := c.Request().Context()

	// Lookup owner by email
	owner, err := h.store.Queries.GetOwnerByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if email exists or not
		return UnauthorizedError(c, "Invalid email or password")
	}

	// Verify password
	if err := h.authService.VerifyPassword(owner.PasswordHash, req.Password); err != nil {
		return UnauthorizedError(c, "Invalid email or password")
	}

	// Generate tokens
	accessToken, err := h.authService.GenerateAccessToken(uuidToString(owner.OwnerUuid))
	if err != nil {
		return InternalError(c, "Failed to generate access token")
	}

	refreshToken, err := h.authService.GenerateRefreshToken()
	if err != nil {
		return InternalError(c, "Failed to generate refresh token")
	}

	// Store refresh token
	tokenHash := h.authService.HashRefreshToken(refreshToken)
	expiresAt := time.Now().Add(service.RefreshTokenLifetime)
	
	ownerUUID, _ := stringToUUID(uuidToString(owner.OwnerUuid))
	_, err = h.store.Queries.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		OwnerUuid: ownerUUID,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{
			Time:  expiresAt,
			Valid: true,
		},
		UserAgent: pgtype.Text{
			String: c.Request().UserAgent(),
			Valid:  true,
		},
		IpAddress: pgtype.Text{
			String: c.RealIP(),
			Valid:  true,
		},
	})
	if err != nil {
		return InternalError(c, "Failed to store refresh token")
	}

	// Set refresh token cookie
	h.setRefreshTokenCookie(c, refreshToken)

	// Return response
	return c.JSON(http.StatusOK, AuthResponse{
		Owner:       ownerToResponse(owner),
		AccessToken: accessToken,
	})
}

// Refresh handles POST /api/auth/refresh
func (h *AuthHandler) Refresh(c echo.Context) error {
	// Get refresh token from cookie or request body
	refreshToken := h.getRefreshTokenFromRequest(c)
	if refreshToken == "" {
		return UnauthorizedError(c, "Refresh token required")
	}

	ctx := c.Request().Context()

	// Hash and lookup token
	tokenHash := h.authService.HashRefreshToken(refreshToken)
	tokenRecord, err := h.store.Queries.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return UnauthorizedError(c, "Invalid or expired refresh token")
	}

	// Generate new access token
	ownerUUID := uuidToString(tokenRecord.OwnerUuid)
	accessToken, err := h.authService.GenerateAccessToken(ownerUUID)
	if err != nil {
		return InternalError(c, "Failed to generate access token")
	}

	// Rotate refresh token: revoke old, create new
	newRefreshToken, err := h.authService.GenerateRefreshToken()
	if err != nil {
		return InternalError(c, "Failed to generate refresh token")
	}

	newTokenHash := h.authService.HashRefreshToken(newRefreshToken)
	expiresAt := time.Now().Add(service.RefreshTokenLifetime)

	// Create new token
	newTokenRecord, err := h.store.Queries.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		OwnerUuid: tokenRecord.OwnerUuid,
		TokenHash: newTokenHash,
		ExpiresAt: pgtype.Timestamptz{
			Time:  expiresAt,
			Valid: true,
		},
		UserAgent: pgtype.Text{
			String: c.Request().UserAgent(),
			Valid:  true,
		},
		IpAddress: pgtype.Text{
			String: c.RealIP(),
			Valid:  true,
		},
	})
	if err != nil {
		return InternalError(c, "Failed to store refresh token")
	}

	// Mark old token as replaced
	err = h.store.Queries.RotateRefreshToken(ctx, sqlc.RotateRefreshTokenParams{
		RefreshTokenUuid:        tokenRecord.RefreshTokenUuid,
		ReplacedByTokenUuid:     newTokenRecord.RefreshTokenUuid,
	})
	if err != nil {
		// Log error but continue (token rotation is best-effort)
		// The old token will expire naturally
	}

	// Revoke old token
	err = h.store.Queries.RevokeRefreshToken(ctx, tokenRecord.RefreshTokenUuid)
	if err != nil {
		// Log error but continue
	}

	// Set new refresh token cookie
	h.setRefreshTokenCookie(c, newRefreshToken)

	// Return response
	return c.JSON(http.StatusOK, RefreshResponse{
		AccessToken: accessToken,
	})
}

// Logout handles POST /api/auth/logout
func (h *AuthHandler) Logout(c echo.Context) error {
	// Get refresh token from cookie or request body
	refreshToken := h.getRefreshTokenFromRequest(c)
	if refreshToken != "" {
		ctx := c.Request().Context()
		tokenHash := h.authService.HashRefreshToken(refreshToken)
		
		// Revoke token
		err := h.store.Queries.RevokeRefreshTokenByHash(ctx, tokenHash)
		if err != nil {
			// Log error but continue (token might not exist)
		}
	}

	// Clear refresh token cookie
	h.clearRefreshTokenCookie(c)

	return c.NoContent(http.StatusNoContent)
}

// Me handles GET /api/auth/me
func (h *AuthHandler) Me(c echo.Context) error {
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

	// Lookup owner
	owner, err := h.store.Queries.GetOwnerByID(ctx, ownerUUID)
	if err != nil {
		return NotFoundError(c, "Owner not found")
	}

	// Return response
	return c.JSON(http.StatusOK, map[string]OwnerResponse{
		"owner": ownerToResponse(owner),
	})
}
