package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"tinychess/internal/game"
	"tinychess/internal/storage"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Only an explicit test database is allowed. Every test owns an isolated schema.
func recoveryDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		t.Fatal("TEST_DATABASE_URL must be a PostgreSQL URL")
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("cannot connect to test database")
	}
	adminPool, err := admin.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { adminPool.Close() })
	schema := "tinychess_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := admin.Exec(`CREATE SCHEMA "` + schema + `"`).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := admin.Exec(`DROP SCHEMA "` + schema + `" CASCADE`).Error; err != nil {
			t.Error(err)
		}
	})
	query := u.Query()
	query.Set("search_path", schema)
	u.RawQuery = query.Encode()
	db, err := storage.New(u.String())
	if err != nil {
		t.Fatal("cannot migrate isolated test schema")
	}
	db = db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })
	return db
}

func recoveryRequest(handler http.HandlerFunc, id, method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.SetPathValue("gameId", id)
	w := httptest.NewRecorder()
	handler(w, r)
	return w
}

func recoveryCreate(t *testing.T, h *Handler) string {
	t.Helper()
	w := recoveryRequest(h.HandleCreateGame, "", "POST", "/api/games", "")
	var result struct {
		ID string `json:"id"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil || result.ID == "" {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	return result.ID
}

func recoverySnapshot(t *testing.T, h *Handler, id, client string) game.ClientState {
	t.Helper()
	w := recoveryRequest(h.HandleSnapshot, id, "GET", "/snapshot?clientId="+client, "")
	var result game.ClientState
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil {
		t.Fatalf("snapshot: %d %s", w.Code, w.Body.String())
	}
	return result
}

func recoveryPlayers(t *testing.T, h *Handler, id string) map[string]string {
	t.Helper()
	players := map[string]string{}
	for _, client := range []string{"owner-secret", "friend-secret"} {
		state := recoverySnapshot(t, h, id, client)
		if state.Color == nil {
			t.Fatal("missing player seat")
		}
		players[*state.Color] = client
	}
	if len(players) != 2 {
		t.Fatal("duplicate seat color")
	}
	return players
}

func recoveryMove(h *Handler, id, client, uci string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(game.MoveRequest{ClientID: client, UCI: uci})
	return recoveryRequest(h.HandleMove, id, "POST", "/move", string(body))
}

func requireMove(t *testing.T, h *Handler, id, client, uci string) {
	t.Helper()
	w := recoveryMove(h, id, client, uci)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"ok":true`) {
		t.Fatalf("move %s: %d %s", uci, w.Code, w.Body.String())
	}
}

func TestPostgresRestartRecovery(t *testing.T) {
	db := recoveryDB(t)
	h := NewHandlerWithStore(game.NewHub(), db)
	id := recoveryCreate(t, h)
	// A game created but not yet visited must also survive a restart.
	h = NewHandlerWithStore(game.NewHub(), db)
	players := recoveryPlayers(t, h, id)
	for i, uci := range []string{"d2d4", "e7e5", "g1f3", "e5d4", "f3d4"} {
		color := "w"
		if i%2 == 1 {
			color = "b"
		}
		requireMove(t, h, id, players[color], uci)
	}
	before := recoverySnapshot(t, h, id, players["w"])
	h = NewHandlerWithStore(game.NewHub(), db)
	spectator := recoverySnapshot(t, h, id, "spectator-arrives-first")
	if spectator.Role != "spectator" || spectator.Color != nil {
		t.Fatal("spectator stole a restored seat")
	}
	after := recoverySnapshot(t, h, id, players["w"])
	if before.FEN != after.FEN || before.PGN != after.PGN || !reflect.DeepEqual(before.UCI, after.UCI) || *after.Color != "w" {
		t.Fatal("game or player identity changed after restart")
	}
	if w := recoveryMove(h, id, "spectator-arrives-first", "b8c6"); !strings.Contains(w.Body.String(), `"ok":false`) {
		t.Fatal("spectator moved")
	}
	requireMove(t, h, id, players["b"], "b8c6")
	var count int64
	if err := db.Model(&storage.Move{}).Where("game_id = ?", id).Count(&count).Error; err != nil || count != 6 {
		t.Fatalf("move rows: %d, %v", count, err)
	}
	review := recoveryRequest(h.HandleGetGame, id, "GET", "/review", "")
	if review.Code != 200 || strings.Contains(review.Body.String(), "secret") || strings.Contains(review.Body.String(), "live_state") {
		t.Fatal("review unavailable or leaked seat credentials")
	}
	// Reconnect via the real streaming endpoint, not only the snapshot endpoint.
	h = NewHandlerWithStore(game.NewHub(), db)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { r.SetPathValue("gameId", id); h.HandleSSE(w, r) }))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, _ := http.NewRequestWithContext(ctx, "GET", server.URL+"?clientId="+players["b"], nil)
	resp, err := server.Client().Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	line, err := bufio.NewReader(resp.Body).ReadString('\n')
	var initial game.ClientState
	if err != nil || json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &initial) != nil || initial.Color == nil || *initial.Color != "b" || len(initial.UCI) != 6 {
		t.Fatalf("SSE recovery: %s, %v", line, err)
	}
}

