package http

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/michelemendel/times.place/internal/service"
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

// GetOwnerUUIDFromContext extracts owner UUID from Echo context
func GetOwnerUUIDFromContext(c echo.Context) (string, error) {
	ownerUUID, ok := c.Get(ContextKeyOwnerUUID).(string)
	if !ok || ownerUUID == "" {
		return "", echo.NewHTTPError(401, "Owner UUID not found in context")
	}
	return ownerUUID, nil
}
