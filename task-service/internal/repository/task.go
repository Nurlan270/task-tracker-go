package repository

import (
	"context"

	"github.com/Nurlan270/task-tracker-go/task-service/internal/models"
)

type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	Update(ctx context.Context, task *models.Task) error
	SoftDelete(ctx context.Context, task *models.Task) error
	Restore(ctx context.Context, task *models.Task) error
	BatchDelete(ctx context.Context, userID uint64, ids []uint64) error
	List(ctx context.Context, userID uint64, taskType models.TaskType) ([]models.Task, error)
	ToggleCompletion(ctx context.Context, task *models.Task) error
}
