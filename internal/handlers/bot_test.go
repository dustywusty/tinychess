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

func TestBotBattleHTTPSettings(t *testing.T) {
	h := NewHandler(game.NewHub())
	for _, input := range []string{
		`{"botId":"pip","clientId":"owner","color":"w","playerBotId":"unknown"}`,
		`{"botId":"pip","clientId":"owner","color":"w","moveDelayMs":-1}`,
		`{"botId":"pip","clientId":"owner","color":"w","moveDelayMs":5001}`,
		`{"botId":"pip","clientId":"owner","color":"w","moveDelayMs":1.5}`,
		`{"playerBotId":"pip"}`, `{"moveDelayMs":0}`,
	} {
		if response := recoveryRequest(h.HandleCreateGame, "", "POST", "/api/games", input); response.Code != 400 {
			t.Fatalf("accepted invalid settings: %s", input)
		}
	}
	response := recoveryRequest(h.HandleCreateGame, "", "POST", "/api/games", `{"botId":"ada","clientId":"owner","color":"w","playerBotId":"max","moveDelayMs":0}`)
	if response.Code != 200 {
		t.Fatal(response.Body.String())
	}
	var created struct{ ID string }
	_ = json.Unmarshal(response.Body.Bytes(), &created)
	state := recoverySnapshot(t, h, created.ID, "visitor")
	if state.Role != "spectator" || state.Bot.PlayerBotID != "max" || state.Bot.MoveDelayMs == nil || *state.Bot.MoveDelayMs != 0 {
		t.Fatal("battle settings missing from snapshot")
	}
	for _, input := range []string{
		`{"clientId":"owner","uci":"e2e4","botMove":true,"expectedPly":0}`,
		`{"clientId":"owner","uci":"e7e5","botMove":true,"expectedPly":1}`,
	} {
		response = recoveryRequest(h.HandleMove, created.ID, "POST", "/move", input)
		if response.Code != 200 || !strings.Contains(response.Body.String(), `"ok":true`) {
			t.Fatal(response.Body.String())
		}
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
