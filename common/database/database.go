package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Connector interface {
	GetDSN() string
}

func MustNew(conn Connector) *gorm.DB {
	db, err := New(conn)
	if err != nil {
		panic(err)
	}

	return db
}

func New(conn Connector) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(conn.GetDSN()), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Discard,
	})
	if err != nil {
		return &gorm.DB{}, fmt.Errorf("postgres: failed to connect: %w", err)
	}

	return db, nil
}
