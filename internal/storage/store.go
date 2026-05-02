package storage

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound is returned by Store reads when a row is missing.
var ErrNotFound = errors.New("storage: not found")

// Store abstracts persistence so handlers can run against either a Postgres
// (gorm) backend or an in-memory backend without code changes.
type Store interface {
	EnsureGame(ctx context.Context, id string, startedAt time.Time) error
	RecordMove(ctx context.Context, gameID string, ply int, uci, fen string) error
	FinishGame(ctx context.Context, gameID, result, pgn, fen string, endedAt time.Time, moveCount int) error
	UpdateGamePosition(ctx context.Context, gameID, fen, pgn string, moveCount int) error
	GetGame(ctx context.Context, id string) (*Game, error)
	GetEvals(ctx context.Context, gameID string) ([]GameEval, error)
	AppendEvals(ctx context.Context, gameID string, evals []GameEval) error
}

// parseGameID converts a free-form game ID into a uuid.UUID. Returns
// (uuid.Nil, false) for IDs that aren't UUIDs (e.g. test fixtures like "g1").
// Both Store implementations use this so write paths silently no-op and read
// paths return ErrNotFound when given a non-UUID id.
func parseGameID(id string) (uuid.UUID, bool) {
	gid, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, false
	}
	return gid, true
}
