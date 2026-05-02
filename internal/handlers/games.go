package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"tinychess/internal/storage"
)

// GameDTO is the JSON shape returned by GET /api/games/:id.
type GameDTO struct {
	ID         string `json:"id"`
	FEN        string `json:"fen"`
	PGN        string `json:"pgn"`
	Result     string `json:"result"`
	StartedAt  string `json:"startedAt"`
	EndedAt    string `json:"endedAt,omitempty"`
	MoveCount  int    `json:"moveCount"`
	WhiteSession string `json:"whiteSession,omitempty"`
	BlackSession string `json:"blackSession,omitempty"`
}

// EvalDTO is the JSON shape returned by GET /api/games/:id/evals.
type EvalDTO struct {
	Ply      int    `json:"ply"`
	FEN      string `json:"fen"`
	EvalCp   *int   `json:"evalCp,omitempty"`
	MateIn   *int   `json:"mateIn,omitempty"`
	BestMove string `json:"bestMove,omitempty"`
}

// HandleGetGame returns persisted metadata for a game (PGN, result, etc.)
// from the storage layer. Returns 404 if the game is not yet persisted.
func (h *Handler) HandleGetGame(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("gameId")
	g, err := h.Store.GetGame(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	dto := GameDTO{
		ID:           g.ID.String(),
		FEN:          g.FEN,
		PGN:          g.PGN,
		Result:       g.Result,
		StartedAt:    g.StartedAt.UTC().Format("2006-01-02T15:04:05Z"),
		MoveCount:    g.MoveCount,
		WhiteSession: g.WhiteSession,
		BlackSession: g.BlackSession,
	}
	if g.EndedAt != nil {
		dto.EndedAt = g.EndedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	WriteJSON(w, http.StatusOK, dto)
}

// HandleGetEvals returns the cached eval array for a game.
func (h *Handler) HandleGetEvals(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("gameId")
	rows, err := h.Store.GetEvals(r.Context(), id)
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	out := make([]EvalDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, EvalDTO{
			Ply:      row.Ply,
			FEN:      row.FEN,
			EvalCp:   row.EvalCp,
			MateIn:   row.MateIn,
			BestMove: row.BestMove,
		})
	}
	WriteJSON(w, http.StatusOK, out)
}

// HandleAppendEvals upserts a batch of eval rows for a game. Idempotent on
// (gameId, ply).
func (h *Handler) HandleAppendEvals(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("gameId")
	var body []EvalDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	rows := make([]storage.GameEval, 0, len(body))
	for _, e := range body {
		rows = append(rows, storage.GameEval{
			Ply:      e.Ply,
			FEN:      e.FEN,
			EvalCp:   e.EvalCp,
			MateIn:   e.MateIn,
			BestMove: e.BestMove,
		})
	}
	if err := h.Store.AppendEvals(r.Context(), id, rows); err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "count": len(rows)})
}
