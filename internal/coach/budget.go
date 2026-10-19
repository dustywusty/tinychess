package coach

import "time"

type budget struct {
	Start time.Time
	Count int
}

// A user request consumes one slot, including routing and any ensuing search.
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
	if b.Count >= 60 || len(s.budgets) >= 2000 {
		return false
	}
	if b.Start.IsZero() {
		b.Start = now
	}
	b.Count++
	s.budgets[key] = b
	return true
}

func (s *Service) consume(gameID string, r *Request) bool {
	if r.accounted {
		return true
	}
	if !s.allow(gameID + ":" + r.ClientID) {
		return false
	}
	r.accounted = true
	return true
}
