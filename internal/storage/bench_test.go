package storage

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
)

// benchStorage runs Get/Set benchmarks against any Storage implementation.
func benchStorage(b *testing.B, name string, s Storage) {
	val := []byte("hello")

	b.Run(name+"/Set", func(b *testing.B) {
		for i := b.Loop(); i; i = b.Loop() {
			s.Set("key", val)
		}
	})

	// Pre-populate for Get
	s.Set("key", val)

	b.Run(name+"/Get", func(b *testing.B) {
		for i := b.Loop(); i; i = b.Loop() {
			s.Get("key")
		}
	})

	b.Run(name+"/SetGet", func(b *testing.B) {
		for i := b.Loop(); i; i = b.Loop() {
			s.Set("key", val)
			s.Get("key")
		}
	})

	// Parallel read-heavy workload (90% Get, 10% Set)
	s.Set("key", val)
	b.Run(name+"/ParallelReadHeavy", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				if i%10 == 0 {
					s.Set("key", val)
				} else {
					s.Get("key")
				}
				i++
			}
		})
	})

	// Parallel write-heavy workload (90% Set, 10% Get)
	b.Run(name+"/ParallelWriteHeavy", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				if i%10 == 0 {
					s.Get("key")
				} else {
					s.Set("key", val)
				}
				i++
			}
		})
	})

	// Benchmark with many distinct keys to stress map growth
	b.Run(name+"/SetDistinctKeys", func(b *testing.B) {
		keys := make([]string, 10000)
		for i := range keys {
			keys[i] = "key:" + strconv.Itoa(i)
		}
		b.ResetTimer()
		for i := b.Loop(); i; i = b.Loop() {
			for _, k := range keys {
				s.Set(k, val)
			}
		}
	})

	// Parallel contention on many distinct keys (better for sharded)
	b.Run(name+"/ParallelDistinctKeys", func(b *testing.B) {
		var mu sync.Mutex
		counter := 0
		nextID := func() int {
			mu.Lock()
			id := counter
			counter++
			mu.Unlock()
			return id
		}
		b.RunParallel(func(pb *testing.PB) {
			id := nextID()
			key := fmt.Sprintf("worker:%d", id)
			for pb.Next() {
				s.Set(key, val)
				s.Get(key)
			}
		})
	})
}

func BenchmarkInMemoryStorage(b *testing.B) {
	s := NewInMemoryStorage()
	benchStorage(b, "InMemory", s)
}

func BenchmarkInMemoryShardedStorage(b *testing.B) {
	s := NewInMemoryShardedStorage()
	benchStorage(b, "Sharded", s)
}
