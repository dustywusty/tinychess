package game

import (
	"strings"
	"testing"
	"time"

	"github.com/notnil/chess"
)

// helper to create a new Game with necessary fields
func newTestGame() *Game {
	return &Game{
		g:         chess.NewGame(chess.UseNotation(chess.UCINotation{})),
		Watchers:  make(map[chan []byte]struct{}),
		LastReact: make(map[string]time.Time),
	}
}

func TestMakeMoveValid(t *testing.T) {
	g := newTestGame()
	if err := g.MakeMove("e2e4"); err != nil {
		t.Fatalf("expected move to be valid, got error: %v", err)
	}
}

func TestMakeMoveIllegal(t *testing.T) {
	g := newTestGame()
	if err := g.MakeMove("e2e5"); err == nil {
		t.Fatalf("expected error for illegal move, got nil")
	}
}

func TestMakeMoveInvalidUCI(t *testing.T) {
	g := newTestGame()
	if err := g.MakeMove("invalid"); err == nil {
		t.Fatalf("expected error for invalid UCI, got nil")
	}
}

func TestCheckmateState(t *testing.T) {
	g := newTestGame()
	moves := []string{"f2f3", "e7e5", "g2g4", "d8h4"}
	for _, m := range moves {
		if err := g.MakeMove(m); err != nil {
			t.Fatalf("move %s failed: %v", m, err)
		}
	}
	g.Mu.Lock()
	st := g.StateLocked()
	g.Mu.Unlock()
	if st.Status == "" {
		t.Fatalf("expected status to be set after checkmate")
	}
	if !strings.Contains(strings.ToLower(st.Status), "checkmate") {
		t.Fatalf("expected checkmate in status, got %s", st.Status)
	}
}
