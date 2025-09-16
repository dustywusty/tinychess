package handlers

import (
	"context"
	"testing"

	"tinychess/internal/game"
)

func newGame(t *testing.T) *game.Game {
	hub := game.NewHub(nil)
	g, _, err := hub.Get(context.Background(), "test", "")
	if err != nil {
		t.Fatalf("failed to create game: %v", err)
	}
	return g
}

func TestAppendPromotionIfPawnRank8(t *testing.T) {
	g := newGame(t)
	if got := appendPromotionIfPawn(g, "a7a8"); got != "a7a8q" {
		t.Fatalf("expected a7a8q got %s", got)
	}
}

func TestAppendPromotionIfPawnRank1(t *testing.T) {
	g := newGame(t)
	if got := appendPromotionIfPawn(g, "a7a1"); got != "a7a1q" {
		t.Fatalf("expected a7a1q got %s", got)
	}
}

func TestNoPromotionForQueen(t *testing.T) {
	g := newGame(t)
	if got := appendPromotionIfPawn(g, "d1d8"); got != "d1d8" {
		t.Fatalf("queen move modified: %s", got)
	}
}

func TestNoPromotionForRook(t *testing.T) {
	g := newGame(t)
	if got := appendPromotionIfPawn(g, "a8a1"); got != "a8a1" {
		t.Fatalf("rook move modified: %s", got)
	}
}
