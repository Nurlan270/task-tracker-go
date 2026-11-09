package repository

import (
	"context"

	"github.com/Nurlan270/task-tracker-go/auth-service/internal/models"
)

type TokenRepository interface {
	CreateRefresh(ctx context.Context, token *models.RefreshToken) error
	GetRefreshByHashWithUser(ctx context.Context, hash string) (models.RefreshToken, error)
	MarkRefreshAsRevokedByHash(ctx context.Context, hash string) error
	DeleteExpiredAndRevokedTokens(ctx context.Context) (int, error)
}
