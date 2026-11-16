package models

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	ID     uint `gorm:"primaryKey"`
	UserID uint `gorm:"not null;index:idx_tasks_user_id_completed_at"`

	Title       string `gorm:"not null;size:100"`
	Description string `gorm:"size:5000"`

	CompletedAt time.Time `gorm:"index:idx_tasks_user_id_completed_at"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
