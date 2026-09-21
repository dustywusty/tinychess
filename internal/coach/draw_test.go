package coach

import (
	"context"
	"os"
	"testing"

	"github.com/corentings/chess/v2"
)

func TestRepetitionEndsTheDisplayedContinuation(t *testing.T) {
	var history []string
	cycle := []string{"g1f3", "g8f6", "f3g1", "f6g8"}
	for range 3 {
		history = append(history, cycle...)
	}
	history = append(history, cycle[:3]...)
	line, err := ValidateLine(history, Line{Moves: []string{"f6g8", "e2e4"}, Depth: 8, Score: Score{"cp", 0}})
	if err != nil {
		t.Fatal(err)
	}
	if len(line.Moves) != 1 || len(line.FENs) != 2 || line.Draws[1] != chess.FivefoldRepetition.String() {
		t.Fatalf("continued past automatic draw: %+v", line)
	}
	for _, fact := range lineEvidence(line, 0) {
		if fact.Kind == "draw" {
			t.Fatal("future draw leaked at origin")
		}
	}
	found := false
	for _, fact := range lineEvidence(line, 1) {
		if fact.Kind == "draw" && fact.FEN == line.FENs[1] {
			found = true
		}
	}
	if !found {
		t.Fatal("draw missing from final position evidence")
	}
	terminal := append(append([]string{}, history...), "f6g8")
	if _, err := Replay(append(terminal, "e2e4")); err == nil {
		t.Fatal("accepted moves after automatic draw")
	}
	// An invalid binary makes this assertion also prove rejection precedes engine execution.
	if _, err := NewEngine("/missing-engine").Analyze(context.Background(), terminal, ""); err == nil || err.Error() != "the position is terminal; review an earlier move" {
		t.Fatal(err)
	}
}

// Legal start-position sequence ending in stalemate:
// https://en.wikipedia.org/wiki/List_of_world_records_in_chess#Shortest_stalemate
var stalemateHistory = []string{"e2e3", "a7a5", "d1h5", "a8a6", "h5a5", "h7h5", "a5c7", "a6h6", "h2h4", "f7f6", "c7d7", "e8f7", "d7b7", "d8d3", "b7b8", "d3h7", "b8c8", "f7g6", "c8e6"}

func TestStalemateEvidenceAndTermination(t *testing.T) {
	last := len(stalemateHistory) - 1
	line, err := ValidateLine(stalemateHistory[:last], Line{Moves: stalemateHistory[last:], Depth: 8, Score: Score{"cp", 0}})
	if err != nil {
		t.Fatal(err)
	}
	if line.Draws[1] != chess.Stalemate.String() {
		t.Fatalf("missing stalemate: %+v", line)
	}
	g, err := Replay(stalemateHistory)
	if err != nil || g.Outcome() != chess.Draw || len(g.ValidMoves()) != 0 {
		t.Fatalf("invalid terminal replay: %v", err)
	}
	for _, fact := range lineEvidence(line, 1) {
		if fact.Kind == "checkmate" {
			t.Fatal("stalemate described as mate")
		}
	}
}

func TestNativeEngineStalemateCandidate(t *testing.T) {
	path := os.Getenv("COACH_TEST_ENGINE")
	if path == "" {
		t.Skip("set COACH_TEST_ENGINE to test pinned native Arasan")
	}
	last := len(stalemateHistory) - 1
	line, err := NewEngine(path).Analyze(context.Background(), stalemateHistory[:last], stalemateHistory[last])
	if err != nil {
		t.Fatal(err)
	}
	if line.Score != (Score{"cp", 0}) || line.Draws[1] != chess.Stalemate.String() || len(line.Moves) != 1 {
		t.Fatalf("incorrect stalemate analysis: %+v", line)
	}
}
