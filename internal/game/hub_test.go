package game

import (
	"testing"
	"time"

	"github.com/corentings/chess/v2"
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
	g, _ := h.Get("test", "")

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
	g, _ := h.Get("g1", "owner")

	if g.OwnerID != "owner" {
		t.Fatalf("expected owner id to be set")
	}
	ownerColor := g.OwnerColor
	if c, ok := g.Clients["owner"]; !ok || c != ownerColor {
		t.Fatalf("owner not recorded with correct color")
	}

	g, _ = h.Get("g1", "client2")
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

func TestTwoClientsReceiveOppositeColors(t *testing.T) {
	h := NewHub()
	g, _ := h.Get("g2", "c1")
	g, _ = h.Get("g2", "c2")

	c1 := g.Clients["c1"]
	c2 := g.Clients["c2"]

	if (c1 != chess.White && c1 != chess.Black) || (c2 != chess.White && c2 != chess.Black) {
		t.Fatalf("clients received invalid colors: %v and %v", c1, c2)
	}
	if c1 == c2 {
		t.Fatalf("expected clients to have opposite colors, both got %v", c1)
	}

	g, _ = h.Get("g2", "")
	if len(g.Clients) != 2 {
		t.Fatalf("spectator should not be assigned a color")
	}
}

func TestColorPersistsAfterOwnerLeaves(t *testing.T) {
	h := NewHub()

	// owner joins
	g, _ := h.Get("g3", "owner")
	// second player joins
	g, _ = h.Get("g3", "player")

	initialColor := g.Clients["player"]

	// owner releases themselves
	g.RemoveClient("owner")

	// player refreshes (rejoins)
	g, col := h.Get("g3", "player")
	if col == nil || *col != initialColor {
		t.Fatalf("expected player to retain color %v, got %v", initialColor, col)
	}

	if g.OwnerID != "player" {
		t.Fatalf("expected player to become owner after release")
	}

	if g.OwnerColor != initialColor {
		t.Fatalf("owner color not updated to player's color")
	}

	// new client should receive opposite color
	g, col2 := h.Get("g3", "newbie")
	if col2 == nil || *col2 == initialColor {
		t.Fatalf("expected new client to receive opposite color")
	}
}
