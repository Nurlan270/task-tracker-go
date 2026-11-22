package request

import "github.com/Nurlan270/task-tracker-go/task-service/internal/models"

type CreateTask struct {
	UserID      uint64 `json:"user_id"     validate:"required"`
	Title       string `json:"title"       validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=5000"`
}

type UpdateTask struct {
	TaskID      uint64 `json:"task_id"     validate:"required"`
	UserID      uint64 `json:"user_id"     validate:"required"`
	Title       string `json:"title"       validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=5000"`
}

type SoftDeleteTask struct {
	TaskID  uint64 `json:"task_id"  validate:"required"`
	UserID  uint64 `json:"user_id"  validate:"required"`
	IsForce bool   `json:"is_force"`
}

type BatchDeleteTask struct {
	IDs    []uint64 `json:"ids" validate:"required"`
	UserID uint64   `json:"user_id"  validate:"required"`
}

type RestoreTask struct {
	TaskID uint64 `json:"task_id" validate:"required"`
	UserID uint64 `json:"user_id" validate:"required"`
}

type ListTask struct {
	UserID uint64          `json:"user_id" validate:"required"`
	Type   models.TaskType `json:"type"    validate:"required,taskType"`
}

type ToggleCompletion struct {
	TaskID uint64 `json:"task_id" validate:"required"`
	UserID uint64 `json:"user_id" validate:"required"`
}
