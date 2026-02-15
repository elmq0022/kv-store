package storage

type Storage interface {
	Set(k string, v []byte) error
	Get(k string) ([]byte, error)
	Del() error
	Incr(k string) (int64, error)
}
