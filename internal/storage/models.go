package storage

import (
	"time"

	"github.com/google/uuid"
)

// Game represents a chess game.
type Game struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FEN         string
	PGN         string
	OwnerID     uuid.UUID `gorm:"type:uuid;index"`
	OwnerColor  string
	Status      string
	Result      string
	Active      bool `gorm:"index"`
	CompletedAt *time.Time
	LastSeen    time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Sessions    []GameSession
	Moves       []Move
}

// GameSession represents an instance of a game session.
type GameSession struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	GameID    uuid.UUID `gorm:"type:uuid;index"`
	Game      Game      `gorm:"constraint:OnDelete:CASCADE;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Users     []UserSession
}

// UserSession links a user to a game session.
type UserSession struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	GameID        uuid.UUID `gorm:"type:uuid;index"`
	GameSessionID uuid.UUID `gorm:"type:uuid;index"`
	UserID        uuid.UUID `gorm:"type:uuid;index;uniqueIndex:idx_game_user"`
	Color         string
	Role          string
	Active        bool
	LastSeen      time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Move stores a single move in a game.
type Move struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	GameID    uuid.UUID `gorm:"type:uuid;index"`
	UserID    uuid.UUID `gorm:"type:uuid;index"`
	Number    int
	UCI       string
	Color     string
	CreatedAt time.Time
}
