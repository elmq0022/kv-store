package storage

import "errors"

var (
	ErrKeyNotFound     = errors.New("key not found")
	ErrIntegerOverflow = errors.New("integer overflow")
)

type Storage interface {
	Set(k string, v []byte) error
	Get(k string) ([]byte, error)
	Del(keys ...string) (int, error)
	Incr(k string) (int64, error)
}
