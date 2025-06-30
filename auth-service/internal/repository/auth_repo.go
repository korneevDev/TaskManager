package repository

import (
	"github.com/korneevDev/auth-service/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *UserRepository) SaveRefreshToken(userID uint, token string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("refresh_token", token).Error
}

func (r *UserRepository) GetUserByRefreshToken(token string) (*models.User, error) {
	var user models.User
	err := r.db.Where("refresh_token = ?", token).First(&user).Error
	return &user, err
}
