package coach

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/corentings/chess/v2"
	"github.com/google/uuid"
)

type Request struct {
	teachingPlan *TeachingPlan
	conversation []TeachingTurn
	accounted    bool
	RequestID    string `json:"requestId"`
	ClientID     string `json:"clientId"`
	Action       string `json:"action"`
	Text         string `json:"text"`
	Ply          int    `json:"ply"`
	LessonID     string `json:"lessonId"`
	Revision     int    `json:"revision"`
	Branch       string `json:"branch"`
	Step         int    `json:"step"`
	Candidate    string `json:"candidate"`
	NoSpoilers   bool   `json:"noSpoilers"`
}
type Choice struct {
	Move  string `json:"move"`
	ID    string `json:"id"`
	Label string `json:"label"`
}
type Frame struct {
	Evidence []Evidence       `json:"evidence"`
	Analysis AnalysisMetadata `json:"analysis"`
	Quality  string           `json:"quality"`
	Caption  string           `json:"caption"`
	FEN      string           `json:"fen"`
	Arrows   []Arrow          `json:"arrows"`
	Score    Score            `json:"score"`
}
type Projection struct {
	Teaching        *Teaching          `json:"teaching,omitempty"`
	Conversation    []TeachingTurn     `json:"conversation,omitempty"`
	CurrentPosition bool               `json:"currentPosition,omitempty"`
	Recommendation  *Recommendation    `json:"recommendation,omitempty"`
	Assessment      *MoveAssessment    `json:"assessment,omitempty"`
	Evidence        []Evidence         `json:"evidence,omitempty"`
	Analysis        *AnalysisMetadata  `json:"analysis,omitempty"`
	Explain         bool               `json:"-"`
	Views           map[string][]Frame `json:"views,omitempty"`
	ID              string             `json:"id"`
	Revision        int                `json:"revision"`
	PositionID      string             `json:"positionId"`
	Ply             int                `json:"ply"`
	FEN             string             `json:"fen"`
	Mode            string             `json:"mode"`
	Branch          string             `json:"branch"`
	Step            int                `json:"step"`
	Steps           int                `json:"steps"`
	Caption         string             `json:"caption"`
	Arrows          []Arrow            `json:"arrows"`
	Choices         []Choice           `json:"choices,omitempty"`
	Reveal          int                `json:"reveal"`
	Quality         string             `json:"quality"`
	Score           *Score             `json:"score,omitempty"`
}
type Arrow struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label"`
}
type Lesson struct {
	Teaching                                                               *Teaching
	Conversation                                                           []TeachingTurn
	CurrentPosition                                                        bool
	CreationRequestID, CreationFingerprint, LastRequestID, LastFingerprint string
	LastChoices                                                            []Choice
	LastRequest, LastAction                                                string
	ExploredHistory                                                        []string
	Explored                                                               *Line
	ID, GameID, ClientID, PositionID                                       string
	History                                                                []string
	Ply, Revision, Reveal, Step                                            int
	Played, Alternative                                                    Line
	Branch, Mode, Caption                                                  string
	Created                                                                time.Time
}
type Router interface {
	Select(context.Context, string, map[string]string, any) (string, error)
}
type Service struct {
	budgets map[string]budget
	Engine  Analyzer
	Router  Router
	mu      sync.Mutex
	lessons map[string]*Lesson
}

