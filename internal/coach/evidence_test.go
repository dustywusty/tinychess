package coach

import "testing"

func TestEvidenceOnlyDescribesRevealedSteps(t *testing.T) {
	line, err := ValidateLine([]string{"f2f3", "e7e5"}, Line{Moves: []string{"g2g4", "d8h4"}, Score: Score{"mate", -1}, Depth: 10})
	if err != nil {
		t.Fatal(err)
	}
	for _, fact := range lineEvidence(line, 1) {
		if fact.Kind == "checkmate" {
			t.Fatal("future checkmate revealed before its step")
		}
		if fact.FEN == "" || fact.ID == "" {
			t.Fatal("unlinked fact")
		}
	}
	seen := false
	for _, fact := range lineEvidence(line, 2) {
		if fact.Kind == "checkmate" {
			seen = true
			if fact.FEN != line.FENs[2] || fact.Provenance != "rules" {
				t.Fatal(fact)
			}
		}
	}
	if !seen {
		t.Fatal("verified checkmate absent")
	}
}
