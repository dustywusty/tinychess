package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"tinychess/internal/game"
	"tinychess/internal/storage"
)

// operate is the only path to live state in database mode. The hub remains a
// local fan-out cache, never a fallback authority when Postgres is unavailable.
func (h *Handler) operate(ctx context.Context, id string, create bool, action func(*game.Game) bool) (*game.Game, error) {
	g, _ := h.Hub.Get(id, "")
	g.OpMu.Lock()
	defer g.OpMu.Unlock()
	publish := false
	apply := func(candidate *game.Game) {
		if action != nil {
			publish = action(candidate)
		}
	}
	if h.DB == nil {
		apply(g)
	} else {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		candidate, err := storage.TransactGame(ctx, h.DB, id, create, apply)
		if err != nil {
			return nil, err
		}
		g.ApplyCommitted(candidate)
	}
	if publish {
		g.Broadcast()
	}
	return g, nil
}

func gameStorageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrNotFound):
		WriteJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "Game not found."})
	case errors.Is(err, storage.ErrUnrecoverable):
		WriteJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "This game cannot be restored. Its saved history has not been changed."})
	default:
		// Do not send database errors or credentials to clients.
		WriteJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "Game storage is unavailable. Reconnect before trying again."})
	}
}
