package storage

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// New initializes the database connection and performs migrations.
func New(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Game{}, &GameSession{}, &UserSession{}, &Move{}); err != nil {
		return nil, err
	}
	return db, nil
}
