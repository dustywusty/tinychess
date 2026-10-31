package coach

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestProviderDeadlinesReleaseCapacity(t *testing.T) {
	client := &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}
	for range 6 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		_, err := (&Jev{Key: "fixture", Client: client}).Select(ctx, "back", map[string]string{"back": "Back"}, nil)
		cancel()
		if err == nil || !strings.Contains(err.Error(), "use the buttons") {
			t.Fatal(err)
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond)
		p := Projection{Reveal: 3, Caption: "Compare the legal continuations."}
		got := (&Narrator{Key: "fixture", Model: "fixture", Client: client}).Explain(ctx, p, "why")
		cancel()
		if got != p.Caption {
			t.Fatal("deadline lost the evidence fallback")
		}
	}
	if len(providerSlots) != 0 {
		t.Fatal("provider deadline leaked concurrency capacity")
	}
}

func TestIncompleteNarrationRestoresEvidenceCaption(t *testing.T) {
	for _, ending := range []string{"response.incomplete", "response.failed", "error", "EOF"} {
		t.Run(ending, func(t *testing.T) {
			stream := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"Partial explanation. \"}\n\n"
			if ending != "EOF" {
				stream += "data: {\"type\":\"" + ending + "\"}\n\n"
			}
			n := &Narrator{Key: "fixture", Model: "fixture", Client: &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(stream))}, nil
			})}}
			var partial string
			p := Projection{Reveal: 3, Caption: "Compare the legal continuations."}
			got := n.Stream(context.Background(), p, "why", func(s string) { partial += s })
			if partial == "" || got != p.Caption {
				t.Fatal("incomplete output replaced the evidence fallback", got)
			}
		})
	}
}

func TestUnsupportedMotifNeverReachesStream(t *testing.T) {
	stream := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"This move fo\"}\n\n" +
		"data: {\"type\":\"response.output_text.delta\",\"delta\":\"rks the king and queen. \"}\n\n" +
		"data: {\"type\":\"response.completed\"}\n\n"
	n := &Narrator{Key: "fixture", Model: "fixture", Client: &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(stream))}, nil
	})}}
	p := Projection{Reveal: 3, Caption: "Compare the legal continuations."}
	var output string
	got := n.Stream(context.Background(), p, "invent a fork", func(s string) { output += s })
	if got != p.Caption || output != "" {
		t.Fatal("unsupported split-token claim leaked", output)
	}
}

func TestUnprovenForcedOutcomeUsesFallback(t *testing.T) {
	for _, text := range []string{"The engine evaluation shows a forced mate sequence.", "This is a forced win."} {
		if !unsupportedNarration(text) {
			t.Fatal("unproven outcome accepted", text)
		}
	}
}
