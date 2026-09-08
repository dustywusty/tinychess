package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/corentings/chess/v2"
)

// PersistentState contains only durable state. Connections and cooldowns are
// transient. Version protects against silently interpreting a future format.
type PersistentState struct {
	Version    int               `json:"version"`
	UCI        []string          `json:"uci"`
	OwnerID    string            `json:"ownerId"`
	OwnerColor string            `json:"ownerColor"`
	Clients    map[string]string `json:"clients"`
	LastSeen   time.Time         `json:"lastSeen"`
}

func NewGame() *Game {
	color := chess.White
	if rand.Intn(2) == 0 {
		color = chess.Black
	}
	return &Game{
		g: chess.NewGame(), OwnerColor: color, LastSeen: time.Now(),
		Watchers:  make(map[chan []byte]struct{}),
		LastReact: make(map[string]time.Time), Clients: make(map[string]chess.Color),
	}
}

func (g *Game) PersistentState() PersistentState {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	p := PersistentState{Version: 1, UCI: g.MovesUCI(), OwnerID: g.OwnerID,
		OwnerColor: g.OwnerColor.String(), Clients: make(map[string]string), LastSeen: g.LastSeen}
	for id, color := range g.Clients {
		p.Clients[id] = color.String()
	}
	return p
}

// Restore replays the complete legal history to retain repetition, castling,
// en passant, promotions, and terminal outcomes. Invalid state fails closed.
func Restore(p PersistentState) (*Game, error) {
	if p.Version != 1 || p.LastSeen.IsZero() || len(p.Clients) > 2 {
		return nil, fmt.Errorf("invalid recovery state")
	}
	parseColor := func(value string) (chess.Color, error) {
		switch value {
		case "w":
			return chess.White, nil
		case "b":
			return chess.Black, nil
		default:
			return chess.NoColor, fmt.Errorf("invalid recovery color")
		}
	}
	ownerColor, err := parseColor(p.OwnerColor)
	if err != nil {
		return nil, err
	}
	g := NewGame()
	g.OwnerID, g.OwnerColor, g.LastSeen = p.OwnerID, ownerColor, p.LastSeen
	seen := make(map[chess.Color]bool)
	for id, value := range p.Clients {
		color, err := parseColor(value)
		if err != nil || id == "" || seen[color] {
			return nil, fmt.Errorf("invalid recovery seats")
		}
		seen[color] = true
		g.Clients[id] = color
	}
	if p.OwnerID != "" {
		if color, ok := g.Clients[p.OwnerID]; !ok || color != ownerColor {
			return nil, fmt.Errorf("invalid recovery owner")
		}
	} else if len(p.Clients) == 2 {
		return nil, fmt.Errorf("missing recovery owner")
	}
	for i, uci := range p.UCI {
		if err := g.makeMoveLocked(uci); err != nil {
			return nil, fmt.Errorf("invalid recovery move at ply %d", i+1)
		}
	}
	return g, nil
}

// ApplyCommitted replaces durable state without disconnecting local watchers.
// The caller holds OpMu and must call this only after the database commits.
func (g *Game) ApplyCommitted(from *Game) {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	g.g, g.Clients = from.g, from.Clients
	g.OwnerID, g.OwnerColor, g.LastSeen = from.OwnerID, from.OwnerColor, from.LastSeen
}
