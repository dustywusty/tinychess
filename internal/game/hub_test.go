package game

import (
	"context"
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
	h := NewHub(nil)
	g, _, err := h.Get(context.Background(), "test", "")
	if err != nil {
		t.Fatalf("get: %v", err)
	}

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
	h := NewHub(nil)
	g, _, err := h.Get(context.Background(), "g1", "owner")
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if g.OwnerID != "owner" {
		t.Fatalf("expected owner id to be set")
	}
	ownerColor := g.OwnerColor
	if c, ok := g.Clients["owner"]; !ok || c != ownerColor {
		t.Fatalf("owner not recorded with correct color")
	}

	g, _, err = h.Get(context.Background(), "g1", "client2")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
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
	h := NewHub(nil)
	if _, _, err := h.Get(context.Background(), "g2", "c1"); err != nil {
		t.Fatalf("get: %v", err)
	}
	g, _, err := h.Get(context.Background(), "g2", "c2")
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	c1 := g.Clients["c1"]
	c2 := g.Clients["c2"]

	if (c1 != chess.White && c1 != chess.Black) || (c2 != chess.White && c2 != chess.Black) {
		t.Fatalf("clients received invalid colors: %v and %v", c1, c2)
	}
	if c1 == c2 {
		t.Fatalf("expected clients to have opposite colors, both got %v", c1)
	}

	g, _, err = h.Get(context.Background(), "g2", "")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(g.Clients) != 2 {
		t.Fatalf("spectator should not be assigned a color")
	}
}

func TestColorPersistsAfterOwnerLeaves(t *testing.T) {
	h := NewHub(nil)

	// owner joins
	if _, _, err := h.Get(context.Background(), "g3", "owner"); err != nil {
		t.Fatalf("get: %v", err)
	}
	// second player joins
	g, _, err := h.Get(context.Background(), "g3", "player")
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	initialColor := g.Clients["player"]

	// owner releases themselves
	g.RemoveClient("owner")

	// player refreshes (rejoins)
	g, col, err := h.Get(context.Background(), "g3", "player")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
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
	_, col2, err := h.Get(context.Background(), "g3", "newbie")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if col2 == nil || *col2 == initialColor {
		t.Fatalf("expected new client to receive opposite color")
	}
}
