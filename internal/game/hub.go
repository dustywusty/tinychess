package game

import (
	"math/rand"
	"time"

	"github.com/notnil/chess"
)

// NewHub creates a new game hub with cleanup goroutine
func NewHub() *Hub {
	h := &Hub{Games: make(map[string]*Game)}
	// cleanup goroutine
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			h.Mu.Lock()
			for id, g := range h.Games {
				g.Mu.Lock()
				idle := time.Since(g.LastSeen) > 24*time.Hour
				g.Mu.Unlock()
				if idle {
					delete(h.Games, id)
				}
			}
			h.Mu.Unlock()
		}
	}()
	return h
}

// Get retrieves an existing game or creates a new one. If a clientId is provided,
// the first one becomes the owner and is assigned a random color. Subsequent
// clients are assigned the opposite color.
func (h *Hub) Get(id, clientId string) *Game {
	h.Mu.Lock()
	g, ok := h.Games[id]
	if !ok {
		color := chess.White
		if rand.Intn(2) == 0 {
			color = chess.Black
		}
		g = &Game{
			g:          chess.NewGame(chess.UseNotation(chess.UCINotation{})),
			Watchers:   make(map[chan []byte]struct{}),
			LastReact:  make(map[string]time.Time),
			Clients:    make(map[string]chess.Color),
			LastSeen:   time.Now(),
			OwnerColor: color,
		}
		if clientId != "" {
			g.OwnerID = clientId
			g.Clients[clientId] = g.OwnerColor
		}
		h.Games[id] = g
		h.Mu.Unlock()
		return g
	}
	h.Mu.Unlock()

	if clientId != "" {
		g.Mu.Lock()
		if g.OwnerID == "" {
			g.OwnerID = clientId
			g.Clients[clientId] = g.OwnerColor
		} else if _, exists := g.Clients[clientId]; !exists {
			var color chess.Color
			if g.OwnerColor == chess.White {
				color = chess.Black
			} else {
				color = chess.White
			}
			g.Clients[clientId] = color
		}
		g.Mu.Unlock()
	}
	return g
}
