package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorResponse represents a consistent error response format
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error code and message
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error codes
const (
	ErrorCodeValidation       = "validation_error"
	ErrorCodeUnauthorized     = "unauthorized"
	ErrorCodeForbidden        = "forbidden"
	ErrorCodeEmailNotVerified = "email_not_verified"
	ErrorCodeTooManyRequests  = "too_many_requests"
	ErrorCodeNotFound         = "not_found"
	ErrorCodeConflict         = "conflict"
	ErrorCodeInternal         = "internal"
)

// ErrorResponseHelper creates an error response
func ErrorResponseHelper(c echo.Context, statusCode int, code, message string) error {
	return c.JSON(statusCode, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// ValidationError returns a 400 validation error
func ValidationError(c echo.Context, message string) error {
	return ErrorResponseHelper(c, http.StatusBadRequest, ErrorCodeValidation, message)
}

// UnauthorizedError returns a 401 unauthorized error
func UnauthorizedError(c echo.Context, message string) error {
	if message == "" {
		message = "Unauthorized"
	}
	return ErrorResponseHelper(c, http.StatusUnauthorized, ErrorCodeUnauthorized, message)
}

// ForbiddenError returns a 403 forbidden error
func ForbiddenError(c echo.Context, message string) error {
	if message == "" {
		message = "Forbidden"
	}
	return ErrorResponseHelper(c, http.StatusForbidden, ErrorCodeForbidden, message)
}

// EmailNotVerifiedError returns a 403 with code email_not_verified (so frontend can show verify-email UI).
func EmailNotVerifiedError(c echo.Context, message string) error {
	if message == "" {
		message = "Please verify your email address to make changes."
	}
	return ErrorResponseHelper(c, http.StatusForbidden, ErrorCodeEmailNotVerified, message)
}

// TooManyRequestsError returns a 429 too many requests error (e.g. resend cooldown).
func TooManyRequestsError(c echo.Context, message string) error {
	if message == "" {
		message = "Too many requests"
	}
	return ErrorResponseHelper(c, http.StatusTooManyRequests, ErrorCodeTooManyRequests, message)
}

// NotFoundError returns a 404 not found error
func NotFoundError(c echo.Context, message string) error {
	if message == "" {
		message = "Not found"
	}
	return ErrorResponseHelper(c, http.StatusNotFound, ErrorCodeNotFound, message)
}

// ConflictError returns a 409 conflict error
func ConflictError(c echo.Context, message string) error {
	if message == "" {
		message = "Conflict"
	}
	return ErrorResponseHelper(c, http.StatusConflict, ErrorCodeConflict, message)
}

// InternalError returns a 500 internal server error
func InternalError(c echo.Context, message string) error {
	if message == "" {
		message = "Internal server error"
	}
	return ErrorResponseHelper(c, http.StatusInternalServerError, ErrorCodeInternal, message)
}
