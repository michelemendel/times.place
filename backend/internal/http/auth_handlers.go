package http

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	sqlc "github.com/michelemendel/times.place/db/sqlc"
	"github.com/michelemendel/times.place/internal/mailer"
	"github.com/michelemendel/times.place/internal/service"
	"github.com/michelemendel/times.place/internal/store"
	"github.com/michelemendel/times.place/utils"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	store       *store.Store
	authService *service.AuthService
	mailer      mailer.Sender
}

// NewAuthHandler creates a new auth handler. mailer may be nil (verification emails will be skipped).
func NewAuthHandler(store *store.Store, authService *service.AuthService, m mailer.Sender) *AuthHandler {
	return &AuthHandler{
		store:       store,
		authService: authService,
		mailer:      m,
	}
}

// Request/Response types

type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Mobile   string `json:"mobile"` // optional
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"` // Optional fallback if not in cookie
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type AuthResponse struct {
	Owner       OwnerResponse `json:"owner"`
	AccessToken string        `json:"access_token"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

type OwnerResponse struct {
	OwnerUUID     string `json:"owner_uuid"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Mobile        string `json:"mobile"`
	EmailVerified bool   `json:"email_verified"`
	IsAdmin       bool   `json:"is_admin"`
	VenueLimit    int32  `json:"venue_limit"`
	CreatedAt     string `json:"created_at"`
	ModifiedAt    string `json:"modified_at"`
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
		OwnerUUID:     uuidToString(owner.OwnerUuid),
		Name:          owner.Name,
		Email:         owner.Email,
		Mobile:        owner.Mobile,
		EmailVerified: owner.EmailVerifiedAt.Valid,
		IsAdmin:       owner.IsAdmin,
		VenueLimit:    owner.VenueLimit,
		CreatedAt:     timestamptzToString(owner.CreatedAt),
		ModifiedAt:    timestamptzToString(owner.ModifiedAt),
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
		VenueLimit:   int32(FreeTierMaxVenues()),
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

	// Create email verification token and send email (best-effort; do not fail registration)
	rawToken, err := utils.GenerateToken()
	if err == nil {
		tokenHash := h.authService.HashRefreshToken(rawToken)
		expiresAt := time.Now().Add(24 * time.Hour)
		_, err = h.store.Queries.CreateEmailVerificationToken(ctx, sqlc.CreateEmailVerificationTokenParams{
			OwnerUuid: ownerUUID,
			TokenHash: tokenHash,
			ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
		})
		if err == nil && h.mailer != nil {
			baseURL := os.Getenv("VERIFICATION_BASE_URL")
			if baseURL == "" {
				log.Printf("ERROR: VERIFICATION_BASE_URL is not set")
			}
			link := baseURL + "/verify-email?token=" + url.QueryEscape(rawToken)
			if sendErr := h.mailer.SendVerificationEmail(owner.Email, link); sendErr != nil {
				log.Printf("failed to send verification email: %v", sendErr)
			}
		}
	}

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
		RefreshTokenUuid:    tokenRecord.RefreshTokenUuid,
		ReplacedByTokenUuid: newTokenRecord.RefreshTokenUuid,
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

	// Venue count and limit (for frontend upgrade prompt)
	venueCount, _ := h.store.Queries.CountVenuesByOwner(ctx, ownerUUID)
	venueLimit := int64(owner.VenueLimit)

	// Return response
	return c.JSON(http.StatusOK, map[string]any{
		"owner":       ownerToResponse(owner),
		"venue_count": venueCount,
		"venue_limit": venueLimit,
	})
}

// DeleteMe handles DELETE /api/auth/me
func (h *AuthHandler) DeleteMe(c echo.Context) error {
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
		return ForbiddenError(c, "Demo accounts cannot be deleted")
	}

	ownerUUID, err := stringToUUID(ownerUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	err = h.store.Queries.DeleteOwner(ctx, ownerUUID)
	if err != nil {
		return InternalError(c, "Failed to delete account")
	}

	h.clearRefreshTokenCookie(c)

	return c.NoContent(http.StatusNoContent)
}

// VerifyEmail handles GET /api/auth/verify-email?token=...
// Public endpoint: validates token, sets email_verified_at, deletes token.
func (h *AuthHandler) VerifyEmail(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return ValidationError(c, "token is required")
	}

	ctx := c.Request().Context()
	tokenHash := h.authService.HashRefreshToken(token)

	row, err := h.store.Queries.GetEmailVerificationTokenByHash(ctx, tokenHash)
	if err != nil {
		return ValidationError(c, "Invalid or expired verification link")
	}

	ownerUUID := row.OwnerUuid
	if err := h.store.Queries.SetOwnerEmailVerified(ctx, ownerUUID); err != nil {
		return InternalError(c, "Failed to verify email")
	}
	_ = h.store.Queries.DeleteEmailVerificationTokenByHash(ctx, tokenHash)

	return c.JSON(http.StatusOK, map[string]string{"message": "Email verified"})
}

