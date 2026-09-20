package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"tinychess/internal/storage"
)

const publicGamePageSize = 8

// HandleListGames lists service-wide activity without touching games or seats.
func (h *Handler) HandleListGames(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status != "active" && status != "completed" {
		http.Error(w, "status must be active or completed", http.StatusBadRequest)
		return
	}
	offset := 0
	if raw := r.URL.Query().Get("offset"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 || value > 1000000 {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}
		offset = value
	}
	cutoff := time.Now().Add(-30 * 24 * time.Hour)
	completed := status == "completed"
	rows := make([]storage.PublicGame, 0)
	more := false
	if h.DB != nil {
		var err error
		rows, more, err = storage.ListGames(h.DB.WithContext(r.Context()), completed, cutoff, offset, publicGamePageSize)
		if err != nil {
			http.Error(w, "storage error", http.StatusInternalServerError)
			return
		}
	} else {
		h.Hub.Mu.Lock()
		for id, g := range h.Hub.Games {
			g.Mu.Lock()
			state := g.StateLocked()
			updatedAt := g.LastSeen
			g.Mu.Unlock()
			finished := state.Status != ""
			if finished != completed || (!finished && updatedAt.Before(cutoff)) {
				continue
			}
			rows = append(rows, storage.PublicGame{ID: id, Result: resultFromPGN(state.PGN), MoveCount: len(state.UCI), UpdatedAt: updatedAt})
		}
		h.Hub.Mu.Unlock()
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].UpdatedAt.Equal(rows[j].UpdatedAt) {
				return rows[i].ID < rows[j].ID
			}
			return rows[i].UpdatedAt.After(rows[j].UpdatedAt)
		})
		start := min(offset, len(rows))
		end := min(start+publicGamePageSize, len(rows))
		more = end < len(rows)
		rows = rows[start:end]
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteJSON(w, http.StatusOK, struct {
		Games   []storage.PublicGame `json:"games"`
		HasMore bool                 `json:"hasMore"`
	}{rows, more})
}
