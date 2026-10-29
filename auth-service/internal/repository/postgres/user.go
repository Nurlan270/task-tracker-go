package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	errs "github.com/Nurlan270/task-tracker-go/auth-service/internal/errors"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/models"

	ctxlib "github.com/Nurlan270/task-tracker-go/common/context"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user models.User) (uint, error) {
	db := ctxlib.GetDB(ctx, r.db)

	err := gorm.G[models.User](db).Create(ctx, &user)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return 0, errs.ErrUserAlreadyExists
		}

		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return user.ID, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (models.User, error) {
	db := ctxlib.GetDB(ctx, r.db)

	user, err := gorm.G[models.User](db).Where("email = $1", email).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, errs.ErrUserNotFound
		}

		return models.User{}, fmt.Errorf("failed to find user by email: %w", err)
	}

	return user, nil
}
