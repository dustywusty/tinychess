package game

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/corentings/chess/v2"
	"github.com/google/uuid"

	"tinychess/internal/storage"
)

// NewHub creates a new game hub with an optional backing store.
func NewHub(store *storage.Store) *Hub {
	h := &Hub{Games: make(map[string]*Game), Store: store}
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

func newGameInstance(id string) *Game {
	color := randomColor()
	return &Game{
		ID:         id,
		g:          chess.NewGame(),
		Watchers:   make(map[chan []byte]struct{}),
		LastReact:  make(map[string]time.Time),
		Clients:    make(map[string]chess.Color),
		LastSeen:   time.Now(),
		OwnerColor: color,
	}
}

func randomColor() chess.Color {
	if rand.Intn(2) == 0 {
		return chess.Black
	}
	return chess.White
}

func colorFromString(s string) chess.Color {
	switch s {
	case "black":
		return chess.Black
	case "white":
		return chess.White
	default:
		return chess.NoColor
	}
}

func (g *Game) assignColor(clientID string) *chess.Color {
	if clientID == "" {
		return nil
	}
	g.Mu.Lock()
	defer g.Mu.Unlock()

	if col, ok := g.Clients[clientID]; ok {
		if g.OwnerID == "" {
			g.OwnerID = clientID
			g.OwnerColor = col
		}
		c := col
		return &c
	}

	if g.OwnerID == "" {
		if g.OwnerColor == chess.NoColor {
			g.OwnerColor = chess.White
		}
		g.OwnerID = clientID
		g.Clients[clientID] = g.OwnerColor
		c := g.OwnerColor
		return &c
	}

	if len(g.Clients) < 2 {
		var color chess.Color
		if g.OwnerColor == chess.White {
			color = chess.Black
		} else {
			color = chess.White
		}
		g.Clients[clientID] = color
		c := color
		return &c
	}

	return nil
}

func (h *Hub) hydrateGame(ctx context.Context, g *Game) error {
	if h.Store == nil {
		return nil
	}
	gameID, err := uuid.Parse(g.ID)
	if err != nil {
		return nil
	}
	persisted, err := h.Store.LoadGame(ctx, gameID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil
		}
		return err
	}

	if persisted.Game.FEN != "" {
		if opt, err := chess.FEN(persisted.Game.FEN); err == nil {
			g.g = chess.NewGame(opt)
		}
	}

	g.LastSeen = persisted.Game.LastSeen
	if g.LastSeen.IsZero() {
		g.LastSeen = time.Now()
	}

	if persisted.Game.OwnerID != uuid.Nil {
		g.OwnerID = persisted.Game.OwnerID.String()
	}
	if col := colorFromString(persisted.Game.OwnerColor); col != chess.NoColor {
		g.OwnerColor = col
	}

	for _, player := range persisted.Players {
		if !player.Active || player.UserID == uuid.Nil {
			continue
		}
		col := colorFromString(player.Color)
		if col == chess.NoColor {
			continue
		}
		g.Clients[player.UserID.String()] = col
	}

	if g.OwnerID == "" && persisted.Game.OwnerID != uuid.Nil {
		g.OwnerID = persisted.Game.OwnerID.String()
	}

	return nil
}

// Get retrieves an existing game or creates a new in-memory copy. If a client ID
// is provided, the player will be assigned a color (if available). The assigned
// color is returned when applicable.
func (h *Hub) Get(ctx context.Context, id, clientID string) (*Game, *chess.Color, error) {
	h.Mu.Lock()
	g, ok := h.Games[id]
	if !ok {
		g = newGameInstance(id)
		if err := h.hydrateGame(ctx, g); err != nil {
			h.Mu.Unlock()
			return nil, nil, err
		}
		h.Games[id] = g
	}
	h.Mu.Unlock()

	var assigned *chess.Color
	if clientID != "" {
		assigned = g.assignColor(clientID)
		if assigned != nil && h.Store != nil {
			gameUUID, err := uuid.Parse(id)
			if err == nil {
				userUUID, err := uuid.Parse(clientID)
				if err == nil {
					role := "player"
					if g.OwnerID == clientID {
						role = "owner"
					}
					now := time.Now()
					if err := h.Store.EnsureUserSession(ctx, gameUUID, userUUID, assigned.String(), role, now); err != nil {
						return g, assigned, err
					}
				}
			}
		}
	}

	return g, assigned, nil
}

// CreateGame creates a brand-new game, stores it if a backing store exists, and
// returns the identifier and assigned owner color.
func (h *Hub) CreateGame(ctx context.Context, ownerID string) (string, chess.Color, error) {
	ownerID = strings.TrimSpace(ownerID)
	if ownerID == "" {
		return "", chess.NoColor, errors.New("missing owner id")
	}
	ownerUUID, err := uuid.Parse(ownerID)
	if err != nil {
		return "", chess.NoColor, err
	}

	id := uuid.NewString()
	g := newGameInstance(id)
	g.OwnerID = ownerID
	g.Clients[ownerID] = g.OwnerColor

	h.Mu.Lock()
	h.Games[id] = g
	h.Mu.Unlock()

	if h.Store != nil {
		gameUUID, err := uuid.Parse(id)
		if err != nil {
			h.Mu.Lock()
			delete(h.Games, id)
			h.Mu.Unlock()
			return "", chess.NoColor, err
		}
		if err := h.Store.CreateGame(ctx, gameUUID, ownerUUID, g.OwnerColor.String(), g.LastSeen); err != nil {
			h.Mu.Lock()
			delete(h.Games, id)
			h.Mu.Unlock()
			return "", chess.NoColor, err
		}
		if err := h.Store.EnsureUserSession(ctx, gameUUID, ownerUUID, g.OwnerColor.String(), "owner", g.LastSeen); err != nil {
			h.Mu.Lock()
			delete(h.Games, id)
			h.Mu.Unlock()
			return "", chess.NoColor, err
		}
		g.Mu.Lock()
		state := g.StateLocked()
		g.Mu.Unlock()
		active := true
		fen := state.FEN
		pgn := state.PGN
		status := state.Status
		if err := h.Store.SaveGameState(ctx, gameUUID, storage.GameStateUpdate{
			FEN:      &fen,
			PGN:      &pgn,
			Status:   &status,
			Active:   &active,
			LastSeen: &g.LastSeen,
		}); err != nil {
			h.Mu.Lock()
			delete(h.Games, id)
			h.Mu.Unlock()
			return "", chess.NoColor, err
		}
	}

	return id, g.OwnerColor, nil
}
