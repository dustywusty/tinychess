package game

import (
	"encoding/json"
	"reflect"
	"sync"
	"testing"

	"github.com/corentings/chess/v2"
)

func TestConcurrentGameConstructionAndRecovery(t *testing.T) {
	var wg sync.WaitGroup
	for range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 10 {
				g := NewGame()
				if err := g.MakeMove("e2e4"); err != nil {
					t.Error(err)
					return
				}
				restored, err := Restore(g.PersistentState())
				if err != nil {
					t.Error(err)
					return
				}
				if err := restored.MakeMove("e7e5"); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func TestRecoveryRetainsHistoryAndRules(t *testing.T) {
	for _, tc := range []struct {
		name  string
		moves []string
		next  string
	}{
		{"captures", []string{"d2d4", "e7e5", "g1f3", "e5d4", "f3d4"}, "b8c6"},
		{"en passant", []string{"e2e4", "a7a6", "e4e5", "d7d5"}, "e5d6"},
		{"castling", []string{"e2e4", "e7e5", "g1f3", "b8c6", "f1c4", "g8f6"}, "e1g1"},
		{"promotion", []string{"a2a4", "h7h5", "a4a5", "h5h4", "a5a6", "h4h3", "a6b7", "h3g2", "b7a8q"}, "g2h1n"},
		{"checkmate", []string{"f2f3", "e7e5", "g2g4", "d8h4"}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGame()
			g.AssignClient("original-owner")
			g.AssignClient("original-friend")
			for _, uci := range tc.moves {
				if err := g.MakeMove(uci); err != nil {
					t.Fatalf("%s: %v", uci, err)
				}
			}
			encoded, err := json.Marshal(g.PersistentState())
			if err != nil {
				t.Fatal(err)
			}
			var saved PersistentState
			if err := json.Unmarshal(encoded, &saved); err != nil {
				t.Fatal(err)
			}
			restored, err := Restore(saved)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(g.StateLocked(), restored.StateLocked()) {
				t.Fatal("restored position/history differs")
			}
			if !reflect.DeepEqual(g.Clients, restored.Clients) || g.OwnerID != restored.OwnerID || g.OwnerColor != restored.OwnerColor {
				t.Fatal("restored seats differ")
			}
			if restored.AssignClient("spectator-first-after-restart") != nil {
				t.Fatal("spectator stole a seat")
			}
			if tc.next != "" {
				if err := restored.MakeMove(tc.next); err != nil {
					t.Fatal(err)
				}
			} else if restored.StateLocked().Status == "" || restored.MakeMove("a2a3") == nil {
				t.Fatal("terminal game reopened")
			}
		})
	}
}

func TestRecoveryRetainsRepetition(t *testing.T) {
	g := NewGame()
	cycle := []string{"g1f3", "g8f6", "f3g1", "f6g8"}
	for range 3 {
		for _, uci := range cycle {
			if err := g.MakeMove(uci); err != nil {
				t.Fatal(err)
			}
		}
	}
	restored, err := Restore(g.PersistentState())
	if err != nil {
		t.Fatal(err)
	}
	for _, uci := range cycle {
		if err := restored.MakeMove(uci); err != nil {
			t.Fatal(err)
		}
	}
	if restored.g.Outcome() != chess.Draw || restored.g.Method() != chess.FivefoldRepetition {
		t.Fatalf("repetition lost: %s", restored.StateLocked().Status)
	}
	if restored.MakeMove("e2e4") == nil {
		t.Fatal("accepted move after automatic draw")
	}
}

func TestRecoveryRejectsInvalidState(t *testing.T) {
	for name, corrupt := range map[string]func(*PersistentState){
		"version": func(p *PersistentState) { p.Version++ },
		"history": func(p *PersistentState) { p.UCI = []string{"e2e5"} },
		"color":   func(p *PersistentState) { p.OwnerColor = "purple" },
		"owner":   func(p *PersistentState) { p.OwnerID = "absent" },
		"seats":   func(p *PersistentState) { p.Clients = map[string]string{"a": "w", "b": "w"} },
	} {
		t.Run(name, func(t *testing.T) {
			p := NewGame().PersistentState()
			corrupt(&p)
			if _, err := Restore(p); err == nil {
				t.Fatal("accepted invalid recovery state")
			}
		})
	}
}

func TestApplyCommittedPreservesConnectionsAndCooldowns(t *testing.T) {
	g := NewGame()
	g.AssignClient("owner")
	ch := make(chan []byte, 1)
	g.AddWatcher(ch)
	g.CanReact("owner")
	candidate, err := Restore(g.PersistentState())
	if err != nil {
		t.Fatal(err)
	}
	if err := candidate.MakeMove("e2e4"); err != nil {
		t.Fatal(err)
	}
	if len(g.MovesUCI()) != 0 {
		t.Fatal("candidate changed live state before commit")
	}
	g.ApplyCommitted(candidate)
	g.Broadcast()
	select {
	case <-ch:
	default:
		t.Fatal("watcher lost during commit")
	}
	if allowed, _ := g.CanReact("owner"); allowed {
		t.Fatal("cooldown lost during commit")
	}
	p := g.PersistentState()
	delete(p.Clients, "owner")
	if len(g.Clients) != 1 {
		t.Fatal("snapshot aliases live seats")
	}
}
