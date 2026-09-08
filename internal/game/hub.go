package game

import (
	"time"

	"github.com/corentings/chess/v2"
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
				idle := time.Since(g.LastSeen) > 24*time.Hour && len(g.Watchers) == 0
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
	if !ok {
		g = NewGame()
		h.Games[id] = g
	}
	g.Touch()
	h.Mu.Unlock()
	return g, g.AssignClient(clientId)
}

// AssignClient restores an existing seat or claims an available seat.
func (g *Game) AssignClient(clientId string) *chess.Color {
	var assigned *chess.Color
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

	return assigned
}
