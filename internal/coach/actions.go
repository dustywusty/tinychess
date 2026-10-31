package coach

// Shared action meanings keep evaluation and production routing in agreement.
func presentationOptions() map[string]string {
	return map[string]string{
		"best_now":    "Recommend the best next move from the current live game position, rather than reviewing a past move or advancing a hypothetical line",
		"played":      "Replay the move the player actually played",
		"alternative": "Show what the engine recommends instead of the played move",
		"compare":     "Compare the played move and the engine alternative",
		"next":        "Show the opponent response or advance one move in the displayed continuation",
		"back":        "Go back one move in the displayed continuation",
		"hint":        "Give a hint without revealing the answer",
		"exercise":    "Let the player solve the position or find the answer themselves by trying a move on the board",
		"return":      "Leave the lesson and return to the actual game",
		"explain":     "Explain why the selected move is good or weak using the visible evidence; do not assume it is bad",
		"clarify":     "More than one supplied legal move matches the requested target; ask the player to choose",
		"unsupported": "No supported action fits, or the request asks for live move submission, data changes, or permission changes",
	}
}
func initialOptions() map[string]string {
	return map[string]string{
		"best_now": "Recommend the best next move from the current live game position, rather than reviewing a past move or advancing a hypothetical line", "review": "Open a review or explanation of the selected played move", "hint": "Give only a hint for the selected played move", "unsupported": "Request is unrelated or needs an existing lesson", "clarify": "The selected move does not resolve the requested target"}
}
