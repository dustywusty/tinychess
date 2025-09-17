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
	if err := db.Exec("DROP INDEX IF EXISTS idx_game_user").Error; err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Game{}, &GameSession{}, &UserSession{}, &Move{}); err != nil {
		return nil, err
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_user_sessions_game_user ON user_sessions (game_id, user_id)").Error; err != nil {
		return nil, err
	}
	return db, nil
}
