package repository_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/korneevDev/task-service/internal/handlers"
	"github.com/korneevDev/task-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskRepository реализует TaskRepositoryInterface для тестов
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Create(task *models.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepository) GetByIDWithOwner(id uint, userID uint) (*models.Task, error) {
	args := m.Called(id, userID)
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepository) UpdateForUser(task *models.Task, userID uint) error {
	args := m.Called(task, userID)
	return args.Error(0)
}

func (m *MockTaskRepository) DeleteForUser(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func (m *MockTaskRepository) ListByUser(userID uint) ([]models.Task, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Task), args.Error(1)
}

type HandlerTestSuite struct {
	router       *gin.Engine
	repo         *MockTaskRepository
	jwtSecret    string
	validToken   string
	invalidToken string
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.repo = &MockTaskRepository{}
	suite.jwtSecret = "test-secret"
	suite.router = gin.Default()

	handler := handlers.NewTaskHandler(suite.repo, suite.jwtSecret)

	suite.router.POST("/tasks", handler.Create)
	suite.router.GET("/tasks", handler.List)
	suite.router.GET("/tasks/:id", handler.GetTaskByID)
	suite.router.PUT("/tasks/:id", handler.UpdateTask)
	suite.router.DELETE("/tasks/:id", handler.DeleteTask)

	// Генерируем тестовые токены
	suite.validToken = generateTestToken(1, suite.jwtSecret)
	suite.invalidToken = "invalid.token.here"
}

