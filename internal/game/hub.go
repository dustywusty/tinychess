package game

import (
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

// Get retrieves an existing game or creates a new one
func (h *Hub) Get(id string) *Game {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	if g, ok := h.Games[id]; ok {
		return g
	}
	ng := &Game{
		g:         chess.NewGame(chess.UseNotation(chess.UCINotation{})),
		Watchers:  make(map[chan []byte]struct{}),
		LastReact: make(map[string]time.Time),
		Clients:   make(map[string]time.Time),
		LastSeen:  time.Now(),
	}
	h.Games[id] = ng
	return ng
}
