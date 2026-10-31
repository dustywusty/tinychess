package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"tinychess/internal/coach"
)

// HandleCoachStream adapts the policy-checked application service to AG-UI.
// Incoming AG-UI state and tool definitions are never execution authority.
func (h *Handler) HandleCoachStream(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ThreadID       string `json:"threadId"`
		RunID          string `json:"runId"`
		ForwardedProps struct {
			Coach coach.Request `json:"coach"`
		} `json:"forwardedProps"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&input) != nil || input.RunID == "" || input.ThreadID == "" {
		http.Error(w, "Invalid agent request", 400)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Accel-Buffering", "no")
	emit := func(event any) {
		encoded, _ := json.Marshal(event)
		_, _ = fmt.Fprintf(w, "data: %s\n\n", encoded)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
	emit(map[string]any{"type": "RUN_STARTED", "threadId": input.ThreadID, "runId": input.RunID})
	projection, status, message := h.coachRequest(r.Context(), r.PathValue("gameId"), input.ForwardedProps.Coach)
	if status != 200 {
		emit(map[string]any{"type": "RUN_ERROR", "message": message})
		return
	}
	emit(map[string]any{"type": "STATE_SNAPSHOT", "snapshot": map[string]any{"lesson": projection}})
	if projection.Explain && projection.Reveal == 3 && h.Narrator != nil {
		messageID := input.RunID + "-explanation"
		emit(map[string]any{"type": "TEXT_MESSAGE_START", "messageId": messageID, "role": "assistant"})
		caption := h.Narrator.Stream(r.Context(), projection, input.ForwardedProps.Coach.Text, func(delta string) {
			emit(map[string]any{"type": "TEXT_MESSAGE_CONTENT", "messageId": messageID, "delta": delta})
		})
		emit(map[string]any{"type": "TEXT_MESSAGE_END", "messageId": messageID})
		if r.Context().Err() == nil {
			projection.Caption = caption
			emit(map[string]any{"type": "STATE_SNAPSHOT", "snapshot": map[string]any{"lesson": projection}})
		}
	}
	emit(map[string]any{"type": "RUN_FINISHED", "threadId": input.ThreadID, "runId": input.RunID})
}
