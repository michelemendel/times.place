package http

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/michelemendel/times.place/internal/service"
	"github.com/michelemendel/times.place/internal/store"
	"github.com/michelemendel/times.place/utils"
)

// ContextKeyOwnerUUID is the key for storing owner UUID in Echo context
const ContextKeyOwnerUUID = "owner_uuid"

// JWTAuthMiddleware creates middleware that validates JWT tokens and extracts owner UUID
func JWTAuthMiddleware(authService *service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return UnauthorizedError(c, "Missing authorization header")
			}

			// Check for Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return UnauthorizedError(c, "Invalid authorization header format")
			}

			tokenString := parts[1]
			if tokenString == "" {
				return UnauthorizedError(c, "Missing token")
			}

			// Parse and validate token
			ownerUUID, err := authService.ParseAccessToken(tokenString)
			if err != nil {
				return UnauthorizedError(c, "Invalid or expired token")
			}

			// Store owner UUID in context
			c.Set(ContextKeyOwnerUUID, ownerUUID)

			return next(c)
		}
	}
}

// AdminOnlyMiddleware ensures the authenticated user is an admin
func AdminOnlyMiddleware(store *store.Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ownerUUIDStr, err := GetOwnerUUIDFromContext(c)
			if err != nil {
				return err
			}

			// 2. Convert to pgtype.UUID
			ownerUUID, err := utils.StringToUUID(ownerUUIDStr)
			if err != nil {
				return UnauthorizedError(c, "Invalid owner UUID")
			}

			// 3. Fetch owner from DB
			owner, err := store.Queries.GetOwnerByID(c.Request().Context(), ownerUUID)
			if err != nil {
				// If owner not found (deleted?) implies unauthorized
				return UnauthorizedError(c, "User not found")
			}

			// 4. Check is_admin flag
			if !owner.IsAdmin {
				return ForbiddenError(c, "Admin access required")
			}

			return next(c)
		}
	}
}

// GetOwnerUUIDFromContext extracts owner UUID from Echo context
func GetOwnerUUIDFromContext(c echo.Context) (string, error) {
	ownerUUID, ok := c.Get(ContextKeyOwnerUUID).(string)
	if !ok || ownerUUID == "" {
		return "", echo.NewHTTPError(401, "Owner UUID not found in context")
	}
	return ownerUUID, nil
}
