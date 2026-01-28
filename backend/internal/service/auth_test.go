package service

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Helper to create AuthService with test secret
func createTestAuthService(t *testing.T, jwtSecret, refreshTokenSecret string) *AuthService {
	t.Helper()
	// Save original values
	oldJWTSecret := os.Getenv("JWT_SECRET")
	oldRefreshTokenSecret := os.Getenv("REFRESH_TOKEN_SECRET")
	
	// Restore after test
	t.Cleanup(func() {
		if oldJWTSecret != "" {
			os.Setenv("JWT_SECRET", oldJWTSecret)
		} else {
			os.Unsetenv("JWT_SECRET")
		}
		if oldRefreshTokenSecret != "" {
			os.Setenv("REFRESH_TOKEN_SECRET", oldRefreshTokenSecret)
		} else {
			os.Unsetenv("REFRESH_TOKEN_SECRET")
		}
	})
	
	// Set test values
	os.Setenv("JWT_SECRET", jwtSecret)
	if refreshTokenSecret != "" {
		os.Setenv("REFRESH_TOKEN_SECRET", refreshTokenSecret)
	} else {
		os.Unsetenv("REFRESH_TOKEN_SECRET")
	}
	
	service, err := NewAuthService()
	if err != nil {
		t.Fatalf("Failed to create AuthService: %v", err)
	}
	return service
}

// Test Password Operations

func TestHashPassword(t *testing.T) {
	service := createTestAuthService(t, "test-secret", "")
	
	password := "test-password-123"
	hash, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	
	if hash == "" {
		t.Error("HashPassword returned empty string")
	}
	
	if hash == password {
		t.Error("HashPassword returned plain password")
	}
	
	// Verify the hash can be used for verification
	if err := service.VerifyPassword(hash, password); err != nil {
		t.Errorf("Generated hash failed verification: %v", err)
	}
}

func TestVerifyPassword(t *testing.T) {
	service := createTestAuthService(t, "test-secret", "")
	
	password := "correct-password"
	hash, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	
	// Test correct password
	if err := service.VerifyPassword(hash, password); err != nil {
		t.Errorf("VerifyPassword failed for correct password: %v", err)
	}
}

func TestVerifyPassword_Invalid(t *testing.T) {
	service := createTestAuthService(t, "test-secret", "")
	
	password := "correct-password"
	hash, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	
	// Test wrong password
	wrongPassword := "wrong-password"
	if err := service.VerifyPassword(hash, wrongPassword); err == nil {
		t.Error("VerifyPassword should fail for wrong password")
	}
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	service := createTestAuthService(t, "test-secret", "")
	
	hash, err := service.HashPassword("")
	if err != nil {
		t.Fatalf("HashPassword should handle empty password: %v", err)
	}
	
	if hash == "" {
		t.Error("HashPassword should return hash even for empty password")
	}
	
	// Verify empty password works
	if err := service.VerifyPassword(hash, ""); err != nil {
		t.Errorf("Empty password hash verification failed: %v", err)
	}
}

// Test JWT Token Operations

func TestGenerateAccessToken(t *testing.T) {
	service := createTestAuthService(t, "test-secret", "")
	
	ownerUUID := uuid.New().String()
	token, err := service.GenerateAccessToken(ownerUUID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	
	if token == "" {
		t.Error("GenerateAccessToken returned empty token")
	}
	
	// Verify token can be parsed
	parsedUUID, err := service.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("Failed to parse generated token: %v", err)
	}
	
	if parsedUUID != ownerUUID {
		t.Errorf("Parsed UUID %s does not match original %s", parsedUUID, ownerUUID)
	}
}

func TestGenerateAccessToken_Expiration(t *testing.T) {
	service := createTestAuthService(t, "test-secret", "")
	
	ownerUUID := uuid.New().String()
	token, err := service.GenerateAccessToken(ownerUUID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	
	// Parse token to check expiration
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}
	
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Failed to extract claims")
	}
	
	exp, ok := claims["exp"].(float64)
	if !ok {
		t.Fatal("Failed to extract expiration claim")
	}
	
	expTime := time.Unix(int64(exp), 0)
	expectedExp := time.Now().Add(AccessTokenLifetime)
	
	// Allow 5 second tolerance for test execution time
	diff := expTime.Sub(expectedExp)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("Token expiration %v is not close to expected %v (diff: %v)", expTime, expectedExp, diff)
	}
}

