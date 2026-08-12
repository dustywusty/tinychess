package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/corentings/chess/v2"
)

// WriteJSON writes a JSON response with the given status code
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// parseSquare converts a coordinate string like "e2" into a chess.Square.
func parseSquare(s string) chess.Square {
	if len(s) != 2 {
		return chess.NoSquare
	}
	file := s[0] - 'a'
	rank := s[1] - '1'
	if file > 7 || rank > 7 {
		return chess.NoSquare
	}
	return chess.Square(rank*8 + file)
}
