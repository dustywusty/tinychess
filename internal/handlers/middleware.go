package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/corentings/chess/v2"
	"tinychess/internal/game"
)

// WriteJSON writes a JSON response with the given status code
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// appendPromotionIfPawn appends a queen promotion suffix if the move is a pawn
// reaching the last rank. Non-pawn moves are returned unchanged.
func appendPromotionIfPawn(g *game.Game, uci string) string {
	if len(uci) != 4 {
		return uci
	}

	to := uci[2:]
	if to[1] != '1' && to[1] != '8' {
		return uci
	}

	sq := parseSquare(uci[:2])
	if sq == chess.NoSquare {
		return uci
	}

	g.Mu.Lock()
	state := g.StateLocked()
	g.Mu.Unlock()

	fenOpt, err := chess.FEN(state.FEN)
	if err != nil {
		return uci
	}
	tmp := chess.NewGame(fenOpt)
	piece := tmp.Position().Board().Piece(sq)

	if piece.Type() == chess.Pawn {
		return uci + "q"
	}
	return uci
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

// isAllowedEmoji checks if an emoji is in the allowed list
func isAllowedEmoji(emoji string) bool {
	allowed := map[string]struct{}{
		"👍": {}, "👎": {}, "❤️": {}, "😠": {}, "😢": {}, "🎉": {}, "👏": {},
		"😂": {}, "🤣": {}, "😎": {}, "🤔": {}, "😏": {}, "🙃": {}, "😴": {}, "🫡": {}, "🤯": {}, "🤡": {},
		"♟️": {}, "♞": {}, "♝": {}, "♜": {}, "♛": {}, "♚": {}, "⏱️": {}, "🏳️": {}, "🔄": {}, "🏆": {},
		"🔥": {}, "💀": {}, "🩸": {}, "⚡": {}, "🚀": {}, "🕳️": {}, "🎯": {}, "💥": {}, "🧠": {},
		"🍿": {}, "☕": {}, "🐢": {}, "🐇": {}, "🤝": {}, "🤬": {},
		"🪦": {}, "🐌": {}, "🎭": {}, "🐙": {}, "🦄": {}, "🐒": {},
	}
	_, ok := allowed[emoji]
	return ok
}
