package repository

import (
	"context"

	"github.com/Nurlan270/task-tracker-go/task-service/internal/models"
)

type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, task *models.Task, isForce bool) error
	List(ctx context.Context, userID uint) ([]models.Task, error)
	ToggleCompletion(ctx context.Context, task *models.Task) error
}
