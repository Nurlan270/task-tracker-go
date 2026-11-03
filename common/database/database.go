package database

import (
	"database/sql"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Connector interface {
	GetDSN() string
	GetMaxOpenConns() int
	GetMaxIdleConns() int
}

func MustNew(conn Connector) (*gorm.DB, *sql.DB) {
	db, sqlDB, err := New(conn)
	if err != nil {
		panic(err)
	}

	return db, sqlDB
}

func New(conn Connector) (*gorm.DB, *sql.DB, error) {
	//	Connect
	db, err := gorm.Open(postgres.Open(conn.GetDSN()), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Discard,
	})
	if err != nil {
		return &gorm.DB{}, &sql.DB{}, fmt.Errorf("failed to connect: %w", err)
	}

	//	Get raw SQL for configuration
	sqlDB, err := db.DB()
	if err != nil {
		return &gorm.DB{}, &sql.DB{}, fmt.Errorf("failed to get raw sql.DB: %w", err)
	}

	// Configure connection pooling
	sqlDB.SetMaxOpenConns(conn.GetMaxOpenConns())
	sqlDB.SetMaxIdleConns(conn.GetMaxIdleConns())
	sqlDB.SetConnMaxLifetime(0) // infinite
	sqlDB.SetConnMaxIdleTime(0) // infinite

	return db, sqlDB, nil
}