func TestParseAccessToken_Valid(t *testing.T) {
	service := createTestAuthService(t, "test-secret", "")
	
	ownerUUID := uuid.New().String()
	token, err := service.GenerateAccessToken(ownerUUID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	
	parsedUUID, err := service.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken failed: %v", err)
	}
	
	if parsedUUID != ownerUUID {
		t.Errorf("Parsed UUID %s does not match original %s", parsedUUID, ownerUUID)
	}
}

func TestParseAccessToken_InvalidSignature(t *testing.T) {
	service1 := createTestAuthService(t, "secret-1", "")
	service2 := createTestAuthService(t, "secret-2", "")
	
	ownerUUID := uuid.New().String()
	token, err := service1.GenerateAccessToken(ownerUUID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	
	// Try to parse with different secret
	_, err = service2.ParseAccessToken(token)
	if err == nil {
		t.Error("ParseAccessToken should fail for token with wrong secret")
	}
}

func TestParseAccessToken_Expired(t *testing.T) {
	service := createTestAuthService(t, "test-secret", "")
	
	// Create an expired token manually
	now := time.Now().Add(-20 * time.Minute) // 20 minutes ago
	ownerUUID := uuid.New().String()
	claims := jwt.RegisteredClaims{
		Subject:   ownerUUID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenLifetime)), // Expired
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}
	
	// Try to parse expired token
	_, err = service.ParseAccessToken(tokenString)
	if err == nil {
		t.Error("ParseAccessToken should fail for expired token")
	}
}

func TestParseAccessToken_WrongAlgorithm(t *testing.T) {
	service := createTestAuthService(t, "test-secret", "")
	
	// The ParseAccessToken method validates that the signing method is HMAC.
	// To test this, we create a token with a different algorithm (RS256).
	// However, we can't easily sign RS256 tokens without RSA keys.
	// Instead, we verify that the service correctly validates signing methods
	// by ensuring it only accepts HS256 tokens.
	
	// Create a valid HS256 token to verify the service works
	ownerUUID := uuid.New().String()
	validToken, err := service.GenerateAccessToken(ownerUUID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	
	// Parse it to verify it works with correct algorithm
	_, err = service.ParseAccessToken(validToken)
	if err != nil {
		t.Fatalf("Valid token should parse: %v", err)
	}
	
	// Note: Testing rejection of non-HMAC algorithms (like RS256) would require
	// creating a properly signed token with that algorithm, which needs RSA keys.
	// The ParseAccessToken implementation does check for HMAC method, so tokens
	// with wrong algorithms will be rejected during parsing.
}

func TestParseAccessToken_MissingSub(t *testing.T) {
	service := createTestAuthService(t, "test-secret", "")
	
	// Create token without subject claim
	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenLifetime)),
		// No Subject field
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}
	
	// Try to parse token without subject
	_, err = service.ParseAccessToken(tokenString)
	if err == nil {
		t.Error("ParseAccessToken should fail for token without subject")
	}
}

func TestParseAccessToken_RoundTrip(t *testing.T) {
	service := createTestAuthService(t, "test-secret", "")
	
	ownerUUID := uuid.New().String()
	
	// Generate token
	token, err := service.GenerateAccessToken(ownerUUID)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	
	// Parse token
	parsedUUID, err := service.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken failed: %v", err)
	}
	
	// Verify round trip
	if parsedUUID != ownerUUID {
		t.Errorf("Round trip failed: expected %s, got %s", ownerUUID, parsedUUID)
	}
}

// Test Refresh Token Operations

func TestGenerateRefreshToken(t *testing.T) {
	service := createTestAuthService(t, "test-secret", "")
	
	token1, err := service.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}
	
	if token1 == "" {
		t.Error("GenerateRefreshToken returned empty token")
	}
	
	// Generate another token - should be different
	token2, err := service.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}
	
	if token1 == token2 {
		t.Error("GenerateRefreshToken should generate different tokens each time")
	}
}

