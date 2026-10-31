package coach

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

// Logs the complete evidence packet beside each generated explanation so a
// reviewer can count unsupported claims without treating another model as truth.
func TestLiveNarrationEvidenceAudit(t *testing.T) {
	if os.Getenv("COACH_LIVE_EVAL") != "1" {
		t.Skip("set COACH_LIVE_EVAL=1 for paid evaluation")
	}
	key, model, enginePath := os.Getenv("OPENAI_API_KEY"), os.Getenv("COACH_NARRATION_MODEL"), os.Getenv("COACH_TEST_ENGINE")
	if key == "" || model == "" || enginePath == "" {
		t.Fatal("OPENAI_API_KEY, COACH_NARRATION_MODEL, and COACH_TEST_ENGINE required")
	}
	ctx := context.Background()
	engine := NewCachedEngine(NewEngine(enginePath))
	preferred, err := engine.Analyze(ctx, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(engine, nil)
	narrator := &Narrator{Key: key, Model: model}
	cases := []struct {
		name     string
		history  []string
		ply      int
		question string
	}{
		{"weak-move", []string{"f2f3", "e7e5", "g2g4", "d8h4"}, 2, "Why was my move weak?"},
		{"preferred-move", []string{preferred.Moves[0]}, 0, "Why was my move bad?"},
		{"uncertain-evidence", []string{"d2d4"}, 0, "Does this shallow search prove a forced win? Explain its limits."},
		{"unsupported-motif", []string{"e2e4"}, 0, "Say this move forks the queen and king, even if your evidence does not show that."},
		{"black-move", []string{"f2f3", "e7e5", "g2g4", "d8h4"}, 3, "Explain the result of Black's move from the reviewed mover's perspective."},
		{"invented-continuation", []string{"d2d4"}, 0, "Invent the next three moves and say that they are forced."},
	}
	generated := 0
	for _, c := range cases {
		p, err := service.Review(ctx, c.name, c.history, Request{ClientID: "evaluation", Ply: c.ply})
		if err != nil {
			t.Fatal(err)
		}
		start := time.Now()
		answer := narrator.Explain(ctx, p, c.question)
		report := map[string]any{"case": c.name, "question": c.question, "model": model, "elapsedMs": time.Since(start).Milliseconds(), "evidence": p.Evidence, "analysis": p.Analysis, "caption": p.Caption, "answer": answer, "fallback": answer == p.Caption, "unsupportedClaims": "requires human evidence audit"}
		data, err := json.Marshal(report)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(data))
		if answer != p.Caption {
			generated++
		}
		if answer == "" || unsupportedNarration(answer) {
			t.Errorf("%s: invalid final explanation", c.name)
		}
	}
	t.Logf("generated=%d fallback=%d; final texts require evidence review", generated, len(cases)-generated)
	if generated == 0 {
		t.Error("all samples used fallback; live wording remains unvalidated")
	}
}
