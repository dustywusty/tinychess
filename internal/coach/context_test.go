package coach

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

type contextEngine struct {
	histories [][]string
	score     Score
}

func (e *contextEngine) Analyze(_ context.Context, h []string, move string) (Line, error) {
	e.histories = append(e.histories, append([]string{}, h...))
	if move == "" {
		move = "e2e4"
	}
	return ValidateLine(h, Line{Moves: []string{move}, Score: e.score, Depth: 10})
}

type contextRouter struct {
	t                 *testing.T
	expectedFEN, move string
}

func (r contextRouter) Select(_ context.Context, _ string, options map[string]string, state any) (string, error) {
	view := state.(map[string]any)
	if view["fen"] != r.expectedFEN {
		r.t.Fatalf("router got a different board: %v", view)
	}
	for id, label := range options {
		if strings.Contains(label, r.move) {
			return id, nil
		}
	}
	r.t.Fatalf("displayed-position move %s missing", r.move)
	return "", nil
}
func TestExploreUsesDisplayedHistoryAndPreservesScorePerspective(t *testing.T) {
	for _, score := range []Score{{"cp", 35}, {"mate", 3}} {
		t.Run(score.Kind, func(t *testing.T) {
			engine := &contextEngine{score: score}
			s := NewService(engine, nil)
			ctx := context.Background()
			history := []string{"d2d4"}
			p, err := s.Review(ctx, "game", history, Request{ClientID: "owner"})
			if err != nil {
				t.Fatal(err)
			}
			position, _ := Replay([]string{"e2e4"})
			s.Router = contextRouter{t, position.Position().String(), "c7c5"}
			p, err = s.Action(ctx, "game", history, Request{ClientID: "owner", LessonID: p.ID, Revision: p.Revision, Action: "language", Text: "What if they play c5?", Branch: "alternative", Step: 1})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(engine.histories[len(engine.histories)-1], []string{"e2e4"}) {
				t.Fatal(engine.histories)
			}
			if p.Score == nil || p.Score.Value != -score.Value || p.Score.Kind != score.Kind {
				t.Fatalf("wrong mover perspective: %+v", p.Score)
			}
			// A second exploration must retain the first exploration's prefix.
			p, err = s.Action(ctx, "game", history, Request{ClientID: "owner", LessonID: p.ID, Revision: p.Revision, Action: "candidate", Candidate: "g1f3", Branch: "explored", Step: 1})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(engine.histories[len(engine.histories)-1], []string{"e2e4", "c7c5"}) {
				t.Fatal(engine.histories)
			}
			if p.Score.Value != score.Value {
				t.Fatal("second exploration perspective was flipped")
			}
			p, err = s.Action(ctx, "game", history, Request{ClientID: "owner", LessonID: p.ID, Revision: p.Revision, Action: "exercise"})
			if err != nil {
				t.Fatal(err)
			}
			origin, _ := Replay(nil)
			if p.FEN != origin.Position().String() || p.Branch != "played" || p.Reveal != 0 {
				t.Fatalf("exercise did not reset origin: %+v", p)
			}
			calls := len(engine.histories)
			_, err = s.Action(ctx, "game", history, Request{ClientID: "owner", LessonID: p.ID, Revision: p.Revision, Action: "candidate", Candidate: "e2e4"})
			if err == nil || len(engine.histories) != calls {
				t.Fatal("hidden exploration reached engine")
			}
		})
	}
}

type fixedRouter string

func (r fixedRouter) Select(_ context.Context, _ string, options map[string]string, _ any) (string, error) {
	return string(r), nil
}
func TestExplicitHintCannotRevealThroughMisroutedExplanation(t *testing.T) {
	engine := &contextEngine{score: Score{"cp", 40}}
	s := NewService(engine, fixedRouter("explain"))
	ctx := context.Background()
	history := []string{"d2d4"}
	p, err := s.Review(ctx, "game", history, Request{ClientID: "owner"})
	if err != nil {
		t.Fatal(err)
	}
	p, err = s.Action(ctx, "game", history, Request{ClientID: "owner", LessonID: p.ID, Revision: p.Revision, Action: "language", Text: "Just a hint; don't tell me the move.", Branch: "played", Step: 1})
	if err != nil {
		t.Fatal(err)
	}
	if p.Reveal != 0 || p.Score != nil || p.Analysis != nil || len(p.Evidence) != 0 || len(p.Views) != 0 || len(p.Arrows) != 0 {
		t.Fatalf("explicit hint leaked analysis: %+v", p)
	}
	s.Router = fixedRouter("alternative")
	if _, err = s.Action(ctx, "game", history, Request{ClientID: "owner", LessonID: p.ID, Revision: p.Revision, Action: "language", Text: "only a hint"}); err == nil {
		t.Fatal("model selected a forbidden spoiler action")
	}
}
func TestLanguageExplanationSignalsNarrationWithoutExposingHiddenEvidence(t *testing.T) {
	s := NewService(&contextEngine{score: Score{"cp", 15}}, fixedRouter("explain"))
	ctx := context.Background()
	history := []string{"e2e4"}
	p, err := s.Review(ctx, "game", history, Request{ClientID: "owner"})
	if err != nil {
		t.Fatal(err)
	}
	p, err = s.Action(ctx, "game", history, Request{ClientID: "owner", LessonID: p.ID, Revision: p.Revision, Action: "language", Text: "Explain this line"})
	if err != nil || !p.Explain || len(p.Evidence) == 0 {
		t.Fatalf("missing narration evidence: %+v %v", p, err)
	}
}
