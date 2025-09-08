package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"tinychess/internal/game"

	"github.com/notnil/chess"
)

func TestHandleRelease(t *testing.T) {
	hub := game.NewHub()
	h := NewHandler(hub)
	g, _ := hub.Get("g1", "owner")
	g.Clients["other"] = chess.Black

	req := httptest.NewRequest("POST", "/release/g1", strings.NewReader(`{"clientId":"owner","targetId":"other"}`))
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
	h := NewHandler(hub)
	g, _ := hub.Get("g2", "owner")
	g.Clients["other"] = chess.Black

	req := httptest.NewRequest("POST", "/release/g2", strings.NewReader(`{"clientId":"notowner","targetId":"other"}`))
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
