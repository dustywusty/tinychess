package handlers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/corentings/chess/v2"
	"github.com/google/uuid"

	"tinychess/internal/game"
)

// TestEndToEndGamePlay spins up the HTTP server, creates a game, joins it as two
// players via the SSE endpoint, and plays through a quick checkmate sequence
// using automated moves for both colors.
func TestEndToEndGamePlay(t *testing.T) {
	hub := game.NewHub(nil)
	handler := NewHandler(hub, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("/new", handler.HandleNew)
	mux.HandleFunc("/sse/", handler.HandleSSE)
	mux.HandleFunc("/move/", handler.HandleMove)
	mux.HandleFunc("/react/", handler.HandleReact)
	mux.HandleFunc("/release/", handler.HandleRelease)
	mux.HandleFunc("/forget/", handler.HandleForget)
	mux.HandleFunc("/api/stats", handler.HandleStats)
	mux.HandleFunc("/", handler.HandlePage)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := srv.Client()

	if resp, err := client.Get(srv.URL + "/"); err != nil {
		t.Fatalf("fetch home page: %v", err)
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected home status: %d", resp.StatusCode)
	} else {
		resp.Body.Close()
	}

	ownerID := uuid.NewString()
	newBody := fmt.Sprintf(`{"userId":"%s"}`, ownerID)
	resp, err := client.Post(srv.URL+"/new", "application/json", strings.NewReader(newBody))
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	defer resp.Body.Close()

	var newResp struct {
		OK    bool   `json:"ok"`
		ID    string `json:"id"`
		Color string `json:"color"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&newResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if !newResp.OK {
		t.Fatalf("create response not ok: %+v", newResp)
	}
	if newResp.ID == "" {
		t.Fatalf("empty game id")
	}

	gameURL := srv.URL + "/" + newResp.ID
	if pageResp, err := client.Get(gameURL); err != nil {
		t.Fatalf("fetch game page: %v", err)
	} else if pageResp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected game page status: %d", pageResp.StatusCode)
	} else {
		pageResp.Body.Close()
	}

	ownerColor := parseColor(t, newResp.Color)
	colorToClient := map[chess.Color]string{
		ownerColor: ownerID,
	}

	spectatorID := uuid.NewString()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sseReq, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/sse/"+newResp.ID+"?clientId="+spectatorID, nil)
	if err != nil {
		t.Fatalf("create sse request: %v", err)
	}
	sseReq.Header.Set("Accept", "text/event-stream")
	sseResp, err := client.Do(sseReq)
	if err != nil {
		t.Fatalf("sse request: %v", err)
	}
	defer sseResp.Body.Close()

	var secondColor chess.Color
	reader := bufio.NewReader(sseResp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if ctx.Err() != nil {
				t.Fatalf("timeout waiting for color assignment: %v", ctx.Err())
			}
			t.Fatalf("read sse line: %v", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		if payload == "" {
			continue
		}
		var clientState game.ClientState
		if err := json.Unmarshal([]byte(payload), &clientState); err != nil {
			t.Fatalf("decode client state: %v", err)
		}
		if clientState.Color == nil {
			t.Fatalf("expected assigned color for second client")
		}
		secondColor = parseColor(t, *clientState.Color)
		colorToClient[secondColor] = spectatorID
		break
	}
	cancel()

	if len(colorToClient) != 2 {
		t.Fatalf("expected two assigned colors, got %d", len(colorToClient))
	}

	local := chess.NewGame()
	movePlan := []struct {
		color chess.Color
		uci   string
	}{
		{chess.White, "f2f3"},
		{chess.Black, "e7e5"},
		{chess.White, "g2g4"},
		{chess.Black, "d8h4"},
	}

	var lastState game.GameState
	notation := chess.UCINotation{}
	for _, planned := range movePlan {
		playerID, ok := colorToClient[planned.color]
		if !ok {
			t.Fatalf("missing client for color %s", planned.color)
		}

		body := map[string]string{
			"uci":      planned.uci,
			"clientId": playerID,
		}
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal move payload: %v", err)
		}

		moveResp, err := client.Post(srv.URL+"/move/"+newResp.ID, "application/json", bytes.NewReader(payload))
		if err != nil {
			t.Fatalf("submit move %s: %v", planned.uci, err)
		}

		var respBody struct {
			OK    bool           `json:"ok"`
			State game.GameState `json:"state"`
		}
		if err := json.NewDecoder(moveResp.Body).Decode(&respBody); err != nil {
			moveResp.Body.Close()
			t.Fatalf("decode move response: %v", err)
		}
		moveResp.Body.Close()

		if !respBody.OK {
			t.Fatalf("move rejected for %s: %+v", planned.uci, respBody)
		}

		mv, err := notation.Decode(local.Position(), planned.uci)
		if err != nil {
			t.Fatalf("decode local move %s: %v", planned.uci, err)
		}
		if err := local.Move(mv, nil); err != nil {
			t.Fatalf("apply local move %s: %v", planned.uci, err)
		}
		lastState = respBody.State
	}

	if outcome := local.Outcome(); outcome == chess.NoOutcome {
		t.Fatalf("expected local game to end in checkmate")
	}

	if !strings.Contains(strings.ToLower(lastState.Status), "checkmate") {
		t.Fatalf("expected checkmate status, got %q", lastState.Status)
	}

	expectedFEN := local.Position().String()
	if lastState.FEN != expectedFEN {
		t.Fatalf("fen mismatch: got %q want %q", lastState.FEN, expectedFEN)
	}
}

func parseColor(t *testing.T, s string) chess.Color {
	t.Helper()
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "white", "w":
		return chess.White
	case "black", "b":
		return chess.Black
	default:
		t.Fatalf("unknown color %q", s)
		return chess.NoColor
	}
}
