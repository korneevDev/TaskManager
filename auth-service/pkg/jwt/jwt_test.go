package jwt

import (
	"strings"
	"testing"
	"time"

	"github.com/korneevDev/auth-service/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAndValidateAccessToken(t *testing.T) {
	secret := "test_secret_123"
	user := &models.User{
		ID:       1,
		Username: "testuser",
	}
	expiry := 15 * time.Minute

	t.Run("valid token", func(t *testing.T) {
		token, err := GenerateAccessToken(user, expiry, secret)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := ValidateToken(token, secret)
		assert.NoError(t, err)
		assert.Equal(t, "1", claims["sub"])
	})

	t.Run("expired token", func(t *testing.T) {
		token, err := GenerateAccessToken(user, -15*time.Minute, secret)
		assert.NoError(t, err)

		_, err = ValidateToken(token, secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token is expired")
	})
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	secret := "test_secret_456"
	user := &models.User{
		ID:       2,
		Username: "testuser2",
	}
	expiry := 24 * time.Hour

	token, err := GenerateRefreshToken(user, expiry, secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateToken(token, secret)
	assert.NoError(t, err)
	assert.Equal(t, "2", claims["sub"])
}

func TestInvalidTokens(t *testing.T) {
	secret := "test_secret_789"

	t.Run("empty token", func(t *testing.T) {
		_, err := ValidateToken("", secret)
		assert.Error(t, err)
	})

	t.Run("malformed token", func(t *testing.T) {
		_, err := ValidateToken("invalid.token.here", secret)
		assert.Error(t, err)
	})

	t.Run("wrong signature", func(t *testing.T) {
		user := &models.User{ID: 3}
		token, _ := GenerateAccessToken(user, 15*time.Minute, secret)
		_, err := ValidateToken(token, "wrong_secret")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signature is invalid")
	})
}

func TestTokenStructure(t *testing.T) {
	secret := "test_secret_012"
	user := &models.User{
		ID:       4,
		Username: "testuser4",
	}

	token, err := GenerateAccessToken(user, 30*time.Minute, secret)
	assert.NoError(t, err)

	parts := strings.Split(token, ".")
	assert.Equal(t, 3, len(parts), "JWT should have 3 parts")
}
