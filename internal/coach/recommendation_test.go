package coach

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

type nextEngine struct{ histories [][]string }

func (e *nextEngine) Analyze(_ context.Context, history []string, candidate string) (Line, error) {
	e.histories = append(e.histories, append([]string{}, history...))
	if candidate == "" {
		candidate = "e2e4"
		if len(history)%2 == 1 {
			candidate = "e7e5"
		}
	}
	return ValidateLine(history, Line{Moves: []string{candidate}, Score: Score{"cp", 35}, Depth: 10})
}

type nextRouter struct{}

func (nextRouter) Select(_ context.Context, _ string, options map[string]string, _ any) (string, error) {
	return "best_now", nil
}

func TestRecommendationUsesLiveRootWithoutInventingPlayedMove(t *testing.T) {
	for _, history := range [][]string{nil, {"g1f3"}, {"g1f3", "g8f6"}} {
		e := &nextEngine{}
		s := NewService(e, nextRouter{})
		ctx := context.Background()
		r := Request{ClientID: "owner", Action: "best_now", RequestID: "first"}
		p, err := s.Review(ctx, "game", history, r)
		if err != nil {
			t.Fatal(err)
		}
		root, _ := Replay(history)
		if !p.CurrentPosition || p.Assessment != nil || p.Recommendation == nil || p.Ply != len(history) || p.PositionID != Identity(history) || p.FEN != root.Position().String() || p.Step != 0 {
			t.Fatalf("wrong current position: %+v", p)
		}
		if !strings.Contains(p.Caption, "pawn") || strings.Contains(p.Caption, "YOU PLAYED") {
			t.Fatal(p.Caption)
		}
		if !reflect.DeepEqual(e.histories[0], append([]string{}, history...)) {
			t.Fatal(e.histories)
		}
		retry, err := s.Review(ctx, "game", history, r)
		if err != nil || retry.ID != p.ID || len(e.histories) != 1 {
			t.Fatal("retry repeated search", err)
		}
		if _, err = s.Action(ctx, "game", history, Request{ClientID: "visitor", LessonID: p.ID, Revision: 1, Action: "alternative"}); err == nil {
			t.Fatal("cross-owner access")
		}
		preview, err := s.Action(ctx, "game", history, Request{ClientID: "owner", LessonID: p.ID, Revision: 1, Action: "alternative"})
		if err != nil || preview.Step != 1 || preview.FEN == p.FEN {
			t.Fatal("preview failed", err)
		}
		// Follow-up questions remain valid on the analyzed root after a live move.
		newer := append(append([]string{}, history...), s.lessons[p.ID].Alternative.Moves[0])
		follow, err := s.Action(ctx, "game", newer, Request{ClientID: "owner", LessonID: p.ID, Revision: preview.Revision, Action: "explain"})
		if err != nil || follow.PositionID != p.PositionID {
			t.Fatal("lost pinned context", err)
		}
	}
}

func TestRecommendationHintsAndTerminalPositions(t *testing.T) {
	s := NewService(&nextEngine{}, nil)
	p, err := s.Review(context.Background(), "game", nil, Request{ClientID: "owner", Action: "best_now", NoSpoilers: true})
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(p)
	if p.Recommendation != nil || p.Views != nil || strings.Contains(string(body), "e2e4") {
		t.Fatal("hidden answer leaked", string(body))
	}
	_, err = s.Review(context.Background(), "finished", []string{"f2f3", "e7e5", "g2g4", "d8h4"}, Request{ClientID: "owner", Action: "best_now"})
	if err == nil || !strings.Contains(err.Error(), "ended") {
		t.Fatal("terminal position accepted", err)
	}
}

func TestBestNowFromPastLessonUsesFullHistory(t *testing.T) {
	e := &nextEngine{}
	s := NewService(e, nextRouter{})
	history := []string{"g1f3", "g8f6"}
	p, err := s.Review(context.Background(), "game", history, Request{ClientID: "owner", Ply: 0})
	if err != nil {
		t.Fatal(err)
	}
	p, err = s.Action(context.Background(), "game", history, Request{ClientID: "owner", LessonID: p.ID, Revision: p.Revision, Action: "language", Text: "What is my next best move?"})
	if err != nil || !p.CurrentPosition || p.Ply != 2 || !reflect.DeepEqual(e.histories[len(e.histories)-1], history) {
		t.Fatal("used past review", p, err)
	}
}
