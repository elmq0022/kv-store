package parser

import (
	"bufio"
	"errors"
	"io"
	"strconv"
)

const (
	typeBulkString   byte = '$'
	typeArray        byte = '*'
	typeSimpleString byte = '+'
	typeInteger      byte = ':'
	typeError        byte = '-'
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
	case typeBulkString:
		val.Bytes, err = parseBulkString(br)
	case typeSimpleString:
		val.Bytes, err = readLine(br)
	case typeInteger:
		val.Bytes, err = readLine(br)
	case typeError:
		val.Bytes, err = readLine(br)
	case typeArray:
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
	b, err := br.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(b) < 2 || b[len(b)-2] != '\r' {
		return nil, errors.New("expected CRLF line terminator")
	}
	return b[:len(b)-2], nil
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

	vals := make([]Value, n)
	for i := range n {
		vals[i], err = Parse(br)
		if err != nil {
			return nil, err
		}
	}

	return vals, nil
}
