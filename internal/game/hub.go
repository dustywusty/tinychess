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
// clients are assigned the opposite color. The returned color indicates the
// assigned color for the given client, or nil if the client is a spectator or no
// clientId was provided.
func (h *Hub) Get(id, clientId string) (*Game, *chess.Color) {
	h.Mu.Lock()
	g, ok := h.Games[id]
	var assigned *chess.Color
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
			c := g.OwnerColor
			assigned = &c
		}
		h.Games[id] = g
		h.Mu.Unlock()
		return g, assigned
	}
	h.Mu.Unlock()

	if clientId != "" {
		g.Mu.Lock()
		if col, exists := g.Clients[clientId]; exists {
			if g.OwnerID == "" {
				g.OwnerID = clientId
				g.OwnerColor = col
			}
			c := col
			assigned = &c
		} else if g.OwnerID == "" {
			g.OwnerID = clientId
			g.Clients[clientId] = g.OwnerColor
			c := g.OwnerColor
			assigned = &c
		} else if len(g.Clients) < 2 {
			var color chess.Color
			if g.OwnerColor == chess.White {
				color = chess.Black
			} else {
				color = chess.White
			}
			g.Clients[clientId] = color
			c := color
			assigned = &c
		}
		g.Mu.Unlock()
	}

	return g, assigned
}
