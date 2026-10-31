package coach

import (
	"context"
	"os"
	"testing"
)

func TestNativeEngine(t *testing.T) {
	path := os.Getenv("COACH_TEST_ENGINE")
	if path == "" {
		t.Skip("set COACH_TEST_ENGINE to test pinned native Arasan")
	}
	engine := NewEngine(path)
	line, err := engine.Analyze(context.Background(), []string{"f2f3", "e7e5", "g2g4"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if line.Analysis.ID == "" || line.Analysis.RootSide != "b" || line.Analysis.Depth != line.Depth || line.Analysis.SearchElapsedMs < 0 {
		t.Fatalf("missing search provenance: %+v", line.Analysis)
	}
	if line.Moves[0] != "d8h4" || line.Score.Kind != "mate" || line.Score.Value != 1 {
		t.Fatalf("expected mate, got %+v", line)
	}
	line, err = engine.Analyze(context.Background(), nil, "d2d4")
	if err != nil {
		t.Fatal(err)
	}
	if line.Depth < 8 || line.Analysis.SearchElapsedMs < 1000 {
		t.Fatalf("restricted analysis stopped too early: depth=%d elapsed=%d", line.Depth, line.Analysis.SearchElapsedMs)
	}
	if line.Moves[0] != "d2d4" {
		t.Fatal("searchmoves ignored")
	}
}
