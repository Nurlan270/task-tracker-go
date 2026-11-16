package postgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/Nurlan270/task-tracker-go/task-service/internal/errors"
	"github.com/Nurlan270/task-tracker-go/task-service/internal/models"

	ctxlib "github.com/Nurlan270/task-tracker-go/common/context"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *models.Task) error {
	const op = "repository.postgres.task.Create"

	db := ctxlib.GetDB(ctx, r.db)

	if err := gorm.G[models.Task](db).Create(ctx, task); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *TaskRepository) Update(ctx context.Context, task *models.Task) error {
	const op = "repository.postgres.task.Update"

	db := ctxlib.GetDB(ctx, r.db)

	updated, err := gorm.G[models.Task](db).
		Where("id = ?", task.ID).
		Where("user_id = ?", task.UserID).
		Select("title, description").
		Updates(ctx, models.Task{Title: task.Title, Description: task.Description})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if updated == 0 {
		//	Specified task not found
		return errors.ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, task *models.Task, isForce bool) error {
	const op = "repository.postgres.task.Delete"

	db := ctxlib.GetDB(ctx, r.db)

	query := db.Where("id = ?", task.ID).
		Where("user_id = ?", task.UserID)

	if isForce {
		query = query.Unscoped() //	Force delete
	}

	result := query.Delete(task)

	if result.Error != nil {
		return fmt.Errorf("%s: %w", op, result.Error)
	}

	if result.RowsAffected == 0 {
		//	Specified task not found
		return errors.ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) List(ctx context.Context, userID uint) ([]models.Task, error) {
	const op = "repository.postgres.task.List"

	db := ctxlib.GetDB(ctx, r.db)

	list, err := gorm.G[models.Task](db).
		Where("user_id = ?", userID).
		Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return list, nil
}

func (r *TaskRepository) MarkAsCompleted(ctx context.Context, task *models.Task) error {
	const op = "repository.postgres.task.MarkAsCompleted"

	db := ctxlib.GetDB(ctx, r.db)

	updated, err := gorm.G[models.Task](db).
		Where("id = ?", task.ID).
		Where("user_id = ?", task.UserID).
		Update(ctx, "completed_at", time.Now())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if updated == 0 {
		//	Specified task not found
		return errors.ErrTaskNotFound
	}

	return nil
}
