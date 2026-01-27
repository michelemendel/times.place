package service

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/michelemendel/times.place/utils"
	"golang.org/x/crypto/bcrypt"
)

const (
	// AccessTokenLifetime is the lifetime of access tokens (15 minutes)
	AccessTokenLifetime = 15 * time.Minute
	// RefreshTokenLifetime is the lifetime of refresh tokens (30 days)
	RefreshTokenLifetime = 30 * 24 * time.Hour
)

// AuthService handles authentication-related operations
type AuthService struct {
	jwtSecret           string
	refreshTokenSecret  string
}

// NewAuthService creates a new auth service
func NewAuthService() (*AuthService, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET environment variable is required")
	}

	refreshTokenSecret := os.Getenv("REFRESH_TOKEN_SECRET")
	if refreshTokenSecret == "" {
		// Fall back to JWT_SECRET if REFRESH_TOKEN_SECRET is not set
		refreshTokenSecret = jwtSecret
	}

	return &AuthService{
		jwtSecret:          jwtSecret,
		refreshTokenSecret: refreshTokenSecret,
	}, nil
}

// HashPassword hashes a password using bcrypt with cost 12
func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against a hash
func (s *AuthService) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GenerateAccessToken generates a JWT access token for an owner
func (s *AuthService) GenerateAccessToken(ownerUUID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   ownerUUID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenLifetime)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ParseAccessToken parses and validates an access token, returning the owner UUID
func (s *AuthService) ParseAccessToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	ownerUUID, ok := claims["sub"].(string)
	if !ok {
		return "", errors.New("missing owner UUID in token")
	}

	return ownerUUID, nil
}

// GenerateRefreshToken generates a random refresh token
func (s *AuthService) GenerateRefreshToken() (string, error) {
	// Use the existing utility function for token generation
	return utils.GenerateToken()
}

// HashRefreshToken hashes a refresh token using SHA256
func (s *AuthService) HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}
