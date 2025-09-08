package game

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/notnil/chess"
	"tinychess/internal/logging"
)

// Touch updates the last seen timestamp for a game
func (g *Game) Touch() {
	g.Mu.Lock()
	g.LastSeen = time.Now()
	g.Mu.Unlock()
}

// MovesUCI returns the list of moves in UCI notation
func (g *Game) MovesUCI() []string {
	ms := g.g.Moves()
	out := make([]string, 0, len(ms))
	tmp := chess.NewGame()
	uci := chess.UCINotation{}
	for _, m := range ms {
		s := uci.Encode(tmp.Position(), m)
		out = append(out, s)
		if mv2, err := uci.Decode(tmp.Position(), s); err == nil {
			_ = tmp.Move(mv2)
		}
	}
	return out
}

// StateLocked returns the current game state (must be called with lock held)
func (g *Game) StateLocked() GameState {
	pos := g.g.Position()
	fen := pos.String()
	turn := pos.Turn().String()
	status := ""
	if g.g.Outcome() != chess.NoOutcome {
		status = fmt.Sprintf("%s by %s", g.g.Outcome().String(), g.g.Method().String())
	}
	pgn := g.g.String()
	return GameState{
		Kind:     "state",
		FEN:      fen,
		Turn:     turn,
		Status:   status,
		PGN:      pgn,
		UCI:      g.MovesUCI(),
		LastSeen: g.LastSeen.UnixMilli(),
		Watchers: len(g.Watchers),
	}
}

// Broadcast sends the current game state to all watchers
func (g *Game) Broadcast() {
	g.Mu.Lock()
	state := g.StateLocked()
	data, _ := json.Marshal(state)
	for ch := range g.Watchers {
		select {
		case ch <- data:
		default:
		}
	}
	g.Mu.Unlock()
}

// MakeMove attempts to make a move and returns the result
func (g *Game) MakeMove(uci string) error {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	return g.g.MoveStr(uci)
}

// Reset resets the game to the starting position
func (g *Game) Reset() {
	g.Mu.Lock()
	g.g = chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	// Debug: Print initial game state
	pos := g.g.Position()
	logging.Debugf("Game reset - FEN: %s, Castling: %s", pos.String(), pos.CastleRights())
	g.Mu.Unlock()
}

// AddWatcher adds a new watcher channel
func (g *Game) AddWatcher(ch chan []byte) {
	g.Mu.Lock()
	g.Watchers[ch] = struct{}{}
	g.Mu.Unlock()
}

// RemoveWatcher removes a watcher channel
func (g *Game) RemoveWatcher(ch chan []byte) {
	g.Mu.Lock()
	delete(g.Watchers, ch)
	g.Mu.Unlock()
}

// CanReact checks if a sender can send a reaction (cooldown check)
func (g *Game) CanReact(sender string) (bool, int) {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	now := time.Now()
	if t, ok := g.LastReact[sender]; ok && now.Sub(t) < 5*time.Second {
		wait := int(5 - now.Sub(t).Seconds())
		return false, wait
	}

	g.LastReact[sender] = now
	return true, 0
}

// BroadcastReaction sends a reaction to all watchers
func (g *Game) BroadcastReaction(payload ReactionPayload) {
	g.Mu.Lock()
	data, _ := json.Marshal(payload)
	for ch := range g.Watchers {
		select {
		case ch <- data:
		default:
		}
	}
	g.Mu.Unlock()
}