func TestPostgresConcurrentMoves(t *testing.T) {
	db := recoveryDB(t)
	h := NewHandlerWithStore(game.NewHub(), db)
	id := recoveryCreate(t, h)
	players := recoveryPlayers(t, h, id)
	start := make(chan struct{})
	results := make(chan *httptest.ResponseRecorder, 2)
	var wg sync.WaitGroup
	for _, uci := range []string{"e2e4", "d2d4"} {
		wg.Add(1)
		go func(uci string) {
			defer wg.Done()
			other := NewHandlerWithStore(game.NewHub(), db)
			<-start
			results <- recoveryMove(other, id, players["w"], uci)
		}(uci)
	}
	close(start)
	wg.Wait()
	close(results)
	accepted := 0
	for w := range results {
		if w.Code != 200 {
			t.Fatalf("move response: %d %s", w.Code, w.Body.String())
		}
		if strings.Contains(w.Body.String(), `"ok":true`) {
			accepted++
		}
	}
	if accepted != 1 {
		t.Fatalf("accepted %d simultaneous white moves", accepted)
	}
	if len(recoverySnapshot(t, h, id, players["w"]).UCI) != 1 {
		t.Fatal("lost or duplicated move")
	}
}

func TestPostgresFinishedGameRemainsFinished(t *testing.T) {
	db := recoveryDB(t)
	h := NewHandlerWithStore(game.NewHub(), db)
	id := recoveryCreate(t, h)
	players := recoveryPlayers(t, h, id)
	for i, uci := range []string{"f2f3", "e7e5", "g2g4", "d8h4"} {
		color := "w"
		if i%2 == 1 {
			color = "b"
		}
		requireMove(t, h, id, players[color], uci)
	}
	var before storage.Game
	if err := db.First(&before, "id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	if before.Result != "0-1" || before.EndedAt == nil {
		t.Fatal("missing terminal metadata")
	}
	h = NewHandlerWithStore(game.NewHub(), db)
	state := recoverySnapshot(t, h, id, players["w"])
	if state.Status == "" || len(state.UCI) != 4 {
		t.Fatal("finished game reopened")
	}
	w := recoveryMove(h, id, players["w"], "a2a3")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"ok":false`) {
		t.Fatal("accepted move after checkmate")
	}
	var after storage.Game
	if err := db.First(&after, "id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	if after.EndedAt == nil || !before.EndedAt.Equal(*after.EndedAt) || after.MoveCount != 4 {
		t.Fatal("finished metadata changed")
	}
}

func TestPostgresCommitFailureDoesNotPublish(t *testing.T) {
	db := recoveryDB(t)
	h := NewHandlerWithStore(game.NewHub(), db)
	id := recoveryCreate(t, h)
	players := recoveryPlayers(t, h, id)
	live, _ := h.Hub.Get(id, "")
	ch := make(chan []byte, 4)
	live.AddWatcher(ch)
	// A deferred trigger fails at COMMIT, after both SQL writes succeed.
	for _, sql := range []string{
		`CREATE FUNCTION reject_test_commit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'simulated commit failure'; END $$`,
		`CREATE CONSTRAINT TRIGGER reject_test_commit AFTER INSERT ON moves DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION reject_test_commit()`,
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	w := recoveryMove(h, id, players["w"], "e2e4")
	if w.Code != 503 {
		t.Fatalf("expected failed commit, got %d %s", w.Code, w.Body.String())
	}
	if len(live.PersistentState().UCI) != 0 {
		t.Fatal("failed commit changed live board")
	}
	select {
	case <-ch:
		t.Fatal("failed commit published an event")
	default:
	}
	var count int64
	if err := db.Model(&storage.Move{}).Where("game_id = ?", id).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("rollback lost: %d %v", count, err)
	}
	if len(recoverySnapshot(t, h, id, players["w"]).UCI) != 0 {
		t.Fatal("failed commit changed durable board")
	}
	if err := db.Exec(`DROP TRIGGER reject_test_commit ON moves`).Error; err != nil {
		t.Fatal(err)
	}
	requireMove(t, h, id, players["w"], "e2e4")
	select {
	case <-ch:
	default:
		t.Fatal("committed move did not publish")
	}
}

func TestPostgresReleasedSeatSurvivesRestart(t *testing.T) {
	db := recoveryDB(t)
	h := NewHandlerWithStore(game.NewHub(), db)
	id := recoveryCreate(t, h)
	players := recoveryPlayers(t, h, id)
	w := recoveryRequest(h.HandleRelease, id, "POST", "/release", `{"clientId":"owner-secret","targetId":"friend-secret"}`)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"ok":true`) {
		t.Fatal(w.Body.String())
	}
	h = NewHandlerWithStore(game.NewHub(), db)
	replacement := recoverySnapshot(t, h, id, "replacement")
	if replacement.Color == nil || players[*replacement.Color] != "friend-secret" {
		t.Fatal("released seat not recovered")
	}
	if recoverySnapshot(t, h, id, "friend-secret").Role != "spectator" {
		t.Fatal("removed player regained occupied seat")
	}
}

