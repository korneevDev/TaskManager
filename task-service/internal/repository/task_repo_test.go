package repository

import (
	"testing"

	"github.com/korneevDev/task-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type RepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo *TaskRepository
}

func (suite *RepositoryTestSuite) SetupTest() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		suite.FailNow("failed to connect database")
	}

	// Мигрируем схему
	err = db.AutoMigrate(&models.Task{})
	if err != nil {
		suite.FailNow("failed to migrate database")
	}

	suite.db = db
	suite.repo = NewTaskRepository(db)
}

func (suite *RepositoryTestSuite) TearDownTest() {
	_ = suite.db.Migrator().DropTable(&models.Task{})
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

func (suite *RepositoryTestSuite) TestCreateTask() {
	task := &models.Task{
		Title:       "Test task",
		Description: "Test description",
		UserID:      1,
	}

	err := suite.repo.Create(task)
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), task.ID)
}

func (suite *RepositoryTestSuite) TestGetByIDWithOwner() {
	// Создаем тестовую задачу
	task := &models.Task{
		Title:       "Test task",
		Description: "Test description",
		UserID:      1,
	}
	err := suite.repo.Create(task)
	assert.NoError(suite.T(), err)

	// Получаем задачу
	foundTask, err := suite.repo.GetByIDWithOwner(task.ID, task.UserID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), task.ID, foundTask.ID)
	assert.Equal(suite.T(), task.Title, foundTask.Title)

	// Пытаемся получить несуществующую задачу
	_, err = suite.repo.GetByIDWithOwner(999, task.UserID)
	assert.Error(suite.T(), err)

	// Пытаемся получить задачу другого пользователя
	_, err = suite.repo.GetByIDWithOwner(task.ID, 2)
	assert.Error(suite.T(), err)
}

func (suite *RepositoryTestSuite) TestUpdateForUser() {
	// Создаем тестовую задачу
	task := &models.Task{
		Title:       "Test task",
		Description: "Test description",
		UserID:      1,
	}
	err := suite.repo.Create(task)
	assert.NoError(suite.T(), err)

	// Обновляем задачу
	updatedTask := &models.Task{
		ID:          task.ID,
		Title:       "Updated title",
		Description: "Updated description",
		Status:      "in_progress",
	}
	err = suite.repo.UpdateForUser(updatedTask, task.UserID)
	assert.NoError(suite.T(), err)

	// Проверяем обновление
	foundTask, err := suite.repo.GetByIDWithOwner(task.ID, task.UserID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated title", foundTask.Title)
	assert.Equal(suite.T(), "Updated description", foundTask.Description)
	assert.Equal(suite.T(), "in_progress", foundTask.Status)

	// Пытаемся обновить несуществующую задачу
	err = suite.repo.UpdateForUser(&models.Task{ID: 999}, task.UserID)
	assert.Error(suite.T(), err)

	// Пытаемся обновить задачу другого пользователя
	err = suite.repo.UpdateForUser(updatedTask, 2)
	assert.Error(suite.T(), err)
}

func (suite *RepositoryTestSuite) TestDeleteForUser() {
	// Создаем тестовую задачу
	task := &models.Task{
		Title:       "Test task",
		Description: "Test description",
		UserID:      1,
	}
	err := suite.repo.Create(task)
	assert.NoError(suite.T(), err)

	// Удаляем задачу
	err = suite.repo.DeleteForUser(task.ID, task.UserID)
	assert.NoError(suite.T(), err)

	// Проверяем, что задача удалена
	_, err = suite.repo.GetByIDWithOwner(task.ID, task.UserID)
	assert.Error(suite.T(), err)

	// Пытаемся удалить несуществующую задачу
	err = suite.repo.DeleteForUser(999, task.UserID)
	assert.Error(suite.T(), err)

	// Пытаемся удалить задачу другого пользователя
	err = suite.repo.DeleteForUser(task.ID, 2)
	assert.Error(suite.T(), err)
}

func (suite *RepositoryTestSuite) TestListByUser() {
	// Создаем тестовые задачи для пользователя 1
	tasksUser1 := []models.Task{
		{Title: "Task 1", UserID: 1},
		{Title: "Task 2", UserID: 1},
	}
	for _, task := range tasksUser1 {
		err := suite.repo.Create(&task)
		assert.NoError(suite.T(), err)
	}

	// Создаем тестовые задачи для пользователя 2
	taskUser2 := models.Task{Title: "Task 3", UserID: 2}
	err := suite.repo.Create(&taskUser2)
	assert.NoError(suite.T(), err)

	// Получаем задачи пользователя 1
	foundTasks, err := suite.repo.ListByUser(1)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), foundTasks, 2)
	for _, task := range foundTasks {
		assert.Equal(suite.T(), uint(1), task.UserID)
	}

	// Получаем задачи пользователя 2
	foundTasks, err = suite.repo.ListByUser(2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), foundTasks, 1)
	assert.Equal(suite.T(), "Task 3", foundTasks[0].Title)

	// Получаем задачи несуществующего пользователя
	foundTasks, err = suite.repo.ListByUser(999)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), foundTasks)
}
