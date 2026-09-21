package coach

import (
	"fmt"
	"strings"
)

// Evidence is derived only from rules-validated lines and an identified search.
// No motif or forced-outcome claims are inferred from a principal variation.
type Evidence struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Text       string `json:"text"`
	FEN        string `json:"fen"`
	Step       int    `json:"step"`
	Provenance string `json:"provenance"`
	AnalysisID string `json:"analysisId,omitempty"`
	Score      *Score `json:"score,omitempty"`
}

func lineEvidence(line Line, step int) []Evidence {
	prefix := line.Analysis.ID
	if prefix == "" {
		prefix = Identity(line.Moves)
	}
	facts := []Evidence{{ID: prefix + ":evaluation", Kind: "engine-evaluation", Text: "Evaluation of this candidate line from its search root, normalized to the reviewed mover.", FEN: line.FENs[0], Provenance: "engine", AnalysisID: line.Analysis.ID, Score: &line.Score}}
	for i := 0; i < step; i++ {
		base := fmt.Sprintf("%s:step:%d", prefix, i+1)
		facts = append(facts, Evidence{ID: base + ":legal", Kind: "legal-move", Text: line.SAN[i] + " is a legal move from this position.", FEN: line.FENs[i], Step: i, Provenance: "rules"})
		if strings.Contains(line.SAN[i], "x") {
			facts = append(facts, Evidence{ID: base + ":capture", Kind: "capture", Text: line.SAN[i] + " makes a legal capture.", FEN: line.FENs[i], Step: i, Provenance: "rules"})
		}
		if strings.HasSuffix(line.SAN[i], "#") {
			facts = append(facts, Evidence{ID: base + ":mate", Kind: "checkmate", Text: "The demonstrated continuation ends in checkmate at this position.", FEN: line.FENs[i+1], Step: i + 1, Provenance: "rules"})
		} else if strings.HasSuffix(line.SAN[i], "+") {
			facts = append(facts, Evidence{ID: base + ":check", Kind: "check", Text: "The move gives check at this position.", FEN: line.FENs[i+1], Step: i + 1, Provenance: "rules"})
		}
		if method, ok := line.Draws[i+1]; ok {
			facts = append(facts, Evidence{ID: base + ":draw", Kind: "draw", Text: "The demonstrated continuation ends in a draw by " + method + ".", FEN: line.FENs[i+1], Step: i + 1, Provenance: "rules"})
		}
	}
	return facts
}
