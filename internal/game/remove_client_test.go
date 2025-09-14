package game

import (
	"testing"

	"github.com/corentings/chess/v2"
)

func TestRemoveClient(t *testing.T) {
	g := &Game{
		Clients:    make(map[string]chess.Color),
		OwnerID:    "owner",
		OwnerColor: chess.White,
	}
	g.Clients["owner"] = chess.White
	g.Clients["other"] = chess.Black

	g.RemoveClient("other")
	if _, ok := g.Clients["other"]; ok {
		t.Fatalf("expected other client to be removed")
	}
	if g.OwnerID != "owner" {
		t.Fatalf("owner id should remain unchanged")
	}

	g.RemoveClient("owner")
	if g.OwnerID != "" {
		t.Fatalf("owner id should be cleared when owner removed")
	}
	if _, ok := g.Clients["owner"]; ok {
		t.Fatalf("owner should be removed from clients map")
	}
}
