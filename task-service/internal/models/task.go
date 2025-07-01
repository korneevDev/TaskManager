package models

import "time"

type Task struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Status      string    `json:"status" gorm:"default:'pending'"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	UserID      uint      `json:"user_id"`
}

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusProgress  TaskStatus = "in_progress"
	StatusCompleted TaskStatus = "completed"
)
