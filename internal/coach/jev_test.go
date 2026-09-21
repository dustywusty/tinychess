package coach

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestJevRejectsUnregisteredActions(t *testing.T) {
	j := &Jev{Key: "test", Client: &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		data, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(data), `"type":"choice"`) {
			t.Fatal("not a Choice request")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"answers":{"action":{"type":"choice","choice":"submitLiveMove"}}}`))}, nil
	})}}
	if _, err := j.Select(context.Background(), "make my move", map[string]string{"back": "Back"}, nil); err == nil {
		t.Fatal("executed arbitrary model action")
	}
}
func TestNarrationNeverReceivesHiddenAnswers(t *testing.T) {
	n := &Narrator{Key: "test", Model: "test", Client: &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		t.Fatal("hidden lesson sent to narration")
		return nil, nil
	})}}
	p := Projection{Reveal: 0, Caption: "Look for checks."}
	if n.Explain(context.Background(), p, "tell me everything") != p.Caption {
		t.Fatal("missing fallback")
	}
}

func TestNarrationStreamsOnlyOutputTextAndFallsBackOnFailure(t *testing.T) {
	n := &Narrator{Key: "test", Model: "configured", Client: &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		data, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(data), `"stream":true`) {
			t.Fatal("stream not requested")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("data: {\"type\":\"response.output_text.delta\",\"delta\":\"Compare the moves.\"}\n\ndata: {\"type\":\"response.completed\",\"response\":{\"usage\":{}}}\n\n"))}, nil
	})}}
	var chunks []string
	p := Projection{Reveal: 3, Caption: "Fallback"}
	text := n.Stream(context.Background(), p, "why", func(s string) { chunks = append(chunks, s) })
	if text != "Compare the moves." || len(chunks) != 1 {
		t.Fatal(text, chunks)
	}
	n.Client = &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
	if n.Explain(context.Background(), p, "why") != "Fallback" {
		t.Fatal("provider failure lost fallback")
	}
}
