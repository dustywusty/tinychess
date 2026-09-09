package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"tinychess/internal/game"
)

func TestHandleSnapshotAssignsPlayersThenSpectator(t *testing.T) {
	h := NewHandler(game.NewHub())

	roles := []string{"player", "player", "spectator"}
	for index, wantRole := range roles {
		req := httptest.NewRequest("GET", "/api/games/g1/snapshot?clientId=c"+string(rune('1'+index)), nil)
		req.SetPathValue("gameId", "g1")
		response := httptest.NewRecorder()
		h.HandleSnapshot(response, req)

		var state game.ClientState
		if err := json.NewDecoder(response.Body).Decode(&state); err != nil {
			t.Fatalf("decode snapshot %d: %v", index, err)
		}
		if state.Role != wantRole {
			t.Fatalf("snapshot %d role = %q, want %q", index, state.Role, wantRole)
		}
		if wantRole == "player" && state.Color == nil {
			t.Fatalf("snapshot %d missing player color", index)
		}
		if wantRole == "spectator" && state.Color != nil {
			t.Fatalf("snapshot %d spectator received color %q", index, *state.Color)
		}
	}
}

func TestHandleSnapshotReconnectKeepsSeat(t *testing.T) {
	h := NewHandler(game.NewHub())

	request := func() game.ClientState {
		req := httptest.NewRequest("GET", "/api/games/g2/snapshot?clientId=returning", nil)
		req.SetPathValue("gameId", "g2")
		response := httptest.NewRecorder()
		h.HandleSnapshot(response, req)
		var state game.ClientState
		if err := json.NewDecoder(response.Body).Decode(&state); err != nil {
			t.Fatalf("decode: %v", err)
		}
		return state
	}

	first := request()
	second := request()
	if first.Role != "player" || second.Role != "player" || first.Color == nil || second.Color == nil || *first.Color != *second.Color {
		t.Fatalf("seat changed: first=%+v second=%+v", first, second)
	}
}
