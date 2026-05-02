package storage

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// memStore is an in-memory Store. State lives in process; nothing survives
// a restart. Suitable for development, tests, and single-host deployments
// where durability isn't required.
type memStore struct {
	mu     sync.RWMutex
	games  map[uuid.UUID]*Game
	moves  map[uuid.UUID][]Move          // gameID → moves in insertion order
	plies  map[uuid.UUID]map[int]struct{} // gameID → set of recorded ply numbers
	evals  map[uuid.UUID][]GameEval       // gameID → evals in insertion order
	evalIx map[uuid.UUID]map[int]int      // gameID → ply → index into evals[gameID]
}

// NewMemStore returns an in-memory Store with empty state.
func NewMemStore() Store {
	return &memStore{
		games:  make(map[uuid.UUID]*Game),
		moves:  make(map[uuid.UUID][]Move),
		plies:  make(map[uuid.UUID]map[int]struct{}),
		evals:  make(map[uuid.UUID][]GameEval),
		evalIx: make(map[uuid.UUID]map[int]int),
	}
}

func (s *memStore) EnsureGame(_ context.Context, id string, startedAt time.Time) error {
	gid, ok := parseGameID(id)
	if !ok {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.games[gid]; exists {
		return nil
	}
	now := time.Now()
	s.games[gid] = &Game{
		ID:        gid,
		StartedAt: startedAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return nil
}

func (s *memStore) RecordMove(_ context.Context, gameID string, ply int, uci, fen string) error {
	gid, ok := parseGameID(gameID)
	if !ok {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.plies[gid][ply]; exists {
		return nil
	}
	if s.plies[gid] == nil {
		s.plies[gid] = make(map[int]struct{})
	}
	s.moves[gid] = append(s.moves[gid], Move{
		ID:        uuid.New(),
		GameID:    gid,
		Number:    ply,
		UCI:       uci,
		FEN:       fen,
		CreatedAt: time.Now(),
	})
	s.plies[gid][ply] = struct{}{}
	return nil
}

func (s *memStore) FinishGame(_ context.Context, gameID, result, pgn, fen string, endedAt time.Time, moveCount int) error {
	gid, ok := parseGameID(gameID)
	if !ok {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	g, exists := s.games[gid]
	if !exists {
		return nil
	}
	endedAtCopy := endedAt
	g.Result = result
	g.PGN = pgn
	g.FEN = fen
	g.EndedAt = &endedAtCopy
	g.MoveCount = moveCount
	g.UpdatedAt = time.Now()
	return nil
}

func (s *memStore) UpdateGamePosition(_ context.Context, gameID, fen, pgn string, moveCount int) error {
	gid, ok := parseGameID(gameID)
	if !ok {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	g, exists := s.games[gid]
	if !exists {
		return nil
	}
	g.FEN = fen
	g.PGN = pgn
	g.MoveCount = moveCount
	g.UpdatedAt = time.Now()
	return nil
}

func (s *memStore) GetGame(_ context.Context, id string) (*Game, error) {
	gid, ok := parseGameID(id)
	if !ok {
		return nil, ErrNotFound
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	g, exists := s.games[gid]
	if !exists {
		return nil, ErrNotFound
	}
	cp := *g
	return &cp, nil
}

func (s *memStore) GetEvals(_ context.Context, gameID string) ([]GameEval, error) {
	gid, ok := parseGameID(gameID)
	if !ok {
		return nil, nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.evals[gid]
	if len(src) == 0 {
		return nil, nil
	}
	out := make([]GameEval, len(src))
	copy(out, src)
	sort.Slice(out, func(i, j int) bool { return out[i].Ply < out[j].Ply })
	return out, nil
}

func (s *memStore) AppendEvals(_ context.Context, gameID string, evals []GameEval) error {
	if len(evals) == 0 {
		return nil
	}
	gid, ok := parseGameID(gameID)
	if !ok {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.evalIx[gid] == nil {
		s.evalIx[gid] = make(map[int]int)
	}
	now := time.Now()
	for _, e := range evals {
		if idx, exists := s.evalIx[gid][e.Ply]; exists {
			existing := &s.evals[gid][idx]
			existing.FEN = e.FEN
			existing.EvalCp = e.EvalCp
			existing.MateIn = e.MateIn
			existing.BestMove = e.BestMove
			continue
		}
		row := e
		row.GameID = gid
		if row.ID == uuid.Nil {
			row.ID = uuid.New()
		}
		if row.CreatedAt.IsZero() {
			row.CreatedAt = now
		}
		s.evals[gid] = append(s.evals[gid], row)
		s.evalIx[gid][row.Ply] = len(s.evals[gid]) - 1
	}
	return nil
}
