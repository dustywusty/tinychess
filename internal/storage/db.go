package storage

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// New initializes the database connection and performs migrations.
func New(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{Logger: logger.New(log.New(os.Stderr, "", log.LstdFlags), logger.Config{
		SlowThreshold: time.Second, LogLevel: logger.Warn,
		IgnoreRecordNotFoundError: true, ParameterizedQueries: true,
	})})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Game{}, &GameSession{}, &UserSession{}, &Move{}, &GameEval{}); err != nil {
		if pool, poolErr := db.DB(); poolErr == nil {
			pool.Close()
		}
		return nil, err
	}
	return db, nil
}
