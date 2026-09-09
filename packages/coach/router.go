package coach

import (
	"strings"
	"unicode/utf8"
)

// Route intentionally recognizes a small vocabulary. Unknown prompts are
// rejected locally instead of being forwarded to a general-purpose model.
func Route(text string) (Intent, string) {
	if utf8.RuneCountInString(text) > MaxPromptRunes {
		return IntentOffTopic, "too_long"
	}

	query := normalize(text)
	if query == "" {
		return IntentOffTopic, "empty"
	}
	if containsAny(query, "weather", "news", "stock price", "write code", "recipe", "politics", "movie") {
		return IntentOffTopic, "off_topic_keyword"
	}
	if containsAny(query, "another hint", "next hint", "more hint") {
		return IntentAnotherHint, "matched"
	}
	if strings.Contains(query, "hint") {
		return IntentHint, "matched"
	}
	if containsAny(query, "why was that bad", "why is that bad", "my mistake", "my blunder", "why bad") {
		return IntentWhyBad, "matched"
	}
	if containsAny(query, "why was that good", "why is that good", "why good") {
		return IntentWhyGood, "matched"
	}
	if containsAny(query, "last move", "move i just played", "move they just played") {
		return IntentExplainLastMove, "matched"
	}
	if containsAny(query, "explain the position", "what is happening", "evaluate this position") {
		return IntentExplainPosition, "matched"
	}
	if containsAny(query, "what should i notice", "what am i missing", "what matters here", "where should i look") {
		return IntentWhatShouldINotice, "matched"
	}
	if containsAny(query, "chess rule", "castling", "en passant", "promotion", "stalemate", "threefold", "fifty-move", "legal move", "in check") {
		return IntentChessRule, "matched"
	}
	return IntentOffTopic, "unsupported_intent"
}

func containsAny(text string, fragments ...string) bool {
	for _, fragment := range fragments {
		if strings.Contains(text, fragment) {
			return true
		}
	}
	return false
}
