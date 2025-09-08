package handlers

import (
	"testing"

	"tinychess/internal/game"
)

func newGame() *game.Game {
	hub := game.NewHub()
	return hub.Get("test", "")
}

func TestAppendPromotionIfPawnRank8(t *testing.T) {
	g := newGame()
	if got := appendPromotionIfPawn(g, "a7a8"); got != "a7a8q" {
		t.Fatalf("expected a7a8q got %s", got)
	}
}

func TestAppendPromotionIfPawnRank1(t *testing.T) {
	g := newGame()
	if got := appendPromotionIfPawn(g, "a7a1"); got != "a7a1q" {
		t.Fatalf("expected a7a1q got %s", got)
	}
}

func TestNoPromotionForQueen(t *testing.T) {
	g := newGame()
	if got := appendPromotionIfPawn(g, "d1d8"); got != "d1d8" {
		t.Fatalf("queen move modified: %s", got)
	}
}

func TestNoPromotionForRook(t *testing.T) {
	g := newGame()
	if got := appendPromotionIfPawn(g, "a8a1"); got != "a8a1" {
		t.Fatalf("rook move modified: %s", got)
	}
}
