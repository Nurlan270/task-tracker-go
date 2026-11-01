package repository

import (
	"context"

	"github.com/Nurlan270/task-tracker-go/auth-service/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (uint, error)
	FindByEmail(ctx context.Context, email string) (models.User, error)
}
