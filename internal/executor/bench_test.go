package executor

import (
	"testing"

	"github.com/elmq0022/kv-store/internal/resp"
	"github.com/elmq0022/kv-store/internal/storage"
)

func makeSetCmd(key, val string) resp.Value {
	return resp.Value{
		Type: resp.TypeArray,
		Array: []resp.Value{
			{Type: resp.TypeBulkString, Bytes: []byte("SET")},
			{Type: resp.TypeBulkString, Bytes: []byte(key)},
			{Type: resp.TypeBulkString, Bytes: []byte(val)},
		},
	}
}

func makeGetCmd(key string) resp.Value {
	return resp.Value{
		Type: resp.TypeArray,
		Array: []resp.Value{
			{Type: resp.TypeBulkString, Bytes: []byte("GET")},
			{Type: resp.TypeBulkString, Bytes: []byte(key)},
		},
	}
}

func BenchmarkExecuteSet(b *testing.B) {
	e := NewExecutor(storage.NewInMemoryStorage())
	cmd := makeSetCmd("foo", "bar")
	for b.Loop() {
		e.Execute(cmd)
	}
}

func BenchmarkExecuteGet(b *testing.B) {
	e := NewExecutor(storage.NewInMemoryStorage())
	e.Execute(makeSetCmd("foo", "bar"))
	cmd := makeGetCmd("foo")
	for b.Loop() {
		e.Execute(cmd)
	}
}

// Measures the full SET+GET cycle through the executor
func BenchmarkExecuteSetThenGet(b *testing.B) {
	e := NewExecutor(storage.NewInMemoryStorage())
	setCmd := makeSetCmd("foo", "bar")
	getCmd := makeGetCmd("foo")
	for b.Loop() {
		e.Execute(setCmd)
		e.Execute(getCmd)
	}
}

// Compare sharded vs non-sharded through the executor
func BenchmarkExecuteGet_Sharded(b *testing.B) {
	e := NewExecutor(storage.NewInMemoryShardedStorage())
	e.Execute(makeSetCmd("foo", "bar"))
	cmd := makeGetCmd("foo")
	for b.Loop() {
		e.Execute(cmd)
	}
}

func BenchmarkExecuteSet_Sharded(b *testing.B) {
	e := NewExecutor(storage.NewInMemoryShardedStorage())
	cmd := makeSetCmd("foo", "bar")
	for b.Loop() {
		e.Execute(cmd)
	}
}

// Parallel SET+GET through executor to measure lock contention
func BenchmarkExecuteParallel(b *testing.B) {
	e := NewExecutor(storage.NewInMemoryShardedStorage())
	setCmd := makeSetCmd("foo", "bar")
	getCmd := makeGetCmd("foo")
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				e.Execute(setCmd)
			} else {
				e.Execute(getCmd)
			}
			i++
		}
	})
}
