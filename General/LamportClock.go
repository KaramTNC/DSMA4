package lamportclock

import (
	"sync"
)

type SafeClock struct {
	time int32
	mu   sync.Mutex
}

func (s *SafeClock) Iterate() {
	s.mu.Lock()
	s.time++
	s.mu.Unlock()
}

func (s *SafeClock) GetTime() int32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.time
}

func (s *SafeClock) MatchTime(otherTime int32) {
	s.mu.Lock()
	s.time = max(s.time, otherTime) + 1
	s.mu.Unlock()
}
