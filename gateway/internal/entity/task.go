package entity

import (
	"time"

	taskV1 "github.com/Nurlan270/task-tracker-go/protos/gen/proto/task/v1"
)

type Task struct {
	ID          uint64     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type TaskType string

const (
	TaskTypeUncompleted TaskType = "uncompleted"
	TaskTypeCompleted   TaskType = "completed"
	TaskTypeDeleted     TaskType = "deleted"
)

// ToProto converts TaskType from string to proto representation.
func (t TaskType) ToProto() taskV1.TaskType {
	switch t {
	case TaskTypeUncompleted:
		return taskV1.TaskType_UNCOMPLETED
	case TaskTypeCompleted:
		return taskV1.TaskType_COMPLETED
	case TaskTypeDeleted:
		return taskV1.TaskType_DELETED
	default:
		panic("unable to convert to proto: invalid task type")
	}
}
