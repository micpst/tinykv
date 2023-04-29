package syncset

import (
	"sync"
)

type SyncSet struct {
	mu   sync.Mutex
	data map[string]struct{}
}

func New() *SyncSet {
	return &SyncSet{
		data: make(map[string]struct{}),
	}
}

func (s *SyncSet) Remove(key string) {
	s.mu.Lock()
	delete(s.data, key)
	s.mu.Unlock()
}

func (s *SyncSet) Add(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[key]; ok {
		return false
	}

	s.data[key] = struct{}{}
	return true
}
