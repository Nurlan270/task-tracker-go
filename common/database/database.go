package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBConfig interface {
	GetConnString() string
}

func MustNew(cfg DBConfig) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.GetConnString()), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Discard,
	})
	if err != nil {
		panic(fmt.Sprintf("database: failed to connect: %s", err))
	}

	return db
}
