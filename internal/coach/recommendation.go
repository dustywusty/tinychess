package coach

import (
	"fmt"
	"math"
	"strings"
)

type Recommendation struct {
	Move    string `json:"move"`
	Summary string `json:"summary"`
	Score   Score  `json:"score"`
}

func recommend(line Line) *Recommendation {
	move := line.Moves[0]
	piece := "pawn"
	if named, ok := map[byte]string{'N': "knight", 'B': "bishop", 'R': "rook", 'Q': "queen", 'K': "king", 'O': "king"}[line.SAN[0][0]]; ok {
		piece = named
	}
	summary := fmt.Sprintf("Try %s: move the %s from %s to %s.", line.SAN[0], piece, move[:2], move[2:4])
	if strings.HasSuffix(line.SAN[0], "#") {
		summary += " That is checkmate."
	} else if strings.HasSuffix(line.SAN[0], "+") {
		summary += " It gives check."
	} else if strings.Contains(line.SAN[0], "x") {
		summary += " It captures material."
	}
	if line.Score.Kind == "cp" {
		side := "White"
		if strings.Fields(line.FENs[0])[1] == "b" {
			side = "Black"
		}
		if line.Score.Value < 0 {
			if side == "White" {
				side = "Black"
			} else {
				side = "White"
			}
		}
		if math.Abs(float64(line.Score.Value)) < 50 {
			summary += " The engine evaluates the position as roughly balanced."
		} else {
			summary += fmt.Sprintf(" The engine puts %s ahead by about %.1f pawns.", side, math.Abs(float64(line.Score.Value))/100)
		}
	}
	return &Recommendation{Move: line.SAN[0], Summary: summary, Score: line.Score}
}
