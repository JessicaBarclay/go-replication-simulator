package store

import (
	"sync"
	"time"
)

type Record struct {
	Key       string
	Value     string
	Timestamp time.Time
}

type Store struct {
	data map[string]Record
	mu   sync.RWMutex
}

func NewStore() *Store {
	return &Store{data: make(map[string]Record)}
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = Record{
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
	}
}

func (s *Store) Get(key string) (Record, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.data[key]
	return rec, ok
}
