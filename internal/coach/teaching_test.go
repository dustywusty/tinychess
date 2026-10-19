package coach

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type teachingRouter struct {
	plan  TeachingPlan
	state any
}

func (r *teachingRouter) Select(context.Context, string, map[string]string, any) (string, error) {
	return r.plan.Action, nil
}
func (r *teachingRouter) Plan(_ context.Context, _ string, _ map[string]string, state any) (TeachingPlan, error) {
	r.state = state
	return r.plan, nil
}

func TestTeachingChangesViewsAndPreservesHiddenAnswers(t *testing.T) {
	router := &teachingRouter{plan: TeachingPlan{"explain", "compare", "played", "practice"}}
	s := NewService(&fakeEngine{}, router)
	history := []string{"d2d4", "d7d5"}
	ctx := context.Background()
	p, err := s.Review(ctx, "game", history, Request{ClientID: "owner"})
	if err != nil {
		t.Fatal(err)
	}
	ask := func(text string) Projection {
		t.Helper()
		next, err := s.Action(ctx, "game", history, Request{ClientID: "owner", LessonID: p.ID, Revision: p.Revision, Action: "language", Text: text})
		if err != nil {
			t.Fatal(err)
		}
		p = next
		return p
	}
	ask("Why was mine worse?")
	if p.Branch != "compare" || p.Step != 0 || p.Teaching.View != "compare" || p.Teaching.Prompt.Label != "Let me try" {
		t.Fatal(p)
	}
	router.plan = TeachingPlan{"explain", "move", "recommended", "reply"}
	ask("I don't understand e4")
	if p.Branch != "alternative" || p.Step != 0 || !reflect.DeepEqual(p.Teaching.Focus, []string{"e2"}) || !strings.Contains(p.Caption, "pawn from e2 to e4") {
		t.Fatal(p)
	}
	if len(router.state.(map[string]any)["conversation"].([]TeachingTurn)) != 1 {
		t.Fatal("missing prior question")
	}
	router.plan = TeachingPlan{"next", "reply", "recommended", "practice"}
	ask("What happens next?")
	if p.Step != 2 || !reflect.DeepEqual(p.Teaching.Focus, []string{"e5"}) {
		t.Fatal(p)
	}
	// A malicious/inconsistent presentation cannot override an explicit hidden-answer request.
	router.plan = TeachingPlan{"explain", "compare", "recommended", "compare"}
	ask("Don't tell me the answer; let me try")
	if p.Reveal != 0 || p.Assessment != nil || p.Views != nil || p.Score != nil || len(p.Teaching.Focus) != 0 || p.Mode != "exercising" {
		t.Fatal("answer leaked", p)
	}
	if router.state.(map[string]any)["recommendedMove"] != "" {
		t.Fatal("hidden recommendation sent to planner")
	}
	if !reflect.DeepEqual(history, []string{"d2d4", "d7d5"}) {
		t.Fatal("live history changed")
	}
}

func TestJevTeachingUsesOneCallAndRejectsInvalidPresentation(t *testing.T) {
	for _, view := range []string{"move", "executeJavascript"} {
		calls := 0
		j := &Jev{Key: "test", Client: &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
			calls++
			body, _ := io.ReadAll(r.Body)
			var input struct {
				Questions map[string]any `json:"questions"`
			}
			json.Unmarshal(body, &input)
			if len(input.Questions) != 4 {
				t.Fatal("teaching decisions not batched", string(body))
			}
			response := `{"answers":{"action":{"type":"choice","choice":"explain"},"view":{"type":"choice","choice":"` + view + `"},"target":{"type":"choice","choice":"recommended"},"followup":{"type":"choice","choice":"reply"}}}`
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(response))}, nil
		})}}
		_, err := j.Plan(context.Background(), "I don't understand e4", presentationOptions(), nil)
		if (view == "move") != (err == nil) || calls != 1 {
			t.Fatal("invalid teaching validation", err, calls)
		}
	}
}

func TestOneBudgetSlotForRoutingAndAnalysis(t *testing.T) {
	router := &teachingRouter{plan: TeachingPlan{"best_now", "move", "recommended", "why"}}
	s := NewService(&fakeEngine{}, router)
	ctx := context.Background()
	r := Request{ClientID: "owner", Action: "language", Text: "What should I play next?"}
	action, err := s.InitialTeaching(ctx, "game", nil, &r)
	if err != nil {
		t.Fatal(err)
	}
	r.Action = action
	if _, err = s.Review(ctx, "game", nil, r); err != nil {
		t.Fatal(err)
	}
	if s.budgets["game:owner"].Count != 1 {
		t.Fatal("question charged twice", s.budgets)
	}
}

func TestUnsupportedTeachingNeverChangesTheBoard(t *testing.T) {
	router := &teachingRouter{plan: TeachingPlan{"unsupported", "reply", "recommended", "practice"}}
	s := NewService(&fakeEngine{}, router)
	ctx := context.Background()
	history := []string{"d2d4"}
	p, _ := s.Review(ctx, "game", history, Request{ClientID: "owner"})
	result, err := s.Action(ctx, "game", history, Request{ClientID: "owner", LessonID: p.ID, Revision: p.Revision, Action: "language", Text: "Play the move for me"})
	if err != nil || result.FEN != p.FEN || result.Teaching != nil {
		t.Fatal("unsupported request changed view", result, err)
	}
}

func TestBestMoveCannotBecomeAnUnrequestedExercise(t *testing.T) {
	s := NewService(&fakeEngine{}, nil)
	p, err := s.Review(context.Background(), "game", nil, Request{ClientID: "owner", Action: "best_now", teachingPlan: &TeachingPlan{"best_now", "exercise", "recommended", "why"}})
	if err != nil || p.Reveal != 3 || p.Recommendation == nil || p.Teaching.View != "move" {
		t.Fatal("best move was hidden", p, err)
	}
}

func TestNarrationDoesNotInventComparisonDelta(t *testing.T) {
	n := &Narrator{Key: "test", Model: "test", Client: &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("data: {\"type\":\"response.output_text.delta\",\"delta\":\"Nc3 is better by about 0.94 pawns.\"}\n\ndata: {\"type\":\"response.completed\"}\n\n"))}, nil
	})}}
	p := Projection{Reveal: 3, Assessment: &MoveAssessment{}, Caption: "Nc3 avoids the demonstrated mating line."}
	if got := n.Explain(context.Background(), p, "Why was mine worse?"); got != p.Caption {
		t.Fatal("invented comparison", got)
	}
}
