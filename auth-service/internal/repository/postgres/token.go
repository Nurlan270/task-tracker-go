package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/Nurlan270/task-tracker-go/auth-service/internal/models"

	ctxlib "github.com/Nurlan270/task-tracker-go/common/context"
)

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) CreateRefresh(ctx context.Context, token models.RefreshToken) error {
	db := ctxlib.GetDB(ctx, r.db)

	if err := gorm.G[models.RefreshToken](db).Create(ctx, &token); err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *TokenRepository) DeleteRefreshByHash(ctx context.Context, hash string) error {
	db := ctxlib.GetDB(ctx, r.db)

	if _, err := gorm.G[models.RefreshToken](db).Where("hash = $1", hash).Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}
