package handlers

import (
	"net/http"
	"time"

	"tinychess/internal/storage"
)

// HandleGameCounts returns service-wide totals without exposing individual games
// or changing activity timestamps and player seats.
func (h *Handler) HandleGameCounts(w http.ResponseWriter, r *http.Request) {
	cutoff := time.Now().Add(-30 * 24 * time.Hour)
	var counts storage.GameCounts
	if h.DB != nil {
		var err error
		counts, err = storage.CountGames(h.DB.WithContext(r.Context()), cutoff)
		if err != nil {
			http.Error(w, "storage error", http.StatusInternalServerError)
			return
		}
	} else {
		h.Hub.Mu.Lock()
		for _, g := range h.Hub.Games {
			g.Mu.Lock()
			state := g.StateLocked()
			updatedAt := g.LastSeen
			g.Mu.Unlock()
			if state.Status != "" {
				counts.Completed++
			} else if !updatedAt.Before(cutoff) {
				counts.Active++
			}
		}
		h.Hub.Mu.Unlock()
	}
	WriteJSON(w, http.StatusOK, counts)
}
