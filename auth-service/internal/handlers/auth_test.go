package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/korneevDev/auth-service/internal/models"
	"github.com/korneevDev/auth-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.Default()
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name         string
		payload      interface{}
		mockSetup    func(*repository.MockUserRepository)
		expectedCode int
		expectedBody string
	}{
		{
			name: "successful registration",
			payload: models.User{
				Username: "testuser",
				Password: "testpass",
			},
			mockSetup: func(m *repository.MockUserRepository) {
				m.On("CreateUser", mock.AnythingOfType("*models.User")).
					Return(nil).
					Once()
			},
			expectedCode: http.StatusCreated,
			expectedBody: `{"message":"User created"}`,
		},
		{
			name:         "invalid request",
			payload:      "invalid",
			mockSetup:    func(m *repository.MockUserRepository) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"Invalid request"}`,
		},
		{
			name: "duplicate username",
			payload: models.User{
				Username: "existing",
				Password: "testpass",
			},
			mockSetup: func(m *repository.MockUserRepository) {
				m.On("CreateUser", mock.AnythingOfType("*models.User")).
					Return(gorm.ErrDuplicatedKey).
					Once()
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"Failed to create user"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repository.MockUserRepository)
			tt.mockSetup(mockRepo)

			h := NewAuthHandler(
				mockRepo,
				"testsecret",
				15*time.Minute,
				24*time.Hour,
			)

			router := setupRouter()
			router.POST("/register", h.Register)

			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	hashedPass, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)

	tests := []struct {
		name         string
		payload      interface{}
		mockSetup    func(*repository.MockUserRepository)
		expectedCode int
		checkBody    func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful login",
			payload: models.LoginRequest{
				Username: "testuser",
				Password: "correct",
			},
			mockSetup: func(m *repository.MockUserRepository) {
				m.On("GetUserByUsername", "testuser").
					Return(&models.User{
						ID:       1,
						Username: "testuser",
						Password: string(hashedPass),
					}, nil).
					Once()
				m.On("SaveRefreshToken", uint(1), mock.AnythingOfType("string")).
					Return(nil).
					Once()
			},
			expectedCode: http.StatusOK,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp["access_token"])
				assert.NotEmpty(t, resp["refresh_token"])
			},
		},
		{
			name: "invalid credentials",
			payload: models.LoginRequest{
				Username: "testuser",
				Password: "wrong",
			},
			mockSetup: func(m *repository.MockUserRepository) {
				m.On("GetUserByUsername", "testuser").
					Return(&models.User{
						Username: "testuser",
						Password: string(hashedPass),
					}, nil).
					Once()
			},
			expectedCode: http.StatusUnauthorized,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.JSONEq(t, `{"error":"Invalid login or pass"}`, w.Body.String())
			},
		},
		{
			name: "user not found",
			payload: models.LoginRequest{
				Username: "nonexistent",
				Password: "any",
			},
			mockSetup: func(m *repository.MockUserRepository) {
				m.On("GetUserByUsername", "nonexistent").
					Return(nil, gorm.ErrRecordNotFound).
					Once()
			},
			expectedCode: http.StatusUnauthorized,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.JSONEq(t, `{"error":"Invalid login or pass"}`, w.Body.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repository.MockUserRepository)
			tt.mockSetup(mockRepo)

			h := NewAuthHandler(
				mockRepo,
				"testsecret",
				15*time.Minute,
				24*time.Hour,
			)

			router := setupRouter()
			router.POST("/login", h.Login)

			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			tt.checkBody(t, w)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Refresh(t *testing.T) {
	tests := []struct {
		name         string
		payload      interface{}
		mockSetup    func(*repository.MockUserRepository)
		expectedCode int
		checkBody    func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful refresh",
			payload: map[string]string{
				"refresh_token": "valid.token.here",
			},
			mockSetup: func(m *repository.MockUserRepository) {
				m.On("GetUserByRefreshToken", "valid.token.here").
					Return(&models.User{
						ID:       1,
						Username: "testuser",
					}, nil).
					Once()
			},
			expectedCode: http.StatusOK,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp["access_token"])
			},
		},
		{
			name: "invalid token",
			payload: map[string]string{
				"refresh_token": "invalid.token",
			},
			mockSetup: func(m *repository.MockUserRepository) {
				m.On("GetUserByRefreshToken", "invalid.token").
					Return(nil, gorm.ErrRecordNotFound).
					Once()
			},
			expectedCode: http.StatusUnauthorized,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.JSONEq(t, `{"error":"Invalid refresh token"}`, w.Body.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repository.MockUserRepository)
			tt.mockSetup(mockRepo)

			h := NewAuthHandler(
				mockRepo,
				"testsecret",
				15*time.Minute,
				24*time.Hour,
			)

			router := setupRouter()
			router.POST("/refresh", h.Refresh)

			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.checkBody != nil {
				tt.checkBody(t, w)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Register_DBError(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	mockRepo.On("CreateUser", mock.AnythingOfType("*models.User")).
		Return(gorm.ErrInvalidDB).
		Once()

	h := NewAuthHandler(
		mockRepo,
		"testsecret",
		15*time.Minute,
		24*time.Hour,
	)

	router := setupRouter()
	router.POST("/register", h.Register)

	payload := models.User{
		Username: "testuser",
		Password: "testpass",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"Failed to create user"}`, w.Body.String())
	mockRepo.AssertExpectations(t)
}
