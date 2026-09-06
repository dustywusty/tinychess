package handlers

import (
	"testing"

	"github.com/corentings/chess/v2"
)

func TestParseSquare(t *testing.T) {
	if got := parseSquare("e2"); got != chess.E2 {
		t.Fatalf("parseSquare(e2) = %v", got)
	}
	for _, value := range []string{"", "e", "i2", "a9", "22"} {
		if got := parseSquare(value); got != chess.NoSquare {
			t.Fatalf("parseSquare(%q) = %v, want NoSquare", value, got)
		}
	}
}
