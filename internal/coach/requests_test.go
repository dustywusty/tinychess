package coach

import (
	"context"
	"testing"
	"time"
)

func TestRequestRetryReturnsCommittedResultWithoutRepeatingAction(t *testing.T) {
	engine := &fakeEngine{}
	s := NewService(engine, nil)
	ctx := context.Background()
	history := []string{"e2e4"}
	review := Request{ClientID: "owner", Action: "review", RequestID: "review-1"}
	p, err := s.Review(ctx, "game", history, review)
	if err != nil {
		t.Fatal(err)
	}
	calls := engine.calls
	retry, err := s.Review(ctx, "game", history, review)
	if err != nil || retry.ID != p.ID || engine.calls != calls {
		t.Fatal("creation retry repeated search", err)
	}
	next := Request{ClientID: "owner", LessonID: p.ID, Revision: p.Revision, Action: "next", RequestID: "next-1"}
	p, err = s.Action(ctx, "game", history, next)
	if err != nil {
		t.Fatal(err)
	}
	retry, err = s.Action(ctx, "game", history, next)
	if err != nil || retry.Revision != p.Revision || retry.Step != 2 {
		t.Fatal("retry advanced twice", retry, err)
	}
	next.Action = "back"
	if _, err = s.Action(ctx, "game", history, next); err == nil {
		t.Fatal("request ID reused with different action")
	}
}

type barrierEngine struct {
	started chan struct{}
	release chan struct{}
}

func (e barrierEngine) Analyze(ctx context.Context, h []string, c string) (Line, error) {
	e.started <- struct{}{}
	select {
	case <-e.release:
	case <-ctx.Done():
		return Line{}, ctx.Err()
	}
	return ValidateLine(h, Line{Moves: []string{c}, Depth: 10, Score: Score{"cp", 10}})
}
func TestConcurrentRequestsCommitOneViewRevision(t *testing.T) {
	for _, sameID := range []bool{false, true} {
		t.Run(map[bool]string{false: "competing", true: "duplicate"}[sameID], func(t *testing.T) {
			s := NewService(&fakeEngine{}, nil)
			ctx := context.Background()
			history := []string{"e2e4"}
			p, err := s.Review(ctx, "game", history, Request{ClientID: "owner"})
			if err != nil {
				t.Fatal(err)
			}
			engine := barrierEngine{make(chan struct{}, 2), make(chan struct{})}
			s.Engine = engine
			results := make(chan error, 2)
			for i := 0; i < 2; i++ {
				id := "one"
				if i == 1 && !sameID {
					id = "two"
				}
				go func(id string) {
					_, err := s.Action(ctx, "game", history, Request{ClientID: "owner", LessonID: p.ID, Revision: 1, RequestID: id, Action: "candidate", Candidate: "g1f3", Branch: "played", Step: 0})
					results <- err
				}(id)
			}
			for i := 0; i < 2; i++ {
				select {
				case <-engine.started:
				case <-time.After(time.Second):
					t.Fatal("search did not start")
				}
			}
			close(engine.release)
			successes := 0
			for i := 0; i < 2; i++ {
				if <-results == nil {
					successes++
				}
			}
			expected := 1
			if sameID {
				expected = 2
			}
			if successes != expected {
				t.Fatalf("got %d successes, expected %d", successes, expected)
			}
			if s.lessons[p.ID].Revision != 2 {
				t.Fatal("two revisions committed")
			}
		})
	}
}
