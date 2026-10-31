package coach

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEngineCancellationReleasesCapacity(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hung-engine")
	// A single process stays blocked waiting for input, with no child processes.
	if err := os.WriteFile(path, []byte("#!/bin/sh\nwhile read line; do :; done\n"), 0700); err != nil {
		t.Fatal(err)
	}
	engine := NewEngine(path)
	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		_, err := engine.Analyze(ctx, nil, "")
		cancel()
		if err == nil {
			t.Fatal("hung engine returned invented analysis")
		}
		if len(engine.slots) != 0 {
			t.Fatal("cancelled process retained capacity")
		}
	}
}
