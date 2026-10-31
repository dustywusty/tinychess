package coach

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

// Live provider experiments are opt-in and never part of deterministic CI.
func TestLiveJevCorpus(t *testing.T) {
	if os.Getenv("COACH_LIVE_EVAL") != "1" {
		t.Skip("set COACH_LIVE_EVAL=1 and TYPESAFE_API_KEY for paid evaluation")
	}
	key := os.Getenv("TYPESAFE_API_KEY")
	if key == "" {
		t.Fatal("TYPESAFE_API_KEY required")
	}
	corpus := "testdata/jev-corpus.json"
	if os.Getenv("COACH_VALIDATION_SET") == "1" {
		corpus = "testdata/jev-validation.json"
	}
	data, err := os.ReadFile(corpus)
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Text          string `json:"text"`
		Expected      string `json:"expected"`
		Initial       bool   `json:"initial,omitempty"`
		RecentRequest string `json:"recentRequest,omitempty"`
		RecentAction  string `json:"recentAction,omitempty"`
	}
	if json.Unmarshal(data, &cases) != nil {
		t.Fatal("invalid corpus")
	}
	router := &Jev{Key: key, Model: os.Getenv("TYPESAFE_MODEL")}
	options := presentationOptions()
	options["candidate-0"] = "Knight on f3 captures pawn on e5"
	options["candidate-1"] = "Knight on c3 captures pawn on d5"
	position, err := Replay([]string{"e2e4", "d7d5", "b1c3", "e7e5", "g1f3", "g8f6"})
	if err != nil {
		t.Fatal(err)
	}
	correct, routes, targets, targetCases, clarifications, clarificationCases := 0, 0, 0, 0, 0, 0
	start := time.Now()
	criticalPassed := true
	for _, c := range cases {
		choices := options
		view := map[string]any{"branch": "alternative", "step": 1, "fen": position.Position().String(), "sideToMove": "w", "recentRequest": c.RecentRequest, "recentAction": c.RecentAction}
		if c.Initial {
			choices = initialOptions()
			view = map[string]any{"selectedPly": 0, "lesson": nil}
		}
		caseStart := time.Now()
		got, err := router.Select(context.Background(), c.Text, choices, view)
		route := func(id string) string {
			if strings.HasPrefix(id, "candidate-") {
				return "candidate"
			}
			return id
		}
		if err == nil && route(got) == route(c.Expected) {
			routes++
		}
		if strings.HasPrefix(c.Expected, "candidate-") {
			targetCases++
			if err == nil && got == c.Expected {
				targets++
			}
		}
		if c.Expected == "clarify" {
			clarificationCases++
			if err == nil && got == c.Expected {
				clarifications++
			}
		}
		if err == nil && got == c.Expected {
			correct++
		} else if strings.HasPrefix(c.Expected, "candidate-") || c.Expected == "clarify" || c.Expected == "hint" || c.Expected == "unsupported" {
			criticalPassed = false
		}
		t.Logf("request=%q expected=%s got=%s elapsed_ms=%d error=%v", c.Text, c.Expected, got, time.Since(caseStart).Milliseconds(), err)
	}
	t.Logf("route_accuracy=%d/%d target_accuracy=%d/%d clarification_accuracy=%d/%d", routes, len(cases), targets, targetCases, clarifications, clarificationCases)
	t.Logf("route_and_target_accuracy=%d/%d elapsed_ms=%d; token usage is logged by stage; cost requires configured provider pricing", correct, len(cases), time.Since(start).Milliseconds())
	passed := correct == len(cases)
	if os.Getenv("COACH_VALIDATION_SET") == "1" {
		passed = criticalPassed && correct*20 >= len(cases)*19
	}
	if !passed {
		t.Error("inspect routing failures before enabling language controls")
	}
}
