package coach

import (
	"encoding/json"
	"testing"
)

func TestDecodeChatRequest(t *testing.T) {
	// Simulates what Hashbrown sends for the initial request
	body := `{
		"operation": "generate",
		"model": "claude-sonnet-4-6",
		"system": "You are a chess coach...",
		"messages": [{"role": "user", "content": "I just finished playing this game."}],
		"tools": [
			{
				"name": "get_game_metadata",
				"description": "Return high-level metadata",
				"parameters": {"type": "object", "description": "No input"}
			},
			{
				"name": "get_position_at",
				"description": "Return the FEN at a given ply",
				"parameters": {"type": "object", "properties": {"ply": {"type": "integer"}}, "required": ["ply"]}
			},
			{
				"name": "output",
				"description": "This should be your final tool call.",
				"parameters": {"type": "object", "properties": {"ui": {"type": "array"}}, "required": ["ui"]}
			}
		],
		"toolChoice": "required"
	}`
	var req ChatRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if req.Operation != "generate" {
		t.Errorf("operation = %q, want generate", req.Operation)
	}
	if len(req.Tools) != 3 {
		t.Errorf("tools = %d, want 3", len(req.Tools))
	}
	if req.ToolChoice != "required" {
		t.Errorf("toolChoice = %q, want required", req.ToolChoice)
	}
}

func TestDecodeAssistantMessageWithToolCalls(t *testing.T) {
	body := `{
		"operation": "generate",
		"model": "claude-sonnet-4-6",
		"system": "You are a coach",
		"messages": [
			{"role": "user", "content": "Review my game"},
			{
				"role": "assistant",
				"content": "",
				"toolCalls": [{
					"id": "call_123",
					"index": 0,
					"type": "function",
					"function": {"name": "get_position_at", "arguments": "{\"ply\":5}"}
				}]
			},
			{
				"role": "tool",
				"content": "{\"ply\":5,\"fen\":\"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR\"}",
				"toolCallId": "call_123",
				"toolName": "get_position_at"
			}
		],
		"tools": [],
		"toolChoice": "required"
	}`
	var req ChatRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(req.Messages) != 3 {
		t.Errorf("messages = %d, want 3", len(req.Messages))
	}
	if len(req.Messages[1].ToolCalls) != 1 {
		t.Errorf("toolCalls = %d, want 1", len(req.Messages[1].ToolCalls))
	}
	if req.Messages[2].Role != "tool" {
		t.Errorf("role = %q, want tool", req.Messages[2].Role)
	}
	if req.Messages[2].ToolCallID != "call_123" {
		t.Errorf("toolCallId = %q, want call_123", req.Messages[2].ToolCallID)
	}
}

func TestDecodeToolResultWithObjectContent(t *testing.T) {
	// Hashbrown sends tool result content as an object, not a string
	body := `{
		"operation": "generate",
		"messages": [
			{"role": "user", "content": "hello"},
			{
				"role": "assistant",
				"content": "",
				"toolCalls": [{"id": "call_1", "type": "function", "function": {"name": "get_game_metadata", "arguments": ""}}]
			},
			{
				"role": "tool",
				"content": {"status": "rejected", "reason": {}},
				"toolCallId": "call_1",
				"toolName": "get_game_metadata"
			}
		],
		"tools": []
	}`
	var req ChatRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	// Object content should be returned as raw JSON
	got := req.Messages[2].ContentString()
	if got != `{"status": "rejected", "reason": {}}` {
		t.Errorf("content = %q", got)
	}
	// String content should be unquoted
	got = req.Messages[0].ContentString()
	if got != "hello" {
		t.Errorf("content = %q, want hello", got)
	}
}

func TestDecodeToolCallWithIndex(t *testing.T) {
	// The client sends 'index' on tool calls but our struct doesn't have it.
	// Go should silently ignore unknown fields.
	body := `{
		"id": "call_456",
		"index": 0,
		"type": "function",
		"function": {"name": "output", "arguments": "{\"ui\":[]}"}
	}`
	var tc ToolCall
	if err := json.Unmarshal([]byte(body), &tc); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if tc.ID != "call_456" {
		t.Errorf("id = %q, want call_456", tc.ID)
	}
}
