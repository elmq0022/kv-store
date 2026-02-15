package executor_test

import (
	"errors"
	"testing"

	"github.com/elmq0022/kv-store/internal/executor"
	"github.com/elmq0022/kv-store/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// call records a single method invocation on the spy.
type call struct {
	Method string
	Args   []any
}

// spyStorage implements storage.Storage and records every call.
type spyStorage struct {
	calls []call

	// Return values to stub per method.
	getVal  []byte
	getErr  error
	setErr  error
	delVal  int
	delErr  error
	incrVal int64
	incrErr error
}

func (s *spyStorage) Set(k string, v []byte) error {
	s.calls = append(s.calls, call{Method: "Set", Args: []any{k, v}})
	return s.setErr
}

func (s *spyStorage) Get(k string) ([]byte, error) {
	s.calls = append(s.calls, call{Method: "Get", Args: []any{k}})
	return s.getVal, s.getErr
}

func (s *spyStorage) Del(keys ...string) (int, error) {
	args := make([]any, len(keys))
	for i, k := range keys {
		args[i] = k
	}
	s.calls = append(s.calls, call{Method: "Del", Args: args})
	return s.delVal, s.delErr
}

func (s *spyStorage) Incr(k string) (int64, error) {
	s.calls = append(s.calls, call{Method: "Incr", Args: []any{k}})
	return s.incrVal, s.incrErr
}

// helpers to build parser.Value inputs
func cmd(args ...string) parser.Value {
	vals := make([]parser.Value, len(args))
	for i, a := range args {
		vals[i] = parser.Value{Type: parser.TypeBulkString, Bytes: []byte(a)}
	}
	return parser.Value{Type: parser.TypeArray, Array: vals}
}

func TestExecute_NonArrayInput(t *testing.T) {
	spy := &spyStorage{}
	e := executor.New(spy)

	got, err := e.Execute(parser.Value{Type: parser.TypeBulkString, Bytes: []byte("hello")})
	require.NoError(t, err)
	assert.Equal(t, parser.TypeError, got.Type)
	assert.Equal(t, "ERR expected array", string(got.Bytes))
	assert.Empty(t, spy.calls)
}

func TestExecute_EmptyArray(t *testing.T) {
	spy := &spyStorage{}
	e := executor.New(spy)

	got, err := e.Execute(parser.Value{Type: parser.TypeArray, Array: []parser.Value{}})
	require.NoError(t, err)
	assert.Equal(t, parser.TypeError, got.Type)
	assert.Equal(t, "ERR empty command", string(got.Bytes))
	assert.Empty(t, spy.calls)
}

func TestExecute_UnknownCommand(t *testing.T) {
	spy := &spyStorage{}
	e := executor.New(spy)

	got, err := e.Execute(cmd("foobar"))
	require.NoError(t, err)
	assert.Equal(t, parser.TypeError, got.Type)
	assert.Contains(t, string(got.Bytes), "ERR unknown command 'foobar'")
	assert.Empty(t, spy.calls)
}

func TestPing(t *testing.T) {
	spy := &spyStorage{}
	e := executor.New(spy)

	t.Run("no args", func(t *testing.T) {
		got, err := e.Execute(cmd("ping"))
		require.NoError(t, err)
		assert.Equal(t, parser.TypeSimpleString, got.Type)
		assert.Equal(t, "pong", string(got.Bytes))
	})

	t.Run("with message", func(t *testing.T) {
		got, err := e.Execute(cmd("ping", "hello"))
		require.NoError(t, err)
		assert.Equal(t, parser.TypeBulkString, got.Type)
		assert.Equal(t, "hello", string(got.Bytes))
	})

	t.Run("too many args", func(t *testing.T) {
		got, err := e.Execute(cmd("ping", "a", "b"))
		require.NoError(t, err)
		assert.Equal(t, parser.TypeError, got.Type)
	})

	assert.Empty(t, spy.calls, "ping should not touch storage")
}

func TestEcho(t *testing.T) {
	spy := &spyStorage{}
	e := executor.New(spy)

	t.Run("echoes message", func(t *testing.T) {
		got, err := e.Execute(cmd("echo", "hello world"))
		require.NoError(t, err)
		assert.Equal(t, "hello world", string(got.Bytes))
	})

	t.Run("wrong arg count", func(t *testing.T) {
		got, err := e.Execute(cmd("echo"))
		require.NoError(t, err)
		assert.Equal(t, parser.TypeError, got.Type)
	})

	assert.Empty(t, spy.calls, "echo should not touch storage")
}

func TestSet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		spy := &spyStorage{}
		e := executor.New(spy)

		got, err := e.Execute(cmd("set", "mykey", "myval"))
		require.NoError(t, err)
		assert.Equal(t, parser.TypeSimpleString, got.Type)
		assert.Equal(t, "OK", string(got.Bytes))

		require.Len(t, spy.calls, 1)
		assert.Equal(t, "Set", spy.calls[0].Method)
		assert.Equal(t, "mykey", spy.calls[0].Args[0])
		assert.Equal(t, []byte("myval"), spy.calls[0].Args[1])
	})

	t.Run("storage error", func(t *testing.T) {
		spy := &spyStorage{setErr: errors.New("disk full")}
		e := executor.New(spy)

		_, err := e.Execute(cmd("set", "k", "v"))
		assert.EqualError(t, err, "disk full")
	})

	t.Run("wrong arg count", func(t *testing.T) {
		spy := &spyStorage{}
		e := executor.New(spy)

		got, err := e.Execute(cmd("set", "onlykey"))
		require.NoError(t, err)
		assert.Equal(t, parser.TypeError, got.Type)
		assert.Empty(t, spy.calls)
	})
}

func TestGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		spy := &spyStorage{getVal: []byte("thevalue")}
		e := executor.New(spy)

		got, err := e.Execute(cmd("get", "mykey"))
		require.NoError(t, err)
		assert.Equal(t, parser.TypeBulkString, got.Type)
		assert.Equal(t, "thevalue", string(got.Bytes))

		require.Len(t, spy.calls, 1)
		assert.Equal(t, "Get", spy.calls[0].Method)
		assert.Equal(t, "mykey", spy.calls[0].Args[0])
	})

	t.Run("storage error", func(t *testing.T) {
		spy := &spyStorage{getErr: errors.New("not found")}
		e := executor.New(spy)

		_, err := e.Execute(cmd("get", "k"))
		assert.EqualError(t, err, "not found")
	})

	t.Run("wrong arg count", func(t *testing.T) {
		spy := &spyStorage{}
		e := executor.New(spy)

		got, err := e.Execute(cmd("get"))
		require.NoError(t, err)
		assert.Equal(t, parser.TypeError, got.Type)
		assert.Empty(t, spy.calls)
	})
}

func TestDel(t *testing.T) {
	t.Run("single key", func(t *testing.T) {
		spy := &spyStorage{delVal: 1}
		e := executor.New(spy)

		got, err := e.Execute(cmd("del", "k1"))
		require.NoError(t, err)
		assert.Equal(t, parser.TypeInteger, got.Type)
		assert.Equal(t, "1", string(got.Bytes))

		require.Len(t, spy.calls, 1)
		assert.Equal(t, "Del", spy.calls[0].Method)
		assert.Equal(t, []any{"k1"}, spy.calls[0].Args)
	})

	t.Run("multiple keys", func(t *testing.T) {
		spy := &spyStorage{delVal: 3}
		e := executor.New(spy)

		got, err := e.Execute(cmd("del", "a", "b", "c"))
		require.NoError(t, err)
		assert.Equal(t, "3", string(got.Bytes))

		require.Len(t, spy.calls, 1)
		assert.Equal(t, []any{"a", "b", "c"}, spy.calls[0].Args)
	})

	t.Run("storage error", func(t *testing.T) {
		spy := &spyStorage{delErr: errors.New("oops")}
		e := executor.New(spy)

		_, err := e.Execute(cmd("del", "k"))
		assert.EqualError(t, err, "oops")
	})

	t.Run("no keys", func(t *testing.T) {
		spy := &spyStorage{}
		e := executor.New(spy)

		got, err := e.Execute(cmd("del"))
		require.NoError(t, err)
		assert.Equal(t, parser.TypeError, got.Type)
		assert.Empty(t, spy.calls)
	})
}

func TestIncr(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		spy := &spyStorage{incrVal: 42}
		e := executor.New(spy)

		got, err := e.Execute(cmd("incr", "counter"))
		require.NoError(t, err)
		assert.Equal(t, parser.TypeInteger, got.Type)
		assert.Equal(t, "42", string(got.Bytes))

		require.Len(t, spy.calls, 1)
		assert.Equal(t, "Incr", spy.calls[0].Method)
		assert.Equal(t, "counter", spy.calls[0].Args[0])
	})

	t.Run("storage error", func(t *testing.T) {
		spy := &spyStorage{incrErr: errors.New("not an integer")}
		e := executor.New(spy)

		_, err := e.Execute(cmd("incr", "k"))
		assert.EqualError(t, err, "not an integer")
	})

	t.Run("wrong arg count", func(t *testing.T) {
		spy := &spyStorage{}
		e := executor.New(spy)

		got, err := e.Execute(cmd("incr"))
		require.NoError(t, err)
		assert.Equal(t, parser.TypeError, got.Type)
		assert.Empty(t, spy.calls)
	})
}
