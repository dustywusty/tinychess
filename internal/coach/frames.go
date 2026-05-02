package coach

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// WriteFrame serializes a Frame and writes it with Hashbrown's 4-byte
// big-endian length prefix.
func WriteFrame(w io.Writer, frame Frame) error {
	b, err := json.Marshal(frame)
	if err != nil {
		return fmt.Errorf("marshal frame: %w", err)
	}
	var prefix [4]byte
	binary.BigEndian.PutUint32(prefix[:], uint32(len(b)))
	if _, err := w.Write(prefix[:]); err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

// TextChunk builds a Frame carrying a text delta from the assistant.
func TextChunk(text string) Frame {
	return Frame{
		Type: "chunk",
		Chunk: &Chunk{
			Choices: []Choice{
				{
					Index: 0,
					Delta: Delta{Role: "assistant", Content: text},
				},
			},
		},
	}
}

// FinishFrame signals end-of-stream.
func FinishFrame() Frame {
	return Frame{Type: "finish"}
}

// ErrorFrame signals a stream error.
func ErrorFrame(message string) Frame {
	return Frame{Type: "error", Error: message}
}

// StreamFrames pumps frames from a channel onto an HTTP response,
// flushing after each write so the client sees streaming progress.
// Returns the first write error encountered, or nil.
func StreamFrames(w http.ResponseWriter, frames <-chan Frame) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("coach: ResponseWriter is not a Flusher")
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Cache-Control", "no-cache")
	for frame := range frames {
		if err := WriteFrame(w, frame); err != nil {
			return err
		}
		flusher.Flush()
	}
	return nil
}
