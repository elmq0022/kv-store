package resp

import (
	"bufio"
	"errors"
	"io"
	"strconv"
)

const (
	TypeBulkString   byte = '$'
	TypeArray        byte = '*'
	TypeSimpleString byte = '+'
	TypeInteger      byte = ':'
	TypeError        byte = '-'

	maxBulkStringLen = 512 * 1024 * 1024 // 512 MB
	maxArrayLen      = 100_000           // 1M elements
	maxLineLen       = 64 * 1024         // 64 KB
)

type Value struct {
	Type  byte
	Bytes []byte
	Array []Value
}

type Decoder struct {
	reader *bufio.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	return &Decoder{
		reader: br,
	}
}

func (p *Decoder) Decode() (Value, error) {
	bytecode, err := p.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	val := Value{}
	val.Type = bytecode

	switch bytecode {
	case TypeBulkString:
		val.Bytes, err = p.parseBulkString()
	case TypeSimpleString, TypeInteger, TypeError:
		val.Bytes, err = p.readLine()
	case TypeArray:
		val.Array, err = p.parseArray()
	default:
		return Value{}, errors.New("unknown type prefix: " + string(bytecode))
	}

	return val, err
}

func (p *Decoder) parseBulkString() ([]byte, error) {
	b, err := p.readLine()
	if err != nil {
		return nil, err
	}

	nWant, err := strconv.Atoi(string(b))
	if err != nil {
		return nil, err
	}

	if nWant == -1 {
		return nil, nil
	}

	if nWant < -1 {
		return nil, errors.New("invalid length for bulk string")
	}

	if nWant > maxBulkStringLen {
		return nil, errors.New("bulk string length exceeds maximum")
	}

	buf := make([]byte, nWant)
	_, err = io.ReadFull(p.reader, buf)
	if err != nil {
		return nil, err
	}
	code, err := p.reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if code != '\r' {
		return nil, errors.New("expected CRLF after bulk string data")
	}
	code, err = p.reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if code != '\n' {
		return nil, errors.New("expected CRLF after bulk string data")
	}

	return buf, nil
}

func (p *Decoder) readLine() ([]byte, error) {
	buf := make([]byte, 0, 64)
	for {
		b, err := p.reader.ReadByte()
		if err != nil {
			return nil, err
		}
		if b == '\n' {
			if len(buf) == 0 || buf[len(buf)-1] != '\r' {
				return nil, errors.New("expected CRLF line terminator")
			}
			return buf[:len(buf)-1], nil
		}
		buf = append(buf, b)
		if len(buf) > maxLineLen {
			return nil, errors.New("line length exceeds maximum")
		}
	}
}

func (p *Decoder) parseArray() ([]Value, error) {
	l, err := p.readLine()
	if err != nil {
		return nil, err
	}
	n, err := strconv.Atoi(string(l))
	if err != nil {
		return nil, err
	}

	if n == -1 {
		return nil, nil
	}

	if n < -1 {
		return nil, errors.New("invalid int for array size")
	}

	if n > maxArrayLen {
		return nil, errors.New("array size exceeds maximum")
	}

	vals := make([]Value, n)
	for i := range n {
		vals[i], err = p.Decode()
		if err != nil {
			return nil, err
		}
	}

	return vals, nil
}
