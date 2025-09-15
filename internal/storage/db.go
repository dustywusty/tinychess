package storage

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// New initializes the database connection and performs migrations.
func New(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Game{}, &GameSession{}, &UserSession{}, &Move{}); err != nil {
		return nil, err
	}
	return db, nil
}
