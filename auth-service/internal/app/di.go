package app

import (
	"context"

	"github.com/Nurlan270/task-tracker-go/common/database"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Nurlan270/task-tracker-go/auth-service/internal/config"
)

type diContainer struct {
	//	App configuration
	cfg *config.Config

	//	Core
	pool *pgxpool.Pool
}

func newDIContainer(cfg *config.Config) diContainer {
	return diContainer{cfg: cfg}
}

func (c *diContainer) DBPool() *pgxpool.Pool {
	if c.pool == nil {
		c.pool = database.MustNewPool(context.Background(), c.cfg.Database)
	}

	return c.pool
}
