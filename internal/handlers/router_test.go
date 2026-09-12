package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tinychess/internal/game"
)

func TestPublicConfig(t *testing.T) {
	for _, key := range []string{"", "secret-not-for-the-browser"} {
		t.Run(key, func(t *testing.T) {
			t.Setenv("OPENAI_API_KEY", key)
			router := NewRouter(game.NewHub(), "test-version")
			r := httptest.NewRecorder()
			router.ServeHTTP(r, httptest.NewRequest("GET", "/api/config", nil))
			var config map[string]any
			if err := json.Unmarshal(r.Body.Bytes(), &config); err != nil {
				t.Fatal(err)
			}
			if r.Code != http.StatusOK || len(config) != 2 || config["version"] != "test-version" || config["coachEnabled"] != (key != "") {
				t.Fatalf("unexpected public config: %d %s", r.Code, r.Body.String())
			}
			if r.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("runtime config must not be cached")
			}
		})
	}
}

func TestAPIRouting(t *testing.T) {
	hub := game.NewHub()
	router := NewRouter(hub, "test")
	for _, path := range []string{"/", "/a-game", "/move/a-game", "/api/missing"} {
		r := httptest.NewRecorder()
		router.ServeHTTP(r, httptest.NewRequest("GET", path, nil))
		if r.Code != http.StatusNotFound {
			t.Errorf("%s: expected 404, got %d", path, r.Code)
		}
	}
	r := httptest.NewRecorder()
	router.ServeHTTP(r, httptest.NewRequest("GET", "/healthz", nil))
	if r.Code != http.StatusOK || r.Body.String() != "ok\n" || len(hub.Games) != 0 {
		t.Fatal("health check must succeed without creating a game")
	}
	r = httptest.NewRecorder()
	router.ServeHTTP(r, httptest.NewRequest("POST", "/api/move/routed-game", strings.NewReader(`{}`)))
	if r.Code != http.StatusBadRequest || !strings.Contains(r.Body.String(), "missing client id") {
		t.Fatalf("move route did not reach handler: %d %s", r.Code, r.Body.String())
	}
	if _, ok := hub.Games["routed-game"]; !ok {
		t.Fatal("API prefix was not removed before passing game ID to the handler")
	}
}
