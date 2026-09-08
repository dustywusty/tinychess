package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"tinychess/internal/game"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrUnrecoverable = errors.New("game recovery data is unavailable or invalid")

// TransactGame serializes operations on one durable game. The callback uses a
// private candidate: no memory state or SSE event changes until COMMIT succeeds.
func TransactGame(ctx context.Context, db *gorm.DB, id string, create bool, action func(*game.Game)) (*game.Game, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrNotFound
	}
	var candidate *game.Game
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row Game
		if create {
			candidate = game.NewGame()
			row = Game{ID: uid, StartedAt: time.Now()}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, "id = ?", uid).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrNotFound
				}
				return err
			}
			if row.LiveState == nil {
				return ErrUnrecoverable
			}
			var state game.PersistentState
			if err := json.Unmarshal([]byte(*row.LiveState), &state); err != nil {
				return ErrUnrecoverable
			}
			var err error
			candidate, err = game.Restore(state)
			if err != nil {
				return ErrUnrecoverable
			}
			// Detect disagreement with the persisted review position, not merely
			// syntactically valid JSON. Never silently reset or truncate a game.
			position := candidate.StateLocked()
			if position.FEN != row.FEN || position.PGN != row.PGN || len(position.UCI) != row.MoveCount {
				return ErrUnrecoverable
			}
		}
		before := candidate.PersistentState()
		if action != nil {
			action(candidate)
		}
		candidate.Touch()
		after := candidate.PersistentState()
		position := candidate.StateLocked()
		if len(after.UCI) < len(before.UCI) || len(after.UCI) > len(before.UCI)+1 {
			return fmt.Errorf("invalid durable move transition")
		}
		if len(after.UCI) > len(before.UCI) {
			move := Move{GameID: uid, Number: len(after.UCI), UCI: after.UCI[len(after.UCI)-1], FEN: position.FEN}
			if err := tx.Create(&move).Error; err != nil {
				return err
			}
		}
		encoded, err := json.Marshal(after)
		if err != nil {
			return err
		}
		result := ""
		var endedAt *time.Time
		if position.Status != "" {
			if row.EndedAt != nil {
				endedAt = row.EndedAt
			} else {
				now := time.Now()
				endedAt = &now
			}
			// The status begins with the engine's result token.
			for _, token := range []string{"1-0", "0-1", "1/2-1/2"} {
				if len(position.Status) >= len(token) && position.Status[:len(token)] == token {
					result = token
					break
				}
			}
		}
		return tx.Model(&row).Updates(map[string]any{
			"live_state": string(encoded), "fen": position.FEN, "pgn": position.PGN,
			// Seat credentials stay in private live_state, never review metadata.
			"move_count": len(after.UCI),
			"result":     result, "ended_at": endedAt,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return candidate, nil
}
