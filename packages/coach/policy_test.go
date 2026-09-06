package coach

import (
	"strings"
	"testing"
)

func TestOffTopicNeverAllowsExplanation(t *testing.T) {
	session := Session{}
	decision := session.Decide(Request{PositionKey: "fen-1", Text: "What's the weather?"})
	if decision.Intent != IntentOffTopic || decision.AllowExplanation || decision.NeedsAnalysis {
		t.Fatalf("unexpected decision: %+v", decision)
	}
}

func TestProgressiveHintsUseCachedArtifactsAndStopAtThree(t *testing.T) {
	session := Session{}
	artifacts := &Artifacts{Hints: [MaxHintsPerMove]string{"concept", "attention", "direction"}}

	for level, want := range []string{"concept", "attention", "direction"} {
		decision := session.Decide(Request{
			PositionKey: "fen-1",
			Text:        []string{"hint", "another hint", "one more hint"}[level],
			Artifacts:   artifacts,
		})
		if decision.HintLevel != level+1 || decision.Canned != want || decision.NeedsAnalysis {
			t.Fatalf("level %d: %+v", level+1, decision)
		}
	}

	decision := session.Decide(Request{PositionKey: "fen-1", Text: "next hint", Artifacts: artifacts})
	if decision.Reason != "hint_limit" || decision.Canned == "" {
		t.Fatalf("limit decision: %+v", decision)
	}
}

func TestMissingHintsRequestOneArtifactGeneration(t *testing.T) {
	session := Session{}
	decision := session.Decide(Request{PositionKey: "fen-1", Text: "hint"})
	if !decision.NeedsAnalysis || decision.AllowExplanation {
		t.Fatalf("unexpected decision: %+v", decision)
	}
}

func TestQuestionAndRepeatedLimits(t *testing.T) {
	session := Session{}
	first := session.Decide(Request{PositionKey: "fen-1", Text: "Why was that bad?"})
	if !first.AllowExplanation {
		t.Fatalf("first question should be allowed: %+v", first)
	}
	repeated := session.Decide(Request{PositionKey: "fen-1", Text: "  WHY was that bad?  "})
	if repeated.Intent != IntentRepeated || repeated.AllowExplanation {
		t.Fatalf("repeat should be canned: %+v", repeated)
	}
	second := session.Decide(Request{PositionKey: "fen-1", Text: "Explain the position"})
	if !second.AllowExplanation {
		t.Fatalf("second distinct question should be allowed: %+v", second)
	}
	third := session.Decide(Request{PositionKey: "fen-1", Text: "What should I notice?"})
	if third.Reason != "question_limit" || third.AllowExplanation {
		t.Fatalf("third question should be capped: %+v", third)
	}
}

func TestPositionChangeResetsLimits(t *testing.T) {
	session := Session{}
	_ = session.Decide(Request{PositionKey: "fen-1", Text: "Why was that bad?"})
	_ = session.Decide(Request{PositionKey: "fen-1", Text: "Explain the position"})
	decision := session.Decide(Request{PositionKey: "fen-2", Text: "Why was that bad?"})
	if !decision.AllowExplanation || session.QuestionsThisMove != 1 {
		t.Fatalf("limits did not reset: decision=%+v session=%+v", decision, session)
	}
}

func TestLongPromptIsRejectedLocally(t *testing.T) {
	intent, reason := Route(strings.Repeat("x", MaxPromptRunes+1))
	if intent != IntentOffTopic || reason != "too_long" {
		t.Fatalf("got %s/%s", intent, reason)
	}
}
