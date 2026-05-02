package coach

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// AnthropicProvider streams chat responses from the Anthropic API.
type AnthropicProvider struct {
	client anthropic.Client
	model  anthropic.Model
}

// NewAnthropicProvider builds a provider with the given API key. Defaults to
// Claude Opus 4.7 — overridden by COACH_MODEL env var if set to a recognised
// model id. Sonnet 4.6 is a reasonable lower-cost alternative for this
// workload.
func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	model := anthropic.ModelClaudeOpus4_7
	// COACH_MODEL accepts a free-form model id so future model launches don't
	// require code changes.
	return &AnthropicProvider{
		client: anthropic.NewClient(option.WithAPIKey(apiKey)),
		model:  model,
	}
}

// WithModel overrides the model id (mostly for tests).
func (p *AnthropicProvider) WithModel(m anthropic.Model) *AnthropicProvider {
	p.model = m
	return p
}

// StreamChat builds a Messages API request, opens the streaming response,
// and pushes text deltas onto the returned Frame channel. The system prompt
// containing the game context is marked cache_control: ephemeral so repeated
// chat turns within one review session reuse the same prefix.
func (p *AnthropicProvider) StreamChat(
	ctx context.Context,
	req ChatRequest,
	gctx *GameContext,
) (<-chan Frame, error) {
	out := make(chan Frame, 64)

	systemBlocks := buildSystemBlocks(gctx)
	messages, err := buildMessages(req.Messages)
	if err != nil {
		close(out)
		return nil, err
	}
	if len(messages) == 0 {
		close(out)
		return nil, fmt.Errorf("coach: no user messages")
	}

	params := anthropic.MessageNewParams{
		Model:     p.model,
		MaxTokens: 8192,
		System:    systemBlocks,
		Messages:  messages,
		Thinking: anthropic.ThinkingConfigParamUnion{
			OfAdaptive: &anthropic.ThinkingConfigAdaptiveParam{},
		},
	}

	stream := p.client.Messages.NewStreaming(ctx, params)

	go func() {
		defer close(out)
		out <- Frame{Type: "chunk", Chunk: &Chunk{Choices: []Choice{{Index: 0, Delta: Delta{Role: "assistant"}}}}}
		for stream.Next() {
			event := stream.Current()
			switch ev := event.AsAny().(type) {
			case anthropic.ContentBlockDeltaEvent:
				switch d := ev.Delta.AsAny().(type) {
				case anthropic.TextDelta:
					if d.Text != "" {
						out <- TextChunk(d.Text)
					}
				}
			}
		}
		if err := stream.Err(); err != nil {
			out <- ErrorFrame(err.Error())
			return
		}
		out <- FinishFrame()
	}()

	return out, nil
}

// buildSystemBlocks composes the system prompt. The game-context portion sits
// behind a cache_control breakpoint so repeated turns hit the cache.
func buildSystemBlocks(gctx *GameContext) []anthropic.TextBlockParam {
	base := `You are a chess coach reviewing a completed game with the player who just finished it.

Your job is to walk them through 2-4 critical moments where the evaluation swung significantly. For each moment:
- Address the player in second person ("you", "your move").
- Be encouraging but honest — name the mistake clearly without being harsh.
- Explain the chess reason (tactical or strategic), not just the eval delta.
- Suggest the better continuation and why.

Keep individual replies tight (2-5 sentences per moment). Use standard algebraic notation for moves (e.g. Nxe5, Bb4+). When asked follow-up questions, focus on the position the user is asking about; don't recap the whole game.`

	if gctx == nil {
		return []anthropic.TextBlockParam{
			{Text: base, CacheControl: anthropic.NewCacheControlEphemeralParam()},
		}
	}

	contextText := renderGameContext(gctx)
	// Two blocks: stable prelude + stable game context. Cache_control on the
	// last block caches everything from the prefix up to that point. The
	// chat messages (volatile) come after.
	return []anthropic.TextBlockParam{
		{Text: base},
		{Text: contextText, CacheControl: anthropic.NewCacheControlEphemeralParam()},
	}
}

func renderGameContext(gctx *GameContext) string {
	var sb strings.Builder
	sb.WriteString("GAME UNDER REVIEW\n")
	if gctx.GameID != "" {
		fmt.Fprintf(&sb, "ID: %s\n", gctx.GameID)
	}
	if gctx.PlayerColor != "" {
		fmt.Fprintf(&sb, "Player's color: %s\n", gctx.PlayerColor)
	}
	if gctx.Result != "" {
		fmt.Fprintf(&sb, "Result: %s\n", gctx.Result)
	}
	sb.WriteString("\nPGN:\n")
	if gctx.PGN != "" {
		sb.WriteString(gctx.PGN)
	} else {
		sb.WriteString("(unavailable)")
	}
	sb.WriteString("\n")

	if len(gctx.Evals) > 0 {
		sb.WriteString("\nPER-PLY EVALUATION (centipawns from white POV; +500 = +5.00):\n")
		for _, e := range gctx.Evals {
			fmt.Fprintf(&sb, "  ply=%d", e.Ply)
			if e.EvalCp != nil {
				fmt.Fprintf(&sb, " cp=%+d", *e.EvalCp)
			}
			if e.MateIn != nil {
				fmt.Fprintf(&sb, " mate=%+d", *e.MateIn)
			}
			if e.BestMove != "" {
				fmt.Fprintf(&sb, " best=%s", e.BestMove)
			}
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// buildMessages converts the Hashbrown-style messages into the Anthropic
// SDK's MessageParam shape. Skips system roles (we build the system prompt
// separately) and rejects an empty conversation.
func buildMessages(in []Message) ([]anthropic.MessageParam, error) {
	out := make([]anthropic.MessageParam, 0, len(in))
	for _, m := range in {
		switch m.Role {
		case "user":
			out = append(out, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
		case "assistant":
			out = append(out, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		case "system":
			// Drop — system prompt is set via params.System.
		default:
			return nil, fmt.Errorf("coach: unknown message role %q", m.Role)
		}
	}
	return out, nil
}
