package coach

import "time"

type budget struct {
	Start time.Time
	Count int
}

// Every paid or engine-backed action consumes the same per-seat minute budget.
func (s *Service) allow(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for k, b := range s.budgets {
		if now.Sub(b.Start) > time.Minute {
			delete(s.budgets, k)
		}
	}
	b := s.budgets[key]
	if b.Count >= 12 || len(s.budgets) >= 2000 {
		return false
	}
	if b.Start.IsZero() {
		b.Start = now
	}
	b.Count++
	s.budgets[key] = b
	return true
}
