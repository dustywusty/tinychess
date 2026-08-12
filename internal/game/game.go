package game

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/corentings/chess/v2"
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
	for _, m := range ms {
		out = append(out, m.String())
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
	return g.makeMoveLocked(uci)
}

// MakeMoveFor validates seat ownership and turn before applying a move. The
// accepted UCI is returned because promotion may be normalized to a queen.
func (g *Game) MakeMoveFor(clientID, uci string) (string, error) {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	playerColor, ok := g.Clients[clientID]
	if !ok {
		return "", fmt.Errorf("unknown client")
	}

	position := g.g.Position()
	if len(uci) == 4 && (uci[3] == '1' || uci[3] == '8') {
		candidate, err := chess.UCINotation{}.Decode(position, uci)
		if err == nil && position.Board().Piece(candidate.S1()).Type() == chess.Pawn {
			uci += "q"
		}
	}

	move, err := chess.UCINotation{}.Decode(position, uci)
	if err != nil {
		return "", err
	}
	piece := position.Board().Piece(move.S1())
	if piece == chess.NoPiece || piece.Color() != playerColor {
		return "", fmt.Errorf("wrong color")
	}
	if position.Turn() != playerColor {
		return "", fmt.Errorf("not your turn")
	}
	if err := g.makeMoveLocked(uci); err != nil {
		return "", err
	}
	g.LastSeen = time.Now()
	return uci, nil
}

func (g *Game) makeMoveLocked(uci string) error {
	mv, err := chess.UCINotation{}.Decode(g.g.Position(), uci)
	if err != nil {
		return err
	}
	valid := false
	for _, m := range g.g.ValidMoves() {
		if m.S1() == mv.S1() && m.S2() == mv.S2() && m.Promo() == mv.Promo() {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("illegal move")
	}
	return g.g.Move(mv, nil)
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

// RemoveClient removes a client from the game. If the client was the owner,
// the owner slot is cleared so another client can claim it later.
func (g *Game) RemoveClient(id string) {
	g.Mu.Lock()
	delete(g.Clients, id)
	if g.OwnerID == id {
		g.OwnerID = ""
	}
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
