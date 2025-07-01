package repository

import (
	"github.com/korneevDev/auth-service/internal/models"
	"gorm.io/gorm"
)

// UserRepository определяет интерфейс репозитория
type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByUsername(username string) (*models.User, error)
	SaveRefreshToken(userID uint, token string) error
	GetUserByRefreshToken(token string) (*models.User, error)
}

// userRepositoryImpl - реализация интерфейса UserRepository
type userRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository создает новую реализацию UserRepository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepositoryImpl) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *userRepositoryImpl) SaveRefreshToken(userID uint, token string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("refresh_token", token).Error
}

func (r *userRepositoryImpl) GetUserByRefreshToken(token string) (*models.User, error) {
	var user models.User
	err := r.db.Where("refresh_token = ?", token).First(&user).Error
	return &user, err
}
