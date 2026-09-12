package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerPort(t *testing.T) {
	for input, want := range map[string]string{"": "8080", "3000": "3000", "65535": "65535", "08080": "8080"} {
		got, err := serverPort(input)
		if err != nil || got != want {
			t.Fatalf("serverPort(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
	for _, input := range []string{"0", "-1", "65536", "localhost:8080", "http", " 8080"} {
		if _, err := serverPort(input); err == nil {
			t.Fatalf("serverPort(%q) must fail", input)
		}
	}
}

func TestHealthHandler(t *testing.T) {
	r := httptest.NewRecorder()
	healthHandler(r, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if r.Code != http.StatusOK || r.Body.String() != `{"ok":true}` || r.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("unexpected health response: %v", r)
	}
}
