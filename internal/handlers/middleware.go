package handlers

import (
	"encoding/json"
	"net/http"
)

// WriteJSON writes a JSON response with the given status code
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// isPromotionToLastRank checks if a 4-character UCI move is a promotion to the last rank
func isPromotionToLastRank(uci string) bool {
	if len(uci) != 4 {
		return false
	}
	to := uci[2:]
	return to[1] == '1' || to[1] == '8'
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
