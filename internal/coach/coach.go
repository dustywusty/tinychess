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
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// ChatRequest is the inbound shape from the browser (Hashbrown's wire format).
type ChatRequest struct {
	Operation      string          `json:"operation,omitempty"`
	Model          string          `json:"model,omitempty"`
	System         string          `json:"system,omitempty"`
	Messages       []Message       `json:"messages"`
	Tools          []ToolDef       `json:"tools,omitempty"`
	ToolChoice     string          `json:"toolChoice,omitempty"`
	ResponseFormat json.RawMessage `json:"responseFormat,omitempty"`
	ThreadID       string          `json:"threadId,omitempty"`
	GameID         string          `json:"gameId,omitempty"`
	ClientID       string          `json:"clientId,omitempty"`
}

// ToolDef is a tool definition from the Hashbrown client. Parameters is
// a pass-through JSON Schema.
type ToolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// Message is a single conversation turn. Assistant messages may carry
// ToolCalls; tool-result messages carry ToolCallID.
// Content is json.RawMessage because Hashbrown sends tool-result content
// as a JSON object, not a string.
type Message struct {
	Role       string          `json:"role"`
	Content    json.RawMessage `json:"content,omitempty"`
	ToolCalls  []ToolCall      `json:"toolCalls,omitempty"`
	ToolCallID string          `json:"toolCallId,omitempty"`
}

// ContentString returns Content as a plain string. If the raw JSON is a
// quoted string it is unquoted; objects/arrays are returned as raw JSON text.
func (m Message) ContentString() string {
	if len(m.Content) == 0 {
		return ""
	}
	var s string
	if json.Unmarshal(m.Content, &s) == nil {
		return s
	}
	return string(m.Content)
}

// ToolCall represents a tool invocation on an assistant message.
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction is the name + arguments of a tool call.
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
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
	Role      string          `json:"role,omitempty"`
	Content   string          `json:"content,omitempty"`
	ToolCalls []ToolCallDelta `json:"toolCalls,omitempty"`
}

// ToolCallDelta is a streaming fragment of a tool call.
type ToolCallDelta struct {
	Index    int                    `json:"index"` // no omitempty — 0 is meaningful
	ID       string                 `json:"id,omitempty"`
	Type     string                 `json:"type,omitempty"`
	Function *ToolCallFunctionDelta `json:"function,omitempty"`
}

// ToolCallFunctionDelta carries incremental name/arguments for a tool call.
// Arguments must not use omitempty — the initial empty string prevents the
// client from concatenating "undefined" when merging deltas.
type ToolCallFunctionDelta struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments"`
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
