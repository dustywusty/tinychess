package storage

import (
	"context"
	"errors"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// gormStore persists games to a SQL database via GORM.
type gormStore struct {
	db *gorm.DB
}

// NewGormStore opens a Postgres connection at dsn, runs schema migrations,
// and returns a Store backed by GORM.
func NewGormStore(dsn string) (Store, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Game{}, &GameSession{}, &UserSession{}, &Move{}, &GameEval{}); err != nil {
		return nil, err
	}
	return &gormStore{db: db}, nil
}

func (s *gormStore) EnsureGame(ctx context.Context, id string, startedAt time.Time) error {
	gid, ok := parseGameID(id)
	if !ok {
		return nil
	}
	row := Game{
		ID:        gid,
		StartedAt: startedAt,
	}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}

func (s *gormStore) RecordMove(ctx context.Context, gameID string, ply int, uci, fen string) error {
	gid, ok := parseGameID(gameID)
	if !ok {
		return nil
	}
	row := Move{
		GameID: gid,
		Number: ply,
		UCI:    uci,
		FEN:    fen,
	}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}

func (s *gormStore) FinishGame(ctx context.Context, gameID, result, pgn, fen string, endedAt time.Time, moveCount int) error {
	gid, ok := parseGameID(gameID)
	if !ok {
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
	return s.db.WithContext(ctx).Model(&Game{}).Where("id = ?", gid).Updates(updates).Error
}

func (s *gormStore) UpdateGamePosition(ctx context.Context, gameID, fen, pgn string, moveCount int) error {
	gid, ok := parseGameID(gameID)
	if !ok {
		return nil
	}
	updates := map[string]any{
		"fen":        fen,
		"pgn":        pgn,
		"move_count": moveCount,
		"updated_at": time.Now(),
	}
	return s.db.WithContext(ctx).Model(&Game{}).Where("id = ?", gid).Updates(updates).Error
}

func (s *gormStore) GetGame(ctx context.Context, id string) (*Game, error) {
	gid, ok := parseGameID(id)
	if !ok {
		return nil, ErrNotFound
	}
	var g Game
	if err := s.db.WithContext(ctx).First(&g, "id = ?", gid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &g, nil
}

func (s *gormStore) GetEvals(ctx context.Context, gameID string) ([]GameEval, error) {
	gid, ok := parseGameID(gameID)
	if !ok {
		return nil, nil
	}
	var rows []GameEval
	if err := s.db.WithContext(ctx).Where("game_id = ?", gid).Order("ply ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *gormStore) AppendEvals(ctx context.Context, gameID string, evals []GameEval) error {
	if len(evals) == 0 {
		return nil
	}
	gid, ok := parseGameID(gameID)
	if !ok {
		return nil
	}
	for i := range evals {
		evals[i].GameID = gid
	}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "game_id"}, {Name: "ply"}},
		DoUpdates: clause.AssignmentColumns([]string{"fen", "eval_cp", "mate_in", "best_move"}),
	}).Create(&evals).Error
}
