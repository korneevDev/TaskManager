package service

import (
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo      repository.AuthRepository
	jwtSecret string
}

func NewAuthService(repo repository.AuthRepository, jwtSecret string) *AuthService {
	return &AuthService{repo: repo, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(email, password, name string) (*models.User, error) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	userID, err := s.repo.CreateUser(email, string(hashedPassword), name)
	if err != nil {
		return nil, err
	}
	return &models.User{ID: userID, Email: email, Name: name}, nil
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return "", errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}

	// Генерация JWT (используйте github.com/golang-jwt/jwt)
	token := generateJWT(user.ID, s.jwtSecret)
	return token, nil
}
