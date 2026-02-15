package parser

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
	maxArrayLen      = 1024 * 1024       // 1M elements
	maxLineLen       = 64 * 1024         // 64 KB
)

type Value struct {
	Type  byte
	Bytes []byte
	Array []Value
}

func Parse(r io.Reader) (Value, error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	var err error

	bytecode, err := br.ReadByte()
	if err != nil {
		return Value{}, err
	}

	val := Value{}
	val.Type = bytecode

	switch bytecode {
	case TypeBulkString:
		val.Bytes, err = parseBulkString(br)
	case TypeSimpleString, TypeInteger, TypeError:
		val.Bytes, err = readLine(br)
	case TypeArray:
		val.Array, err = parseArray(br)
	default:
		return Value{}, errors.New("unknown type prefix: " + string(bytecode))
	}

	return val, err
}

func parseBulkString(br *bufio.Reader) ([]byte, error) {
	b, err := readLine(br)
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

	p := make([]byte, nWant)
	_, err = io.ReadFull(br, p)
	if err != nil {
		return nil, err
	}
	code, err := br.ReadByte()
	if err != nil {
		return nil, err
	}
	if code != '\r' {
		return nil, errors.New("expected CRLF after bulk string data")
	}
	code, err = br.ReadByte()
	if err != nil {
		return nil, err
	}
	if code != '\n' {
		return nil, errors.New("expected CRLF after bulk string data")
	}

	return p, nil
}

func readLine(br *bufio.Reader) ([]byte, error) {
	buf := make([]byte, 0, 64)
	for {
		b, err := br.ReadByte()
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

func parseArray(br *bufio.Reader) ([]Value, error) {
	l, err := readLine(br)
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
		vals[i], err = Parse(br)
		if err != nil {
			return nil, err
		}
	}

	return vals, nil
}
