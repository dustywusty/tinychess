package game

import (
	"testing"
	"time"
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
	g := h.Get("test")

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
