package game

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestBotSeatAndHistory(t *testing.T) {
	g := NewGame()
	if err := g.ConfigureBot("pip", "secret", "w"); err != nil {
		t.Fatal(err)
	}
	if g.AssignClient("visitor") != nil {
		t.Fatal("visitor claimed bot seat")
	}
	zero := 0
	if _, err := g.SubmitMove(MoveRequest{ClientID: "secret", UCI: "e2e4", BotMove: true, ExpectedPly: &zero}); err == nil {
		t.Fatal("bot moved human pieces")
	}
	if _, err := g.MakeMoveFor("secret", "e2e4"); err != nil {
		t.Fatal(err)
	}
	one := 1
	for _, request := range []MoveRequest{
		{ClientID: "visitor", UCI: "e7e5", BotMove: true, ExpectedPly: &one},
		{ClientID: "secret", UCI: "e7e5", BotMove: true},
		{ClientID: "secret", UCI: "e7e5", BotMove: true, ExpectedPly: &zero},
		{ClientID: "secret", UCI: "e7e5"},
		{ClientID: "secret", UCI: "e7e4", BotMove: true, ExpectedPly: &one},
	} {
		if _, err := g.SubmitMove(request); err == nil {
			t.Fatalf("accepted forbidden request: %+v", request)
		}
	}
	var accepted atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := g.SubmitMove(MoveRequest{ClientID: "secret", UCI: "e7e5", BotMove: true, ExpectedPly: &one}); err == nil {
				accepted.Add(1)
			}
		}()
	}
	wg.Wait()
	if accepted.Load() != 1 {
		t.Fatalf("accepted %d duplicate moves", accepted.Load())
	}
	restored, err := Restore(g.PersistentState())
	if err != nil {
		t.Fatal(err)
	}
	if restored.Bot == nil || restored.Bot.ID != "pip" || restored.AssignClient("visitor") != nil {
		t.Fatal("bot metadata/seat not recovered")
	}
	restored.RemoveClient("secret")
	if restored.OwnerID != "secret" {
		t.Fatal("bot owner released")
	}
	restored.Mu.Lock()
	state := restored.StateLocked()
	restored.Mu.Unlock()
	if len(state.UCI) != 2 || !strings.Contains(state.PGN, "e4 e5") || state.Turn != "w" {
		t.Fatalf("incorrect history: %+v", state)
	}
}

func TestBotCheckmate(t *testing.T) {
	g := NewGame()
	if err := g.ConfigureBot("ada", "owner", "w"); err != nil {
		t.Fatal(err)
	}
	for ply, uci := range []string{"f2f3", "e7e5", "g2g4", "d8h4"} {
		if _, err := g.SubmitMove(MoveRequest{ClientID: "owner", UCI: uci, BotMove: ply%2 == 1, ExpectedPly: &ply}); err != nil {
			t.Fatal(err)
		}
	}
	four := 4
	if _, err := g.SubmitMove(MoveRequest{ClientID: "owner", UCI: "a2a3", ExpectedPly: &four}); err == nil {
		t.Fatal("move after mate")
	}
	restored, err := Restore(g.PersistentState())
	if err != nil {
		t.Fatal(err)
	}
	restored.Mu.Lock()
	state := restored.StateLocked()
	restored.Mu.Unlock()
	if state.Status == "" || !strings.Contains(state.PGN, "0-1") {
		t.Fatal("mate not recovered")
	}
}

func TestBotRecoveryRejectsInvalidMetadata(t *testing.T) {
	g := NewGame()
	_ = g.ConfigureBot("max", "owner", "b")
	for _, bot := range []*BotOpponent{
		{ID: "unknown", Color: "w", PolicyVersion: 1}, {ID: "pip", Color: "b", PolicyVersion: 1},
		{ID: "pip", Color: "w", PolicyVersion: 100}, {ID: "pip", Color: "invalid", PolicyVersion: 1},
	} {
		state := g.PersistentState()
		state.Bot = bot
		if _, err := Restore(state); err == nil {
			t.Fatalf("accepted %+v", bot)
		}
	}
}
