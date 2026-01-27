package http

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// CustomValidator wraps the validator
type CustomValidator struct {
	validator *validator.Validate
}

// NewCustomValidator creates a new custom validator
func NewCustomValidator() *CustomValidator {
	v := validator.New()
	
	// Register custom email validation
	v.RegisterValidation("email", validateEmail)
	
	return &CustomValidator{validator: v}
}

// Validate implements echo.Validator interface
func (cv *CustomValidator) Validate(i interface{}) error {
	if cv.validator == nil {
		cv.validator = validator.New()
		cv.validator.RegisterValidation("email", validateEmail)
	}
	return cv.validator.Struct(i)
}

// validateEmail validates email format
func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	if email == "" {
		return false
	}
	// Simple email regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
