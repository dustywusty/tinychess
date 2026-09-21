package coach

import (
	"fmt"
	"math"
	"strings"
)

type MoveAssessment struct {
	Grade       string   `json:"grade"`
	Tone        string   `json:"tone"`
	Played      string   `json:"played"`
	Best        string   `json:"best"`
	PlayedScore Score    `json:"playedScore"`
	BestScore   Score    `json:"bestScore"`
	Summary     string   `json:"summary"`
	Difference  *float64 `json:"difference,omitempty"`
}

func assessMove(l *Lesson) *MoveAssessment {
	a := &MoveAssessment{Grade: "Analysis limited", Tone: "neutral", Played: l.Played.SAN[0], Best: l.Alternative.SAN[0], PlayedScore: l.Played.Score, BestScore: l.Alternative.Score}
	if l.Played.Moves[0] == l.Alternative.Moves[0] {
		a.Grade = "Best move"
		a.Tone = "good"
		a.Summary = fmt.Sprintf("%s is the engine’s first choice here. There isn’t a better move in this search.", a.Played)
	} else if l.Played.Depth >= 8 && l.Alternative.Depth >= 8 && a.PlayedScore.Kind == "cp" && a.BestScore.Kind == "cp" {
		loss := math.Max(0, float64(a.BestScore.Value-a.PlayedScore.Value)/100)
		a.Difference = &loss
		a.Tone = "good"
		a.Grade = "Good move"
		a.Summary = fmt.Sprintf("%s and %s leave you with a similar position.", a.Played, a.Best)
		if loss > .5 {
			a.Grade = "Inaccuracy"
			a.Tone = "caution"
		}
		if loss > 1 {
			a.Grade = "Mistake"
			a.Tone = "warning"
		}
		if loss > 2.5 {
			a.Grade = "Blunder"
		}
		if loss > .5 {
			a.Summary = fmt.Sprintf("%s was stronger than %s here — an advantage worth about %.1f pawns more.", a.Best, a.Played, loss)
		}
	} else if a.PlayedScore.Kind == "mate" && a.PlayedScore.Value < 0 && (a.BestScore.Kind == "cp" || a.BestScore.Value > 0) {
		a.Grade = "Allows checkmate"
		a.Tone = "warning"
		a.Summary = fmt.Sprintf("%s allows a checkmating line. %s avoids that line in this search.", a.Played, a.Best)
	} else {
		a.Summary = fmt.Sprintf("The engine prefers %s, but this search isn’t deep enough to grade %s reliably.", a.Best, a.Played)
	}
	if len(l.Played.SAN) > 1 && strings.HasSuffix(l.Played.SAN[1], "#") {
		a.Summary = fmt.Sprintf("After %s, your opponent can play %s — checkmate.", a.Played, l.Played.SAN[1])
		if l.Played.Moves[0] != l.Alternative.Moves[0] {
			a.Summary += fmt.Sprintf(" Try %s instead.", a.Best)
		}
	} else if len(l.Played.SAN) > 1 && strings.Contains(l.Played.SAN[1], "x") && a.Tone == "warning" {
		a.Summary += fmt.Sprintf(" Your opponent can reply %s, capturing a piece.", l.Played.SAN[1])
	}
	return a
}
