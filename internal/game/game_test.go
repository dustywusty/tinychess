package game

import (
	"strings"
	"testing"
	"time"

	"github.com/corentings/chess/v2"
)

// helper to create a new Game with necessary fields
func newTestGame() *Game {
	return &Game{
		g:         chess.NewGame(),
		Watchers:  make(map[chan []byte]struct{}),
		LastReact: make(map[string]time.Time),
		Clients:   make(map[string]chess.Color),
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

func TestMakeMoveForValidatesSeatAndTurn(t *testing.T) {
	g := newTestGame()
	g.Clients["white"] = chess.White
	g.Clients["black"] = chess.Black

	if _, err := g.MakeMoveFor("missing", "e2e4"); err == nil {
		t.Fatal("unknown client move should fail")
	}
	if _, err := g.MakeMoveFor("black", "e7e5"); err == nil {
		t.Fatal("out-of-turn move should fail")
	}
	if accepted, err := g.MakeMoveFor("white", "e2e4"); err != nil || accepted != "e2e4" {
		t.Fatalf("white move = %q, %v", accepted, err)
	}
}

func TestMakeMoveForDefaultsPromotionToQueen(t *testing.T) {
	option, err := chess.FEN("7k/P7/8/8/8/8/8/7K w - - 0 1")
	if err != nil {
		t.Fatalf("fen: %v", err)
	}
	g := newTestGame()
	g.g = chess.NewGame(option)
	g.Clients["white"] = chess.White

	accepted, err := g.MakeMoveFor("white", "a7a8")
	if err != nil {
		t.Fatalf("promotion: %v", err)
	}
	if accepted != "a7a8q" {
		t.Fatalf("accepted = %q, want a7a8q", accepted)
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
