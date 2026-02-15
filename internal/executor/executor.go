package executor

import (
	"errors"

	"github.com/elmq0022/kv-store/internal/parser"
	"github.com/elmq0022/kv-store/internal/storage"
)

const (
	cmdSet  = "set"
	cmdGet  = "get"
	cmdDel  = "del"
	cmdIncr = "incr"
	cmdEcho = "echo"
	cmdPing = "ping"
)

type Executor struct {
	storage storage.Storage
}

func (e *Executor) Execute(val parser.Value) (parser.Value, error) {
	if val.Type != parser.TypeArray {
		return parser.Value{Type: parser.TypeError, Bytes: []byte("ERR expected array")}, nil
	}
	if len(val.Array) < 1 {
		return parser.Value{Type: parser.TypeError, Bytes: []byte("ERR empty command")}, nil
	}

	cmd := val.Array[0].Bytes

	switch string(cmd) {
	case cmdSet:
		return e.set(val.Array[1:])
	case cmdGet:
		return e.get(val.Array[1:])
	case cmdDel:
		return e.del(val.Array[1:])
	case cmdIncr:
		return e.incr(val.Array[1:])
	case cmdEcho:
		return e.echo(val.Array[1:])
	case cmdPing:
		return e.ping(val.Array[1:])
	default:
		return parser.Value{Type: parser.TypeError, Bytes: []byte("ERR unknown command '" + string(cmd) + "'")}, nil
	}
}

func (e *Executor) set(args []parser.Value) (parser.Value, error) {
	return parser.Value{}, nil
}

func (e *Executor) get(args []parser.Value) (parser.Value, error) {
	return parser.Value{}, nil
}

func (e *Executor) del(args []parser.Value) (parser.Value, error) {
	return parser.Value{}, nil
}

func (e *Executor) incr(args []parser.Value) (parser.Value, error) {
	return parser.Value{}, nil
}

func (e *Executor) echo(args []parser.Value) (parser.Value, error) {
	if len(args) != 1 {
		return parser.Value{}, errors.New("bad command")
	}
	return args[0], nil
}

func (e *Executor) ping(args []parser.Value) (parser.Value, error) {
	if len(args) > 1 {
		return parser.Value{}, errors.New("bad command")
	}
	if len(args) > 0 {
		return parser.Value{
			Type:  parser.TypeBulkString,
			Bytes: args[0].Bytes,
		}, nil
	}
	return parser.Value{
		Type:  parser.TypeSimpleString,
		Bytes: []byte("pong"),
	}, nil
}
