package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/korneevDev/task-service/internal/models"
	"github.com/korneevDev/task-service/internal/repository"
)

type TaskHandler struct {
	repo      repository.TaskRepository
	jwtSecret string
}

func NewTaskHandler(repo repository.TaskRepository, jwtSecret string) *TaskHandler {
	return &TaskHandler{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (h *TaskHandler) ExtractUserID(c *gin.Context) (uint, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return 0, errors.New("authorization header is required")
	}

	// Проверяем формат "Bearer {token}"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, errors.New("expected authorization header format: Bearer {token}")
	}

	tokenString := parts[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, jwt.ErrInvalidKey
	}

	userID, err := claims.GetSubject()
	if err != nil {
		return 0, err
	}

	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint(id), nil
}

// Create godoc
// @Summary Создать задачу
// @Tags tasks
// @Security BearerAuth
// @Param task body models.Task true "Данные задачи"
// @Success 201 {object} models.Task
// @Router /tasks [post]
func (h *TaskHandler) Create(c *gin.Context) {
	userID, err := h.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.UserID = userID
	if err := h.repo.Create(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// List godoc
// @Summary Получить список задач
// @Tags tasks
// @Security BearerAuth
// @Success 200 {array} models.Task
// @Router /tasks [get]
func (h *TaskHandler) List(c *gin.Context) {
	userID, err := h.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	tasks, err := h.repo.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// GetTaskByID godoc
// @Summary Получить задачу по ID
// @Security BearerAuth
// @Param id path int true "ID задачи"
// @Success 200 {object} models.Task
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetTaskByID(c *gin.Context) {
	userID, err := h.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := h.repo.GetByIDWithOwner(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// UpdateTask godoc
// @Summary Обновить задачу
// @Security BearerAuth
// @Param id path int true "ID задачи"
// @Param task body models.Task true "Данные для обновления"
// @Success 200 {object} models.Task
// @Router /tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	userID, err := h.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.ID = uint(id)
	if err := h.repo.UpdateForUser(&task, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	updatedTask, _ := h.repo.GetByIDWithOwner(uint(id), userID)
	c.JSON(http.StatusOK, updatedTask)
}

// DeleteTask godoc
// @Summary Удалить задачу
// @Security BearerAuth
// @Param id path int true "ID задачи"
// @Success 204
// @Router /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	userID, err := h.ExtractUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := h.repo.DeleteForUser(uint(id), userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
