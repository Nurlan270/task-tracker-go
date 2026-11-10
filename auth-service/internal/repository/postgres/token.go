package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	errs "github.com/Nurlan270/task-tracker-go/auth-service/internal/errors"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/models"

	ctxlib "github.com/Nurlan270/task-tracker-go/common/context"
)

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) CreateRefresh(ctx context.Context, token *models.RefreshToken) error {
	db := ctxlib.GetDB(ctx, r.db)

	if err := gorm.G[models.RefreshToken](db).Create(ctx, token); err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *TokenRepository) GetRefreshByHashWithUser(
	ctx context.Context,
	hash string,
) (models.RefreshToken, error) {
	db := ctxlib.GetDB(ctx, r.db)

	token, err := gorm.G[models.RefreshToken](db).
		Preload("User", nil).
		Where("token_hash = ?", hash).
		Where("revoked_at IS NULL").
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.RefreshToken{}, errs.ErrRefreshTokenNotFound
		}

		return models.RefreshToken{}, fmt.Errorf("failed to fetch refresh token: %w", err)
	}

	return token, nil
}

func (r *TokenRepository) MarkRefreshAsRevokedByHash(ctx context.Context, hash string) error {
	db := ctxlib.GetDB(ctx, r.db)

	updated, err := gorm.G[models.RefreshToken](db).
		Where("token_hash = ?", hash).
		Update(ctx, "revoked_at", time.Now())
	if err != nil {
		return fmt.Errorf("failed to mark refresh token as revoked: %w", err)
	}

	if updated == 0 {
		return errs.ErrRefreshTokenNotFound
	}

	return nil
}

func (r *TokenRepository) DeleteExpiredOrRevokedTokens(ctx context.Context) (int, error) {
	db := ctxlib.GetDB(ctx, r.db)

	deleted, err := gorm.G[models.RefreshToken](db).
		Where("expires_at <= ?", time.Now()).
		Or("revoked_at IS NOT NULL").
		Delete(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	return deleted, nil
}
