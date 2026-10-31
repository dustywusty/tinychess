package coach

import (
	"context"
	"errors"
)

var teachingViews = map[string]string{"keep": "Keep the displayed position", "move": "Demonstrate one move from its starting position and focus its piece", "compare": "Show both choices from the same starting position", "reply": "Show the selected move and the opponent's first possible reply", "exercise": "Hide the answer and invite a move on the board"}
var teachingTargets = map[string]string{"displayed": "The line currently displayed", "played": "The move actually played and its line", "recommended": "The recommended move and its line"}
var teachingFollowups = map[string]string{"why": "Ask why this move helps", "reply": "Explore the opponent's reply", "practice": "Try solving the position", "compare": "Compare with the other choice", "none": "No suggested follow-up"}

type TeachingPlan struct {
	Action   string `json:"action"`
	View     string `json:"view"`
	Target   string `json:"target"`
	Followup string `json:"followup"`
}
type TeachingTurn struct {
	Question string `json:"question"`
	Action   string `json:"action"`
	View     string `json:"view"`
}
type TeachingPrompt struct {
	Label    string `json:"label"`
	Question string `json:"question"`
}
type Teaching struct {
	View   string          `json:"view"`
	Title  string          `json:"title"`
	Focus  []string        `json:"focus"`
	Prompt *TeachingPrompt `json:"prompt,omitempty"`
}
type TeachingPlanner interface {
	Plan(context.Context, string, map[string]string, any) (TeachingPlan, error)
}

func planTeaching(ctx context.Context, router Router, text string, options map[string]string, view any) (*TeachingPlan, error) {
	if planner, ok := router.(TeachingPlanner); ok {
		plan, err := planner.Plan(ctx, text, options, view)
		if err != nil {
			return nil, err
		}
		if _, ok := options[plan.Action]; !ok {
			return nil, errors.New("unknown teaching action")
		}
		if _, ok := teachingViews[plan.View]; !ok {
			return nil, errors.New("unknown teaching view")
		}
		if _, ok := teachingTargets[plan.Target]; !ok {
			return nil, errors.New("unknown teaching target")
		}
		if _, ok := teachingFollowups[plan.Followup]; !ok {
			return nil, errors.New("unknown teaching follow-up")
		}
		return &plan, nil
	}
	action, err := router.Select(ctx, text, options, view)
	return &TeachingPlan{Action: action, View: "keep", Target: "displayed", Followup: "none"}, err
}

// The model selects finite choices. Only application code can move the view,
// reveal evidence, or select highlight squares from a validated engine line.
func applyTeaching(l *Lesson, plan *TeachingPlan, question string) {
	if plan == nil {
		l.Teaching = nil
		return
	}
	if plan.Action == "best_now" && plan.View == "exercise" && l.Reveal == 3 {
		adjusted := *plan
		adjusted.View = "move"
		plan = &adjusted
	}
	l.Conversation = append(append([]TeachingTurn{}, l.Conversation...), TeachingTurn{question, plan.Action, plan.View})
	if len(l.Conversation) > 6 {
		l.Conversation = l.Conversation[len(l.Conversation)-6:]
	}
	if plan.Action == "unsupported" || plan.Action == "clarify" || plan.Action == "return" {
		l.Teaching = nil
		return
	}
	t := &Teaching{View: plan.View, Focus: []string{}}
	if l.Reveal < 3 || plan.View == "exercise" || plan.Action == "hint" || plan.Action == "exercise" {
		l.Reveal = 0
		l.Step = 0
		l.Branch = "played"
		l.Mode = "exercising"
		l.Caption = ""
		t.View = "exercise"
		t.Title = "Your turn to try"
		t.Prompt = &TeachingPrompt{"Keep it to a hint", "Give me only a hint; don't reveal the answer."}
		l.Teaching = t
		return
	}
	branch := l.Branch
	if branch == "compare" {
		branch = "alternative"
	}
	if plan.Target == "recommended" || l.CurrentPosition && plan.Target == "played" {
		branch = "alternative"
	} else if plan.Target == "played" {
		branch = "played"
	}
	if plan.Action == "candidate" || len(plan.Action) > 10 && plan.Action[:10] == "candidate-" {
		branch = l.Branch
	}
	line := activeLine(l, branch)
	switch plan.View {
	case "compare":
		if !l.CurrentPosition && l.Played.Moves[0] != l.Alternative.Moves[0] {
			l.Branch = "compare"
			l.Step = 0
			t.Title = "Two choices, one position"
		} else {
			l.Branch = branch
			l.Step = 0
			t.View = "move"
			t.Title = "The recommended move"
			t.Focus = []string{line.Moves[0][:2]}
		}
	case "move":
		l.Branch = branch
		l.Step = 0
		t.Title = "Follow this piece"
		t.Focus = []string{line.Moves[0][:2]}
	case "reply":
		l.Branch = branch
		l.Step = min(2, len(line.Moves))
		t.Title = "A possible reply"
		if l.Step == 2 {
			t.Focus = []string{line.Moves[1][2:4]}
		} else {
			t.Title = "No reply in this line"
		}
	default:
		t.View = "keep"
		t.Title = "Let's look at this position"
	}
	prompts := map[string]TeachingPrompt{
		"why":      {"Why this move?", "Why does this move help?"},
		"reply":    {"Show the reply", "Show the opponent's first reply to this move."},
		"practice": {"Let me try", "Let me solve this position myself without revealing the answer."},
		"compare":  {"Compare the choices", "Show why the other move was better by comparing both choices."},
	}
	if p, ok := prompts[plan.Followup]; ok && !(l.CurrentPosition && plan.Followup == "compare") {
		t.Prompt = &p
	}
	l.Teaching = t
}
