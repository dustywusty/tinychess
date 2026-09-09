package handlers

import (
	"encoding/json"
	"strings"
	"testing"
	"tinychess/internal/game"
)

func botGame(t *testing.T, h *Handler) string {
	t.Helper()
	response := recoveryRequest(h.HandleCreateGame, "", "POST", "/api/games", `{"botId":"ada","color":"b","clientId":"owner"}`)
	var result struct {
		ID string `json:"id"`
	}
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &result) != nil || result.ID == "" {
		t.Fatal(response.Body.String())
	}
	return result.ID
}

func TestBotHTTPPipeline(t *testing.T) {
	h := NewHandler(game.NewHub())
	id := botGame(t, h)
	spectator := recoverySnapshot(t, h, id, "visitor")
	if spectator.Role != "spectator" || spectator.Bot == nil || spectator.Bot.ID != "ada" {
		t.Fatal("incorrect bot snapshot")
	}
	for _, input := range []string{`{"botId":"nope","clientId":"owner","color":"w"}`, `{"botId":"pip","color":"w"}`, `{"botId":"pip","clientId":"owner","color":"red"}`, `{`} {
		if response := recoveryRequest(h.HandleCreateGame, "", "POST", "/api/games", input); response.Code != 400 {
			t.Fatal("accepted malformed create")
		}
	}
	response := recoveryRequest(h.HandleMove, id, "POST", "/move", `{"clientId":"owner","uci":"e2e4","botMove":true,"expectedPly":0}`)
	if !strings.Contains(response.Body.String(), `"ok":true`) {
		t.Fatal(response.Body.String())
	}
	requireMove(t, h, id, "owner", "e7e5")
	response = recoveryRequest(h.HandleRelease, id, "POST", "/release", `{"clientId":"owner","targetId":"owner"}`)
	if !strings.Contains(response.Body.String(), `"ok":false`) {
		t.Fatal("bot owner can be released")
	}
}

func TestPostgresBotRestart(t *testing.T) {
	db := recoveryDB(t)
	first := NewHandlerWithStore(game.NewHub(), db)
	id := botGame(t, first)
	next := NewHandlerWithStore(game.NewHub(), db)
	state := recoverySnapshot(t, next, id, "visitor")
	if state.Bot == nil || state.Bot.Color != "w" || state.Role != "spectator" {
		t.Fatal("bot game not restored")
	}
	response := recoveryRequest(next.HandleMove, id, "POST", "/move", `{"clientId":"owner","uci":"e2e4","botMove":true,"expectedPly":0}`)
	if !strings.Contains(response.Body.String(), `"ok":true`) {
		t.Fatal(response.Body.String())
	}
	restored := NewHandlerWithStore(game.NewHub(), db)
	state = recoverySnapshot(t, restored, id, "owner")
	if len(state.UCI) != 1 || state.Role != "player" || !strings.Contains(state.PGN, "e4") {
		t.Fatal("bot move history not persisted")
	}
}
