package persist

import (
	"sync"
	"time"
)

const persistFlushDelay = 500 * time.Millisecond

type flushScheduler struct {
	mu    sync.Mutex
	timer *time.Timer
}

func (s *flushScheduler) schedule(flush func() error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.timer != nil {
		s.timer.Stop()
	}
	s.timer = time.AfterFunc(persistFlushDelay, func() {
		_ = flush()
	})
}

func (s *flushScheduler) flushNow(flush func() error) error {
	s.mu.Lock()
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	s.mu.Unlock()
	return flush()
}
