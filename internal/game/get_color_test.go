package game

import "testing"

func TestGetReturnsAssignedColor(t *testing.T) {
	h := NewHub()
	_, c1 := h.Get("g", "c1")
	if c1 == nil {
		t.Fatalf("expected color for first client")
	}
	_, c2 := h.Get("g", "c2")
	if c2 == nil {
		t.Fatalf("expected color for second client")
	}
	if *c1 == *c2 {
		t.Fatalf("expected different colors, got %v and %v", *c1, *c2)
	}
	_, c3 := h.Get("g", "c3")
	if c3 != nil {
		t.Fatalf("expected spectator for third client")
	}
}
