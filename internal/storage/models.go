package storage

import (
	"time"

	"github.com/google/uuid"
)

// Game represents a chess game.
type Game struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FEN          string
	PGN          string
	Result       string     // '1-0' | '0-1' | '1/2-1/2' | '' (in-progress)
	WhiteSession string     // session id of the white player, optional
	BlackSession string     // session id of the black player, optional
	StartedAt    time.Time  // when the game was first observed
	EndedAt      *time.Time // null while in-progress
	MoveCount    int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Sessions     []GameSession
	Moves        []Move
	Evals        []GameEval
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
	Color         string
	Role          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Move stores a single move in a game.
type Move struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	GameID    uuid.UUID `gorm:"type:uuid;index:idx_moves_game_ply,unique"`
	Number    int       `gorm:"index:idx_moves_game_ply,unique"`
	UCI       string
	FEN       string // position AFTER this move
	CreatedAt time.Time
}

// GameEval stores a single Stockfish-evaluated position for game review.
type GameEval struct {
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	GameID   uuid.UUID `gorm:"type:uuid;index:idx_evals_game_ply,unique"`
	Game     Game      `gorm:"constraint:OnDelete:CASCADE;"`
	Ply      int       `gorm:"index:idx_evals_game_ply,unique"` // 0-indexed half-move
	FEN      string
	EvalCp   *int    // centipawns from white POV, null on mate
	MateIn   *int    // positive = white mates, negative = black, null otherwise
	BestMove string  // UCI
	CreatedAt time.Time
}
