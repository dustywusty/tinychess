package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"tinychess/internal/game"

	"github.com/corentings/chess/v2"
)

func TestHandleRelease(t *testing.T) {
	hub := game.NewHub()
	h := newTestHandler(hub)
	g, _ := hub.Get("g1", "owner")
	g.Clients["other"] = chess.Black

	req := httptest.NewRequest("POST", "/api/games/g1/release", strings.NewReader(`{"clientId":"owner","targetId":"other"}`))
	req.SetPathValue("gameId", "g1")
	w := httptest.NewRecorder()
	h.HandleRelease(w, req)

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp["ok"].(bool) {
		t.Fatalf("expected ok true")
	}
	if _, exists := g.Clients["other"]; exists {
		t.Fatalf("expected client to be removed")
	}
}

func TestHandleReleaseNotOwner(t *testing.T) {
	hub := game.NewHub()
	h := newTestHandler(hub)
	g, _ := hub.Get("g2", "owner")
	g.Clients["other"] = chess.Black

	req := httptest.NewRequest("POST", "/api/games/g2/release", strings.NewReader(`{"clientId":"notowner","targetId":"other"}`))
	req.SetPathValue("gameId", "g2")
	w := httptest.NewRecorder()
	h.HandleRelease(w, req)

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["ok"].(bool) {
		t.Fatalf("expected ok false")
	}
	if _, exists := g.Clients["other"]; !exists {
		t.Fatalf("client should still be present")
	}
}
