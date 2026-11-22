package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	errs "github.com/Nurlan270/task-tracker-go/task-service/internal/errors"
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
		Select("title", "description").
		Updates(ctx, models.Task{Title: task.Title, Description: task.Description})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if updated == 0 {
		//	Specified task not found
		return errs.ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) SoftDelete(ctx context.Context, task *models.Task) error {
	const op = "repository.postgres.task.SoftDelete"

	db := ctxlib.GetDB(ctx, r.db)

	result := db.WithContext(ctx).
		Where("id = ?", task.ID).
		Where("user_id = ?", task.UserID).
		Delete(task)
	if result.Error != nil {
		return fmt.Errorf("%s: %w", op, result.Error)
	}

	if result.RowsAffected == 0 {
		//	Specified task not found
		return errs.ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) Restore(ctx context.Context, task *models.Task) error {
	const op = "repository.postgres.task.Restore"

	db := ctxlib.GetDB(ctx, r.db)

	result := db.WithContext(ctx).Model(&models.Task{}).Unscoped().
		Where("id = ?", task.ID).
		Where("user_id = ?", task.UserID).
		Update("deleted_at", nil)
	if result.Error != nil {
		return fmt.Errorf("%s: %w", op, result.Error)
	}

	if result.RowsAffected == 0 {
		//	Specified task not found
		return errs.ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) BatchDelete(ctx context.Context, userID uint64, ids []uint64) error {
	const op = "repository.postgres.task.BatchDelete"

	db := ctxlib.GetDB(ctx, r.db)

	var tasks []models.Task

	for _, id := range ids {
		tasks = append(tasks, models.Task{ID: id})
	}

	result := db.WithContext(ctx).Unscoped().Where("user_id = ?", userID).Delete(&tasks)
	if result.Error != nil {
		return fmt.Errorf("%s: %w", op, result.Error)
	}

	if result.RowsAffected == 0 {
		//	Specified task(s) not found
		return errs.ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) List(
	ctx context.Context,
	userID uint64,
	taskType models.TaskType,
) ([]models.Task, error) {
	const op = "repository.postgres.task.List"

	db := ctxlib.GetDB(ctx, r.db)

	query := db.WithContext(ctx).Where("user_id = ?", userID)

	switch taskType {
	case models.TaskTypeUncompleted:
		//	Uncompleted tasks
		query = query.Where("completed_at IS NULL")
	case models.TaskTypeCompleted:
		//	Completed tasks
		query = query.Where("completed_at IS NOT NULL")
	case models.TaskTypeDeleted:
		//	Deleted tasks
		query = query.Unscoped().Where("deleted_at IS NOT NULL").Order("deleted_at DESC")
	default:
		return nil, fmt.Errorf("%s: %w", op, fmt.Errorf("invalid task type: %d", taskType))
	}

	var list []models.Task

	if err := query.Find(&list).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return list, nil
}

func (r *TaskRepository) ToggleCompletion(ctx context.Context, task *models.Task) error {
	const op = "repository.postgres.task.ToggleCompletion"

	db := ctxlib.GetDB(ctx, r.db)

	query := gorm.G[models.Task](db).
		Where("id = ?", task.ID).
		Where("user_id = ?", task.UserID)

	dbTask, err := query.First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			//	The task not found
			return errs.ErrTaskNotFound
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	if dbTask.CompletedAt == nil {
		//	The task is not completed yet; set completed_at to current time
		if _, err = query.Update(ctx, "completed_at", time.Now()); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	} else {
		//	The task is already completed; set completed_at to nil
		if _, err = query.Update(ctx, "completed_at", nil); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}
