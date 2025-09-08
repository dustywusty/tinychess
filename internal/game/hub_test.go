package game

import (
	"testing"
	"time"

	"github.com/notnil/chess"
)

// runCleanup mimics the hub's cleanup routine for testing purposes.
func runCleanup(h *Hub) {
	h.Mu.Lock()
	for id, game := range h.Games {
		game.Mu.Lock()
		idle := time.Since(game.LastSeen) > 24*time.Hour
		game.Mu.Unlock()
		if idle {
			delete(h.Games, id)
		}
	}
	h.Mu.Unlock()
}

func TestGamePersistenceBeforeCleanup(t *testing.T) {
	h := NewHub()
	g := h.Get("test", "")

	// Simulate a game that was last seen 23 hours ago.
	g.Mu.Lock()
	g.LastSeen = time.Now().Add(-23 * time.Hour)
	g.Mu.Unlock()

	runCleanup(h)

	h.Mu.Lock()
	_, exists := h.Games["test"]
	h.Mu.Unlock()
	if !exists {
		t.Fatalf("game removed before 24 hours of inactivity")
	}

	// Simulate a game that was last seen 25 hours ago.
	g.Mu.Lock()
	g.LastSeen = time.Now().Add(-25 * time.Hour)
	g.Mu.Unlock()

	runCleanup(h)

	h.Mu.Lock()
	_, exists = h.Games["test"]
	h.Mu.Unlock()
	if exists {
		t.Fatalf("game not removed after 24 hours of inactivity")
	}
}

func TestOwnerAndClientColorAssignment(t *testing.T) {
	h := NewHub()
	g := h.Get("g1", "owner")

	if g.OwnerID != "owner" {
		t.Fatalf("expected owner id to be set")
	}
	ownerColor := g.OwnerColor
	if c, ok := g.Clients["owner"]; !ok || c != ownerColor {
		t.Fatalf("owner not recorded with correct color")
	}

	g = h.Get("g1", "client2")
	var expected chess.Color
	if ownerColor == chess.White {
		expected = chess.Black
	} else {
		expected = chess.White
	}
	if c := g.Clients["client2"]; c != expected {
		t.Fatalf("expected client2 color %v, got %v", expected, c)
	}
}

func TestSpectatorReceivesNoColor(t *testing.T) {
	h := NewHub()
	g := h.Get("g2", "p1")
	g = h.Get("g2", "p2")
	g = h.Get("g2", "p3")

	g.Mu.Lock()
	c1, ok1 := g.Clients["p1"]
	c2, ok2 := g.Clients["p2"]
	_, ok3 := g.Clients["p3"]
	g.Mu.Unlock()

	if !ok1 || !ok2 {
		t.Fatalf("expected first two clients to have colors assigned")
	}
	if c1 == c2 {
		t.Fatalf("expected players to have opposite colors")
	}
	if ok3 {
		t.Fatalf("third client should be a spectator without color")
	}
}
