package coach

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
)

// AnthropicProvider streams chat responses from the Anthropic API.
type AnthropicProvider struct {
	client anthropic.Client
	model  anthropic.Model
}

// NewAnthropicProvider builds a provider with the given API key. Defaults to
// Claude Sonnet 4.6 — a good fit for chess-review chat with structured UI
// output. Override via COACH_MODEL env var or by request body model field.
func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	model := anthropic.ModelClaudeSonnet4_6
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
// and pushes deltas onto the returned Frame channel. Supports both plain text
// and tool-call streaming (for Hashbrown's structured-output protocol).
func (p *AnthropicProvider) StreamChat(
	ctx context.Context,
	req ChatRequest,
	gctx *GameContext,
) (<-chan Frame, error) {
	out := make(chan Frame, 64)

	systemBlocks := buildSystemBlocks(req.System, gctx)
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
		Model:     resolveModel(req.Model, p.model),
		MaxTokens: 8192,
		System:    systemBlocks,
		Messages:  messages,
	}

	forceTool := req.ToolChoice == "required"
	if tools := convertTools(req.Tools); len(tools) > 0 {
		params.Tools = tools
	}
	if forceTool {
		params.ToolChoice = anthropic.ToolChoiceUnionParam{
			OfAny: &anthropic.ToolChoiceAnyParam{},
		}
	}
	// Adaptive thinking is incompatible with forced tool_choice.
	if !forceTool {
		params.Thinking = anthropic.ThinkingConfigParamUnion{
			OfAdaptive: &anthropic.ThinkingConfigAdaptiveParam{},
		}
	}

	stream := p.client.Messages.NewStreaming(ctx, params)

	go func() {
		defer close(out)
		out <- StartFrame()
		out <- Frame{Type: "generation-chunk", Chunk: &Chunk{Choices: []Choice{{Index: 0, Delta: Delta{Role: "assistant"}}}}}

		for stream.Next() {
			event := stream.Current()
			switch ev := event.AsAny().(type) {
			case anthropic.ContentBlockStartEvent:
				if tu, ok := ev.ContentBlock.AsAny().(anthropic.ToolUseBlock); ok {
					out <- ToolCallStartChunk(int(ev.Index), tu.ID, tu.Name)
				}

			case anthropic.ContentBlockDeltaEvent:
				switch d := ev.Delta.AsAny().(type) {
				case anthropic.TextDelta:
					if d.Text != "" {
						out <- TextChunk(d.Text)
					}
				case anthropic.InputJSONDelta:
					if d.PartialJSON != "" {
						out <- ToolCallArgsChunk(int(ev.Index), d.PartialJSON)
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

// resolveModel returns the request model if set, otherwise the provider default.
func resolveModel(reqModel string, fallback anthropic.Model) anthropic.Model {
	if reqModel != "" {
		return anthropic.Model(reqModel)
	}
	return fallback
}

// convertTools maps Hashbrown tool definitions to Anthropic SDK params.
func convertTools(defs []ToolDef) []anthropic.ToolUnionParam {
	if len(defs) == 0 {
		return nil
	}
	out := make([]anthropic.ToolUnionParam, 0, len(defs))
	for _, d := range defs {
		schema := anthropic.ToolInputSchemaParam{}
		if len(d.Parameters) > 0 {
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(d.Parameters, &raw); err == nil {
				if props, ok := raw["properties"]; ok {
					var p any
					if json.Unmarshal(props, &p) == nil {
						schema.Properties = p
					}
				}
				if req, ok := raw["required"]; ok {
					var r []string
					if json.Unmarshal(req, &r) == nil {
						schema.Required = r
					}
				}
			}
		}
		tool := anthropic.ToolUnionParamOfTool(schema, d.Name)
		if d.Description != "" {
			tool.OfTool.Description = param.NewOpt(d.Description)
		}
		out = append(out, tool)
	}
	return out
}

// buildSystemBlocks composes the system prompt. When reqSystem is set (from
// Hashbrown's compiled prompt including UI schema), it replaces the hardcoded
// base. The game-context portion sits behind a cache_control breakpoint so
// repeated turns hit the cache.
func buildSystemBlocks(reqSystem string, gctx *GameContext) []anthropic.TextBlockParam {
	base := reqSystem
	if base == "" {
		base = `You are a chess coach reviewing a completed game with the player who just finished it.

Your job is to walk them through 2-4 critical moments where the evaluation swung significantly. For each moment:
- Address the player in second person ("you", "your move").
- Be encouraging but honest — name the mistake clearly without being harsh.
- Explain the chess reason (tactical or strategic), not just the eval delta.
- Suggest the better continuation and why.

Keep individual replies tight (2-5 sentences per moment). Use standard algebraic notation for moves (e.g. Nxe5, Bb4+). When asked follow-up questions, focus on the position the user is asking about; don't recap the whole game.`
	}

	if gctx == nil {
		return []anthropic.TextBlockParam{
			{Text: base, CacheControl: anthropic.NewCacheControlEphemeralParam()},
		}
	}

	contextText := renderGameContext(gctx)
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

// buildMessages converts the Hashbrown-style messages into the Anthropic SDK's
// MessageParam shape. Handles plain text, assistant tool calls, and tool
// result messages. Coalesces consecutive tool-result messages into a single
// user message.
func buildMessages(in []Message) ([]anthropic.MessageParam, error) {
	out := make([]anthropic.MessageParam, 0, len(in))

	for i := 0; i < len(in); i++ {
		m := in[i]
		switch m.Role {
		case "user":
			out = append(out, anthropic.NewUserMessage(anthropic.NewTextBlock(m.ContentString())))

		case "assistant":
			var blocks []anthropic.ContentBlockParamUnion
			if s := m.ContentString(); s != "" {
				blocks = append(blocks, anthropic.NewTextBlock(s))
			}
			for _, tc := range m.ToolCalls {
				args := json.RawMessage(tc.Function.Arguments)
				if len(args) == 0 {
					args = json.RawMessage("{}")
				}
				blocks = append(blocks, anthropic.NewToolUseBlock(
					tc.ID,
					args,
					tc.Function.Name,
				))
			}
			if len(blocks) > 0 {
				out = append(out, anthropic.NewAssistantMessage(blocks...))
			}

		case "tool":
			// Coalesce consecutive tool-result messages into one user message.
			var results []anthropic.ContentBlockParamUnion
			for i < len(in) && in[i].Role == "tool" {
				results = append(results, anthropic.NewToolResultBlock(
					in[i].ToolCallID,
					in[i].ContentString(),
					false,
				))
				i++
			}
			i-- // outer loop will increment
			out = append(out, anthropic.NewUserMessage(results...))

		case "system":
			// Drop — system prompt is set via params.System.

		default:
			return nil, fmt.Errorf("coach: unknown message role %q", m.Role)
		}
	}
	return out, nil
}
