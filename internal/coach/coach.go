// Package coach implements the Phase 2 post-game review chat backend.
//
// The handler at POST /api/coach/chat receives messages from the browser
// (Hashbrown's React client), enriches the conversation with persisted game
// context (PGN, eval array, player color), forwards to the configured LLM
// provider with prompt caching enabled, and streams responses back in
// Hashbrown's binary frame format (4-byte big-endian length prefix + JSON).
//
// Provider is selected via COACH_PROVIDER env var (default "anthropic").
package coach

import (
	"context"
	"errors"
	"fmt"
	"os"
)

// ChatRequest is the inbound shape from the browser.
type ChatRequest struct {
	Messages []Message `json:"messages"`
	GameID   string    `json:"gameId,omitempty"`
	ClientID string    `json:"clientId,omitempty"`
}

// Message is a single conversation turn.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GameContext is fetched from storage and woven into the system prompt for
// review sessions. Stable for the lifetime of a review, so it's the natural
// content for prompt caching.
type GameContext struct {
	GameID       string
	PGN          string
	Result       string
	PlayerColor  string // "white" | "black" | "spectator"
	WhiteSession string
	BlackSession string
	Evals        []EvalEntry
}

// EvalEntry mirrors a single ply's Stockfish evaluation.
type EvalEntry struct {
	Ply      int
	FEN      string
	EvalCp   *int
	MateIn   *int
	BestMove string
}

// Frame is a single streamed chunk in Hashbrown's expected JSON shape.
// Mimics OpenAI's stream delta format so the same JS-side consumer code
// works regardless of provider.
type Frame struct {
	Type  string `json:"type"`            // "chunk" | "finish" | "error"
	Chunk *Chunk `json:"chunk,omitempty"` // populated when Type == "chunk"
	Error string `json:"error,omitempty"` // populated when Type == "error"
}

// Chunk is the payload of a "chunk" frame.
type Chunk struct {
	Choices []Choice `json:"choices"`
}

// Choice is one element of Chunk.Choices.
type Choice struct {
	Index        int     `json:"index"`
	Delta        Delta   `json:"delta"`
	FinishReason *string `json:"finishReason"`
}

// Delta is the per-chunk additive payload.
type Delta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// Provider is the LLM driver. StreamChat returns a channel that delivers
// Frames in order; the channel closes when the stream completes (either
// normally with a "finish" frame or with an "error" frame).
type Provider interface {
	StreamChat(ctx context.Context, req ChatRequest, gctx *GameContext) (<-chan Frame, error)
}

// New returns the configured Provider, or an error if the environment is
// unset or the chosen provider's credentials are missing.
func New() (Provider, error) {
	name := os.Getenv("COACH_PROVIDER")
	if name == "" {
		name = "anthropic"
	}
	switch name {
	case "anthropic":
		key := os.Getenv("ANTHROPIC_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("coach: ANTHROPIC_API_KEY not set")
		}
		return NewAnthropicProvider(key), nil
	default:
		return nil, fmt.Errorf("coach: unknown provider %q (supported: anthropic)", name)
	}
}

// ErrNotConfigured is returned by handlers when the coach can't be
// constructed (missing API key, unknown provider, etc.) so callers can map
// it to 503.
var ErrNotConfigured = errors.New("coach not configured")
