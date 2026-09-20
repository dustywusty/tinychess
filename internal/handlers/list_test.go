package handlers

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"tinychess/internal/game"
	"tinychess/internal/storage"

	"github.com/google/uuid"
)

func listRequest(t *testing.T, h *Handler, query string) ([]storage.PublicGame, bool) {
	t.Helper()
	w := httptest.NewRecorder()
	h.HandleListGames(w, httptest.NewRequest("GET", "/api/games?"+query, nil))
	if w.Code != 200 {
		t.Fatalf("list: %d %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "secret") || strings.Contains(w.Body.String(), "liveState") {
		t.Fatal("private data leaked")
	}
	var response struct {
		Games   []storage.PublicGame
		HasMore bool
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	return response.Games, response.HasMore
}

func TestListGamesMemory(t *testing.T) {
	hub := game.NewHub()
	h := NewHandler(hub)
	now := time.Now()
	for i := 0; i < 10; i++ {
		g := game.NewGame()
		g.LastSeen = now.Add(-time.Duration(i) * time.Hour)
		g.AssignClient("secret")
		hub.Games[fmt.Sprintf("game%d", i)] = g
	}
	old := game.NewGame()
	old.LastSeen = now.Add(-31 * 24 * time.Hour)
	hub.Games["old"] = old
	finished := game.NewGame()
	for _, move := range []string{"f2f3", "e7e5", "g2g4", "d8h4"} {
		if err := finished.MakeMove(move); err != nil {
			t.Fatal(err)
		}
	}
	finished.LastSeen = old.LastSeen
	hub.Games["finished"] = finished
	rows, more := listRequest(t, h, "status=active")
	if len(rows) != 8 || !more || rows[0].ID != "game0" {
		t.Fatalf("first page: %+v %v", rows, more)
	}
	rows, more = listRequest(t, h, "status=active&offset=8")
	if len(rows) != 2 || more || rows[0].ID != "game8" {
		t.Fatalf("second page: %+v %v", rows, more)
	}
	rows, _ = listRequest(t, h, "status=completed")
	if len(rows) != 1 || rows[0].ID != "finished" || rows[0].Result != "0-1" {
		t.Fatalf("completed: %+v", rows)
	}
	if !hub.Games["game0"].LastSeen.Equal(now) || len(hub.Games["game0"].Clients) != 1 {
		t.Fatal("listing mutated game")
	}
	for _, query := range []string{"", "status=unknown", "status=active&offset=-1", "status=active&offset=nope"} {
		w := httptest.NewRecorder()
		h.HandleListGames(w, httptest.NewRequest("GET", "/api/games?"+query, nil))
		if w.Code != 400 {
			t.Fatalf("invalid query accepted: %s", query)
		}
	}
}

func TestListGamesPostgres(t *testing.T) {
	db := recoveryDB(t)
	now := time.Now().UTC().Truncate(time.Second)
	cutoff := now.Add(-30 * 24 * time.Hour)
	secret := `{"ownerId":"secret"}`
	active := storage.Game{ID: uuid.New(), StartedAt: cutoff, UpdatedAt: cutoff, LiveState: &secret}
	old := storage.Game{ID: uuid.New(), StartedAt: cutoff, UpdatedAt: cutoff.Add(-time.Second)}
	finished := storage.Game{ID: uuid.New(), StartedAt: cutoff, UpdatedAt: cutoff.Add(-time.Hour), EndedAt: &cutoff, Result: "1-0"}
	for _, row := range []storage.Game{active, old, finished} {
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	rows, more, err := storage.ListGames(db, false, cutoff, 0, 8)
	if err != nil || more || len(rows) != 1 || rows[0].ID != active.ID.String() {
		t.Fatalf("cutoff: %+v %v %v", rows, more, err)
	}
	rows, _, err = storage.ListGames(db, true, cutoff, 0, 8)
	if err != nil || len(rows) != 1 || rows[0].ID != finished.ID.String() {
		t.Fatalf("completed: %+v %v", rows, err)
	}
	h := NewHandlerWithStore(game.NewHub(), db)
	listRequest(t, h, "status=completed")
}
