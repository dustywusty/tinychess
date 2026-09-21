package coach

import (
	"context"
	"sync"
	"time"
)

type cacheEntry struct {
	Line Line
	At   time.Time
}

// Cache belongs to one fixed analyzer configuration. History includes repetition.
// It never reuses a different search budget or a different engine instance.
type CachedEngine struct {
	Engine  Analyzer
	mu      sync.Mutex
	entries map[string]cacheEntry
}

func NewCachedEngine(engine Analyzer) *CachedEngine {
	return &CachedEngine{Engine: engine, entries: map[string]cacheEntry{}}
}
func (e *CachedEngine) Analyze(ctx context.Context, history []string, candidate string) (Line, error) {
	key := Identity(history) + ":" + candidate
	e.mu.Lock()
	entry, ok := e.entries[key]
	e.mu.Unlock()
	if ok && time.Since(entry.At) < 30*time.Minute {
		return entry.Line, nil
	}
	line, err := e.Engine.Analyze(ctx, history, candidate)
	if err != nil {
		return Line{}, err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.entries) >= 256 {
		oldest := ""
		var at time.Time
		for k, v := range e.entries {
			if oldest == "" || v.At.Before(at) {
				oldest = k
				at = v.At
			}
		}
		delete(e.entries, oldest)
	}
	e.entries[key] = cacheEntry{line, time.Now()}
	return line, nil
}