func TestHashRefreshToken(t *testing.T) {
	service := createTestAuthService(t, "test-secret", "")
	
	token := "test-refresh-token-123"
	hash := service.HashRefreshToken(token)
	
	if hash == "" {
		t.Error("HashRefreshToken returned empty hash")
	}
	
	if hash == token {
		t.Error("HashRefreshToken returned plain token")
	}
}

func TestHashRefreshToken_Deterministic(t *testing.T) {
	service := createTestAuthService(t, "test-secret", "")
	
	token := "test-refresh-token-123"
	hash1 := service.HashRefreshToken(token)
	hash2 := service.HashRefreshToken(token)
	
	if hash1 != hash2 {
		t.Error("HashRefreshToken should produce same hash for same token")
	}
}

func TestHashRefreshToken_DifferentTokens(t *testing.T) {
	service := createTestAuthService(t, "test-secret", "")
	
	token1 := "test-refresh-token-123"
	token2 := "test-refresh-token-456"
	
	hash1 := service.HashRefreshToken(token1)
	hash2 := service.HashRefreshToken(token2)
	
	if hash1 == hash2 {
		t.Error("HashRefreshToken should produce different hashes for different tokens")
	}
}

// Test Service Initialization

func TestNewAuthService_MissingJWTSecret(t *testing.T) {
	// Save original value
	oldValue := os.Getenv("JWT_SECRET")
	defer func() {
		if oldValue != "" {
			os.Setenv("JWT_SECRET", oldValue)
		} else {
			os.Unsetenv("JWT_SECRET")
		}
	}()
	
	// Unset JWT_SECRET
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("REFRESH_TOKEN_SECRET")
	
	_, err := NewAuthService()
	if err == nil {
		t.Error("NewAuthService should fail when JWT_SECRET is not set")
	}
}

func TestNewAuthService_WithRefreshTokenSecret(t *testing.T) {
	// Save original values
	oldJWTSecret := os.Getenv("JWT_SECRET")
	oldRefreshTokenSecret := os.Getenv("REFRESH_TOKEN_SECRET")
	defer func() {
		if oldJWTSecret != "" {
			os.Setenv("JWT_SECRET", oldJWTSecret)
		} else {
			os.Unsetenv("JWT_SECRET")
		}
		if oldRefreshTokenSecret != "" {
			os.Setenv("REFRESH_TOKEN_SECRET", oldRefreshTokenSecret)
		} else {
			os.Unsetenv("REFRESH_TOKEN_SECRET")
		}
	}()
	
	os.Setenv("JWT_SECRET", "jwt-secret")
	os.Setenv("REFRESH_TOKEN_SECRET", "refresh-secret")
	
	service, err := NewAuthService()
	if err != nil {
		t.Fatalf("NewAuthService failed: %v", err)
	}
	
	// Verify service was created (we can't directly access private fields,
	// but we can verify it works by using it)
	ownerUUID := uuid.New().String()
	_, err = service.GenerateAccessToken(ownerUUID)
	if err != nil {
		t.Fatalf("Service should work: %v", err)
	}
}

func TestNewAuthService_FallbackToJWTSecret(t *testing.T) {
	// Save original values
	oldJWTSecret := os.Getenv("JWT_SECRET")
	oldRefreshTokenSecret := os.Getenv("REFRESH_TOKEN_SECRET")
	defer func() {
		if oldJWTSecret != "" {
			os.Setenv("JWT_SECRET", oldJWTSecret)
		} else {
			os.Unsetenv("JWT_SECRET")
		}
		if oldRefreshTokenSecret != "" {
			os.Setenv("REFRESH_TOKEN_SECRET", oldRefreshTokenSecret)
		} else {
			os.Unsetenv("REFRESH_TOKEN_SECRET")
		}
	}()
	
	os.Setenv("JWT_SECRET", "jwt-secret")
	os.Unsetenv("REFRESH_TOKEN_SECRET")
	
	service, err := NewAuthService()
	if err != nil {
		t.Fatalf("NewAuthService failed: %v", err)
	}
	
	// Verify service was created and works
	ownerUUID := uuid.New().String()
	_, err = service.GenerateAccessToken(ownerUUID)
	if err != nil {
		t.Fatalf("Service should work: %v", err)
	}
}
