package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// HashbrownRequest represents the format hashbrown sends
type HashbrownRequest struct {
	Messages []OpenAIMessage `json:"messages"`
	GameID   string          `json:"gameId,omitempty"`
	ClientID string          `json:"clientId,omitempty"`
}

// OpenAIMessage represents a message in the OpenAI chat format
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIRequest represents a request to the OpenAI API
type OpenAIRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

// OpenAIStreamResponse represents a chunk from the OpenAI streaming API
type OpenAIStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// HandleCoach handles chess coach chat requests from hashbrown
func (h *Handler) HandleCoach(w http.ResponseWriter, r *http.Request) {
	fmt.Println("=== HandleCoach called ===")
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("ERROR: OPENAI_API_KEY not configured")
		http.Error(w, "OPENAI_API_KEY not configured", http.StatusInternalServerError)
		return
	}

	var req HashbrownRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("ERROR: Failed to decode JSON: %v\n", err)
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	// Hashbrown sends the messages in a different field structure
	var messages []OpenAIMessage
	if len(req.Messages) > 0 {
		messages = req.Messages
	}

	fmt.Printf("Received request with %d messages\n", len(messages))

	// Get game context from query params if available
	gameID := r.URL.Query().Get("gameId")
	clientID := r.URL.Query().Get("clientId")

	// Or from request body
	if gameID == "" {
		gameID = req.GameID
	}
	if clientID == "" {
		clientID = req.ClientID
	}

	// Inject game context if we have a gameID
	if gameID != "" {
		fmt.Printf("Game context: gameID=%s, clientID=%s\n", gameID, clientID)
		g, col := h.Hub.Get(gameID, clientID)
		g.Mu.Lock()
		state := g.StateLocked()
		g.Mu.Unlock()

		var playerColor string
		if col != nil {
			playerColor = col.String()
		} else {
			playerColor = "spectator"
		}

		systemPrompt := fmt.Sprintf(`You are Scooter, a helpful and friendly chess coach. The player is asking for advice about their current game.

Current position (FEN): %s
Player's color: %s
Turn: %s
Game status: %s

Recent moves (UCI notation): %s

Provide helpful, concise chess advice. Suggest the best move and explain why. Use standard chess notation (like e4, Nf3, etc.) when describing moves.`,
			state.FEN,
			playerColor,
			state.Turn,
			state.Status,
			strings.Join(state.UCI, " "),
		)

		// Prepend system message with game context
		messages = append([]OpenAIMessage{{Role: "system", Content: systemPrompt}}, messages...)
		fmt.Printf("Added system prompt with game context\n")
	}

	// Prepare OpenAI request
	openAIReq := OpenAIRequest{
		Model:    "gpt-4o-mini",
		Messages: messages,
		Stream:   true,
	}

	// Call OpenAI API with streaming
	reqBody, _ := json.Marshal(openAIReq)
	openAIHTTPReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"ok":    false,
			"error": "failed to create request",
		})
		return
	}

	openAIHTTPReq.Header.Set("Content-Type", "application/json")
	openAIHTTPReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	fmt.Println("Calling OpenAI API...")
	resp, err := client.Do(openAIHTTPReq)
	if err != nil {
		fmt.Printf("ERROR: Failed to call OpenAI API: %v\n", err)
		WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"ok":    false,
			"error": "failed to call OpenAI API",
		})
		return
	}
	defer resp.Body.Close()

	fmt.Printf("OpenAI API response status: %d\n", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("ERROR: OpenAI API error response: %s\n", string(body))
		WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"ok":    false,
			"error": fmt.Sprintf("OpenAI API error: %s", string(body)),
		})
		return
	}

	// Set up streaming response (hashbrown expects application/octet-stream)
	flusher, ok := w.(http.Flusher)
	if !ok {
		fmt.Println("ERROR: Streaming not supported")
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Convert OpenAI SSE stream to hashbrown frames
	fmt.Println("Starting to stream response...")
	lineCount := 0
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		// Parse SSE lines that start with "data: "
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				continue
			}

			// Parse the OpenAI chunk
			var chunk OpenAIStreamResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				fmt.Printf("ERROR: Failed to parse chunk: %v\n", err)
				continue
			}

			// Extract the content
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				content := chunk.Choices[0].Delta.Content
				if lineCount <= 3 || lineCount%20 == 0 {
					fmt.Printf("Streaming chunk %d: %s\n", lineCount, content)
				}

				// Create hashbrown frame in the format it expects
				frame := map[string]interface{}{
					"type": "chunk",
					"chunk": map[string]interface{}{
						"choices": []map[string]interface{}{
							{
								"index": 0,
								"delta": map[string]interface{}{
									"content": content,
									"role":    "assistant",
								},
								"finishReason": nil,
							},
						},
					},
				}

				// Encode as length-prefixed binary frame
				frameJSON, _ := json.Marshal(frame)
				frameLen := uint32(len(frameJSON))

				// Write 4-byte length prefix (big-endian)
				lengthBytes := make([]byte, 4)
				lengthBytes[0] = byte(frameLen >> 24)
				lengthBytes[1] = byte(frameLen >> 16)
				lengthBytes[2] = byte(frameLen >> 8)
				lengthBytes[3] = byte(frameLen)

				w.Write(lengthBytes)
				w.Write(frameJSON)
				flusher.Flush()
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("ERROR: Scanner error: %v\n", err)
	}

	// Send finish frame
	finishFrame := map[string]interface{}{
		"type": "finish",
	}
	frameJSON, _ := json.Marshal(finishFrame)
	frameLen := uint32(len(frameJSON))
	lengthBytes := make([]byte, 4)
	lengthBytes[0] = byte(frameLen >> 24)
	lengthBytes[1] = byte(frameLen >> 16)
	lengthBytes[2] = byte(frameLen >> 8)
	lengthBytes[3] = byte(frameLen)
	w.Write(lengthBytes)
	w.Write(frameJSON)
	flusher.Flush()

	fmt.Printf("Finished streaming %d chunks\n", lineCount)
}