func NewService(engine Analyzer, router Router) *Service {
	return &Service{Engine: engine, Router: router, lessons: map[string]*Lesson{}, budgets: map[string]budget{}}
}
func Identity(history []string) string {
	sum := sha256.Sum256([]byte("standard:startpos:" + strings.Join(history, " ")))
	return hex.EncodeToString(sum[:])
}
func (s *Service) Review(ctx context.Context, gameID string, history []string, r Request) (Projection, error) {
	if len(r.RequestID) > 64 {
		return Projection{}, errors.New("invalid request ID")
	}
	current := r.Action == "best_now"
	if current {
		r.Ply = len(history)
	}
	end := r.Ply + 1
	if current {
		end = r.Ply
	}
	fingerprint := requestFingerprint(r)
	r.NoSpoilers = r.NoSpoilers || requestsHiddenAnswer(r.Text)
	if r.Ply < 0 || (!current && r.Ply >= len(history)) || len(history) > 2000 {
		return Projection{}, errors.New("select a played move")
	}
	if r.RequestID != "" {
		s.mu.Lock()
		for _, existing := range s.lessons {
			if existing.GameID == gameID && existing.ClientID == r.ClientID && existing.CreationRequestID == r.RequestID {
				s.mu.Unlock()
				if existing.CreationFingerprint != fingerprint || existing.Revision != 1 || time.Since(existing.Created) > 30*time.Minute || existing.PositionID != Identity(history[:end]) {
					return Projection{}, errors.New("request no longer matches the review")
				}
				return replayResult(existing), nil
			}
		}
		s.mu.Unlock()
	}
	if !s.consume(gameID, &r) {
		return Projection{}, errors.New("coach request limit reached; try again in a minute")
	}
	s.mu.Lock()
	now := time.Now()
	for k, l := range s.lessons {
		if now.Sub(l.Created) > 30*time.Minute {
			delete(s.lessons, k)
		}
	}
	if len(s.lessons) >= 1000 {
		s.mu.Unlock()
		return Projection{}, errors.New("coach lesson capacity reached; try again later")
	}
	s.mu.Unlock()
	prefix := append([]string{}, history[:r.Ply]...)
	if current {
		position, err := Replay(prefix)
		if err != nil {
			return Projection{}, err
		}
		if position.Outcome() != chess.NoOutcome {
			return Projection{}, errors.New("this game has ended; review a played move instead")
		}
	}
	alternative, err := s.Engine.Analyze(ctx, prefix, "")
	if err != nil {
		return Projection{}, err
	}
	played := alternative
	if len(alternative.Moves) == 0 {
		return Projection{}, errors.New("analysis unavailable")
	}
	if !current && alternative.Moves[0] != history[r.Ply] {
		played, err = s.Engine.Analyze(ctx, prefix, history[r.Ply])
		if err != nil {
			return Projection{}, err
		}
	}
	if err := ctx.Err(); err != nil {
		return Projection{}, err
	}
	l := &Lesson{CurrentPosition: current, ID: uuid.NewString(), GameID: gameID, ClientID: r.ClientID, History: prefix, PositionID: Identity(history[:end]), Ply: r.Ply, Revision: 1, Played: played, Alternative: alternative, Reveal: 3, Branch: "played", Mode: "reviewing", Step: 1, Created: now, CreationRequestID: r.RequestID, CreationFingerprint: fingerprint, LastRequestID: r.RequestID, LastFingerprint: fingerprint, LastRequest: r.Text, LastAction: "review"}
	if current {
		l.Branch = "alternative"
		l.Step = 0
		l.LastAction = "best_now"
	}
	if r.NoSpoilers {
		l.Reveal = 0
		l.Step = 0
		l.Mode = "exercising"
	}
	l.Conversation = r.conversation
	applyTeaching(l, r.teachingPlan, r.Text)
	s.mu.Lock()
	if r.RequestID != "" {
		for _, existing := range s.lessons {
			if existing.GameID == gameID && existing.ClientID == r.ClientID && existing.CreationRequestID == r.RequestID {
				s.mu.Unlock()
				if existing.CreationFingerprint != fingerprint || existing.Revision != 1 {
					return Projection{}, errors.New("request no longer matches the review")
				}
				return replayResult(existing), nil
			}
		}
	}
	if len(s.lessons) >= 1000 {
		s.mu.Unlock()
		return Projection{}, errors.New("coach lesson capacity reached; try again later")
	}
	s.lessons[l.ID] = l
	s.mu.Unlock()
	p := project(l)
	p.Explain = r.Text != "" && !r.NoSpoilers
	return p, nil
}
func (s *Service) Action(ctx context.Context, gameID string, history []string, r Request) (Projection, error) {
	s.mu.Lock()
	original := s.lessons[r.LessonID]
	if original == nil || original.GameID != gameID || original.ClientID != r.ClientID || time.Since(original.Created) > 30*time.Minute {
		s.mu.Unlock()
		return Projection{}, errors.New("lesson expired; start a new review")
	}
	l := *original
	s.mu.Unlock()
	if len(r.RequestID) > 64 {
		return Projection{}, errors.New("invalid request ID")
	}
	fingerprint := requestFingerprint(r)
	end := l.Ply + 1
	if l.CurrentPosition {
		end = l.Ply
	}
	if len(history) < end || Identity(history[:end]) != l.PositionID {
		return Projection{}, errors.New("stale review")
	}
	if r.RequestID != "" && r.RequestID == l.LastRequestID {
		if fingerprint != l.LastFingerprint {
			return Projection{}, errors.New("request ID was reused with different arguments")
		}
		return replayResult(&l), nil
	}
	if r.Revision != l.Revision {
		return Projection{}, errors.New("stale review")
	}
	if r.Branch != "" {
		if r.Branch != "played" && r.Branch != "alternative" && r.Branch != "compare" && r.Branch != "explored" {
			return Projection{}, errors.New("invalid branch")
		}
		if (r.Branch == "explored" && l.Explored == nil) || (r.Branch == "compare" && r.Step != 0) {
			return Projection{}, errors.New("invalid branch context")
		}
		line := activeLine(&l, r.Branch)
		if r.Step < 0 || r.Step > len(line.Moves) || (l.Reveal < 3 && r.Step != 0) {
			return Projection{}, errors.New("invalid view step")
		}
		l.Branch = r.Branch
		l.Step = r.Step
	}
	action := r.Action
	if action == "language" {
		r.NoSpoilers = r.NoSpoilers || requestsHiddenAnswer(r.Text)
	}
	if r.NoSpoilers {
		l.Reveal = 0
		l.Step = 0
		l.Branch = "played"
		l.Mode = "exercising"
		l.Caption = ""
	}
	if action == "language" || action == "candidate" || action == "attempt" || action == "explain" {
		if !s.consume(gameID, &r) {
			return Projection{}, errors.New("coach request limit reached; try again in a minute")
		}
	}
	var choices []Choice
	if action == "language" {
		if s.Router == nil {
			return Projection{}, errors.New("language controls are unavailable; use the buttons")
		}
		options := presentationOptions()
		if l.CurrentPosition {
			delete(options, "played")
			delete(options, "compare")
		}
		if l.Reveal < 3 || r.NoSpoilers {
			delete(options, "alternative")
			delete(options, "compare")
			delete(options, "next")
			delete(options, "played")
		}
		g, err := Replay(displayedHistory(&l))
		if err != nil {
			return Projection{}, err
		}
		// Candidate IDs refer only to legal moves in the displayed position.
		candidates := map[string]string{}
		for i, m := range g.ValidMoves() {
			if l.Reveal < 3 || r.NoSpoilers {
				break
			}
			id := fmt.Sprintf("candidate-%d", i)
			u := m.String()
			candidates[id] = u
			san := (chess.AlgebraicNotation{}).Encode(g.Position(), &m)
			options[id] = "Explore legal move " + san + " (" + u + ")"
			choices = append(choices, Choice{ID: id, Label: san + " (" + u + ")", Move: u})
		}
		recommendedMove := ""
		if l.Reveal == 3 && !r.NoSpoilers {
			recommendedMove = l.Alternative.SAN[0]
		}
		plan, err := planTeaching(ctx, s.Router, r.Text, options, map[string]any{"branch": l.Branch, "step": l.Step, "sideToMove": g.Position().Turn().String(), "fen": g.Position().String(), "reviewedMove": l.Played.Moves[0], "noSpoilers": r.NoSpoilers || l.Reveal < 3, "candidatePositionId": Identity(displayedHistory(&l)), "recentRequest": l.LastRequest, "recentAction": l.LastAction, "conversation": l.Conversation, "currentPosition": l.CurrentPosition, "recommendedMove": recommendedMove})
		if err != nil {
			return Projection{}, err
		}
		r.teachingPlan = plan
		selected := plan.Action
		if _, ok := options[selected]; !ok {
			return Projection{}, errors.New("unsupported action")
		}
		if u, ok := candidates[selected]; ok {
			action = "candidate"
			r.Candidate = u
		} else {
			action = selected
		}
	}
	if action == "best_now" {
		r.Action = "best_now"
		r.conversation = l.Conversation
		return s.Review(ctx, gameID, history, r)
	}
	if l.CurrentPosition && (action == "played" || action == "compare") {
		return Projection{}, errors.New("this is a suggestion, not a played move; preview the recommendation instead")
	}
	switch action {
	case "return":
		l.Mode = "live"
		l.Caption = "Returned to the game."
	case "played", "alternative", "compare":
		if r.NoSpoilers {
			return Projection{}, errors.New("turn off hidden answers before revealing moves")
		}
		l.Reveal = 3
		l.Mode = "reviewing"
		l.Branch = action
		l.Step = 1
		if action == "compare" {
			l.Step = 0
		}
		l.Caption = ""
	case "next", "back":
		if l.Branch == "compare" {
			return Projection{}, errors.New("choose Your move or Alternative before stepping")
		}
		if l.Reveal < 3 {
			return Projection{}, errors.New("reveal a line first")
		}
		if action == "next" {
			l.Step++
		} else {
			l.Step--
		}
		l.Caption = ""
	case "hint":
		l.Branch = "played"
		l.Reveal = 0
		l.Step = 0
		l.Mode = "exercising"
		l.Caption = ""
	case "exercise":
		l.Branch = "played"
		l.Reveal = 0
		l.Step = 0
		l.Mode = "exercising"
		l.Caption = ""
	case "reveal":
		if r.NoSpoilers {
			return Projection{}, errors.New("turn off hidden answers first")
		}
		l.Reveal++
		if l.Reveal > 3 {
			l.Reveal = 3
		}
		l.Caption = ""
	case "candidate", "attempt":
		candidateHistory := l.History
		if action == "candidate" {
			if r.NoSpoilers || l.Reveal < 3 {
				return Projection{}, errors.New("reveal the continuation before exploring a line")
			}
			candidateHistory = displayedHistory(&l)
		} else if l.Mode != "exercising" {
			return Projection{}, errors.New("start an exercise before submitting an answer")
		}
		g, err := Replay(candidateHistory)
		if err != nil {
			return Projection{}, err
		}
		if err := play(g, r.Candidate); err != nil {
			return Projection{}, errors.New("illegal exercise move")
		}
		answer, err := s.Engine.Analyze(ctx, candidateHistory, r.Candidate)
		if err != nil {
			return Projection{}, err
		}
		if action == "attempt" {
			l.Caption = acceptance(l.Alternative, answer)
		} else {
			// UCI scores are relative to the search root's mover. Keep the
			// lesson's original mover perspective when exploring an opponent reply.
			if (len(candidateHistory)-len(l.History))%2 != 0 {
				answer.Score.Value = -answer.Score.Value
			}
			l.ExploredHistory = append([]string{}, candidateHistory...)
			l.Explored = &answer
			l.Branch = "explored"
			l.Step = 1
			l.Caption = "Your selected legal continuation."
		}
	case "explain":
		l.Caption = "" // Deterministic evidence caption remains available without narration.
	case "clarify":
		l.Caption = "Which legal move do you mean?"
	case "unsupported":
		l.Caption = "I can review this move, give hints, or navigate this lesson."
	default:
		return Projection{}, errors.New("unsupported action")
	}
	applyTeaching(&l, r.teachingPlan, r.Text)
	line := activeLine(&l, l.Branch)
	if l.Step < 0 {
		l.Step = 0
	}
	if l.Step > len(line.Moves) {
		l.Step = len(line.Moves)
	}
	if err := ctx.Err(); err != nil {
		return Projection{}, err
	}
	if r.Action == "language" {
		l.LastRequest = r.Text
	}
	l.LastAction = action
	l.LastRequestID = r.RequestID
	l.LastFingerprint = fingerprint
	if action == "clarify" {
		l.LastChoices = choices
	} else {
		l.LastChoices = nil
	}
	l.Revision++
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lessons[l.ID] != original || original.Revision != r.Revision {
		current := s.lessons[l.ID]
		if r.RequestID != "" && current != nil && current.LastRequestID == r.RequestID && current.LastFingerprint == fingerprint {
			return replayResult(current), nil
		}
		return Projection{}, errors.New("stale review")
	}
	s.lessons[l.ID] = &l
	p := project(&l)
	p.Explain = action == "explain" || (r.teachingPlan != nil && (r.teachingPlan.View == "compare" || r.teachingPlan.View == "reply" || r.teachingPlan.View == "move"))
	if action == "clarify" {
		p.Choices = choices
	}
	return p, nil
}
func acceptance(best, answer Line) string {
	if best.Score.Kind == "cp" && answer.Score.Kind == "cp" && best.Depth == answer.Depth {
		if best.Score.Value-answer.Score.Value <= 50 {
			return "This is a sound alternative within this search: at most half a pawn behind the engine preference."
		}
		return "This search prefers another move. Try again or reveal a hint."
	}
	if best.Moves[0] == answer.Moves[0] {
		return "This matches the engine preference."
	}
	return "This legal answer needs deeper comparable analysis before I can judge it."
}
func project(l *Lesson) Projection {
	p := Projection{Teaching: l.Teaching, Conversation: l.Conversation, CurrentPosition: l.CurrentPosition, ID: l.ID, Revision: l.Revision, PositionID: l.PositionID, Ply: l.Ply, Mode: l.Mode, Branch: l.Branch, Step: l.Step, Reveal: l.Reveal, Arrows: []Arrow{}, Quality: "Arasan · bounded 1.5 s searches · shallow results are provisional"}
	line := activeLine(l, l.Branch)
	if l.Reveal < 3 {
		line = l.Played
	}
	p.FEN = line.FENs[0]
	if l.Reveal == 0 {
		p.Caption = "Look for checks, captures, and threats. What would you play?"
		if l.Caption != "" {
			p.Caption = l.Caption
		}
		return p
	}
	if l.Reveal == 1 {
		p.Caption = "Consider the piece on " + l.Alternative.Moves[0][:2] + "."
		return p
	}
	if l.Reveal == 2 {
		p.Caption = "Consider " + l.Alternative.SAN[0] + "."
		p.Arrows = append(p.Arrows, arrow(l.Alternative.Moves[0], "Alternative"))
		return p
	}
	if !l.CurrentPosition {
		p.Assessment = assessMove(l)
	}
	p.Views = map[string][]Frame{}
	variations := map[string]Line{"played": l.Played, "alternative": l.Alternative}
	if l.Explored != nil {
		variations["explored"] = *l.Explored
	}
	for branch, variation := range variations {
		label := "Your move"
		if branch == "alternative" {
			label = "Alternative"
		}
		for step, fen := range variation.FENs {
			arrows := []Arrow{}
			if step > 0 {
				arrows = append(arrows, arrow(variation.Moves[step-1], label))
			}
			p.Views[branch] = append(p.Views[branch], Frame{Evidence: lineEvidence(variation, step), Analysis: lessonAnalysis(variation), Quality: quality(variation), FEN: fen, Arrows: arrows, Score: variation.Score, Caption: frameCaption(branch, variation, step)})
		}
	}
	p.Views["compare"] = []Frame{{Evidence: append(lineEvidence(l.Played, 0), lineEvidence(l.Alternative, 0)...), Analysis: lessonAnalysis(l.Played), Quality: quality(l.Played), FEN: l.Played.FENs[0], Arrows: []Arrow{arrow(l.Played.Moves[0], "Your move"), arrow(l.Alternative.Moves[0], "Alternative")}, Score: l.Played.Score, Caption: "Compare the played move and alternative from the same starting position."}}
	p.Steps = len(line.Moves)
	p.FEN = line.FENs[l.Step]
	p.Evidence = lineEvidence(line, l.Step)
	p.Score = &line.Score
	metadata := lessonAnalysis(line)
	p.Analysis = &metadata
	p.Quality = quality(line)
	if l.Step > 0 {
		p.Arrows = append(p.Arrows, arrow(line.Moves[l.Step-1], map[bool]string{true: "Alternative", false: "Your move"}[l.Branch == "alternative"]))
	}
	if l.Branch == "compare" {
		p.Arrows = []Arrow{arrow(l.Played.Moves[0], "Your move"), arrow(l.Alternative.Moves[0], "Alternative")}
	}
	if l.CurrentPosition {
		p.Recommendation = recommend(l.Alternative)
		p.Caption = p.Recommendation.Summary
		if l.Step == 0 {
			p.Arrows = []Arrow{arrow(l.Alternative.Moves[0], "Alternative")}
			p.Evidence = lineEvidence(l.Alternative, 1)
		}
	} else {
		p.Caption = p.Assessment.Summary
	}
	if l.Step > 1 || l.Branch == "explored" {
		p.Caption += " " + frameCaption(l.Branch, line, l.Step)
	}
	if l.Teaching != nil {
		if l.Teaching.View == "move" {
			p.Caption = recommend(line).Summary
			p.Arrows = []Arrow{arrow(line.Moves[0], "Alternative")}
			p.Evidence = lineEvidence(line, 1)
		} else if l.Teaching.View == "compare" {
			p.Evidence = append(lineEvidence(l.Played, 1), lineEvidence(l.Alternative, 1)...)
		}
	}
	if l.Caption != "" {
		p.Caption = l.Caption
	}
	return p
}
func arrow(u, label string) Arrow { return Arrow{u[:2], u[2:4], label} }