// ResendVerification handles POST /api/auth/resend-verification (protected).
// Deletes existing tokens for owner, creates new token, sends email. Rate-limited by caller or middleware if needed.
func (h *AuthHandler) ResendVerification(c echo.Context) error {
	ownerUUIDStr, err := GetOwnerUUIDFromContext(c)
	if err != nil {
		return UnauthorizedError(c, "Unauthorized")
	}

	ctx := c.Request().Context()

	ownerUUID, err := stringToUUID(ownerUUIDStr)
	if err != nil {
		return ValidationError(c, "Invalid owner UUID")
	}

	owner, err := h.store.Queries.GetOwnerByID(ctx, ownerUUID)
	if err != nil {
		return NotFoundError(c, "Owner not found")
	}
	if owner.EmailVerifiedAt.Valid {
		return ValidationError(c, "Email is already verified")
	}

	// Rate limit: don't send another verification email within 60 seconds (improves deliverability)
	latest, err := h.store.Queries.GetLatestVerificationCreatedAtByOwner(ctx, ownerUUID)
	if err != nil && err != pgx.ErrNoRows {
		return InternalError(c, "Failed to check resend limit")
	}
	if err == nil && latest.Valid && time.Since(latest.Time) < 60*time.Second {
		return TooManyRequestsError(c, "Please wait a minute before requesting another verification email.")
	}

	// Delete any existing verification tokens for this owner
	_ = h.store.Queries.DeleteEmailVerificationTokensByOwner(ctx, ownerUUID)

	rawToken, err := utils.GenerateToken()
	if err != nil {
		return InternalError(c, "Failed to create verification token")
	}
	tokenHash := h.authService.HashRefreshToken(rawToken)
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = h.store.Queries.CreateEmailVerificationToken(ctx, sqlc.CreateEmailVerificationTokenParams{
		OwnerUuid: ownerUUID,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return InternalError(c, "Failed to create verification token")
	}

	if h.mailer != nil {
		baseURL := os.Getenv("VERIFICATION_BASE_URL")
		if baseURL == "" {
			log.Printf("ERROR: VERIFICATION_BASE_URL is not set")
		}
		link := baseURL + "/verify-email?token=" + url.QueryEscape(rawToken)
		if sendErr := h.mailer.SendVerificationEmail(owner.Email, link); sendErr != nil {
			log.Printf("resend verification email failed: %v", sendErr)
			return InternalError(c, "Failed to send verification email")
		}
	}

	return c.JSON(http.StatusAccepted, map[string]string{"message": "Verification email sent"})
}

// RequestPasswordReset handles POST /api/auth/forgot-password
func (h *AuthHandler) RequestPasswordReset(c echo.Context) error {
	var req ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return ValidationError(c, "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return ValidationError(c, err.Error())
	}

	ctx := c.Request().Context()
	owner, err := h.store.Queries.GetOwnerByEmail(ctx, req.Email)
	if err != nil {
		// Do not reveal if email exists
		return c.JSON(http.StatusAccepted, map[string]string{"message": "If that email is registered, you will receive a reset link shortly."})
	}

	// Generate reset token
	rawToken, err := utils.GenerateToken()
	if err != nil {
		return InternalError(c, "Failed to generate reset token")
	}
	tokenHash := h.authService.HashRefreshToken(rawToken) // Reuse hash helper
	expiresAt := time.Now().Add(1 * time.Hour)

	// Delete old tokens for this owner
	_ = h.store.Queries.DeletePasswordResetTokensByOwner(ctx, owner.OwnerUuid)

	_, err = h.store.Queries.CreatePasswordResetToken(ctx, sqlc.CreatePasswordResetTokenParams{
		TokenHash: tokenHash,
		OwnerUuid: owner.OwnerUuid,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return InternalError(c, "Failed to store reset token")
	}

	if h.mailer != nil {
		baseURL := os.Getenv("RESET_PASSWORD_BASE_URL")
		if baseURL == "" {
			log.Printf("ERROR: RESET_PASSWORD_BASE_URL is not set")
		}
		link := baseURL + "/reset-password?token=" + url.QueryEscape(rawToken)
		if sendErr := h.mailer.SendPasswordResetEmail(owner.Email, link); sendErr != nil {
			log.Printf("failed to send password reset email: %v", sendErr)
		}
	}

	return c.JSON(http.StatusAccepted, map[string]string{"message": "If that email is registered, you will receive a reset link shortly."})
}

// ResetPassword handles POST /api/auth/reset-password
func (h *AuthHandler) ResetPassword(c echo.Context) error {
	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return ValidationError(c, "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return ValidationError(c, err.Error())
	}

	ctx := c.Request().Context()
	tokenHash := h.authService.HashRefreshToken(req.Token)

	tokenRecord, err := h.store.Queries.GetPasswordResetTokenByHash(ctx, tokenHash)
	if err != nil {
		return ValidationError(c, "Invalid or expired reset link")
	}

	if time.Now().After(tokenRecord.ExpiresAt.Time) {
		_ = h.store.Queries.DeletePasswordResetTokenByHash(ctx, tokenHash)
		return ValidationError(c, "Invalid or expired reset link")
	}

	// Hash new password
	passwordHash, err := h.authService.HashPassword(req.Password)
	if err != nil {
		return InternalError(c, "Failed to process password")
	}

	// Update password
	err = h.store.Queries.UpdateOwnerPassword(ctx, sqlc.UpdateOwnerPasswordParams{
		PasswordHash: passwordHash,
		OwnerUuid:    tokenRecord.OwnerUuid,
	})
	if err != nil {
		return InternalError(c, "Failed to update password")
	}

	// Revoke token
	_ = h.store.Queries.DeletePasswordResetTokenByHash(ctx, tokenHash)

	// Optional: revoke all refresh tokens for this user? Probably a good idea.
	_ = h.store.Queries.RevokeAllTokensForOwner(ctx, tokenRecord.OwnerUuid)

	return c.JSON(http.StatusOK, map[string]string{"message": "Password updated successfully"})
}
