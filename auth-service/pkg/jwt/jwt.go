package jwt

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/korneevDev/auth-service/internal/models"
)

func GenerateAccessToken(user *models.User, accessTokenExpiry time.Duration, secretKey string) (string, error) {
	claims := jwt.MapClaims{
		"sub": strconv.FormatUint(uint64(user.ID), 10),
		"exp": time.Now().Add(accessTokenExpiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func GenerateRefreshToken(user *models.User, refreshTokenExpiry time.Duration, secretKey string) (string, error) {
	claims := jwt.MapClaims{
		"sub": strconv.FormatUint(uint64(user.ID), 10),
		"exp": time.Now().Add(refreshTokenExpiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func ValidateToken(tokenString, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	return token.Claims.(jwt.MapClaims), nil
}
