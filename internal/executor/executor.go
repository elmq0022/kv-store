package executor

import (
	"strconv"

	"github.com/elmq0022/kv-store/internal/resp"
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

func New(s storage.Storage) *Executor {
	return &Executor{storage: s}
}

func (e *Executor) Execute(val resp.Value) (resp.Value, error) {
	if val.Type != resp.TypeArray {
		return resp.Value{Type: resp.TypeError, Bytes: []byte("ERR expected array")}, nil
	}
	if len(val.Array) < 1 {
		return resp.Value{Type: resp.TypeError, Bytes: []byte("ERR empty command")}, nil
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
		return resp.Value{Type: resp.TypeError, Bytes: []byte("ERR unknown command '" + string(cmd) + "'")}, nil
	}
}

func (e *Executor) set(args []resp.Value) (resp.Value, error) {
	if len(args) != 2 {
		return resp.Value{Type: resp.TypeError, Bytes: []byte("ERR wrong number of arguments for 'set' command")}, nil
	}

	k := string(args[0].Bytes)
	v := args[1].Bytes
	if err := e.storage.Set(k, v); err != nil {
		return resp.Value{}, err
	}
	return resp.Value{Type: resp.TypeSimpleString, Bytes: []byte("OK")}, nil
}

func (e *Executor) get(args []resp.Value) (resp.Value, error) {
	if len(args) != 1 {
		return resp.Value{Type: resp.TypeError, Bytes: []byte("ERR wrong number of arguments for 'get' command")}, nil
	}
	k := string(args[0].Bytes)
	v, err := e.storage.Get(k)
	if err != nil {
		return resp.Value{}, err
	}
	return resp.Value{Type: resp.TypeBulkString, Bytes: v}, nil
}

func (e *Executor) del(args []resp.Value) (resp.Value, error) {
	if len(args) < 1 {
		return resp.Value{Type: resp.TypeError, Bytes: []byte("ERR wrong number of arguments for 'del' command")}, nil
	}
	keys := make([]string, len(args))
	for i, a := range args {
		keys[i] = string(a.Bytes)
	}
	n, err := e.storage.Del(keys...)
	if err != nil {
		return resp.Value{}, err
	}
	return resp.Value{Type: resp.TypeInteger, Bytes: []byte(strconv.Itoa(n))}, nil
}

func (e *Executor) incr(args []resp.Value) (resp.Value, error) {
	if len(args) != 1 {
		return resp.Value{Type: resp.TypeError, Bytes: []byte("ERR wrong number of arguments for 'incr' command")}, nil
	}
	k := string(args[0].Bytes)
	n, err := e.storage.Incr(k)
	if err != nil {
		return resp.Value{}, err
	}
	return resp.Value{Type: resp.TypeInteger, Bytes: []byte(strconv.FormatInt(n, 10))}, nil
}

func (e *Executor) echo(args []resp.Value) (resp.Value, error) {
	if len(args) != 1 {
		return resp.Value{Type: resp.TypeError, Bytes: []byte("ERR wrong number of arguments for 'echo' command")}, nil
	}
	return args[0], nil
}

func (e *Executor) ping(args []resp.Value) (resp.Value, error) {
	if len(args) > 1 {
		return resp.Value{Type: resp.TypeError, Bytes: []byte("ERR wrong number of arguments for 'ping' command")}, nil
	}
	if len(args) > 0 {
		return resp.Value{
			Type:  resp.TypeBulkString,
			Bytes: args[0].Bytes,
		}, nil
	}
	return resp.Value{
		Type:  resp.TypeSimpleString,
		Bytes: []byte("pong"),
	}, nil
}
