package coach

import "strings"

// Explicit requests to hide an answer are application constraints, not model
// preferences. Conservatively hide answers even if a hint phrase is negated;
// the player can always use the explicit reveal button to change that choice.
func requestsHiddenAnswer(text string) bool {
	text = strings.ToLower(strings.ReplaceAll(text, "’", "'"))
	for _, phrase := range []string{"hint", "no spoiler", "no-spoiler", "don't tell", "do not tell", "don't reveal", "do not reveal", "without revealing", "keep the answer hidden"} {
		if strings.Contains(text, phrase) {
			return true
		}
	}
	return false
}
