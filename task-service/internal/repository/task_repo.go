package repository

import (
	"errors"

	"github.com/korneevDev/task-service/internal/models"
	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(task *models.Task) error {
	return r.db.Create(task).Error
}

func (r *TaskRepository) GetByIDWithOwner(id uint, userID uint) (*models.Task, error) {
	var task models.Task
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&task).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("task not found or access denied")
	}
	return &task, err
}

func (r *TaskRepository) UpdateForUser(task *models.Task, userID uint) error {
	result := r.db.Model(&models.Task{}).
		Where("id = ? AND user_id = ?", task.ID, userID).
		Updates(task)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("task not found or access denied")
	}
	return nil
}

func (r *TaskRepository) DeleteForUser(id uint, userID uint) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Task{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("task not found or access denied")
	}
	return nil
}

func (r *TaskRepository) ListByUser(userID uint) ([]models.Task, error) {
	var tasks []models.Task
	err := r.db.Where("user_id = ?", userID).Find(&tasks).Error
	return tasks, err
}
