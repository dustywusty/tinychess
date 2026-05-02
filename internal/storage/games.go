package storage

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrNotFound is returned by Get* helpers when the row is missing.
var ErrNotFound = errors.New("storage: not found")

// EnsureGame inserts a Game row keyed by id if it doesn't already exist.
// No-op if db is nil. Returns nil on success or if the row already exists.
func EnsureGame(db *gorm.DB, id string, startedAt time.Time) error {
	if db == nil {
		return nil
	}
	gid, err := uuid.Parse(id)
	if err != nil {
		return nil // non-uuid game id (e.g. test fixture) — skip persistence
	}
	row := Game{
		ID:        gid,
		StartedAt: startedAt,
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}

// RecordMove appends a Move row. No-op if db is nil.
func RecordMove(db *gorm.DB, gameID string, ply int, uciStr, fen string) error {
	if db == nil {
		return nil
	}
	gid, err := uuid.Parse(gameID)
	if err != nil {
		return nil
	}
	row := Move{
		GameID: gid,
		Number: ply,
		UCI:    uciStr,
		FEN:    fen,
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}

// FinishGame updates the Game row when the game ends (or has its outcome
// finalized). Idempotent — re-running is safe. No-op if db is nil.
func FinishGame(db *gorm.DB, gameID, result, pgn, fen string, endedAt time.Time, moveCount int) error {
	if db == nil {
		return nil
	}
	gid, err := uuid.Parse(gameID)
	if err != nil {
		return nil
	}
	updates := map[string]any{
		"result":     result,
		"pgn":        pgn,
		"fen":        fen,
		"ended_at":   endedAt,
		"move_count": moveCount,
		"updated_at": time.Now(),
	}
	return db.Model(&Game{}).Where("id = ?", gid).Updates(updates).Error
}

// UpdateGamePosition keeps the persisted Game's FEN/PGN/MoveCount in sync
// with the live state — useful for in-progress games. No-op if db is nil.
func UpdateGamePosition(db *gorm.DB, gameID, fen, pgn string, moveCount int) error {
	if db == nil {
		return nil
	}
	gid, err := uuid.Parse(gameID)
	if err != nil {
		return nil
	}
	updates := map[string]any{
		"fen":        fen,
		"pgn":        pgn,
		"move_count": moveCount,
		"updated_at": time.Now(),
	}
	return db.Model(&Game{}).Where("id = ?", gid).Updates(updates).Error
}

// GetGame fetches a Game by id. Returns ErrNotFound if missing.
func GetGame(db *gorm.DB, gameID string) (*Game, error) {
	if db == nil {
		return nil, ErrNotFound
	}
	gid, err := uuid.Parse(gameID)
	if err != nil {
		return nil, ErrNotFound
	}
	var g Game
	if err := db.First(&g, "id = ?", gid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &g, nil
}

// GetEvals returns all evals for a game, ordered by ply.
func GetEvals(db *gorm.DB, gameID string) ([]GameEval, error) {
	if db == nil {
		return nil, nil
	}
	gid, err := uuid.Parse(gameID)
	if err != nil {
		return nil, nil
	}
	var rows []GameEval
	if err := db.Where("game_id = ?", gid).Order("ply ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// AppendEvals upserts a batch of evals for a game. Idempotent on (game_id, ply).
func AppendEvals(db *gorm.DB, gameID string, evals []GameEval) error {
	if db == nil || len(evals) == 0 {
		return nil
	}
	gid, err := uuid.Parse(gameID)
	if err != nil {
		return nil
	}
	for i := range evals {
		evals[i].GameID = gid
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "game_id"}, {Name: "ply"}},
		DoUpdates: clause.AssignmentColumns([]string{"fen", "eval_cp", "mate_in", "best_move"}),
	}).Create(&evals).Error
}
