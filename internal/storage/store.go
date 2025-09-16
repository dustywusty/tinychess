package storage

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Store wraps a gorm DB instance and provides helper methods for persisting games.
type Store struct {
	db *gorm.DB
}

// NewStore creates a new store helper from a gorm DB.
func NewStore(db *gorm.DB) *Store {
	if db == nil {
		return nil
	}
	return &Store{db: db}
}

// DB exposes the underlying gorm DB instance.
func (s *Store) DB() *gorm.DB {
	if s == nil {
		return nil
	}
	return s.db
}

// ErrNotFound is returned when a record is not found.
var ErrNotFound = gorm.ErrRecordNotFound

// GameStateUpdate represents a partial update to a game row.
type GameStateUpdate struct {
	FEN         *string
	PGN         *string
	Status      *string
	Result      *string
	Active      *bool
	LastSeen    *time.Time
	CompletedAt *time.Time
}

// CreateGame inserts a new game with the provided identifiers.
func (s *Store) CreateGame(ctx context.Context, id, ownerID uuid.UUID, ownerColor string, lastSeen time.Time) error {
	if s == nil {
		return nil
	}
	game := Game{
		ID:         id,
		OwnerID:    ownerID,
		OwnerColor: ownerColor,
		Active:     true,
		LastSeen:   lastSeen,
	}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&game).Error
}

// SaveGameState applies partial updates to the game row.
func (s *Store) SaveGameState(ctx context.Context, id uuid.UUID, upd GameStateUpdate) error {
	if s == nil {
		return nil
	}
	updates := make(map[string]any)
	if upd.FEN != nil {
		updates["fen"] = *upd.FEN
	}
	if upd.PGN != nil {
		updates["pgn"] = *upd.PGN
	}
	if upd.Status != nil {
		updates["status"] = *upd.Status
	}
	if upd.Result != nil {
		updates["result"] = *upd.Result
	}
	if upd.Active != nil {
		updates["active"] = *upd.Active
	}
	if upd.LastSeen != nil {
		updates["last_seen"] = *upd.LastSeen
	}
	if upd.CompletedAt != nil {
		updates["completed_at"] = *upd.CompletedAt
	}
	if len(updates) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).Model(&Game{}).Where("id = ?", id).Updates(updates).Error
}

// EnsureUserSession upserts a user session record for a game.
func (s *Store) EnsureUserSession(ctx context.Context, gameID, userID uuid.UUID, color, role string, lastSeen time.Time) error {
	if s == nil {
		return nil
	}
	session := UserSession{
		GameID:   gameID,
		UserID:   userID,
		Color:    color,
		Role:     role,
		Active:   true,
		LastSeen: lastSeen,
	}
	return s.db.WithContext(ctx).
		Where("game_id = ? AND user_id = ?", gameID, userID).
		Assign(map[string]any{
			"color":     color,
			"role":      role,
			"active":    true,
			"last_seen": lastSeen,
		}).
		FirstOrCreate(&session).Error
}

// DeactivateUserSession marks the given user session as inactive.
func (s *Store) DeactivateUserSession(ctx context.Context, gameID, userID uuid.UUID) error {
	if s == nil {
		return nil
	}
	return s.db.WithContext(ctx).
		Model(&UserSession{}).
		Where("game_id = ? AND user_id = ?", gameID, userID).
		Updates(map[string]any{"active": false}).Error
}

// RecordMove inserts a move row for the given game.
func (s *Store) RecordMove(ctx context.Context, gameID, userID uuid.UUID, number int, uci, color string) error {
	if s == nil {
		return nil
	}
	move := Move{
		GameID: gameID,
		UserID: userID,
		Number: number,
		UCI:    uci,
		Color:  color,
	}
	return s.db.WithContext(ctx).Create(&move).Error
}

// LoadGame fetches a persisted game and its active sessions.
type PersistedGame struct {
	Game    Game
	Players []UserSession
}

func (s *Store) LoadGame(ctx context.Context, id uuid.UUID) (*PersistedGame, error) {
	if s == nil {
		return nil, gorm.ErrRecordNotFound
	}
	var game Game
	if err := s.db.WithContext(ctx).First(&game, "id = ?", id).Error; err != nil {
		return nil, err
	}
	var players []UserSession
	if err := s.db.WithContext(ctx).
		Where("game_id = ? AND active = ?", id, true).
		Find(&players).Error; err != nil {
		return nil, err
	}
	return &PersistedGame{Game: game, Players: players}, nil
}

// Stats represents aggregate counts for games.
type Stats struct {
	Started   int64 `json:"started"`
	Completed int64 `json:"completed"`
	Active    int64 `json:"active"`
}

// FetchStats aggregates counts for display on the home page.
func (s *Store) FetchStats(ctx context.Context) (Stats, error) {
	var stats Stats
	if s == nil {
		return stats, nil
	}
	if err := s.db.WithContext(ctx).Model(&Game{}).Count(&stats.Started).Error; err != nil {
		return stats, err
	}
	if err := s.db.WithContext(ctx).Model(&Game{}).Where("active = ?", true).Count(&stats.Active).Error; err != nil {
		return stats, err
	}
	if err := s.db.WithContext(ctx).Model(&Game{}).Where("completed_at IS NOT NULL").Count(&stats.Completed).Error; err != nil {
		return stats, err
	}
	return stats, nil
}

// CompleteGame marks a game as finished with the provided status and result.
func (s *Store) CompleteGame(ctx context.Context, id uuid.UUID, status, result string, completedAt time.Time) error {
	if s == nil {
		return nil
	}
	active := false
	return s.SaveGameState(ctx, id, GameStateUpdate{
		Status:      &status,
		Result:      &result,
		Active:      &active,
		CompletedAt: &completedAt,
	})
}

// SetActive updates the active flag for a game without changing other fields.
func (s *Store) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	if s == nil {
		return nil
	}
	return s.SaveGameState(ctx, id, GameStateUpdate{Active: &active})
}

// UpdateLastSeen updates the last seen timestamp for a game.
func (s *Store) UpdateLastSeen(ctx context.Context, id uuid.UUID, lastSeen time.Time) error {
	if s == nil {
		return nil
	}
	return s.SaveGameState(ctx, id, GameStateUpdate{LastSeen: &lastSeen})
}

// ForgetGame marks a game as ended by the owner forgetting it.
func (s *Store) ForgetGame(ctx context.Context, id uuid.UUID, when time.Time) error {
	if s == nil {
		return nil
	}
	status := "Abandoned"
	active := false
	return s.SaveGameState(ctx, id, GameStateUpdate{
		Status:      &status,
		Active:      &active,
		CompletedAt: &when,
	})
}

// DeactivateAllSessions marks all sessions for the game as inactive.
func (s *Store) DeactivateAllSessions(ctx context.Context, gameID uuid.UUID) error {
	if s == nil {
		return nil
	}
	return s.db.WithContext(ctx).Model(&UserSession{}).Where("game_id = ?", gameID).Updates(map[string]any{"active": false}).Error
}

// ErrMissingGame is returned when attempting to operate on a non-existing game.
var ErrMissingGame = errors.New("game not found")
