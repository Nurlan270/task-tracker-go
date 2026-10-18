package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config interface {
	GetConnString() string
}

func MustNewPool(ctx context.Context, cfg Config) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, cfg.GetConnString())
	if err != nil {
		panic(fmt.Sprintf("failed to create new pgx pool: %s", err))
	}

	return pool
}
