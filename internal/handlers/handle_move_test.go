package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"tinychess/internal/game"

	"github.com/corentings/chess/v2"
)

// Test that a move is rejected when the piece is not of the player's color.
func TestHandleMoveWrongColor(t *testing.T) {
	hub := game.NewHub()
	h := NewHandler(hub)
	g, _ := hub.Get("g1", "")
	g.Clients["c1"] = chess.White

	req := httptest.NewRequest("POST", "/api/games/g1/move", strings.NewReader(`{"uci":"a7a6","clientId":"c1"}`))
	req.SetPathValue("gameId", "g1")
	w := httptest.NewRecorder()
	h.HandleMove(w, req)

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["ok"].(bool) {
		t.Fatalf("expected move to be rejected")
	}
}

// Test that a move is rejected when it is not the player's turn.
func TestHandleMoveNotYourTurn(t *testing.T) {
	hub := game.NewHub()
	h := NewHandler(hub)
	g, _ := hub.Get("g2", "")
	g.Clients["c2"] = chess.Black

	req := httptest.NewRequest("POST", "/api/games/g2/move", strings.NewReader(`{"uci":"a7a6","clientId":"c2"}`))
	req.SetPathValue("gameId", "g2")
	w := httptest.NewRecorder()
	h.HandleMove(w, req)

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["ok"].(bool) {
		t.Fatalf("expected move to be rejected")
	}
}

// Test that a valid move by the correct player succeeds.
func TestHandleMoveSuccess(t *testing.T) {
	hub := game.NewHub()
	h := NewHandler(hub)
	g, _ := hub.Get("g3", "")
	g.Clients["c1"] = chess.White

	req := httptest.NewRequest("POST", "/api/games/g3/move", strings.NewReader(`{"uci":"e2e4","clientId":"c1"}`))
	req.SetPathValue("gameId", "g3")
	w := httptest.NewRecorder()
	h.HandleMove(w, req)

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp["ok"].(bool) {
		t.Fatalf("expected move to succeed")
	}
}
