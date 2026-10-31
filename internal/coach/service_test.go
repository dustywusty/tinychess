package coach

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

type fakeEngine struct{ calls int }

func (e *fakeEngine) Analyze(_ context.Context, h []string, c string) (Line, error) {
	e.calls++
	if c == "" {
		c = "e2e4"
	}
	return ValidateLine(h, Line{Moves: []string{c, "e7e5"}, Score: Score{"cp", 20}, Depth: 10})
}
func TestReviewIsolationProjectionAndFreshness(t *testing.T) {
	e := &fakeEngine{}
	s := NewService(e, nil)
	history := []string{"d2d4", "d7d5"}
	p, err := s.Review(context.Background(), "game", history, Request{ClientID: "owner", Ply: 0, NoSpoilers: true})
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(p)
	for _, secret := range []string{"e2e4", "e7e5", "d2d4", "score", "Alternative"} {
		if strings.Contains(string(data), secret) {
			t.Fatalf("hint leaked %s: %s", secret, data)
		}
	}
	if len(p.Arrows) != 0 || p.Step != 0 || p.Steps != 0 {
		t.Fatal("hint revealed a line")
	}
	r := Request{ClientID: "intruder", LessonID: p.ID, Revision: p.Revision, Action: "alternative"}
	if _, err = s.Action(context.Background(), "game", history, r); err == nil {
		t.Fatal("cross-owner access")
	}
	r.ClientID = "owner"
	if _, err = s.Action(context.Background(), "other", history, r); err == nil {
		t.Fatal("cross-game access")
	}
	r.Action = "alternative"
	p, err = s.Action(context.Background(), "game", history, r)
	if err != nil {
		t.Fatal(err)
	}
	if p.FEN == "" || p.Branch != "alternative" {
		t.Fatal(p)
	}
	if _, err = s.Action(context.Background(), "game", history, r); err == nil {
		t.Fatal("replayed event advanced")
	}
	r.Revision = p.Revision
	r.Action = "next"
	p, err = s.Action(context.Background(), "game", append(history, "g1f3"), r)
	if err != nil {
		t.Fatal(err)
	}
	if p.Step != 2 || history[0] != "d2d4" || len(history) != 2 {
		t.Fatal("live state changed")
	}
	r.Revision = p.Revision
	if _, err = s.Action(context.Background(), "game", []string{"e2e4"}, r); err == nil {
		t.Fatal("wrong history accepted")
	}
	if e.calls != 2 {
		t.Fatalf("navigation called engine: %d", e.calls)
	}
}
func TestRulesRejectInvalidPVAndScores(t *testing.T) {
	if _, err := ValidateLine(nil, Line{Moves: []string{"e2e5"}, Depth: 10, Score: Score{"cp", 0}}); err == nil {
		t.Fatal("illegal pv")
	}
	if _, ok := parseInfo("info depth 10 score cp 25 lowerbound pv e2e4"); ok {
		t.Fatal("bound treated as exact")
	}
	l, ok := parseInfo("info depth 10 score mate -2 pv e2e4")
	if !ok || l.Score.Kind != "mate" || l.Score.Value != -2 {
		t.Fatal(l)
	}
	if Identity([]string{"g1f3", "g8f6", "f3g1", "f6g8"}) == Identity(nil) {
		t.Fatal("repetition history omitted")
	}
}
func TestGoodMoveAndAlternativeExercise(t *testing.T) {
	e := &fakeEngine{}
	s := NewService(e, nil)
	p, err := s.Review(context.Background(), "game", []string{"e2e4"}, Request{ClientID: "owner"})
	if err != nil {
		t.Fatal(err)
	}
	if p.Assessment == nil || p.Assessment.Grade != "Best move" {
		t.Fatal(p.Caption)
	}
	best := Line{Moves: []string{"e2e4"}, Score: Score{"cp", 40}, Depth: 10}
	alt := Line{Moves: []string{"d2d4"}, Score: Score{"cp", 20}, Depth: 10}
	if !strings.Contains(acceptance(best, alt), "sound alternative") {
		t.Fatal("sound alternative rejected")
	}
	alt.Score = Score{"mate", 2}
	if !strings.Contains(acceptance(best, alt), "deeper") {
		t.Fatal("mate subtracted as cp")
	}
}

func TestSpecialMovesStayLegal(t *testing.T) {
	cases := [][]string{
		{"e2e4", "a7a6", "e4e5", "d7d5", "e5d6"},
		{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4", "g8f6", "e1g1"},
		{"a2a4", "h7h5", "a4a5", "h5h4", "a5a6", "h4h3", "a6b7", "h3g2", "b7a8q"},
		{"f2f3", "e7e5", "g2g4", "d8h4"},
	}
	for _, moves := range cases {
		line, err := ValidateLine(moves[:len(moves)-1], Line{Moves: moves[len(moves)-1:], Depth: 8, Score: Score{"cp", 0}})
		if err != nil || len(line.FENs) != 2 || len(line.SAN) != 1 {
			t.Fatalf("%v: %v", moves, err)
		}
	}
}

func TestCandidateExplorationKeepsOriginalMove(t *testing.T) {
	s := NewService(&fakeEngine{}, nil)
	history := []string{"d2d4"}
	p, err := s.Review(context.Background(), "game", history, Request{ClientID: "owner"})
	if err != nil {
		t.Fatal(err)
	}
	p, err = s.Action(context.Background(), "game", history, Request{ClientID: "owner", LessonID: p.ID, Revision: p.Revision, Action: "candidate", Candidate: "g1f3", Branch: "played", Step: 0})
	if err != nil {
		t.Fatal(err)
	}
	p, err = s.Action(context.Background(), "game", history, Request{ClientID: "owner", LessonID: p.ID, Revision: p.Revision, Action: "played"})
	if err != nil {
		t.Fatal(err)
	}
	if p.Arrows[0].From != "d2" || p.Arrows[0].To != "d4" {
		t.Fatal("original move overwritten")
	}
}
