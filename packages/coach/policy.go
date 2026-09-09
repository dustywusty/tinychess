// Package coach enforces the deterministic product policy that sits in front
// of Stockfish and any optional language-model provider.
package coach

import "strings"

const (
	MaxQuestionsPerMove = 2
	MaxHintsPerMove     = 3
	MaxPromptRunes      = 240
)

type Intent string

const (
	IntentHint              Intent = "HINT"
	IntentAnotherHint       Intent = "ANOTHER_HINT"
	IntentWhyBad            Intent = "WHY_BAD"
	IntentWhyGood           Intent = "WHY_GOOD"
	IntentExplainPosition   Intent = "EXPLAIN_POSITION"
	IntentExplainLastMove   Intent = "EXPLAIN_LAST_MOVE"
	IntentWhatShouldINotice Intent = "WHAT_SHOULD_I_NOTICE"
	IntentChessRule         Intent = "CHESS_RULE"
	IntentOffTopic          Intent = "OFF_TOPIC"
	IntentRepeated          Intent = "REPEATED"
)

type Artifacts struct {
	Observation         string
	Hints               [MaxHintsPerMove]string
	MistakeExplanation  string
	BestMoveExplanation string
}

type Request struct {
	PositionKey string
	Text        string
	Artifacts   *Artifacts
}

// Decision tells the transport layer whether it can answer locally or should
// request a bounded explanation/position artifact from the analysis pipeline.
type Decision struct {
	Intent           Intent
	Canned           string
	HintLevel        int
	NeedsAnalysis    bool
	AllowExplanation bool
	Reason           string
}

type Session struct {
	PositionKey       string
	QuestionsThisMove int
	HintsThisMove     int
	LastCoachEvent    Intent
	LastQuestion      string
}

func (s *Session) Decide(req Request) Decision {
	if req.PositionKey != s.PositionKey {
		s.PositionKey = req.PositionKey
		s.QuestionsThisMove = 0
		s.HintsThisMove = 0
		s.LastCoachEvent = ""
		s.LastQuestion = ""
	}

	normalized := normalize(req.Text)
	if normalized != "" && normalized == s.LastQuestion {
		s.LastCoachEvent = IntentRepeated
		return Decision{
			Intent: IntentRepeated,
			Canned: "You've already asked that. Use what you know and make your move.",
			Reason: "same_question_same_position",
		}
	}

	intent, reason := Route(req.Text)
	if normalized != "" {
		s.LastQuestion = normalized
	}
	s.LastCoachEvent = intent

	switch intent {
	case IntentOffTopic:
		message := "I'm just here for the chess. Make your move, or ask me for a hint."
		if reason == "too_long" {
			message = "Keep it to one short chess question about this position."
		}
		return Decision{Intent: intent, Canned: message, Reason: reason}

	case IntentHint, IntentAnotherHint:
		if s.HintsThisMove >= MaxHintsPerMove {
			return Decision{
				Intent: intent,
				Canned: "You've got enough to work with. What would you play?",
				Reason: "hint_limit",
			}
		}
		s.HintsThisMove++
		level := s.HintsThisMove
		decision := Decision{Intent: intent, HintLevel: level}
		if req.Artifacts != nil && req.Artifacts.Hints[level-1] != "" {
			decision.Canned = req.Artifacts.Hints[level-1]
			decision.Reason = "cached_hint"
			return decision
		}
		decision.NeedsAnalysis = true
		decision.Reason = "missing_position_artifacts"
		return decision

	default:
		if s.QuestionsThisMove >= MaxQuestionsPerMove {
			return Decision{
				Intent: intent,
				Canned: "That's enough analysis for this move. What would you play?",
				Reason: "question_limit",
			}
		}
		s.QuestionsThisMove++
		return Decision{
			Intent:           intent,
			AllowExplanation: true,
			Reason:           "bounded_chess_question",
		}
	}
}

func normalize(text string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(text))), " ")
}
