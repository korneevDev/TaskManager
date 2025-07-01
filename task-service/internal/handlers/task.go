package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/korneevDev/task-service/internal/models"
	"github.com/korneevDev/task-service/internal/repository"
)

type TaskHandler struct {
	repo repository.TaskRepository
}

func NewTaskHandler(
	taskRepo repository.TaskRepository,
) *TaskHandler {
	return &TaskHandler{repo: taskRepo}
}

// @Summary Создать задачу
// @Tags tasks
// @Security BearerAuth
// @Param task body models.Task true "Данные задачи"
// @Success 201 {object} models.Task
// @Router /tasks [post]
func (h *TaskHandler) Create(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("userID")
	task.UserID = userID

	if err := h.repo.Create(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// @Summary Получить список задач
// @Tags tasks
// @Security BearerAuth
// @Success 200 {array} models.Task
// @Router /tasks [get]
func (h *TaskHandler) List(c *gin.Context) {
	userID := c.GetUint("userID")
	tasks, err := h.repo.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// Аналогично реализуйте Get, Update, Delete
