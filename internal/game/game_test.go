package game

import (
	"testing"
	"time"

	"github.com/notnil/chess"
)

// helper to create a new Game with necessary fields
func newTestGame() *Game {
	return &Game{
		g:         chess.NewGame(),
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

func TestMakeMoveInvalid(t *testing.T) {
	g := newTestGame()
	if err := g.MakeMove("e2e5"); err == nil {
		t.Fatalf("expected error for illegal move, got nil")
	}
}
