package repository

import (
	"testing"

	"github.com/korneevDev/auth-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo UserRepository // Используем интерфейс, а не конкретную реализацию
}

func (s *UserRepositoryTestSuite) SetupTest() {
	var err error
	s.db, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(s.T(), err)

	err = s.db.AutoMigrate(&models.User{})
	assert.NoError(s.T(), err)

	s.repo = NewUserRepository(s.db) // Возвращает UserRepository (интерфейс)
}

func (s *UserRepositoryTestSuite) TearDownTest() {
	db, _ := s.db.DB()
	db.Close()
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (s *UserRepositoryTestSuite) TestCreateUser() {
	user := &models.User{
		Username: "testuser",
		Password: "testpass",
	}

	err := s.repo.CreateUser(user)
	s.NoError(err)
	s.NotZero(user.ID)
}

func (s *UserRepositoryTestSuite) TestGetUserByUsername() {
	// Создаем тестового пользователя
	testUser := &models.User{
		Username: "existing_user",
		Password: "password",
	}
	err := s.repo.CreateUser(testUser)
	s.NoError(err)

	// Получаем существующего пользователя
	user, err := s.repo.GetUserByUsername("existing_user")
	s.NoError(err)
	s.Equal("existing_user", user.Username)
}

func (s *UserRepositoryTestSuite) TestRefreshTokenOperations() {
	// Создаем пользователя
	user := &models.User{
		Username: "refresh_user",
		Password: "password",
	}
	err := s.repo.CreateUser(user)
	s.NoError(err)

	// Сохраняем refresh токен
	token := "test_refresh_token"
	err = s.repo.SaveRefreshToken(user.ID, token)
	s.NoError(err)

	// Проверяем сохранение токена
	updatedUser, err := s.repo.GetUserByUsername("refresh_user")
	s.NoError(err)
	s.Equal(token, updatedUser.RefreshToken)

	// Получаем пользователя по токену
	foundUser, err := s.repo.GetUserByRefreshToken(token)
	s.NoError(err)
	s.Equal(user.ID, foundUser.ID)
}

func (s *UserRepositoryTestSuite) TestCreateUserDuplicate() {
	user1 := &models.User{
		Username: "duplicate_user",
		Password: "pass1",
	}
	err := s.repo.CreateUser(user1)
	s.NoError(err)
}
