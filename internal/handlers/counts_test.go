package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"tinychess/internal/game"
	"tinychess/internal/storage"

	"github.com/google/uuid"
)

func assertGameCounts(t *testing.T, h *Handler, active, completed float64) {
	t.Helper()
	w := httptest.NewRecorder()
	h.HandleGameCounts(w, httptest.NewRequest("GET", "/api/stats/games", nil))
	if w.Code != 200 {
		t.Fatalf("counts: %d %s", w.Code, w.Body.String())
	}
	var response map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	// Exact response shape prevents accidentally exposing individual game data.
	want := map[string]any{"active": active, "completed": completed}
	if !reflect.DeepEqual(response, want) {
		t.Fatalf("counts: got %+v, want %+v", response, want)
	}
}

func TestGameCountsMemory(t *testing.T) {
	hub := game.NewHub()
	h := NewHandler(hub)
	assertGameCounts(t, h, 0, 0)
	now := time.Now()
	active := game.NewGame()
	active.LastSeen = now.Add(-29 * 24 * time.Hour)
	active.AssignClient("secret")
	hub.Games["private-active-id"] = active
	old := game.NewGame()
	old.LastSeen = now.Add(-31 * 24 * time.Hour)
	hub.Games["private-old-id"] = old
	finished := game.NewGame()
	for _, move := range []string{"f2f3", "e7e5", "g2g4", "d8h4"} {
		if err := finished.MakeMove(move); err != nil {
			t.Fatal(err)
		}
	}
	finished.LastSeen = old.LastSeen
	hub.Games["private-finished-id"] = finished
	assertGameCounts(t, h, 1, 1)
	if !active.LastSeen.Equal(now.Add(-29*24*time.Hour)) || len(active.Clients) != 1 {
		t.Fatal("counting mutated game")
	}
}

func TestGameCountsPostgres(t *testing.T) {
	db := recoveryDB(t)
	assertGameCounts(t, NewHandlerWithStore(game.NewHub(), db), 0, 0)
	cutoff := time.Now().UTC().Truncate(time.Second).Add(-30 * 24 * time.Hour)
	secret := `{"ownerId":"secret"}`
	rows := []storage.Game{
		{ID: uuid.New(), StartedAt: cutoff, UpdatedAt: cutoff, LiveState: &secret},
		{ID: uuid.New(), StartedAt: cutoff, UpdatedAt: cutoff.Add(-time.Second)},
		{ID: uuid.New(), StartedAt: cutoff, UpdatedAt: cutoff.Add(-time.Hour), EndedAt: &cutoff, Result: "1-0"},
		{ID: uuid.New(), StartedAt: cutoff, UpdatedAt: cutoff, Result: "1/2-1/2"},
		{ID: uuid.New(), StartedAt: cutoff, UpdatedAt: cutoff, EndedAt: &cutoff},
	}
	for _, row := range rows {
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}
	counts, err := storage.CountGames(db, cutoff)
	if err != nil || counts.Active != 1 || counts.Completed != 3 {
		t.Fatalf("cutoff: %+v %v", counts, err)
	}
	assertGameCounts(t, NewHandlerWithStore(game.NewHub(), db), 0, 3)
}
