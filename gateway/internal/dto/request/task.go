package request

import "github.com/Nurlan270/task-tracker-go/gateway/internal/entity"

type CreateTask struct {
	UserID      uint64 `json:"user_id"     validate:"required"`
	Title       string `json:"title"       validate:"required"`
	Description string `json:"description"`
}

type UpdateTask struct {
	TaskID      uint64 `json:"task_id"     uri:"id" validate:"required"`
	UserID      uint64 `json:"user_id"              validate:"required"`
	Title       string `json:"title"                validate:"required"`
	Description string `json:"description"`
}

type SoftDeleteTask struct {
	TaskID uint64 `json:"task_id" uri:"id" validate:"required"`
	UserID uint64 `json:"user_id"          validate:"required"`
}

type BatchDeleteTask struct {
	IDs []uint64 `json:"ids" validate:"required"`
}

type RestoreTask struct {
	TaskID uint64 `json:"task_id" uri:"id" validate:"required"`
	UserID uint64 `json:"user_id"          validate:"required"`
}

type ListTask struct {
	UserID uint64          `json:"user_id" validate:"required"`
	Type   entity.TaskType `json:"type"    validate:"required,taskType" form:"type"`
}

type ToggleCompletion struct {
	TaskID uint64 `json:"task_id" uri:"id" validate:"required"`
	UserID uint64 `json:"user_id"          validate:"required"`
}
