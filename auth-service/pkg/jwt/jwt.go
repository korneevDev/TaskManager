package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/korneevDev/auth-service/internal/models"
)

func GenerateAccessToken(user *models.User, accessTokenExpiry time.Duration, secretKey string) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(accessTokenExpiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func GenerateRefreshToken(user *models.User, refreshTokenExpiry time.Duration, secretKey string) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(refreshTokenExpiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}
