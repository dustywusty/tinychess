package handlers

import (
	"tinychess/internal/game"
	"tinychess/internal/storage"
)

// newTestHandler returns a Handler bound to an in-memory store. Tests use
// non-UUID game IDs ("g1", "g2", …); memStore silently no-ops on those for
// writes and returns ErrNotFound for reads, matching gormStore's behavior.
func newTestHandler(hub *game.Hub) *Handler {
	return NewHandler(hub, storage.NewMemStore())
}