func generateTestToken(userID uint, secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": strconv.FormatUint(uint64(userID), 10),
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestHandlers(t *testing.T) {
	suite := new(HandlerTestSuite)
	suite.SetupTest()

	t.Run("Test ExtractUserID", func(t *testing.T) {
		// Тестирование успешного извлечения ID пользователя
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request, _ = http.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer "+suite.validToken)

		handler := handlers.NewTaskHandler(suite.repo, suite.jwtSecret)
		userID, err := handler.ExtractUserID(c)

		assert.NoError(t, err)
		assert.Equal(t, uint(1), userID)

		// Тестирование ошибок
		testCases := []struct {
			name   string
			header string
			errMsg string
		}{
			{"No header", "", "authorization header is required"},
			{"Invalid format", "Token " + suite.validToken, "expected authorization header format: Bearer {token}"},
			{"Invalid token", "Bearer invalid.token", "token contains an invalid number of segments"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request, _ = http.NewRequest("GET", "/", nil)
				if tc.header != "" {
					c.Request.Header.Set("Authorization", tc.header)
				}

				_, err := handler.ExtractUserID(c)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			})
		}
	})

	t.Run("Test Create Task", func(t *testing.T) {
		task := models.Task{
			Title:       "Test task",
			Description: "Test description",
		}
		createdTask := task
		createdTask.ID = 1
		createdTask.UserID = 1

		// Настраиваем мок
		suite.repo.On("Create", mock.AnythingOfType("*models.Task")).
			Run(func(args mock.Arguments) {
				arg := args.Get(0).(*models.Task)
				arg.ID = 1
				arg.UserID = 1
			}).
			Return(nil).Once()

		// Подготавливаем запрос
		body, _ := json.Marshal(task)
		req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+suite.validToken)
		req.Header.Set("Content-Type", "application/json")

		// Выполняем запрос
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		// Проверяем результат
		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.Task
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, createdTask.ID, response.ID)
		assert.Equal(t, createdTask.Title, response.Title)
		assert.Equal(t, createdTask.UserID, response.UserID)

		// Проверяем ошибки
		testCases := []struct {
			name         string
			token        string
			body         string
			expectedCode int
		}{
			{"Invalid token", suite.invalidToken, string(body), http.StatusUnauthorized},
			{"Invalid body", suite.validToken, "{invalid json}", http.StatusBadRequest},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req, _ := http.NewRequest("POST", "/tasks", bytes.NewBufferString(tc.body))
				req.Header.Set("Authorization", "Bearer "+tc.token)
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				suite.router.ServeHTTP(w, req)
				assert.Equal(t, tc.expectedCode, w.Code)
			})
		}
	})

	t.Run("Test List Tasks", func(t *testing.T) {
		tasks := []models.Task{
			{ID: 1, Title: "Task 1", UserID: 1},
			{ID: 2, Title: "Task 2", UserID: 1},
		}

		// Настраиваем мок
		suite.repo.On("ListByUser", uint(1)).Return(tasks, nil).Once()

		// Подготавливаем запрос
		req, _ := http.NewRequest("GET", "/tasks", nil)
		req.Header.Set("Authorization", "Bearer "+suite.validToken)

		// Выполняем запрос
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		// Проверяем результат
		assert.Equal(t, http.StatusOK, w.Code)

		var response []models.Task
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 2)
		assert.Equal(t, tasks[0].Title, response[0].Title)

		// Проверяем ошибки
		t.Run("Invalid token", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/tasks", nil)
			req.Header.Set("Authorization", "Bearer "+suite.invalidToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	})

	t.Run("Test Get Task By ID", func(t *testing.T) {
		task := models.Task{ID: 1, Title: "Test task", UserID: 1}

		// Настраиваем мок для успешного случая
		suite.repo.On("GetByIDWithOwner", uint(1), uint(1)).Return(&task, nil).Once()
		// Настраиваем мок для случая "не найдено"
		suite.repo.On("GetByIDWithOwner", uint(2), uint(1)).Return(&models.Task{}, errors.New("not found")).Once()

		// Подготавливаем запрос
		req, _ := http.NewRequest("GET", "/tasks/1", nil)
		req.Header.Set("Authorization", "Bearer "+suite.validToken)

		// Выполняем запрос
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		// Проверяем результат
		assert.Equal(t, http.StatusOK, w.Code)

		var response models.Task
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, task.ID, response.ID)
		assert.Equal(t, task.Title, response.Title)

		// Проверяем ошибки
		testCases := []struct {
			name         string
			taskID       string
			expectedCode int
		}{
			{"Invalid ID", "abc", http.StatusBadRequest},
			{"Not found", "2", http.StatusNotFound},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req, _ := http.NewRequest("GET", "/tasks/"+tc.taskID, nil)
				req.Header.Set("Authorization", "Bearer "+suite.validToken)

				w := httptest.NewRecorder()
				suite.router.ServeHTTP(w, req)
				assert.Equal(t, tc.expectedCode, w.Code)
			})
		}
	})

	t.Run("Test Update Task", func(t *testing.T) {
		updatedTask := models.Task{
			ID:          1,
			Title:       "Updated title",
			Description: "Updated description",
			Status:      "in_progress",
		}

		// Настраиваем моки
		suite.repo.On("UpdateForUser", &updatedTask, uint(1)).Return(nil).Once()
		suite.repo.On("GetByIDWithOwner", uint(1), uint(1)).Return(&updatedTask, nil).Once()
		suite.repo.On("UpdateForUser", mock.Anything, uint(1)).Return(errors.New("not found")).Once()

		// Подготавливаем запрос
		body, _ := json.Marshal(updatedTask)
		req, _ := http.NewRequest("PUT", "/tasks/1", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+suite.validToken)
		req.Header.Set("Content-Type", "application/json")

		// Выполняем запрос
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		// Проверяем результат
		assert.Equal(t, http.StatusOK, w.Code)

		var response models.Task
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, updatedTask.Title, response.Title)
		assert.Equal(t, updatedTask.Status, response.Status)

		// Проверяем ошибки
		testCases := []struct {
			name         string
			taskID       string
			body         string
			expectedCode int
		}{
			{"Invalid ID", "abc", string(body), http.StatusBadRequest},
			{"Invalid body", "1", "{invalid json}", http.StatusBadRequest},
			{"Not found", "2", string(body), http.StatusNotFound},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req, _ := http.NewRequest("PUT", "/tasks/"+tc.taskID, bytes.NewBufferString(tc.body))
				req.Header.Set("Authorization", "Bearer "+suite.validToken)
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				suite.router.ServeHTTP(w, req)
				assert.Equal(t, tc.expectedCode, w.Code)
			})
		}
	})

	t.Run("Test Delete Task", func(t *testing.T) {
		// Настраиваем моки
		suite.repo.On("DeleteForUser", uint(1), uint(1)).Return(nil).Once()
		suite.repo.On("DeleteForUser", uint(2), uint(1)).Return(errors.New("not found")).Once()

		// Подготавливаем запрос
		req, _ := http.NewRequest("DELETE", "/tasks/1", nil)
		req.Header.Set("Authorization", "Bearer "+suite.validToken)

		// Выполняем запрос
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		// Проверяем результат
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Проверяем ошибки
		testCases := []struct {
			name         string
			taskID       string
			expectedCode int
		}{
			{"Invalid ID", "abc", http.StatusBadRequest},
			{"Not found", "2", http.StatusNotFound},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req, _ := http.NewRequest("DELETE", "/tasks/"+tc.taskID, nil)
				req.Header.Set("Authorization", "Bearer "+suite.validToken)

				w := httptest.NewRecorder()
				suite.router.ServeHTTP(w, req)
				assert.Equal(t, tc.expectedCode, w.Code)
			})
		}
	})
}
