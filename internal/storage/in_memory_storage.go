package storage

import (
	"math"
	"strconv"
	"sync"
)

type InMemoryStorage struct {
	mux sync.RWMutex
	m   map[string][]byte
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		m: make(map[string][]byte),
	}
}

func (s *InMemoryStorage) Get(k string) ([]byte, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	v, ok := s.m[k]
	if !ok {
		return nil, ErrKeyNotFound
	}
	cp := make([]byte, len(v))
	copy(cp, v)
	return cp, nil
}

func (s *InMemoryStorage) Set(k string, v []byte) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	cp := make([]byte, len(v))
	copy(cp, v)
	s.m[k] = cp
	return nil
}

func (s *InMemoryStorage) Del(k ...string) (int, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	count := 0
	for _, item := range k {
		if _, ok := s.m[item]; ok {
			count++
		}
		delete(s.m, item)
	}
	return count, nil
}

func (s *InMemoryStorage) Incr(k string) (int64, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	v, ok := s.m[k]
	if !ok {
		s.m[k] = []byte(strconv.FormatInt(1, 10))
		return 1, nil
	}
	n, err := strconv.ParseInt(string(v), 10, 64)

	if err != nil {
		return 0, err
	}
	if n == math.MaxInt64 {
		return 0, ErrIntegerOverflow
	}
	n++
	s.m[k] = []byte(strconv.FormatInt(n, 10))
	return n, nil
}
