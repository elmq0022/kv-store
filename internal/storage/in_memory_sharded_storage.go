package storage

import "hash/fnv"

const size int64 = 64

type InMemoryShardedStorage struct {
	m []*InMemoryStorage
}

func NewInMemoryShardedStorage() *InMemoryShardedStorage {
	storages := make([]*InMemoryStorage, size)
	for i := range size {
		storages[i] = NewInMemoryStorage()
	}

	return &InMemoryShardedStorage{
		m: storages,
	}
}

func (s *InMemoryShardedStorage) shard(k string) *InMemoryStorage {
	h := fnv.New64a()
	h.Write([]byte(k))
	return s.m[h.Sum64()%uint64(size)]
}

func (s *InMemoryShardedStorage) Get(k string) ([]byte, error) {
	return s.shard(k).Get(k)
}

func (s *InMemoryShardedStorage) Set(k string, v []byte) error {
	return s.shard(k).Set(k, v)
}

func (s *InMemoryShardedStorage) Del(k ...string) (int, error) {
	count := 0
	for _, key := range k {
		n, err := s.shard(key).Del(key)
		if err != nil {
			return count, err
		}
		count += n
	}
	return count, nil
}

func (s *InMemoryShardedStorage) Incr(k string) (int64, error) {
	return s.shard(k).Incr(k)
}
