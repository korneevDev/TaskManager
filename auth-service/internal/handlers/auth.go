package handlers

import (
	"net/http"
	"time"

	"github.com/korneevDev/auth-service/internal/models"
	"github.com/korneevDev/auth-service/internal/repository"
	"github.com/korneevDev/auth-service/pkg/jwt"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo           repository.UserRepository
	jwtSecret          string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

func NewAuthHandler(
	userRepo repository.UserRepository,
	jwtSecret string,
	accessTokenExpiry time.Duration,
	refreshTokenExpiry time.Duration,
) *AuthHandler {
	return &AuthHandler{userRepo: userRepo,
		jwtSecret:          jwtSecret,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

// Register godoc
// @Summary Регистрация пользователя
// @Description Создает нового пользователя
// @Tags auth
// @Accept  json
// @Produce  json
// @Param user body models.RegisterRequest true "Данные пользователя"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)

	if err := h.userRepo.CreateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created"})
}

// Login godoc
// @Summary Аутентификация пользователя
// @Description Возвращает access и refresh токены
// @Tags auth
// @Accept  json
// @Produce  json
// @Param credentials body models.LoginRequest true "Учетные данные"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := h.userRepo.GetUserByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid login or pass"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid login or pass"})
		return
	}

	accessToken, _ := jwt.GenerateAccessToken(user, h.accessTokenExpiry, h.jwtSecret)
	refreshToken, _ := jwt.GenerateRefreshToken(user, h.refreshTokenExpiry, h.jwtSecret)

	h.userRepo.SaveRefreshToken(user.ID, refreshToken)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Refresh godoc
// @Summary Обновление токена
// @Description Возвращает новый access токен
// @Tags auth
// @Accept  json
// @Produce  json
// @Param token body object{refresh_token=string} true "Refresh токен"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := h.userRepo.GetUserByRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	newAccessToken, _ := jwt.GenerateAccessToken(user, h.accessTokenExpiry, h.jwtSecret)

	c.JSON(http.StatusOK, gin.H{
		"access_token": newAccessToken,
	})
}
