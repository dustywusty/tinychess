package game

import (
	"sync"
	"time"

	"github.com/notnil/chess"
)

// Hub manages all active chess games
type Hub struct {
	Mu    sync.Mutex
	Games map[string]*Game
}

// Game represents a single chess game with its state and watchers
type Game struct {
	Mu         sync.Mutex
	g          *chess.Game
	Watchers   map[chan []byte]struct{}
	LastReact  map[string]time.Time
	LastSeen   time.Time
	OwnerID    string
	OwnerColor chess.Color
	Clients    map[string]chess.Color // clientId -> color
}

// MoveRequest represents a move request from a client
type MoveRequest struct {
	UCI      string `json:"uci"`
	ClientID string `json:"clientId"`
}

// ReactionRequest represents a reaction request from a client
type ReactionRequest struct {
	Emoji  string `json:"emoji"`
	Sender string `json:"sender"`
}

// GameState represents the current state of a game
type GameState struct {
	Kind     string   `json:"kind"`
	FEN      string   `json:"fen"`
	Turn     string   `json:"turn"`
	Status   string   `json:"status"`
	PGN      string   `json:"pgn"`
	UCI      []string `json:"uci"`
	LastSeen int64    `json:"lastSeen"`
	Watchers int      `json:"watchers"`
}

// ClientState represents the state sent to a specific client, including their color
type ClientState struct {
	GameState
	Color    *string `json:"color"`
	Role     string  `json:"role"`
	ClientID string  `json:"clientId"`
}

// ReactionPayload represents a reaction broadcast
type ReactionPayload struct {
	Kind   string `json:"kind"`
	Emoji  string `json:"emoji"`
	At     int64  `json:"at"`
	Sender string `json:"sender"`
}
