package handlers

import (
	"context"
	"strings"
	"testing"
	"tinychess/internal/coach"
	"tinychess/internal/game"
)

type coachEngine struct{ calls int }

type recommendRouter struct{}

func (recommendRouter) Select(_ context.Context, _ string, options map[string]string, _ any) (string, error) {
	return "best_now", nil
}

func TestCurrentPositionBeforeFirstMoveAndLanguageRouting(t *testing.T) {
	h := NewHandler(game.NewHub())
	engine := &coachEngine{}
	h.Coach = coach.NewService(engine, recommendRouter{})
	id := botGame(t, h)
	for _, action := range []string{"best_now", "language"} {
		p, status, message := h.coachRequest(context.Background(), id, coach.Request{ClientID: "owner", Action: action, Text: "What should I play next?"})
		if status != 200 || !p.CurrentPosition || p.Ply != 0 || p.Recommendation == nil {
			t.Fatal(status, message, p)
		}
	}
	_, status, _ := h.coachRequest(context.Background(), id, coach.Request{ClientID: "visitor", Action: "best_now"})
	if status != 403 || engine.calls != 2 {
		t.Fatal("current-position policy bypass", status, engine.calls)
	}
}

func (e *coachEngine) Analyze(_ context.Context, history []string, candidate string) (coach.Line, error) {
	e.calls++
	if candidate == "" {
		candidate = "e2e4"
	}
	return coach.ValidateLine(history, coach.Line{Moves: []string{candidate}, Score: coach.Score{Kind: "cp", Value: 10}, Depth: 10})
}
func TestCoachPolicyBeforeAnalysisAndCache(t *testing.T) {
	h := NewHandler(game.NewHub())
	engine := &coachEngine{}
	h.Coach = coach.NewService(engine, nil)
	id := botGame(t, h)
	recoveryRequest(h.HandleMove, id, "POST", "/move", `{"clientId":"owner","uci":"e2e4","botMove":true,"expectedPly":0}`)
	for _, client := range []string{"", "visitor"} {
		_, status, _ := h.coachRequest(context.Background(), id, coach.Request{ClientID: client, Action: "review"})
		if status == 200 {
			t.Fatal("unauthorized review")
		}
	}
	if engine.calls != 0 {
		t.Fatal("engine called before authorization")
	}
	p, status, message := h.coachRequest(context.Background(), id, coach.Request{ClientID: "owner", Action: "review"})
	if status != 200 {
		t.Fatal(message)
	}
	_, status, _ = h.coachRequest(context.Background(), id, coach.Request{ClientID: "visitor", Action: "alternative", LessonID: p.ID, Revision: 1})
	if status != 403 {
		t.Fatal("cached lesson exposed")
	}
	friend := recoverySnapshot(t, h, "friend-game", "friend")
	_, status, _ = h.coachRequest(context.Background(), "friend-game", coach.Request{ClientID: friend.ClientID, Action: "review"})
	if status != 403 {
		t.Fatal("live human assistance permitted")
	}
	body := `{"threadId":"thread","runId":"run","forwardedProps":{"coach":{"clientId":"visitor","action":"review"}}}`
	response := recoveryRequest(h.HandleCoachStream, id, "POST", "/agent", body)
	if !strings.Contains(response.Body.String(), "RUN_ERROR") || strings.Contains(response.Body.String(), "STATE_SNAPSHOT") {
		t.Fatal("AG-UI bypassed policy")
	}
}

type revokingEngine struct{ revoke func() }

func (e revokingEngine) Analyze(_ context.Context, h []string, c string) (coach.Line, error) {
	e.revoke()
	if c == "" {
		c = "f2f3"
	}
	return coach.ValidateLine(h, coach.Line{Moves: []string{c}, Score: coach.Score{Kind: "cp", Value: 10}, Depth: 8})
}
func TestCoachRechecksSeatBeforeReturningDelayedAnalysis(t *testing.T) {
	h := NewHandler(game.NewHub())
	recoverySnapshot(t, h, "completed", "owner")
	g, _ := h.Hub.Get("completed", "")
	for _, move := range []string{"f2f3", "e7e5", "g2g4", "d8h4"} {
		if err := g.MakeMove(move); err != nil {
			t.Fatal(err)
		}
	}
	h.Coach = coach.NewService(revokingEngine{func() { g.RemoveClient("owner") }}, nil)
	projection, status, _ := h.coachRequest(context.Background(), "completed", coach.Request{ClientID: "owner", Action: "review"})
	if status != 403 || projection.ID != "" {
		t.Fatal("revoked seat received delayed lesson", status)
	}
}
