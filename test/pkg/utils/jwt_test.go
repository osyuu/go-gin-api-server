package utils

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/pkg/apperrors"
	"go-gin-api-server/pkg/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testUsername = "testuser"
	testEmail    = "test@example.com"
)

func TestJWTManager_GenerateToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		jwtMgr := utils.NewJWTManager("test-secret", 15*time.Minute)
		username := testUsername
		email := testEmail
		user := &model.User{
			ID:       "user-123",
			Username: &username,
			Email:    &email,
		}

		tokenResponse, err := jwtMgr.GenerateToken(user)

		assert.NoError(t, err)
		assert.NotNil(t, tokenResponse)
		assert.NotEmpty(t, tokenResponse.AccessToken)
		assert.NotEmpty(t, tokenResponse.RefreshToken)
		assert.Equal(t, "Bearer", tokenResponse.TokenType)
		assert.Equal(t, int64(900), tokenResponse.ExpiresIn) // 15 minutes in seconds
	})

	t.Run("EmptySecretKey", func(t *testing.T) {
		jwtMgr := utils.NewJWTManager("", 15*time.Minute)
		user := &model.User{
			ID: "user-123",
		}

		tokenResponse, err := jwtMgr.GenerateToken(user)

		// JWT library allows empty secret key, so this should succeed
		assert.NoError(t, err)
		assert.NotNil(t, tokenResponse)
		assert.NotEmpty(t, tokenResponse.AccessToken)
		assert.NotEmpty(t, tokenResponse.RefreshToken)
	})
}

func TestJWTManager_GenerateAccessToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		jwtMgr := utils.NewJWTManager("test-secret", 15*time.Minute)
		username := testUsername
		email := testEmail
		user := &model.User{
			ID:       "user-123",
			Username: &username,
			Email:    &email,
		}

		accessToken, err := jwtMgr.GenerateAccessToken(user)

		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
	})

	t.Run("EmptySecretKey", func(t *testing.T) {
		jwtMgr := utils.NewJWTManager("", 15*time.Minute)
		user := &model.User{
			ID: "user-123",
		}

		accessToken, err := jwtMgr.GenerateAccessToken(user)

		// JWT library allows empty secret key, so this should succeed
		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
	})
}

func TestJWTManager_ValidateToken(t *testing.T) {
	t.Run("ValidToken", func(t *testing.T) {
		jwtMgr := utils.NewJWTManager("test-secret", 15*time.Minute)
		username := testUsername
		email := testEmail
		user := &model.User{
			ID:       "user-123",
			Username: &username,
			Email:    &email,
		}

		// Generate a valid token
		tokenString, err := jwtMgr.GenerateAccessToken(user)
		assert.NoError(t, err)

		// Validate the token
		claims, err := jwtMgr.ValidateToken(tokenString)

		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, "user-123", claims.UserID)
		assert.Equal(t, utils.JWTIssuer, claims.Issuer)
		assert.Equal(t, "user-123", claims.Subject)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		jwtMgr := utils.NewJWTManager("test-secret", 15*time.Minute)

		claims, err := jwtMgr.ValidateToken("invalid-token")

		assert.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrInvalidToken)
		assert.Nil(t, claims)
	})

	t.Run("EmptyToken", func(t *testing.T) {
		jwtMgr := utils.NewJWTManager("test-secret", 15*time.Minute)

		claims, err := jwtMgr.ValidateToken("")

		assert.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrInvalidToken)
		assert.Nil(t, claims)
	})

	t.Run("WrongSecretKey", func(t *testing.T) {
		// Generate token with one secret
		jwtMgr1 := utils.NewJWTManager("secret1", 15*time.Minute)
		user := &model.User{ID: "user-123"}
		tokenString, err := jwtMgr1.GenerateAccessToken(user)
		assert.NoError(t, err)

		// Try to validate with different secret
		jwtMgr2 := utils.NewJWTManager("secret2", 15*time.Minute)
		claims, err := jwtMgr2.ValidateToken(tokenString)

		assert.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrInvalidToken)
		assert.Nil(t, claims)
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		// Create a very short-lived token
		jwtMgr := utils.NewJWTManager("test-secret", 1*time.Nanosecond)
		user := &model.User{ID: "user-123"}

		tokenString, err := jwtMgr.GenerateAccessToken(user)
		assert.NoError(t, err)

		// Wait for token to expire
		time.Sleep(2 * time.Nanosecond)

		claims, err := jwtMgr.ValidateToken(tokenString)

		assert.Error(t, err)
		assert.ErrorIs(t, err, apperrors.ErrExpiredToken)
		assert.Nil(t, claims)
	})

	t.Run("MalformedToken", func(t *testing.T) {
		jwtMgr := utils.NewJWTManager("test-secret", 15*time.Minute)

		// Test various malformed tokens
		malformedTokens := []string{
			"not.a.token",
			"header.payload",                 // missing signature
			"header.payload.signature.extra", // too many parts
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ", // missing signature
		}

		for _, token := range malformedTokens {
			claims, err := jwtMgr.ValidateToken(token)
			assert.Error(t, err, "Token should be invalid: %s", token)
			assert.ErrorIs(t, err, apperrors.ErrInvalidToken)
			assert.Nil(t, claims)
		}
	})
}

func TestJWTManager_GetTokenDuration(t *testing.T) {
	t.Run("GetDuration", func(t *testing.T) {
		expectedDuration := 30 * time.Minute
		jwtMgr := utils.NewJWTManager("test-secret", expectedDuration)

		duration := jwtMgr.GetTokenDuration()

		assert.Equal(t, expectedDuration, duration)
	})
}

func TestJWTManager_TokenConsistency(t *testing.T) {
	t.Run("GenerateTokenAndValidate", func(t *testing.T) {
		jwtMgr := utils.NewJWTManager("test-secret", 15*time.Minute)
		username := testUsername
		email := testEmail
		user := &model.User{
			ID:       "user-123",
			Username: &username,
			Email:    &email,
		}

		// Generate token response
		tokenResponse, err := jwtMgr.GenerateToken(user)
		assert.NoError(t, err)

		// Validate access token
		accessClaims, err := jwtMgr.ValidateToken(tokenResponse.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, accessClaims.UserID)

		// Validate refresh token
		refreshClaims, err := jwtMgr.ValidateToken(tokenResponse.RefreshToken)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, refreshClaims.UserID)

		// Refresh token should have longer expiration
		assert.True(t, refreshClaims.ExpiresAt.After(accessClaims.ExpiresAt.Time))
	})

	t.Run("GenerateAccessTokenAndValidate", func(t *testing.T) {
		jwtMgr := utils.NewJWTManager("test-secret", 15*time.Minute)
		username := testUsername
		email := testEmail
		user := &model.User{
			ID:       "user-123",
			Username: &username,
			Email:    &email,
		}

		// Generate access token
		accessToken, err := jwtMgr.GenerateAccessToken(user)
		assert.NoError(t, err)

		// Validate the token
		claims, err := jwtMgr.ValidateToken(accessToken)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, utils.JWTIssuer, claims.Issuer)
		assert.Equal(t, user.ID, claims.Subject)
	})
}
