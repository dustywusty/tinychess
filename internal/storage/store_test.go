package storage

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

// runStoreConformance verifies the cross-implementation contract: any Store
// returned by factory must pass these subtests. Add behaviors here, not in
// the impl-specific test functions, so memStore and gormStore stay in lock
// step.
func runStoreConformance(t *testing.T, factory func(t *testing.T) Store) {
	t.Run("ensure_game_idempotent", func(t *testing.T) {
		s := factory(t)
		ctx := context.Background()
		id := uuid.NewString()
		started := time.Now().Add(-time.Hour)
		if err := s.EnsureGame(ctx, id, started); err != nil {
			t.Fatalf("first EnsureGame: %v", err)
		}
		if err := s.EnsureGame(ctx, id, time.Now()); err != nil {
			t.Fatalf("second EnsureGame: %v", err)
		}
		g, err := s.GetGame(ctx, id)
		if err != nil {
			t.Fatalf("GetGame: %v", err)
		}
		if !g.StartedAt.Equal(started) {
			t.Errorf("StartedAt should reflect first call, got %v want %v", g.StartedAt, started)
		}
	})

	t.Run("record_move_idempotent_on_ply", func(t *testing.T) {
		s := factory(t)
		ctx := context.Background()
		id := uuid.NewString()
		_ = s.EnsureGame(ctx, id, time.Now())
		if err := s.RecordMove(ctx, id, 1, "e2e4", "fen1"); err != nil {
			t.Fatalf("RecordMove#1: %v", err)
		}
		// Second insert at same ply: silent no-op (does not error, does not overwrite).
		if err := s.RecordMove(ctx, id, 1, "DIFFERENT", "DIFFERENT"); err != nil {
			t.Fatalf("RecordMove#1 again: %v", err)
		}
	})

	t.Run("get_game_not_found", func(t *testing.T) {
		s := factory(t)
		_, err := s.GetGame(context.Background(), uuid.NewString())
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("non_uuid_id_no_op", func(t *testing.T) {
		s := factory(t)
		ctx := context.Background()
		// Writes silently no-op.
		if err := s.EnsureGame(ctx, "not-a-uuid", time.Now()); err != nil {
			t.Fatalf("EnsureGame on bad id: %v", err)
		}
		if err := s.RecordMove(ctx, "not-a-uuid", 1, "e2e4", "fen"); err != nil {
			t.Fatalf("RecordMove on bad id: %v", err)
		}
		// Reads return ErrNotFound (GetGame) or empty (GetEvals).
		_, err := s.GetGame(ctx, "not-a-uuid")
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("GetGame on bad id: want ErrNotFound, got %v", err)
		}
		evals, err := s.GetEvals(ctx, "not-a-uuid")
		if err != nil || len(evals) != 0 {
			t.Fatalf("GetEvals on bad id: want (nil, nil), got (%v, %v)", evals, err)
		}
	})

	t.Run("update_game_position_preserves_untouched_fields", func(t *testing.T) {
		s := factory(t)
		ctx := context.Background()
		id := uuid.NewString()
		started := time.Now().Add(-time.Hour)
		_ = s.EnsureGame(ctx, id, started)
		if err := s.UpdateGamePosition(ctx, id, "newfen", "newpgn", 5); err != nil {
			t.Fatalf("UpdateGamePosition: %v", err)
		}
		g, err := s.GetGame(ctx, id)
		if err != nil {
			t.Fatalf("GetGame: %v", err)
		}
		if g.FEN != "newfen" || g.PGN != "newpgn" || g.MoveCount != 5 {
			t.Errorf("update did not apply: %+v", g)
		}
		if !g.StartedAt.Equal(started) {
			t.Errorf("StartedAt was clobbered: got %v want %v", g.StartedAt, started)
		}
		if g.EndedAt != nil {
			t.Errorf("EndedAt should remain nil after position update, got %v", g.EndedAt)
		}
	})

	t.Run("finish_game_sets_ended_at_and_result", func(t *testing.T) {
		s := factory(t)
		ctx := context.Background()
		id := uuid.NewString()
		_ = s.EnsureGame(ctx, id, time.Now())
		ended := time.Now()
		if err := s.FinishGame(ctx, id, "1-0", "pgn", "fen", ended, 42); err != nil {
			t.Fatalf("FinishGame: %v", err)
		}
		g, err := s.GetGame(ctx, id)
		if err != nil {
			t.Fatalf("GetGame: %v", err)
		}
		if g.Result != "1-0" {
			t.Errorf("Result: got %q, want %q", g.Result, "1-0")
		}
		if g.EndedAt == nil {
			t.Fatalf("EndedAt should be non-nil after FinishGame")
		}
		if !g.EndedAt.Equal(ended) {
			t.Errorf("EndedAt: got %v, want %v", *g.EndedAt, ended)
		}
		if g.MoveCount != 42 {
			t.Errorf("MoveCount: got %d, want 42", g.MoveCount)
		}
	})

	t.Run("append_evals_upsert_on_game_id_ply", func(t *testing.T) {
		s := factory(t)
		ctx := context.Background()
		id := uuid.NewString()
		_ = s.EnsureGame(ctx, id, time.Now())

		cp1, cp2 := 25, -120
		first := []GameEval{
			{Ply: 0, FEN: "fen0", EvalCp: &cp1, BestMove: "e2e4"},
			{Ply: 2, FEN: "fen2", EvalCp: &cp2, BestMove: "d7d5"},
		}
		if err := s.AppendEvals(ctx, id, first); err != nil {
			t.Fatalf("AppendEvals#1: %v", err)
		}

		// Re-append ply 0 with new values; ply 4 is fresh.
		cp1b, cp4 := 99, 0
		second := []GameEval{
			{Ply: 0, FEN: "fen0-updated", EvalCp: &cp1b, BestMove: "e2e4-updated"},
			{Ply: 4, FEN: "fen4", EvalCp: &cp4, BestMove: "g1f3"},
		}
		if err := s.AppendEvals(ctx, id, second); err != nil {
			t.Fatalf("AppendEvals#2: %v", err)
		}

		got, err := s.GetEvals(ctx, id)
		if err != nil {
			t.Fatalf("GetEvals: %v", err)
		}
		if len(got) != 3 {
			t.Fatalf("expected 3 evals, got %d: %+v", len(got), got)
		}

		// GetEvals must return sorted by Ply ASC.
		for i := 1; i < len(got); i++ {
			if got[i].Ply <= got[i-1].Ply {
				t.Fatalf("expected Ply ASC, got %v", []int{got[0].Ply, got[1].Ply, got[2].Ply})
			}
		}

		if got[0].FEN != "fen0-updated" {
			t.Errorf("ply 0 FEN: got %q want %q", got[0].FEN, "fen0-updated")
		}
		if got[0].EvalCp == nil || *got[0].EvalCp != 99 {
			t.Errorf("ply 0 EvalCp: got %v want 99", got[0].EvalCp)
		}
	})

	t.Run("get_evals_empty", func(t *testing.T) {
		s := factory(t)
		evals, err := s.GetEvals(context.Background(), uuid.NewString())
		if err != nil {
			t.Fatalf("GetEvals: %v", err)
		}
		if len(evals) != 0 {
			t.Errorf("expected empty slice, got %d", len(evals))
		}
	})
}

func TestMemStoreConformance(t *testing.T) {
	runStoreConformance(t, func(t *testing.T) Store { return NewMemStore() })
}

func TestGormStoreConformance(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run gorm conformance against a real Postgres")
	}
	runStoreConformance(t, func(t *testing.T) Store {
		s, err := NewGormStore(dsn)
		if err != nil {
			t.Fatalf("NewGormStore: %v", err)
		}
		gs := s.(*gormStore)
		// Truncate for a clean slate per subtest.
		for _, tbl := range []string{"game_evals", "moves", "user_sessions", "game_sessions", "games"} {
			if err := gs.db.Exec("TRUNCATE TABLE " + tbl + " CASCADE").Error; err != nil {
				t.Fatalf("truncate %s: %v", tbl, err)
			}
		}
		return s
	})
}