func activeLine(l *Lesson, branch string) Line {
	if branch == "alternative" {
		return l.Alternative
	}
	if branch == "explored" && l.Explored != nil {
		return *l.Explored
	}
	return l.Played
}

// Local navigation is reported with a guarded branch/step on the next request.
// Preserve its complete history, including repetitions and earlier explorations.
func displayedHistory(l *Lesson) []string {
	history := l.History
	if l.Reveal < 3 || l.Branch == "compare" {
		return append([]string{}, history...)
	}
	if l.Branch == "explored" {
		history = l.ExploredHistory
	}
	line := activeLine(l, l.Branch)
	return append(append([]string{}, history...), line.Moves[:l.Step]...)
}

func frameCaption(branch string, line Line, step int) string {
	label := map[string]string{"played": "Played line", "alternative": "Alternative", "explored": "Explored line", "compare": "Comparison"}[branch]
	if step == 0 {
		return label + ": starting position."
	}
	if method, ok := line.Draws[step]; ok {
		return label + ": " + strings.Join(line.SAN[:step], " → ") + ". This continuation ends in a draw by " + method + "."
	}
	return label + ": " + strings.Join(line.SAN[:step], " → ") + "."
}

func lessonAnalysis(line Line) AnalysisMetadata {
	m := line.Analysis
	m.ScorePerspective = "reviewed-mover"
	return m
}
