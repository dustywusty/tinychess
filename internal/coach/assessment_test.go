package coach

import (
	"strings"
	"testing"
)

func TestAssessmentUsesMoveLossNotPositionScore(t *testing.T) {
	l := &Lesson{Played: Line{Moves: []string{"e2e4"}, SAN: []string{"e4"}, Depth: 18, Score: Score{"cp", -400}}, Alternative: Line{Moves: []string{"d2d4"}, SAN: []string{"d4"}, Depth: 20, Score: Score{"cp", -390}}}
	if got := assessMove(l); got.Grade != "Good move" {
		t.Fatal("losing position confused with bad move", got)
	}
	l.Played.Score.Value = -700
	if got := assessMove(l); got.Grade != "Blunder" || got.Difference == nil || *got.Difference != 3.1 {
		t.Fatal(got)
	}
	l.Played.Depth = 2
	if got := assessMove(l); got.Grade != "Analysis limited" {
		t.Fatal("shallow search graded", got)
	}
	l.Played = l.Alternative
	if got := assessMove(l); got.Grade != "Best move" || !strings.Contains(got.Summary, "isn’t a better") {
		t.Fatal(got)
	}
}
