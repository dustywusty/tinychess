package game

import (
	"context"
	"testing"
)

func TestGetReturnsAssignedColor(t *testing.T) {
	h := NewHub(nil)
	_, c1, err := h.Get(context.Background(), "g", "c1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if c1 == nil {
		t.Fatalf("expected color for first client")
	}
	_, c2, err := h.Get(context.Background(), "g", "c2")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if c2 == nil {
		t.Fatalf("expected color for second client")
	}
	if *c1 == *c2 {
		t.Fatalf("expected different colors, got %v and %v", *c1, *c2)
	}
	_, c3, err := h.Get(context.Background(), "g", "c3")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if c3 != nil {
		t.Fatalf("expected spectator for third client")
	}
}