func TestPostgresUnavailableOrLegacyNeverResetsGame(t *testing.T) {
	db := recoveryDB(t)
	h := NewHandlerWithStore(game.NewHub(), db)
	id := recoveryCreate(t, h)
	players := recoveryPlayers(t, h, id)
	requireMove(t, h, id, players["w"], "e2e4")
	for _, value := range []any{nil, `{"version":999}`} {
		if err := db.Model(&storage.Game{}).Where("id = ?", id).Update("live_state", value).Error; err != nil {
			t.Fatal(err)
		}
		h = NewHandlerWithStore(game.NewHub(), db)
		w := recoveryRequest(h.HandleSnapshot, id, "GET", "/snapshot?clientId=visitor", "")
		if w.Code != 409 {
			t.Fatalf("corrupt/legacy game response: %d", w.Code)
		}
		var row storage.Game
		if err := db.First(&row, "id = ?", id).Error; err != nil || row.MoveCount != 1 {
			t.Fatal("damaged original game history")
		}
	}
	w := recoveryRequest(h.HandleSnapshot, uuid.NewString(), "GET", "/snapshot?clientId=visitor", "")
	if w.Code != 404 {
		t.Fatalf("unknown game response: %d", w.Code)
	}
	pool, _ := db.DB()
	pool.Close()
	w = recoveryRequest(h.HandleSnapshot, id, "GET", "/snapshot?clientId=visitor", "")
	if w.Code != 503 {
		t.Fatalf("database outage response: %d", w.Code)
	}
	w = recoveryRequest(h.HandleCreateGame, "", "POST", "/api/games", "")
	if w.Code != 503 {
		t.Fatal("created volatile game during database outage")
	}
}
