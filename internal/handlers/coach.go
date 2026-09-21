package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"unicode/utf8"

	"tinychess/internal/coach"
	"tinychess/internal/game"
)

func (h *Handler) coachRequest(ctx context.Context, id string, request coach.Request) (coach.Projection, int, string) {
	if h.Coach == nil {
		return coach.Projection{}, 404, "Coach is not enabled."
	}
	if strings.TrimSpace(request.ClientID) == "" || utf8.RuneCountInString(request.Text) > 240 {
		return coach.Projection{}, 400, "Invalid coach request."
	}
	history, status, message := h.coachContext(ctx, id, request.ClientID)
	if status != 200 {
		return coach.Projection{}, status, message
	}
	var err error
	if request.Action == "language" && request.LessonID == "" {
		if h.Coach.Router == nil {
			return coach.Projection{}, 503, "Language controls are unavailable; use Review move."
		}
		selected, routeErr := h.Coach.InitialIntent(ctx, id, request)
		if routeErr != nil {
			return coach.Projection{}, 422, routeErr.Error()
		}
		if selected != "review" && selected != "hint" && selected != "best_now" {
			return coach.Projection{}, 422, "Select a played move, then ask about that move."
		}
		request.Action = "review"
		if selected == "best_now" {
			request.Action = selected
		}
		request.NoSpoilers = request.NoSpoilers || selected == "hint"
	}
	var result coach.Projection
	if request.Action == "review" || request.Action == "best_now" {
		result, err = h.Coach.Review(ctx, id, history, request)
	} else {
		result, err = h.Coach.Action(ctx, id, history, request)
	}
	if err != nil {
		return coach.Projection{}, 422, err.Error()
	}
	// Analysis may outlive a seat change or a durable state replacement.
	latest, status, message := h.coachContext(ctx, id, request.ClientID)
	if status != 200 {
		return coach.Projection{}, status, message
	}
	end := result.Ply + 1
	if result.CurrentPosition {
		end = result.Ply
	}
	if end < 0 || end > len(latest) || coach.Identity(latest[:end]) != result.PositionID {
		return coach.Projection{}, 409, "The reviewed position changed. Start a new review."
	}
	return result, 200, ""
}

// HandleCoach rechecks eligibility before both analysis and cached lesson access.
func (h *Handler) HandleCoach(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	var request coach.Request
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&request) != nil {
		WriteJSON(w, 400, map[string]string{"error": "Invalid coach request."})
		return
	}
	result, status, message := h.coachRequest(r.Context(), r.PathValue("gameId"), request)
	if status != 200 {
		WriteJSON(w, status, map[string]string{"error": message})
		return
	}
	WriteJSON(w, 200, result)
}

func (h *Handler) coachContext(ctx context.Context, id, clientID string) ([]string, int, string) {
	allowed := false
	var history []string
	_, err := h.operate(ctx, id, false, func(g *game.Game) bool {
		g.Mu.Lock()
		defer g.Mu.Unlock()
		_, seated := g.Clients[clientID]
		state := g.StateLocked()
		allowed = seated && (state.Bot != nil || state.Status != "")
		if allowed {
			history = append([]string{}, state.UCI...)
		}
		return false
	})
	if err != nil {
		return nil, 503, "Game storage is unavailable. Reconnect before trying again."
	}
	if !allowed {
		return nil, 403, "Coaching is available to seated players in computer games or after the game."
	}

	return history, 200, ""
}
