package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"tinychess/internal/coach"
)

// HandleCoachChat streams coach responses to the browser. Body is the
// Hashbrown chat request JSON; response is a sequence of length-prefixed
// JSON frames (see coach.WriteFrame). The route returns 503 if no provider
// is configured (e.g. ANTHROPIC_API_KEY missing).
func (h *Handler) HandleCoachChat(w http.ResponseWriter, r *http.Request) {
	if h.Coach == nil {
		http.Error(w, "coach not configured", http.StatusServiceUnavailable)
		return
	}
	var req coach.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	gctx := h.loadGameContext(r.Context(), req.GameID, req.ClientID)

	frames, err := h.Coach.StreamChat(r.Context(), req, gctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := coach.StreamFrames(w, frames); err != nil {
		// connection lost — nothing useful to report
		return
	}
}

func (h *Handler) loadGameContext(ctx context.Context, gameID, clientID string) *coach.GameContext {
	if gameID == "" {
		return nil
	}
	g, err := h.Store.GetGame(ctx, gameID)
	if err != nil {
		return nil
	}
	playerColor := "spectator"
	if clientID != "" {
		if g.WhiteSession == clientID {
			playerColor = "white"
		} else if g.BlackSession == clientID {
			playerColor = "black"
		}
	}
	evals, _ := h.Store.GetEvals(ctx, gameID)
	out := &coach.GameContext{
		GameID:       g.ID.String(),
		PGN:          g.PGN,
		Result:       g.Result,
		PlayerColor:  playerColor,
		WhiteSession: g.WhiteSession,
		BlackSession: g.BlackSession,
	}
	for _, e := range evals {
		out.Evals = append(out.Evals, coach.EvalEntry{
			Ply:      e.Ply,
			FEN:      e.FEN,
			EvalCp:   e.EvalCp,
			MateIn:   e.MateIn,
			BestMove: e.BestMove,
		})
	}
	return out
}
